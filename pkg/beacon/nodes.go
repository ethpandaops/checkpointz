package beacon

import (
	"context"
	"errors"
	"math/rand"
	"time"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/ethpandaops/checkpointz/pkg/beacon/node"
	sbeacon "github.com/samcm/beacon"
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

		opts := *sbeacon.DefaultOptions()

		opts.HealthCheck.Interval.Duration = time.Second * 5
		opts.HealthCheck.SuccessfulResponses = 2

		snode := sbeacon.NewNode(log.WithField("upstream", config.Name), sconfig, namespace, opts)

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

func (n Nodes) Filter(ctx context.Context, f func(*Node) bool) Nodes {
	nodes := []*Node{}

	for _, node := range n {
		if !f(node) {
			continue
		}

		nodes = append(nodes, node)
	}

	return nodes
}

func (n Nodes) PastFinalizedCheckpoint(ctx context.Context, checkpoint *v1.Finality) Nodes {
	return n.Filter(ctx, func(node *Node) bool {
		finality, err := node.Beacon.GetFinality(ctx)
		if err != nil {
			return false
		}

		if finality.Finalized.Epoch < checkpoint.Finalized.Epoch {
			return false
		}

		return true
	})
}
