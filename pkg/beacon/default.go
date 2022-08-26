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
	"github.com/samcm/beacon/state"
	"github.com/samcm/checkpointz/pkg/beacon/checkpoints"
	"github.com/samcm/checkpointz/pkg/beacon/node"
	"github.com/samcm/checkpointz/pkg/beacon/store"
	"github.com/sirupsen/logrus"
)

type Default struct {
	log logrus.FieldLogger

	nodeConfigs []node.Config
	nodes       Nodes
	broker      *emission.Emitter

	head          *v1.Finality
	servingBundle *v1.Finality

	blocks *store.Block
	states *store.BeaconState

	spec *state.Spec

	metrics *Metrics
}

var _ FinalityProvider = (*Default)(nil)

var (
	topicFinalityHeadUpdated = "finality_head_updated"
)

func NewDefaultProvider(namespace string, log logrus.FieldLogger, nodes []node.Config, maxBlockItems, maxStateItems int) FinalityProvider {
	return &Default{
		nodeConfigs: nodes,
		log:         log.WithField("module", "beacon/default"),
		nodes:       NewNodesFromConfig(log, nodes, namespace),

		head:          &v1.Finality{},
		servingBundle: &v1.Finality{},

		broker: emission.NewEmitter(),
		blocks: store.NewBlock(log, maxBlockItems, namespace),
		states: store.NewBeaconState(log, maxStateItems, namespace),

		metrics: NewMetrics(namespace + "_beacon"),
	}
}

func (d *Default) Start(ctx context.Context) error {
	if err := d.nodes.StartAll(ctx); err != nil {
		return err
	}

	d.OnFinalityCheckpointHeadUpdated(ctx, d.fetchHistoricalCheckpoints)

	s := gocron.NewScheduler(time.Local)

	if _, err := s.Every("5s").Do(func() {
		if err := d.checkFinality(ctx); err != nil {
			d.log.WithError(err).Error("Failed to check finality")
		}
	}); err != nil {
		return err
	}

	if _, err := s.Every("60s").Do(func() {
		if err := d.checkBeaconStateSpec(ctx); err != nil {
			d.log.WithError(err).Error("Failed to check beacon state spec")
		}
	}); err != nil {
		return err
	}

	go func() {
		if err := d.startGenesisLoop(ctx); err != nil {
			d.log.WithError(err).Fatal("Failed to start genesis loop")
		}
	}()

	go func() {
		if err := d.startServingLoop(ctx); err != nil {
			d.log.WithError(err).Fatal("Failed to start serving loop")
		}
	}()

	s.StartAsync()

	return nil
}

func (d *Default) StartAsync(ctx context.Context) {
	go func() {
		if err := d.Start(ctx); err != nil {
			d.log.WithError(err).Error("Failed to start")
		}
	}()
}

