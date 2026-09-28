// Package config handles loading and saving giteasy's user-level settings.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// PrefixMode controls how giteasy derives a default branch-name prefix.
type PrefixMode string

const (
	PrefixGitUserFirstName PrefixMode = "gitUserFirstName"
	PrefixCustom           PrefixMode = "custom"
	PrefixNone             PrefixMode = "none"
)

// DeleteAfterMerge controls whether the source branch is deleted post-merge.
type DeleteAfterMerge string

const (
	DeleteAlways DeleteAfterMerge = "always"
	DeleteNever  DeleteAfterMerge = "never"
	DeleteAsk    DeleteAfterMerge = "ask"
)

type BranchConfig struct {
	PrefixMode   PrefixMode `json:"prefixMode"`
	CustomPrefix string     `json:"customPrefix"`
}

type MergeConfig struct {
	DeleteAfterMerge    DeleteAfterMerge `json:"deleteAfterMerge"`
	ConfirmNonDevToMain bool             `json:"confirmNonDevToMain"`
}

type DefaultBranches struct {
	Primary   string `json:"primary"`   // e.g. "main" or "master"
	Secondary string `json:"secondary"` // e.g. "dev"
}

type Config struct {
	Branch          BranchConfig    `json:"branch"`
	Merge           MergeConfig     `json:"merge"`
	DefaultBranches DefaultBranches `json:"defaultBranches"`
}

// Default returns giteasy's out-of-the-box settings.
func Default() *Config {
	return &Config{
		Branch: BranchConfig{
			PrefixMode:   PrefixNone,
			CustomPrefix: "",
		},
		Merge: MergeConfig{
			DeleteAfterMerge:    DeleteAsk,
			ConfirmNonDevToMain: true,
		},
		DefaultBranches: DefaultBranches{
			Primary:   "main",
			Secondary: "dev",
		},
	}
}

// Path returns the path to giteasy's config file (~/.giteasy/config.json).
func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".giteasy", "config.json"), nil
}

// Load reads the config file, creating a default one on first run.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := Default()
		if saveErr := cfg.Save(); saveErr != nil {
			return cfg, nil // still usable in-memory even if save failed
		}
		return cfg, nil
	}
	if err != nil {
		return nil, err
	}
	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Save writes the config to disk, creating ~/.giteasy if needed.
func (c *Config) Save() error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
