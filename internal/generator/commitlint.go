package generator

import (
	"encoding/json"
	"fmt"

	"github.com/tylerbutler/commit-config-gen/internal/config"
)

func init() {
	Register(&CommitlintGenerator{})
}

// CommitlintGenerator generates .commitlintrc.json for commitlint.
type CommitlintGenerator struct{}

func (g *CommitlintGenerator) Name() string     { return "commitlint" }
func (g *CommitlintGenerator) FileName() string { return ".commitlintrc.json" }

func (g *CommitlintGenerator) Generate(cfg *config.Config, existing []byte) ([]byte, error) {
	typeNames := cfg.TypeNames()

	rules := map[string]any{
		"type-enum": []any{2, "always", typeNames},
	}
	for name, rule := range cfg.CommitlintRules {
		rules[name] = rule
	}

	if existing != nil {
		return mergeCommitlint(existing, rules)
	}
	return freshCommitlint(rules)
}

func freshCommitlint(rules map[string]any) ([]byte, error) {
	doc := map[string]any{
		"extends": []string{"@commitlint/config-conventional"},
		"rules":   rules,
	}
	return marshalJSON(doc)
}

func mergeCommitlint(existing []byte, rules map[string]any) ([]byte, error) {
	var doc map[string]any
	if err := json.Unmarshal(existing, &doc); err != nil {
		return nil, fmt.Errorf("parsing existing .commitlintrc.json: %w", err)
	}

	existingRules, ok := doc["rules"].(map[string]any)
	if !ok {
		existingRules = map[string]any{}
	}
	for k, v := range rules {
		existingRules[k] = v
	}
	doc["rules"] = existingRules

	return marshalJSON(doc)
}

func marshalJSON(v any) ([]byte, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
