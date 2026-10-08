package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const Filename = "devcheck.yml"

// Config holds the parsed contents of a devcheck.yml file.
type Config struct {
	// Require lists extra binaries that must be present on PATH.
	Require []string `yaml:"require"`
	// Skip lists check names (as reported by Check.Name()) to suppress.
	Skip []string `yaml:"skip"`
}

// Load reads devcheck.yml from dir.
// If no config file is present, it returns a zero-value Config and no error.
// Malformed YAML returns an error.
func Load(dir string) (Config, error) {
	path := filepath.Join(dir, Filename)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("reading %s: %w", Filename, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing %s: %w", Filename, err)
	}
	return cfg, nil
}

// SkipSet returns the Skip list as a set for O(1) lookup.
func (c Config) SkipSet() map[string]struct{} {
	s := make(map[string]struct{}, len(c.Skip))
	for _, name := range c.Skip {
		s[name] = struct{}{}
	}
	return s
}
