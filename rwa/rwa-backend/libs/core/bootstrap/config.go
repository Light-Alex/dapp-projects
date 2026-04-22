package bootstrap

import (
	"os"

	"gopkg.in/yaml.v3"
)

// LoadConfig reads a YAML config file and unmarshals it into the given type.
func LoadConfig[T any](configFile string) (*T, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var conf T
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
