package store

import "errors"

type Config struct {
	MaxItems int `yaml:"max_items"`
}

func (c *Config) Validate() error {
	if c.MaxItems < 1 {
		return errors.New("max_items must be at least 1")
	}

	return nil
}
