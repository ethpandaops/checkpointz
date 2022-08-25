package beacon

import (
	"context"
	"errors"
	"fmt"
	"time"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/chuckpreslar/emission"
	"github.com/go-co-op/gocron"
	"github.com/samcm/checkpointz/pkg/beacon/node"
	"github.com/samcm/checkpointz/pkg/beacon/store"
	"github.com/sirupsen/logrus"
)

type Majority struct {
	log logrus.FieldLogger

	nodeConfigs []node.Config
	nodes       Nodes
	broker      *emission.Emitter

	head          *v1.Finality
	servingBundle *v1.Finality

	blocks *store.Block
	states *store.BeaconState

	metrics *Metrics
}

var _ FinalityProvider = (*Majority)(nil)

var (
	topicFinalityHeadUpdated = "finality_head_updated"
)

func NewMajorityProvider(namespace string, log logrus.FieldLogger, nodes []node.Config, maxBlockItems, maxStateItems int) FinalityProvider {
	return &Majority{
		nodeConfigs: nodes,
		log:         log.WithField("module", "beacon/majority"),
		nodes:       NewNodesFromConfig(log, nodes, namespace),

		head:          &v1.Finality{},
		servingBundle: &v1.Finality{},

		broker: emission.NewEmitter(),
		blocks: store.NewBlock(log, maxBlockItems, namespace),
		states: store.NewBeaconState(log, maxStateItems, namespace),

		metrics: NewMetrics(namespace + "_beacon"),
	}
}

func (m *Majority) Start(ctx context.Context) error {
	if err := m.nodes.StartAll(ctx); err != nil {
		return err
	}

	m.OnFinalityCheckpointHeadUpdated(ctx, m.handleFinalityUpdated)
	m.OnFinalityCheckpointHeadUpdated(ctx, m.fetchHistoricalCheckpoints)

	s := gocron.NewScheduler(time.Local)

	if _, err := s.Every("5s").Do(func() {
		if err := m.checkFinality(ctx); err != nil {
			m.log.WithError(err).Error("Failed to check finality")
		}
	}); err != nil {
		return err
	}

	if _, err := s.Every("5s").Do(func() {
		if err := m.checkGenesis(ctx); err != nil {
			m.log.WithError(err).Error("Failed to check for genesis")
		}
	}); err != nil {
		return err
	}

	s.StartAsync()

	return nil
}

func (m *Majority) StartAsync(ctx context.Context) {
	go func() {
		if err := m.Start(ctx); err != nil {
			m.log.WithError(err).Error("Failed to start")
		}
	}()
}

func (m *Majority) Healthy(ctx context.Context) (bool, error) {
	if len(m.nodes.Healthy(ctx)) == 0 {
		return false, nil
	}

	return true, nil
}

func (m *Majority) Syncing(ctx context.Context) (bool, error) {
	if len(m.nodes.NotSyncing(ctx)) == 0 {
		return true, nil
	}

	return false, nil
}

func (m *Majority) Finality(ctx context.Context) (*v1.Finality, error) {
	return m.servingBundle, nil
}

func (m *Majority) checkFinality(ctx context.Context) error {
	aggFinality := []*v1.Finality{}
	readyNodes := m.nodes.Ready(ctx)

	for _, node := range readyNodes {
		finality, err := node.Beacon.GetFinality(ctx)
		if err != nil {
			m.log.Info("Failed to get finality from node", "node", node.Config.Name)

			continue
		}

		aggFinality = append(aggFinality, finality)
	}

	aggregated := NewCheckpoints(aggFinality)

	majority, err := aggregated.Majority()
	if err != nil {
		return err
	}

	if m.head == nil || m.head.Finalized == nil || m.head.Finalized.Root != majority.Finalized.Root {
		m.head = majority
		m.publishFinalityCheckpointHeadUpdated(ctx, majority)
		m.log.WithField("epoch", majority.Finalized.Epoch).WithField("root", fmt.Sprintf("%#x", majority.Finalized.Root)).Info("New finalized head checkpoint")
	}

	return nil
}

