package beacon

import (
	"context"
	"errors"
	"math/rand"

	sbeacon "github.com/samcm/beacon"
	"github.com/samcm/checkpointz/pkg/beacon/node"
	"github.com/sirupsen/logrus"
)

type Node struct {
	Config node.Config
	Beacon sbeacon.Node
}

type Nodes []*Node

func NewNodesFromConfig(log logrus.FieldLogger, configs []node.Config, namespace string) Nodes {
	nodes := make(Nodes, len(configs))

	for i, config := range configs {
		sconfig := &sbeacon.Config{
			Name: config.Name,
			Addr: config.Address,
		}

		snode := sbeacon.NewNode(log.WithField("upstream", config.Name), sconfig, namespace, *sbeacon.DefaultOptions())

		// TODO(sam.calder-mason): Can we re-enable this if we're expecting to use a full beacon node for v1?
		snode.Options().BeaconSubscription.Enabled = false

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

func (n Nodes) Ready(ctx context.Context) Nodes {
	return n.
		Healthy(ctx).
		NotSyncing(ctx)
}

func (n Nodes) RandomNode(ctx context.Context) (*Node, error) {
	nodes := n.Ready(ctx)

	if len(nodes) == 0 {
		return nil, errors.New("no nodes found")
	}

	//nolint:gosec // not critical to worry about/will probably be replaced.
	return nodes[rand.Intn(len(nodes))], nil
}
