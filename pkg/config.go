package pkg

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"time"
)

const (
	// Duration for which a node's cached connected peers remain valid.
	connectedPeersExpiration = 1 * time.Minute
)

// Config holds the RPC endpoints and network definitions.
type Config struct {
	RPCs     map[string]string   `yaml:"rpcs"`
	Networks map[string][]string `yaml:"networks"`
	Interval time.Duration       `yaml:"interval" default:"60s"`
}

// LoadConfig reads and unmarshals the config YAML file.
func LoadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
