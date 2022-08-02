package beacon

import (
	"errors"
	"fmt"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/sirupsen/logrus"
)

var (
	ErrCheckpointBundleNotFound = errors.New("checkpoint bundle not found")
)

type CheckpointBundles struct {
	log     logrus.FieldLogger
	bundles []*CheckpointBundle
}

func NewCheckpointBundles(log logrus.FieldLogger) *CheckpointBundles {
	return &CheckpointBundles{
		log:     log.WithField("module", "beacon/bundles"),
		bundles: []*CheckpointBundle{},
	}
}

func (c *CheckpointBundles) Add(bundle *CheckpointBundle) error {
	root, err := bundle.block.Root()
	if err != nil {
		return err
	}

	c.log.WithField("root", fmt.Sprintf("%#x", root)).Info("Adding checkpoint bundle")

	c.bundles = append(c.bundles, bundle)

	return nil
}

func (c *CheckpointBundles) GetBySlotNumber(slot phase0.Slot) (*CheckpointBundle, error) {
	for _, bundle := range c.bundles {
		s, err := bundle.block.Slot()
		if err != nil {
			continue
		}

		if s == slot {
			return bundle, nil
		}
	}

	return nil, ErrCheckpointBundleNotFound
}

func (c *CheckpointBundles) GetByStateRoot(root phase0.Root) (*CheckpointBundle, error) {
	for _, bundle := range c.bundles {
		s, err := bundle.block.StateRoot()
		if err != nil {
			continue
		}

		if s == root {
			return bundle, nil
		}
	}

	return nil, ErrCheckpointBundleNotFound
}

func (c *CheckpointBundles) GetByRoot(root phase0.Root) (*CheckpointBundle, error) {
	for _, bundle := range c.bundles {
		s, err := bundle.block.Root()
		if err != nil {
			continue
		}

		if s == root {
			return bundle, nil
		}
	}

	return nil, ErrCheckpointBundleNotFound
}
