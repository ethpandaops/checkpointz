package checkpointz

import (
	"fmt"

	"github.com/samcm/checkpointz/pkg/beacon/node"
)

type Config struct {
	GlobalConfig      `yaml:"global"`
	BeaconConfig      `yaml:"beacon"`
	CheckpointzConfig `yaml:"checkpointz"`
}

type GlobalConfig struct {
	ListenAddr   string `yaml:"listenAddr"`
	LoggingLevel string `yaml:"logging"`
	MetricsAddr  string `yaml:"metricsAddr"`
}

type BeaconConfig struct {
	BeaconUpstreams []node.Config `yaml:"upstreams"`
}

//nolint:revive // Already defined 'config'
type CheckpointzConfig struct {
	MaxBlockCacheSize int `yaml:"maxBlockCacheSize"`
	MaxStateCacheSize int `yaml:"maxStateCacheSize"`
}

func (c *Config) Validate() error {
	// Check that all upstreams have different names and addresses
	duplicates := make(map[string]struct{})

	for _, u := range c.BeaconConfig.BeaconUpstreams {
		if _, ok := duplicates[u.Name]; ok {
			return fmt.Errorf("there's a duplicate upstream with the same name: %s", u.Name)
		}

		if _, ok := duplicates[u.Address]; ok {
			return fmt.Errorf("there's a duplicate upstream with the same address: %s", u.Address)
		}

		duplicates[u.Name] = struct{}{}
		duplicates[u.Address] = struct{}{}
	}

	if c.CheckpointzConfig.MaxBlockCacheSize < 2 {
		return fmt.Errorf("maxBlockCacheSize must be at least 2")
	}

	if c.CheckpointzConfig.MaxStateCacheSize < 2 {
		return fmt.Errorf("maxStateCacheSize must be at least 2")
	}

	if c.CheckpointzConfig.MaxStateCacheSize > 20 {
		return fmt.Errorf("maxStateCacheSize must be at most 20")
	}

	return nil
}
