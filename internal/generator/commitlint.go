package generator

import (
	"encoding/json"

	"github.com/tylerbutler/commit-config-gen/internal/config"
)

// CommitlintConfig represents the .commitlintrc.json structure
type CommitlintConfig struct {
	Extends []string       `json:"extends"`
	Rules   map[string]any `json:"rules"`
}

// GenerateCommitlint generates .commitlintrc.json content from config
func GenerateCommitlint(cfg *config.Config) (string, error) {
	typeNames := cfg.TypeNames()

	rules := map[string]any{
		"type-enum": []any{2, "always", typeNames},
	}

	// Merge in additional rules from config
	for name, rule := range cfg.CommitlintRules {
		rules[name] = rule
	}

	commitlintCfg := CommitlintConfig{
		Extends: []string{"@commitlint/config-conventional"},
		Rules:   rules,
	}

	data, err := json.MarshalIndent(commitlintCfg, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data) + "\n", nil
}
