package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ScanPaths   []string          `yaml:"scan_paths"`
	Exclude     []string          `yaml:"exclude"`
	LastPath    string            `yaml:"last_path"`
	RepoStates  map[string]bool   `yaml:"repo_states"` // path -> active
}

const stateFile = ".gitmux-state"

func Load() (*Config, error) {
	// Find config file
	home, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("could not find home directory: %w", err)
	}

	configPaths := []string{
		"./config.yaml",
		filepath.Join(home, ".gitmux.yaml"),
		filepath.Join(home, ".config", "gitmux.yaml"),
	}

	var cfg *Config
	for _, path := range configPaths {
		absPath, err := expandPath(path)
		if err != nil {
			continue
		}

		data, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}

		cfg = &Config{}
		if err := yaml.Unmarshal(data, cfg); err != nil {
			continue
		}

		// Expand ~ in scan paths
		for i, p := range cfg.ScanPaths {
			cfg.ScanPaths[i], _ = expandPath(p)
		}

		return cfg, nil
	}

	// Default config if no config file found
	return &Config{
		ScanPaths: []string{filepath.Join(home, "Documents", "runcloud")},
		Exclude:   []string{"node_modules", "vendor", ".git", "target", "dist", "build"},
	}, nil
}

func expandPath(path string) (string, error) {
	if len(path) > 1 && path[:2] == "~/" {
		home, err := homedir.Dir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, path[2:]), nil
	}
	return filepath.Abs(path)
}

// GetLastPath returns the last scanned directory
func GetLastPath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	statePath := filepath.Join(home, stateFile)
	data, err := os.ReadFile(statePath)
	if err != nil {
		return "", err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return "", err
	}

	return cfg.LastPath, nil
}

// SaveLastPath saves the last scanned directory
func SaveLastPath(path string) error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	statePath := filepath.Join(home, stateFile)

	// Read existing state if it exists
	cfg := &Config{}
	if data, err := os.ReadFile(statePath); err == nil {
		_ = yaml.Unmarshal(data, cfg)
	}

	cfg.LastPath = path

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0644)
}

// GetRepoStates returns the map of repo paths to their active state
func GetRepoStates() (map[string]bool, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, err
	}

	statePath := filepath.Join(home, stateFile)
	data, err := os.ReadFile(statePath)
	if err != nil {
		return make(map[string]bool), nil
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return make(map[string]bool), nil
	}

	if cfg.RepoStates == nil {
		return make(map[string]bool), nil
	}

	return cfg.RepoStates, nil
}

// SetRepoState sets the active state for a specific repo
func SetRepoState(path string, active bool) error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	statePath := filepath.Join(home, stateFile)

	// Read existing state
	cfg := &Config{}
	if data, err := os.ReadFile(statePath); err == nil {
		_ = yaml.Unmarshal(data, cfg)
	}

	if cfg.RepoStates == nil {
		cfg.RepoStates = make(map[string]bool)
	}

	cfg.RepoStates[path] = active

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(statePath, data, 0644)
}
