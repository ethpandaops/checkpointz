package node

type Config struct {
	Name                string `yaml:"name"`
	Address             string `yaml:"address"`
	ProxyTimeoutSeconds uint   `yaml:"proxyTimeoutSeconds"`
	DataProvider        bool   `yaml:"dataProvider"`
}
