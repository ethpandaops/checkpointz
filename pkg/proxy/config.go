package proxy

import "fmt"

type Config struct {
	GlobalConfig    `yaml:"global"`
	BeaconConfig    `yaml:"beacon"`
	ExecutionConfig `yaml:"execution"`
}

type GlobalConfig struct {
	ListenAddr   string `yaml:"listenAddr"`
	LoggingLevel string `yaml:"logging"`
}

type BeaconConfig struct {
	BeaconUpstreams     []BeaconUpstream `yaml:"upstreams"`
	APIAllowPaths       []string         `yaml:"apiAllowPaths"`
	ProxyTimeoutSeconds uint             `yaml:"proxyTimeoutSeconds"`
}

type ExecutionConfig struct {
	ExecutionUpstreams  []ExecutionUpstream `yaml:"upstreams"`
	RPCAllowMethods     []string            `yaml:"rpcAllowMethods"`
	ProxyTimeoutSeconds uint                `yaml:"proxyTimeoutSeconds"`
}

type BeaconUpstream struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
}

type ExecutionUpstream struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
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
