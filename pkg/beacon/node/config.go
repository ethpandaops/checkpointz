package node

type Config struct {
	Name         string `yaml:"name"`
	Address      string `yaml:"address"`
	DataProvider bool   `yaml:"dataProvider"`
}