func (d *Default) startGenesisLoop(ctx context.Context) error {
	if err := d.checkGenesis(ctx); err != nil {
		d.log.WithError(err).Error("Failed to check for genesis")
	}

	for {
		select {
		case <-time.After(time.Second * 15):
			if err := d.checkGenesis(ctx); err != nil {
				d.log.WithError(err).Error("Failed to check for genesis")
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (d *Default) startServingLoop(ctx context.Context) error {
	for {
		select {
		case <-time.After(time.Second * 1):
			if err := d.checkForNewServingCheckpoint(ctx); err != nil {
				d.log.WithError(err).Error("Failed to check for new serving checkpoint")

				time.Sleep(time.Second * 30)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (d *Default) checkForNewServingCheckpoint(ctx context.Context) error {
	// Don't bother checking if we don't know the head yet.
	if d.head == nil {
		return nil
	}

	if d.head.Finalized == nil {
		return nil
	}

	// If head == serving, we're done.
	if d.servingBundle != nil && d.servingBundle.Finalized != nil && d.servingBundle.Finalized.Epoch == d.head.Finalized.Epoch {
		return nil
	}

	if err := d.downloadServingCheckpoint(ctx, d.head); err != nil {
		return err
	}

	return nil
}
func (d *Default) Healthy(ctx context.Context) (bool, error) {
	if len(d.nodes.Healthy(ctx)) == 0 {
		return false, nil
	}

	return true, nil
}

func (d *Default) Syncing(ctx context.Context) (bool, error) {
	if len(d.nodes.NotSyncing(ctx)) == 0 {
		return true, nil
	}

	return false, nil
}

func (d *Default) Finality(ctx context.Context) (*v1.Finality, error) {
	return d.servingBundle, nil
}

func (d *Default) checkFinality(ctx context.Context) error {
	aggFinality := []*v1.Finality{}
	readyNodes := d.nodes.Ready(ctx)

	for _, node := range readyNodes {
		finality, err := node.Beacon.GetFinality(ctx)
		if err != nil {
			d.log.Info("Failed to get finality from node", "node", node.Config.Name)

			continue
		}

		aggFinality = append(aggFinality, finality)
	}

	Default, err := checkpoints.NewMajorityDecider().Decide(aggFinality)
	if err != nil {
		return err
	}

	if d.head == nil || d.head.Finalized == nil || d.head.Finalized.Root != Default.Finalized.Root {
		d.head = Default

		d.publishFinalityCheckpointHeadUpdated(ctx, Default)

		d.log.WithField("epoch", Default.Finalized.Epoch).WithField("root", fmt.Sprintf("%#x", Default.Finalized.Root)).Info("New finalized head checkpoint")

		d.metrics.ObserveHeadEpoch(Default.Finalized.Epoch)
	}

	return nil
}

func (d *Default) checkBeaconStateSpec(ctx context.Context) error {
	// No-Op if we already have a beacon state spec
	if d.spec != nil {
		return nil
	}

	d.log.Debug("Fetching beacon state spec")

	upstream, err := d.nodes.Ready(ctx).DataProviders(ctx).RandomNode(ctx)
	if err != nil {
		return err
	}

	s, err := upstream.Beacon.GetSpec(ctx)
	if err != nil {
		return err
	}

	// store the beacon state spec
	d.spec = s

	d.log.Info("Fetched beacon state spec")

	return nil
}

func (d *Default) checkGenesis(ctx context.Context) error {
	// No-Op if we already have the genesis block AND state stored.
	// Note: this check will constantly touch the genesis block and state in their
	// respective stores, ensuring that we never purge those items.
	block, err := d.blocks.GetBySlot(phase0.Slot(0))
	if err == nil && block != nil {
		stateRoot, errr := block.StateRoot()
		if errr == nil {
			if st, er := d.states.GetByStateRoot(stateRoot); er == nil && st != nil {
				return nil
			}
		}
	}

	d.log.Debug("Fetching genesis block and state")

	readyNodes := d.nodes.Ready(ctx)
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

	upstream, err := d.nodes.Ready(ctx).DataProviders(ctx).RandomNode(ctx)
	if err != nil {
		return err
	}

	if upstream == nil {
		return errors.New("no upstream nodes")
	}

	// Fetch the bundle
	if _, err := d.fetchBundle(ctx, genesisBlockRoot, upstream); err != nil {
		return err
	}

	d.log.WithFields(logrus.Fields{
		"root": fmt.Sprintf("%#x", genesisBlockRoot),
	}).Info("Fetched genesis bundle")

	return nil
}

func (d *Default) OnFinalityCheckpointHeadUpdated(ctx context.Context, cb func(ctx context.Context, checkpoint *v1.Finality) error) {
	d.broker.On(topicFinalityHeadUpdated, func(checkpoint *v1.Finality) {
		if err := cb(ctx, checkpoint); err != nil {
			d.log.WithError(err).Error("Failed to handle finality updated")
		}
	})
}

func (d *Default) publishFinalityCheckpointHeadUpdated(ctx context.Context, checkpoint *v1.Finality) {
	d.broker.Emit(topicFinalityHeadUpdated, checkpoint)
}

func (d *Default) downloadServingCheckpoint(ctx context.Context, checkpoint *v1.Finality) error {
	upstream, err := d.nodes.
		Ready(ctx).
		DataProviders(ctx).
		PastFinalizedCheckpoint(ctx, checkpoint). // Ensure we attempt to fetch the bundle from a node that knows about the checkpoint.
		RandomNode(ctx)
	if err != nil {
		return err
	}

	block, err := d.fetchBundle(ctx, checkpoint.Finalized.Root, upstream)
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

	d.servingBundle = checkpoint
	d.metrics.ObserveServingEpoch(checkpoint.Finalized.Epoch)

	d.log.WithFields(
		logrus.Fields{
			"epoch": checkpoint.Finalized.Epoch,
			"root":  fmt.Sprintf("%#x", checkpoint.Finalized.Root),
		},
	).Info("Serving a new finalized checkpoint bundle")

	return nil
}

func (d *Default) fetchHistoricalCheckpoints(ctx context.Context, checkpoint *v1.Finality) error {
	historicalDistance := uint64(10)

	// Download the previous n epochs worth of epoch boundaries if they don't already exist
	upstream, err := d.nodes.Ready(ctx).DataProviders(ctx).RandomNode(ctx)
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
		bl, err := d.blocks.GetBySlot(slot)
		if err == nil && bl != nil {
			continue
		}

		d.log.Infof("Fetching historical block for slot %d", slot)

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

		d.log.Infof("Fetched historical block for slot %d with state_root of %#x", slot, stateRoot)

		expiresAt := CalculateBlockExpiration(slot, sp.SecondsPerSlot, uint64(sp.SlotsPerEpoch), genesis.GenesisTime, 3*24*time.Hour)

		if err := d.blocks.Add(block, expiresAt); err != nil {
			return err
		}
	}

	return nil
}

func (d *Default) GetBlockBySlot(ctx context.Context, slot phase0.Slot) (*spec.VersionedSignedBeaconBlock, error) {
	block, err := d.blocks.GetBySlot(slot)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, errors.New("block not found")
	}

	return block, nil
}

func (d *Default) GetBlockByRoot(ctx context.Context, root phase0.Root) (*spec.VersionedSignedBeaconBlock, error) {
	block, err := d.blocks.GetByRoot(root)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, errors.New("block not found")
	}

	return block, nil
}

func (d *Default) GetBlockByStateRoot(ctx context.Context, stateRoot phase0.Root) (*spec.VersionedSignedBeaconBlock, error) {
	block, err := d.blocks.GetByStateRoot(stateRoot)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, errors.New("block not found")
	}

	return block, nil
}

func (d *Default) GetBeaconStateBySlot(ctx context.Context, slot phase0.Slot) (*[]byte, error) {
	block, err := d.GetBlockBySlot(ctx, slot)
	if err != nil {
		return nil, err
	}

	stateRoot, err := block.StateRoot()
	if err != nil {
		return nil, err
	}

	return d.states.GetByStateRoot(stateRoot)
}

func (d *Default) GetBeaconStateByStateRoot(ctx context.Context, stateRoot phase0.Root) (*[]byte, error) {
	return d.states.GetByStateRoot(stateRoot)
}

func (d *Default) GetBeaconStateByRoot(ctx context.Context, root phase0.Root) (*[]byte, error) {
	block, err := d.GetBlockByRoot(ctx, root)
	if err != nil {
		return nil, err
	}

	stateRoot, err := block.StateRoot()
	if err != nil {
		return nil, err
	}

	return d.states.GetByStateRoot(stateRoot)
}

func (d *Default) fetchBundle(ctx context.Context, root phase0.Root, upstream *Node) (*spec.VersionedSignedBeaconBlock, error) {
	d.log.Infof("Fetching bundle from node %s with root %#x", upstream.Config.Name, root)

	block, err := d.blocks.GetByRoot(root)
	if err != nil || block == nil {
		// Download the block.
		block, err = upstream.Beacon.FetchBlock(ctx, fmt.Sprintf("%#x", root))
		if err != nil {
			return nil, err
		}

		if block == nil {
			return nil, errors.New("block is nil")
		}
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

	d.log.
		WithField("slot", slot).
		WithField("root", fmt.Sprintf("%#x", blockRoot)).
		WithField("state_root", fmt.Sprintf("%#x", stateRoot)).
		Info("Fetched beacon block")

	expiresAt := time.Now().Add(time.Hour * 2)
	if slot == phase0.Slot(0) {
		expiresAt = time.Now().Add(time.Hour * 999999)
	}

	err = d.blocks.Add(block, expiresAt)
	if err != nil {
		return nil, err
	}

	// If the state already exists, don't bother downloading it again.
	existingState, err := d.states.GetByStateRoot(stateRoot)
	if err == nil && existingState != nil {
		d.log.Infof("Successfully fetched bundle from %s", upstream.Config.Name)

		return block, nil
	}

	beaconState, err := upstream.Beacon.FetchRawBeaconState(ctx, fmt.Sprintf("%#x", stateRoot), "application/octet-stream")
	if err != nil {
		return nil, err
	}

	if beaconState == nil {
		return nil, errors.New("beacon state is nil")
	}

	if err := d.states.Add(stateRoot, &beaconState, expiresAt); err != nil {
		return nil, err
	}

	d.log.Infof("Successfully fetched bundle from %s", upstream.Config.Name)

	return block, nil
}

func (d *Default) UpstreamsStatus(ctx context.Context) (map[string]*UpstreamStatus, error) {
	rsp := make(map[string]*UpstreamStatus)

	for _, node := range d.nodes {
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

func (d *Default) ListFinalizedSlots(ctx context.Context) ([]phase0.Slot, error) {
	slots := []phase0.Slot{}
	if d.spec == nil {
		return slots, errors.New("no upstream beacon state spec available")
	}

	finality, err := d.Finality(ctx)
	if err != nil {
		return slots, err
	}

	latestSlot := phase0.Slot(uint64(finality.Finalized.Epoch) * uint64(d.spec.SlotsPerEpoch))

	for i, val := uint64(latestSlot), uint64(latestSlot)-uint64(d.spec.SlotsPerEpoch)*50; i > val; i -= uint64(d.spec.SlotsPerEpoch) {
		slots = append(slots, phase0.Slot(i))
	}

	return slots, nil
}