func (m *Majority) checkGenesis(ctx context.Context) error {
	// No-Op if we already have the genesis block AND state stored.
	// Note: this check will constantly touch the genesis block and state in their
	// respective stores, ensuring that we never purge those items.
	block, err := m.blocks.GetBySlot(phase0.Slot(0))
	if err == nil && block != nil {
		stateRoot, errr := block.StateRoot()
		if errr == nil {
			if state, er := m.states.GetByStateRoot(stateRoot); er == nil && state != nil {
				return nil
			}
		}
	}

	m.log.Debug("Fetching genesis block and state")

	readyNodes := m.nodes.Ready(ctx)
	if len(readyNodes) == 0 {
		return errors.New("no nodes ready")
	}

	// Grab the genesis root
	randomNode, err := readyNodes.RandomNode(ctx)
	if err != nil {
		return err
	}

	genesisBlock, err := randomNode.Beacon.FetchBlock(ctx, "genesis")
	if err != nil {
		return err
	}

	genesisBlockRoot, err := genesisBlock.Root()
	if err != nil {
		return err
	}

	// Fetch the bundle
	if _, err := m.fetchBundle(ctx, genesisBlockRoot); err != nil {
		return err
	}

	m.log.WithFields(logrus.Fields{
		"root": fmt.Sprintf("%#x", genesisBlockRoot),
	}).Info("Fetched genesis bundle")

	return nil
}

func (m *Majority) OnFinalityCheckpointHeadUpdated(ctx context.Context, cb func(ctx context.Context, checkpoint *v1.Finality) error) {
	m.broker.On(topicFinalityHeadUpdated, func(checkpoint *v1.Finality) {
		if err := cb(ctx, checkpoint); err != nil {
			m.log.WithError(err).Error("Failed to handle finality updated")
		}
	})
}

func (m *Majority) publishFinalityCheckpointHeadUpdated(ctx context.Context, checkpoint *v1.Finality) {
	m.broker.Emit(topicFinalityHeadUpdated, checkpoint)
}

func (m *Majority) handleFinalityUpdated(ctx context.Context, checkpoint *v1.Finality) error {
	block, err := m.fetchBundle(ctx, checkpoint.Finalized.Root)
	if err != nil {
		return err
	}

	// Validate that everything is ok to serve.
	// Lighthouse ref: https://lighthouse-book.sigmaprime.io/checkpoint-sync.html#alignment-requirements
	blockSlot, err := block.Slot()
	if err != nil {
		return fmt.Errorf("failed to get slot from block: %w", err)
	}

	// For simplicity we'll hardcode SLOTS_PER_EPOCH to 32.
	// TODO(sam.calder-mason): Fetch this from a beacon node and store it in the instance.
	const slotsPerEpoch = 32
	if blockSlot%slotsPerEpoch != 0 {
		return fmt.Errorf("block slot is not aligned from an epoch boundary: %d", blockSlot)
	}

	m.servingBundle = checkpoint
	m.metrics.ObserveServingEpoch(checkpoint.Finalized.Epoch)

	m.log.WithFields(
		logrus.Fields{
			"epoch": checkpoint.Finalized.Epoch,
			"root":  fmt.Sprintf("%#x", checkpoint.Finalized.Root),
		},
	).Info("Serving a new finalized checkpoint bundle")

	return nil
}

func (m *Majority) fetchHistoricalCheckpoints(ctx context.Context, checkpoint *v1.Finality) error {
	historicalDistance := uint64(10)

	// Download the previous n epochs worth of epoch boundaries if they don't already exist
	upstream, err := m.nodes.Ready(ctx).DataProviders(ctx).RandomNode(ctx)
	if err != nil {
		return errors.New("no data provider node available")
	}

	sp, err := upstream.Beacon.GetSpec(ctx)
	if err != nil {
		return err
	}

	genesis, err := upstream.Beacon.GetGenesis(ctx)
	if err != nil {
		return err
	}

	// Calculate the epoch boundaries we need to fetch
	// We'll derive the current finalized slot and then work back in intervals of SLOTS_PER_EPOCH.
	currentSlot := uint64(checkpoint.Finalized.Epoch) * uint64(sp.SlotsPerEpoch)
	for i := uint64(1); i < historicalDistance; i++ {
		if currentSlot-(i*uint64(sp.SlotsPerEpoch)) == 0 {
			continue
		}

		slot := phase0.Slot(currentSlot - i*uint64(sp.SlotsPerEpoch))

		// Check if we've already fetched this slot.
		bl, err := m.blocks.GetBySlot(slot)
		if err == nil && bl != nil {
			continue
		}

		m.log.Infof("Fetching historical block for slot %d", slot)

		// Fetch the block for the slot.
		block, err := upstream.Beacon.FetchBlock(ctx, fmt.Sprintf("%v", slot))
		if err != nil {
			return err
		}

		if block == nil {
			continue
		}

		stateRoot, err := block.StateRoot()
		if err != nil {
			return err
		}

		m.log.Infof("Fetched historical block for slot %d with state_root of %#x", slot, stateRoot)

		expiresAt := CalculateBlockExpiration(slot, sp.SecondsPerSlot, uint64(sp.SlotsPerEpoch), genesis.GenesisTime, 3*24*time.Hour)

		if err := m.blocks.Add(block, expiresAt); err != nil {
			return err
		}
	}

	return nil
}

