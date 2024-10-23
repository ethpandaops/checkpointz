package beacon

import (
	"errors"
	"fmt"

	"github.com/ethpandaops/checkpointz/pkg/beacon/store"
)

// Config holds configuration for running a FinalityProvider config
type Config struct {
	// Mode sets the operational mode of the provider.
	Mode OperatingMode `yaml:"mode" default:"light"`
	// Cache holds configuration for the caches.
	Caches CacheConfig `yaml:"caches"`

	// HistoricalEpochCount determines how many historical epochs the provider will cache.
	HistoricalEpochCount int `yaml:"historical_epoch_count" default:"20"`

	// Cache holds configuration for the caches.
	Frontend FrontendConfig `yaml:"frontend"`

	// LightClient holds configuration for serving light client data.
	LightClient LightClientConfig `yaml:"light_client"`
}

// Cache configuration holds configuration for the caches.
type CacheConfig struct {
	// Blocks holds the block cache configuration.
	Blocks store.Config `yaml:"blocks" default:"{\"MaxItems\": 30}"`
	// States holds the state cache configuration.
	States store.Config `yaml:"states" default:"{\"MaxItems\": 5}"`
	// DepositSnapshots holds the deposit snapshot cache configuration.
	DepositSnapshots store.Config `yaml:"deposit_snapshots" default:"{\"MaxItems\": 30}"`
	// BlobSidecars holds the blob sidecar cache configuration.
	BlobSidecars store.Config `yaml:"blob_sidecars" default:"{\"MaxItems\": 30}"`
}

type FrontendConfig struct {
	// Enabled flag enables the frontend assets to be served
	Enabled bool `yaml:"enabled" default:"true"`

	// PublicURL is the public URL where checkpointz will be served from
	PublicURL string `yaml:"public_url"`

	// BrandName is the name of the brand to display in the frontend
	BrandName string `yaml:"brand_name"`

	// BrandImageURL is the URL of the brand image to be displayed on the frontend
	BrandImageURL string `yaml:"brand_image_url"`
}

type LightClientConfig struct {
	// Enabled flag enables the light client data to be served
	Enabled bool `yaml:"enabled" default:"false"`
	// Mode sets the mode of operation for serving light client data.
	Mode LightClientMode `yaml:"mode" default:"proxy"`
}

func (c *Config) Validate() error {
	if c.HistoricalEpochCount < 1 {
		return errors.New("historical_epoch_count must be at least 1")
	}

	if err := c.Caches.Validate(); err != nil {
		return fmt.Errorf("invalid caches config: %s", err)
	}

	if c.HistoricalEpochCount >= c.Caches.Blocks.MaxItems {
		return fmt.Errorf("historical_epoch_count (%d) must be less than caches.blocks.max_items (%d)", c.HistoricalEpochCount, c.Caches.Blocks.MaxItems)
	}

	if c.HistoricalEpochCount > 200 {
		return fmt.Errorf("historical_epoch_count (%d) cannot be higher than 200", c.HistoricalEpochCount)
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
