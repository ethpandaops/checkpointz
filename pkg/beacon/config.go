package beacon

import (
	"errors"
	"fmt"

	"github.com/samcm/checkpointz/pkg/beacon/store"
)

// Config holds configuration for running a FinalityProvider config
type Config struct {
	// Cache holds configuration for the caches.
	Caches CacheConfig `yaml:"caches"`

	// HistoricalEpochCount determines how many historical epochs the provider will cache.
	HistoricalEpochCount int `yaml:"historical_epoch_count"`

	// Cache holds configuration for the caches.
	Frontend FrontendConfig `yaml:"frontend"`
}

// Cache configuration holds configuration for the caches.
type CacheConfig struct {
	// Blocks holds the block cache configuration.
	Blocks store.Config `yaml:"blocks"`
	// States holds the state cache configuration.
	States store.Config `yaml:"states"`
}

type FrontendConfig struct {
	// Enabled flag enables the frontend assets to be served
	Enabled bool `yaml:"enabled"`

	// PublicURL is the public URL where checkpointz will be served from
	PublicURL string `yaml:"public_url"`

	// BrandName is the name of the brand to display in the frontend
	BrandName string `yaml:"brand_name"`

	// BrandImageURL is the URL of the brand image to be displayed on the frontend
	BrandImageURL string `yaml:"brand_image_url"`
}

func (c *Config) Validate() error {
	if c.HistoricalEpochCount < 1 {
		return errors.New("historical_epoch_count must be at least 1")
	}

	if err := c.Caches.Validate(); err != nil {
		return fmt.Errorf("invalid caches config: %s", err)
	}

	return nil
}

func (c *CacheConfig) Validate() error {
	if err := c.Blocks.Validate(); err != nil {
		return fmt.Errorf("invalid blocks config: %s", err)
	}

	if err := c.States.Validate(); err != nil {
		return fmt.Errorf("invalid states config: %s", err)
	}

	if c.Blocks.MaxItems < 3 {
		return errors.New("blocks.max_items must be at least 3")
	}

	if c.States.MaxItems < 3 {
		return errors.New("states.max_items must be at least 3")
	}

	return nil
}