func (m *Majority) GetBlockBySlot(ctx context.Context, slot phase0.Slot) (*spec.VersionedSignedBeaconBlock, error) {
	block, err := m.blocks.GetBySlot(slot)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, errors.New("block not found")
	}

	return block, nil
}

func (m *Majority) GetBlockByRoot(ctx context.Context, root phase0.Root) (*spec.VersionedSignedBeaconBlock, error) {
	block, err := m.blocks.GetByRoot(root)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, errors.New("block not found")
	}

	return block, nil
}

func (m *Majority) GetBlockByStateRoot(ctx context.Context, stateRoot phase0.Root) (*spec.VersionedSignedBeaconBlock, error) {
	block, err := m.blocks.GetByStateRoot(stateRoot)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, errors.New("block not found")
	}

	return block, nil
}

func (m *Majority) GetBeaconStateBySlot(ctx context.Context, slot phase0.Slot) (*[]byte, error) {
	block, err := m.GetBlockBySlot(ctx, slot)
	if err != nil {
		return nil, err
	}

	stateRoot, err := block.StateRoot()
	if err != nil {
		return nil, err
	}

	return m.states.GetByStateRoot(stateRoot)
}

func (m *Majority) GetBeaconStateByStateRoot(ctx context.Context, stateRoot phase0.Root) (*[]byte, error) {
	return m.states.GetByStateRoot(stateRoot)
}

func (m *Majority) GetBeaconStateByRoot(ctx context.Context, root phase0.Root) (*[]byte, error) {
	block, err := m.GetBlockByRoot(ctx, root)
	if err != nil {
		return nil, err
	}

	stateRoot, err := block.StateRoot()
	if err != nil {
		return nil, err
	}

	return m.states.GetByStateRoot(stateRoot)
}

func (m *Majority) fetchBundle(ctx context.Context, root phase0.Root) (*spec.VersionedSignedBeaconBlock, error) {
	m.log.Infof("Fetching a new bundle for root %#x", root)

	// Fetch the bundle from a random data provider node.
	upstream, err := m.nodes.Ready(ctx).DataProviders(ctx).RandomNode(ctx)
	if err != nil {
		return nil, err
	}

	m.log.Infof("Fetching bundle from node %s with root %#x", upstream.Config.Name, root)

	block, err := upstream.Beacon.FetchBlock(ctx, fmt.Sprintf("%#x", root))
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, errors.New("block is nil")
	}

	stateRoot, err := block.StateRoot()
	if err != nil {
		return nil, err
	}

	blockRoot, err := block.Root()
	if err != nil {
		return nil, err
	}

	if blockRoot != root {
		return nil, errors.New("block root does not match")
	}

	slot, err := block.Slot()
	if err != nil {
		return nil, err
	}

	m.log.
		WithField("slot", slot).
		WithField("root", fmt.Sprintf("%#x", blockRoot)).
		WithField("state_root", fmt.Sprintf("%#x", stateRoot)).
		Info("Fetched beacon block")

	expiresAt := time.Now().Add(time.Hour * 2)
	if slot == phase0.Slot(0) {
		expiresAt = time.Now().Add(time.Hour * 999999)
	}

	err = m.blocks.Add(block, expiresAt)
	if err != nil {
		return nil, err
	}

	beaconState, err := upstream.Beacon.FetchRawBeaconState(ctx, fmt.Sprintf("%#x", stateRoot), "application/octet-stream")
	if err != nil {
		return nil, err
	}

	m.log.
		Info("Fetched beacon state")

	if err := m.states.Add(stateRoot, &beaconState, expiresAt); err != nil {
		return nil, err
	}

	m.log.Infof("Successfully fetched bundle from %s", upstream.Config.Name)

	return block, nil
}

func (m *Majority) UpstreamsStatus(ctx context.Context) (map[string]*UpstreamStatus, error) {
	rsp := make(map[string]*UpstreamStatus)

	for _, node := range m.nodes {
		rsp[node.Config.Name] = &UpstreamStatus{
			Name:    node.Config.Name,
			Healthy: false,
		}

		if node.Beacon == nil {
			continue
		}

		finality, err := node.Beacon.GetFinality(ctx)
		if err != nil {
			continue
		}

		rsp[node.Config.Name].Healthy = node.Beacon.GetStatus(ctx).Healthy()

		if finality != nil {
			rsp[node.Config.Name].Finality = finality
		}
	}

	return rsp, nil
}
