package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"gopkg.in/yaml.v3"
)

const filename = ".lk9s.yaml"

type Context struct {
	Name      string `yaml:"name"`
	URL       string `yaml:"url"`
	APIKey    string `yaml:"api-key"`
	APISecret string `yaml:"api-secret"`
	// Write must be explicitly set to allow destructive/mutating actions
	// (delete room, kick participant, mute track, edit permissions, create
	// room) against this context. Contexts default to read-only.
	Write bool `yaml:"write"`
}

type Config struct {
	Contexts []Context `yaml:"contexts"`
	// Warnings are non-fatal problems found while loading, for the caller
	// to show the user.
	Warnings []string `yaml:"-"`
}

// Load reads the config from ~/.lk9s.yaml.
func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}

	return LoadFile(filepath.Join(home, filename))
}

// LoadFile reads and validates the config at path.
func LoadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if len(cfg.Contexts) == 0 {
		return nil, fmt.Errorf("no contexts defined in %s", path)
	}

	names := make(map[string]bool, len(cfg.Contexts))

	for _, ctx := range cfg.Contexts {
		if ctx.Name == "" || ctx.URL == "" || ctx.APIKey == "" || ctx.APISecret == "" {
			return nil, fmt.Errorf("context %q is missing required fields (name, url, api-key, api-secret)", ctx.Name)
		}

		if names[ctx.Name] {
			return nil, fmt.Errorf("duplicate context name %q", ctx.Name)
		}

		names[ctx.Name] = true
	}

	if w := permissionWarning(path); w != "" {
		cfg.Warnings = append(cfg.Warnings, w)
	}

	return &cfg, nil
}

// permissionWarning flags a config file that other users can read, since it
// holds API secrets. File modes carry no such meaning on Windows.
func permissionWarning(path string) string {
	if runtime.GOOS == "windows" {
		return ""
	}

	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm()&0o077 == 0 {
		return ""
	}

	return fmt.Sprintf("%s is accessible by other users (mode %04o); it holds API secrets, run: chmod 600 %s",
		path, info.Mode().Perm(), path)
}
