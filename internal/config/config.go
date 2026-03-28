package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const filename = ".lk9s.yaml"

type Context struct {
	Name      string `yaml:"name"`
	URL       string `yaml:"url"`
	APIKey    string `yaml:"api-key"`
	APISecret string `yaml:"api-secret"`
}

type Config struct {
	Contexts []Context `yaml:"contexts"`
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}

	data, err := os.ReadFile(filepath.Join(home, filename))
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if len(cfg.Contexts) == 0 {
		return nil, fmt.Errorf("no contexts defined in ~/%s", filename)
	}

	for _, ctx := range cfg.Contexts {
		if ctx.Name == "" || ctx.URL == "" || ctx.APIKey == "" || ctx.APISecret == "" {
			return nil, fmt.Errorf("context %q is missing required fields (url, api-key, api-secret)", ctx.Name)
		}
	}

	return &cfg, nil
}
