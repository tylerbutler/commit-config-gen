package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// CommitType defines a single commit type configuration
type CommitType struct {
	Description    string  `json:"description"`
	ChangelogGroup *string `json:"changelog_group"` // nil means excluded from changelog
	Bump           string  `json:"bump,omitempty"`  // "major", "minor", "patch", or "none"
}

// CommitlintRule represents a commitlint rule configuration
type CommitlintRule = []any

// Config represents the commit-types.json structure
type Config struct {
	Schema          string                    `json:"$schema,omitempty"`
	Description     string                    `json:"description,omitempty"`
	Types           map[string]CommitType     `json:"types"`
	ExcludedScopes  []string                  `json:"excluded_scopes,omitempty"`
	CommitlintRules map[string]CommitlintRule `json:"commitlint_rules,omitempty"`
}

// Load reads and parses a commit-types.json file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &cfg, nil
}

// VisibleTypes returns only types that have a changelog group
func (c *Config) VisibleTypes() map[string]CommitType {
	visible := make(map[string]CommitType)
	for name, t := range c.Types {
		if t.ChangelogGroup != nil {
			visible[name] = t
		}
	}
	return visible
}

// TypeNames returns all type names in a consistent order
func (c *Config) TypeNames() []string {
	// Use a predefined order for consistency
	order := []string{"feat", "fix", "improvement", "perf", "refactor", "docs", "style", "test", "build", "ci", "chore", "revert", "data"}
	var names []string
	for _, name := range order {
		if _, ok := c.Types[name]; ok {
			names = append(names, name)
		}
	}
	// Add any types not in the predefined order
	for name := range c.Types {
		found := false
		for _, n := range names {
			if n == name {
				found = true
				break
			}
		}
		if !found {
			names = append(names, name)
		}
	}
	return names
}
