package beacon

import (
	"context"
	"errors"
	"fmt"
	"time"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/chuckpreslar/emission"
	"github.com/go-co-op/gocron"
	"github.com/samcm/checkpointz/pkg/checkpointz/beacon/node"
	"github.com/sirupsen/logrus"
)

type Majority struct {
	FinalityProvider

	log logrus.FieldLogger

	nodeConfigs []node.Config
	nodes       Nodes
	broker      *emission.Emitter

	bundles *CheckpointBundles

	current *v1.Finality
}

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

func (m *Majority) checkFinality(ctx context.Context) error {
	aggFinality := []*v1.Finality{}
	readyNodes := m.nodes.ReadyNodes(ctx)

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
	m.log.Info("Finality updated, checking if we need to fetch a new bundle")

	if err := m.fetchBundle(ctx, checkpoint.Finalized.Root); err != nil {
		return err
	}

	return nil
}

func (m *Majority) fetchBundle(ctx context.Context, root phase0.Root) error {
	// Fetch the bundle from a random data provider node.
	upstream := m.nodes.ReadyNodes(ctx).DataProviders(ctx).RandomNode(ctx)
	if upstream == nil {
		return errors.New("no data provider node available")
	}

	m.log.Infof("Fetching bundle from node %s with root %#x", upstream.Config.Name, root)

	block, err := upstream.Beacon.FetchBlock(ctx, fmt.Sprintf("%#x", root))
	if err != nil {
		return err
	}

	stateRoot, err := block.StateRoot()
	if err != nil {
		return err
	}

	slot, err := block.Slot()
	if err != nil {
		return err
	}

	m.log.
		WithField("slot", slot).
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
}
