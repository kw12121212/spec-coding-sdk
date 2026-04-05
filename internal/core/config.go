package core

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the SDK configuration. Fields will be added by later milestones.
type Config struct{}

// LoadConfig reads and parses the YAML file at the given path, returning a validated Config.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}

	return &cfg, nil
}
