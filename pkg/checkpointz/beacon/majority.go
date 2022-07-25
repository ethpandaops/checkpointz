package beacon

import (
	"context"
	"time"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/go-co-op/gocron"
	"github.com/samcm/checkpointz/pkg/checkpointz/beacon/node"
	"github.com/sirupsen/logrus"
)

type Majority struct {
	FinalityProvider

	log logrus.FieldLogger

	nodeConfigs []node.Config
	nodes       Nodes

	current *v1.Finality
}

func NewMajorityProvider(log logrus.FieldLogger, nodes []node.Config) FinalityProvider {
	return &Majority{
		nodeConfigs: nodes,
		log:         log.WithField("module", "beacon/majority"),
		nodes:       NewNodesFromConfig(log, nodes),
		current:     &v1.Finality{},
	}
}

func (m *Majority) Start(ctx context.Context) error {
	if err := m.nodes.StartAll(ctx); err != nil {
		return err
	}

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

	m.log.Debugf("Fetching finality from %v healthy and synced nodes", len(readyNodes))

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

		m.log.WithField("epoch", finalizedMajority.Finalized.Epoch).WithField("root", string(finalizedMajority.Finalized.Root[:])).Info("New finalized checkpoint")
	}

	return nil
}
