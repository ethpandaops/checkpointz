package store

import (
	"errors"
	"sync"
	"time"

	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/samcm/checkpointz/pkg/cache"
	"github.com/samcm/checkpointz/pkg/eth"
	"github.com/sirupsen/logrus"
)

type Block struct {
	log   logrus.FieldLogger
	store *cache.TTLMap

	slotToBlockRoot      sync.Map
	stateRootToBlockRoot sync.Map
}

func NewBlock(log logrus.FieldLogger, maxItems int, namespace string) *Block {
	c := &Block{
		log:   log.WithField("component", "beacon/store/block"),
		store: cache.NewTTLMap(maxItems, "block", namespace),

		slotToBlockRoot:      sync.Map{},
		stateRootToBlockRoot: sync.Map{},
	}

	c.store.OnItemDeleted(func(key string, value interface{}, expiredAt time.Time) {
		c.log.WithField("block_root", key).WithField("expired_at", expiredAt.String()).Debug("Block was evicted from the cache")

		block, ok := value.(*spec.VersionedSignedBeaconBlock)
		if !ok {
			c.log.WithField("block_root", key).Error("Invalid block type when cleaning up block cache")

			return
		}

		if err := c.cleanupBlock(block); err != nil {
			c.log.WithError(err).Error("Failed to cleanup block")
		}
	})

	c.store.EnableMetrics(namespace)

	return c
}

func (c *Block) Add(block *spec.VersionedSignedBeaconBlock, expiresAt time.Time) error {
	root, err := block.Root()
	if err != nil {
		return err
	}

	slot, err := block.Slot()
	if err != nil {
		return err
	}

	stateRoot, err := block.StateRoot()
	if err != nil {
		return err
	}

	c.store.Add(eth.RootAsString(root), block, expiresAt)

	c.slotToBlockRoot.Store(slot, root)
	c.stateRootToBlockRoot.Store(stateRoot, root)

	c.log.WithFields(
		logrus.Fields{
			"block_root": eth.RootAsString(root),
			"slot":       eth.SlotAsString(slot),
			"state_root": eth.RootAsString(stateRoot),
		},
	).Debug("Added block")

	return nil
}

func (c *Block) cleanupBlock(block *spec.VersionedSignedBeaconBlock) error {
	slot, err := block.Slot()
	if err != nil {
		return err
	}

	stateRoot, err := block.StateRoot()
	if err != nil {
		return err
	}

	c.slotToBlockRoot.Delete(eth.SlotAsString(slot))
	c.stateRootToBlockRoot.Delete(eth.RootAsString(stateRoot))

	return nil
}

func (c *Block) GetByRoot(root phase0.Root) (*spec.VersionedSignedBeaconBlock, error) {
	data, _, err := c.store.Get(eth.RootAsString(root))
	if err != nil {
		return nil, err
	}

	return c.parseBlock(data)
}

func (c *Block) GetByStateRoot(stateRoot phase0.Root) (*spec.VersionedSignedBeaconBlock, error) {
	data, ok := c.stateRootToBlockRoot.Load(stateRoot)
	if !ok {
		return nil, errors.New("block not found")
	}

	root, err := c.parseRoot(data)
	if err != nil {
		return nil, err
	}

	return c.GetByRoot(root)
}

func (c *Block) GetBySlot(slot phase0.Slot) (*spec.VersionedSignedBeaconBlock, error) {
	data, ok := c.slotToBlockRoot.Load(slot)
	if !ok {
		return nil, errors.New("block not found")
	}

	root, err := c.parseRoot(data)
	if err != nil {
		return nil, err
	}

	return c.GetByRoot(root)
}

func (c *Block) parseBlock(data interface{}) (*spec.VersionedSignedBeaconBlock, error) {
	block, ok := data.(*spec.VersionedSignedBeaconBlock)
	if !ok {
		return nil, errors.New("invalid block type")
	}

	return block, nil
}

func (c *Block) parseRoot(data interface{}) (phase0.Root, error) {
	root, ok := data.(phase0.Root)
	if !ok {
		return phase0.Root{}, errors.New("invalid root")
	}

	return root, nil
}
