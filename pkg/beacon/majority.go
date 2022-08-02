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
	"github.com/sirupsen/logrus"
)

type Majority struct {
	log logrus.FieldLogger

	nodeConfigs []node.Config
	nodes       Nodes
	broker      *emission.Emitter

	bundles *CheckpointBundles

	current *v1.Finality
}

var _ FinalityProvider = (*Majority)(nil)

var (
	topicFinalityUpdated = "finality_updated"
)

func NewMajorityProvider(log logrus.FieldLogger, nodes []node.Config) FinalityProvider {
	return &Majority{
		nodeConfigs: nodes,
		log:         log.WithField("module", "beacon/majority"),
		nodes:       NewNodesFromConfig(log, nodes),
		current:     &v1.Finality{},
		broker:      emission.NewEmitter(),
		bundles:     NewCheckpointBundles(log),
	}
}

func (m *Majority) Start(ctx context.Context) error {
	if err := m.nodes.StartAll(ctx); err != nil {
		return err
	}

	m.OnFinalityCheckpointUpdated(ctx, m.handleFinalityUpdated)
	m.OnFinalityCheckpointUpdated(ctx, m.fetchHistoricalCheckpoints)

	s := gocron.NewScheduler(time.Local)

	if _, err := s.Every("5s").Do(func() {
		if err := m.checkFinality(ctx); err != nil {
			m.log.WithError(err).Error("Failed to check finality")
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
	return m.current, nil
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

	finalizedMajority, err := aggregated.Majority()
	if err != nil {
		return err
	}

	if m.current.Finalized == nil || finalizedMajority.Finalized.Epoch != m.current.Finalized.Epoch || finalizedMajority.Finalized.Root != m.current.Finalized.Root {
		m.current = finalizedMajority

		m.publishFinalityCheckpointUpdated(ctx, finalizedMajority)

		m.log.WithField("epoch", finalizedMajority.Finalized.Epoch).WithField("root", fmt.Sprintf("%#x", finalizedMajority.Finalized.Root)).Info("New finalized checkpoint")
	}

	return nil
}

func (m *Majority) OnFinalityCheckpointUpdated(ctx context.Context, cb func(ctx context.Context, checkpoint *v1.Finality) error) {
	m.broker.On(topicFinalityUpdated, func(checkpoint *v1.Finality) {
		if err := cb(ctx, checkpoint); err != nil {
			m.log.WithError(err).Error("Failed to handle finality updated")
		}
	})
}

func (m *Majority) publishFinalityCheckpointUpdated(ctx context.Context, checkpoint *v1.Finality) {
	m.broker.Emit(topicFinalityUpdated, checkpoint)
}

func (m *Majority) handleFinalityUpdated(ctx context.Context, checkpoint *v1.Finality) error {
	return m.fetchBundle(ctx, checkpoint.Finalized.Root)
}

func (m *Majority) fetchHistoricalCheckpoints(ctx context.Context, checkpoint *v1.Finality) error {
	historicalDistance := uint64(10)

	// Download the previous n epochs worth of epoch boundaries if they don't already exist
	upstream, err := m.nodes.Ready(ctx).DataProviders(ctx).RandomNode(ctx)
	if err != nil {
		return errors.New("no data provider node available")
	}

	spec, err := upstream.Beacon.GetSpec(ctx)
	if err != nil {
		return err
	}

	// Calculate the epoch boundaries we need to fetch
	// We'll derive the current finalized slot and then work back in intervals of SLOTS_PER_EPOCH.
	currentSlot := uint64(checkpoint.Finalized.Epoch) * uint64(spec.SlotsPerEpoch)
	for i := uint64(1); i < historicalDistance; i++ {
		if currentSlot-i*uint64(spec.SlotsPerEpoch) < 0 {
			continue
		}

		slot := phase0.Slot(currentSlot - i*uint64(spec.SlotsPerEpoch))

		// Check if we've already fetched this slot.
		bundle, err := m.bundles.GetBySlotNumber(slot)
		if err == nil && bundle.Block() != nil {
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

		if err = m.bundles.Add(NewCheckpointBundle(
			block,
			nil,
		)); err != nil {
			return err
		}
	}

	return nil
}

func (m *Majority) GetBlockBySlot(ctx context.Context, slot phase0.Slot) (*spec.VersionedSignedBeaconBlock, error) {
	bundle, err := m.bundles.GetBySlotNumber(slot)
	if err != nil {
		return nil, err
	}

	if bundle.Block() == nil {
		return nil, errors.New("block not found")
	}

	return bundle.Block(), nil
}

func (m *Majority) GetBlockByRoot(ctx context.Context, root phase0.Root) (*spec.VersionedSignedBeaconBlock, error) {
	bundle, err := m.bundles.GetByRoot(root)
	if err != nil {
		return nil, err
	}

	if bundle.Block() == nil {
		return nil, errors.New("block not found")
	}

	return bundle.Block(), nil
}

func (m *Majority) GetBlockByStateRoot(ctx context.Context, stateRoot phase0.Root) (*spec.VersionedSignedBeaconBlock, error) {
	bundle, err := m.bundles.GetByStateRoot(stateRoot)
	if err != nil {
		return nil, err
	}

	if bundle.Block() == nil {
		return nil, errors.New("block not found")
	}

	return bundle.Block(), nil
}

func (m *Majority) GetBeaconStateBySlot(ctx context.Context, slot phase0.Slot) (*spec.VersionedBeaconState, error) {
	bundle, err := m.bundles.GetBySlotNumber(slot)
	if err != nil {
		return nil, err
	}

	if bundle.State() == nil {
		return nil, errors.New("state not found")
	}

	return bundle.State(), nil
}

func (m *Majority) GetBeaconStateByStateRoot(ctx context.Context, root phase0.Root) (*spec.VersionedBeaconState, error) {
	bundle, err := m.bundles.GetByStateRoot(root)
	if err != nil {
		return nil, err
	}

	if bundle.State() == nil {
		return nil, errors.New("state not found")
	}

	return bundle.State(), nil
}

func (m *Majority) fetchBundle(ctx context.Context, root phase0.Root) error {
	m.log.Infof("Fetching a new bundle for root %#x", root)

	// Fetch the bundle from a random data provider node.
	upstream, err := m.nodes.Ready(ctx).DataProviders(ctx).RandomNode(ctx)
	if err != nil {
		return err
	}

	m.log.Infof("Fetching bundle from node %s with root %#x", upstream.Config.Name, root)

	block, err := upstream.Beacon.FetchBlock(ctx, fmt.Sprintf("%#x", root))
	if err != nil {
		return err
	}

	if block == nil {
		return errors.New("block is nil")
	}

	stateRoot, err := block.StateRoot()
	if err != nil {
		return err
	}

	blockRoot, err := block.Root()
	if err != nil {
		return err
	}

	if blockRoot != root {
		return errors.New("block root does not match")
	}

	slot, err := block.Slot()
	if err != nil {
		return err
	}

	m.log.
		WithField("slot", slot).
		WithField("root", fmt.Sprintf("%#x", blockRoot)).
		WithField("state_root", fmt.Sprintf("%#x", stateRoot)).
		Info("Fetched beacon block")

	beaconState, err := upstream.Beacon.FetchBeaconState(ctx, fmt.Sprintf("%#x", stateRoot))
	if err != nil {
		return err
	}

	m.log.
		WithField("slot", beaconState.Bellatrix.Slot).
		Info("Fetched beacon state")

	if err = m.bundles.Add(NewCheckpointBundle(
		block,
		beaconState,
	)); err != nil {
		return err
	}

	// Fetch the bundle
	bundle, err := m.bundles.GetByStateRoot(stateRoot)
	if err != nil {
		return err
	}

	bundleStateRoot, err := bundle.block.StateRoot()
	if err != nil {
		return err
	}

	if bundleStateRoot != stateRoot {
		return errors.New("bundle state root does not match inserted block state root")
	}

	m.log.Infof("Successfully fetched bundle from %s", upstream.Config.Name)

	return nil
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
