package beacon

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/samcm/checkpointz/pkg/beacon/store"
	"github.com/samcm/checkpointz/pkg/cache"
	"github.com/samcm/checkpointz/pkg/eth"
	"github.com/sirupsen/logrus"
)

type BundleDownloader struct {
	log   logrus.FieldLogger
	nodes Nodes

	rootErrors *cache.TTLMap

	mu    *sync.Mutex
	queue []phase0.Root

	states *store.BeaconState
	blocks *store.Block
}

func NewBundleDownloader(log logrus.FieldLogger, nodes Nodes, states *store.BeaconState, blocks *store.Block) *BundleDownloader {
	return &BundleDownloader{
		log:   log,
		nodes: nodes,

		rootErrors: cache.NewTTLMap(100, "root_errors", "beacon"),

		queue: []phase0.Root{},
		mu:    &sync.Mutex{},

		states: states,
		blocks: blocks,
	}
}

func (d *BundleDownloader) AddToQueue(ctx context.Context, root phase0.Root) error {
	d.log.WithField("root", eth.RootAsString(root)).Debug("Adding root to queue")

	d.mu.Lock()
	defer d.mu.Unlock()

	for _, r := range d.queue {
		if r == root {
			return nil
		}
	}

	d.queue = append(d.queue, root)

	return nil
}

func (d *BundleDownloader) RemoveFromQueue(ctx context.Context, root phase0.Root) {
	d.log.WithField("root", eth.RootAsString(root)).Debug("Removing queue")

	d.mu.Lock()
	defer d.mu.Unlock()

	for i, r := range d.queue {
		if r == root {
			d.queue = append(d.queue[:i], d.queue[i+1:]...)
			break
		}
	}
}

func (d *BundleDownloader) ExistsInQueue(root phase0.Root) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, r := range d.queue {
		if r == root {
			return true
		}
	}

	return false
}

func (d *BundleDownloader) Start(ctx context.Context) error {
	d.log.Debug("Starting bundle downloader")

	select {
	case <-time.After(time.Second * 1):
		if err := d.downloadQueue(ctx); err != nil {
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (d *BundleDownloader) downloadQueue(ctx context.Context) error {
	for _, root := range d.queue {
		if _, _, err := d.rootErrors.Get(eth.RootAsString(root)); err == nil {
			continue
		}

		err := d.downloadBundle(ctx, root)
		if err != nil {
			d.rootErrors.Add(eth.RootAsString(root), err, time.Now().Add(1*time.Minute))

			return err
		}

		d.RemoveFromQueue(ctx, root)
	}

	return nil
}

func (d *BundleDownloader) downloadBundle(ctx context.Context, root phase0.Root) error {
	d.log.Infof("Fetching a new bundle for root %#x", root)

	// Fetch the bundle from a random data provider node.
	upstream, err := d.nodes.Ready(ctx).DataProviders(ctx).RandomNode(ctx)
	if err != nil {
		return err
	}

	d.log.Infof("Fetching bundle from node %s with root %#x", upstream.Config.Name, root)

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
		return err
	}

	beaconState, err := upstream.Beacon.FetchRawBeaconState(ctx, fmt.Sprintf("%#x", stateRoot), "application/octet-stream")
	if err != nil {
		return err
	}

	d.log.
		Info("Fetched beacon state")

	if err := d.states.Add(stateRoot, &beaconState, expiresAt); err != nil {
		return err
	}

	d.log.Infof("Successfully fetched bundle from %s", upstream.Config.Name)

	return nil
}
