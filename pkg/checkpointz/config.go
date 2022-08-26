package checkpointz

import (
	"fmt"

	"github.com/samcm/checkpointz/pkg/beacon"
	"github.com/samcm/checkpointz/pkg/beacon/node"
)

type Config struct {
	GlobalConfig GlobalConfig  `yaml:"global"`
	BeaconConfig BeaconConfig  `yaml:"beacon"`
	Checkpointz  beacon.Config `yaml:"checkpointz"`
}

type GlobalConfig struct {
	ListenAddr   string `yaml:"listenAddr"`
	LoggingLevel string `yaml:"logging"`
	MetricsAddr  string `yaml:"metricsAddr"`
}

type BeaconConfig struct {
	BeaconUpstreams []node.Config `yaml:"upstreams"`
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

	if err := c.Checkpointz.Validate(); err != nil {
		return fmt.Errorf("invalid checkpointz config: %s", err)
	}

	return nil
}
