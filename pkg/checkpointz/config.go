package checkpointz

import (
	"fmt"

	"github.com/samcm/checkpointz/pkg/beacon/node"
)

type Config struct {
	GlobalConfig `yaml:"global"`
	BeaconConfig `yaml:"beacon"`
}

type GlobalConfig struct {
	ListenAddr   string `yaml:"listenAddr"`
	LoggingLevel string `yaml:"logging"`
}

type BeaconConfig struct {
	BeaconUpstreams []node.Config `yaml:"upstreams"`
	NetworkID       uint64        `yaml:"networkID"`
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

	return nil
}
