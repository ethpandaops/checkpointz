package proxy

import "fmt"

type Config struct {
	BeaconConfig `yaml:"beacon"`
	GlobalConfig `yaml:"global"`
}

type GlobalConfig struct {
	ListenAddr   string `yaml:"listenAddr"`
	LoggingLevel string `yaml:"logging"`
}

type BeaconConfig struct {
	BeaconUpstreams     []BeaconUpstream `yaml:"upstreams"`
	APIAllowPath        []string         `yaml:"apiAllowPaths"`
	ProxyTimeoutSeconds uint             `yaml:"proxyTimeoutSeconds"`
}

type BeaconUpstream struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
}

func (c *Config) Validate() error {
	// Check that all upstreams have different names and addresses
	duplicates := make(map[string]struct{})
	for _, u := range c.BeaconConfig.BeaconUpstreams {
		if _, ok := duplicates[u.Name]; ok {
			return fmt.Errorf("there's a duplicate upstream with the same name: %s", u.Name)
		} else {
			duplicates[u.Name] = struct{}{}
		}
		if _, ok := duplicates[u.Address]; ok {
			return fmt.Errorf("there's a duplicate upstream with the same address: %s", u.Address)
		} else {
			duplicates[u.Name] = struct{}{}
		}
	}
	return nil
}
