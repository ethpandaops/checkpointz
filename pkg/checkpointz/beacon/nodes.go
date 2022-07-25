package beacon

import (
	"context"
	"time"

	sbeacon "github.com/samcm/beacon"
	"github.com/samcm/beacon/human"
	"github.com/samcm/checkpointz/pkg/checkpointz/beacon/node"
	"github.com/sirupsen/logrus"
)

type Node struct {
	Config node.Config
	Beacon sbeacon.Node
}

type Nodes []*Node

func NewNodesFromConfig(log logrus.FieldLogger, configs []node.Config) Nodes {
	nodes := make(Nodes, len(configs))

	for i, config := range configs {
		sconfig := &sbeacon.Config{
			Name:        config.Name,
			Addr:        config.Address,
			EventTopics: []string{},
			HealthCheckConfig: sbeacon.HealthCheckConfig{
				Interval:            human.Duration{Duration: time.Second * 5},
				FailedResponses:     3,
				SuccessfulResponses: 2,
			},
		}

		snode := sbeacon.NewNode(log.WithField("upstream", config.Name), sconfig)

		nodes[i] = &Node{
			Config: config,
			Beacon: snode,
		}
	}

	return nodes
}

func (n Nodes) StartAll(ctx context.Context) error {
	for _, node := range n {
		node.Beacon.StartAsync(ctx)
	}

	return nil
}

func (n Nodes) DataProviders(ctx context.Context) Nodes {
	nodes := []*Node{}

	for _, node := range n {
		if !node.Config.DataProvider {
			continue
		}

		nodes = append(nodes, node)
	}

	return nodes
}

func (n Nodes) Healthy(ctx context.Context) Nodes {
	nodes := []*Node{}

	for _, node := range n {
		if !node.Beacon.GetStatus(ctx).Healthy() {
			continue
		}

		nodes = append(nodes, node)
	}

	return nodes
}

func (n Nodes) NotSyncing(ctx context.Context) Nodes {
	nodes := []*Node{}

	for _, node := range n {
		if node.Beacon.GetStatus(ctx).Syncing() {
			continue
		}

		nodes = append(nodes, node)
	}

	return nodes
}

func (n Nodes) Syncing(ctx context.Context) Nodes {
	nodes := []*Node{}

	for _, node := range n {
		if !node.Beacon.GetStatus(ctx).Syncing() {
			continue
		}

		nodes = append(nodes, node)
	}

	return nodes
}

func (n Nodes) ReadyNodes(ctx context.Context) Nodes {
	return n.
		Healthy(ctx).
		NotSyncing(ctx)
}
