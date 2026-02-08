package generator

import (
	"encoding/json"
	"fmt"

	"github.com/tylerbutler/commit-config-gen/internal/config"
)

func init() {
	Register(&ConventionalChangelogGenerator{})
}

// ConventionalChangelogGenerator generates .versionrc.json for conventional-changelog.
type ConventionalChangelogGenerator struct{}

func (g *ConventionalChangelogGenerator) Name() string     { return "conventional-changelog" }
func (g *ConventionalChangelogGenerator) FileName() string { return ".versionrc.json" }

func (g *ConventionalChangelogGenerator) Generate(cfg *config.Config, existing []byte) ([]byte, error) {
	types := buildVersionRCTypes(cfg)

	if existing != nil {
		return mergeVersionRC(existing, types)
	}
	return freshVersionRC(types)
}

type versionRCType struct {
	Type    string `json:"type"`
	Section string `json:"section,omitempty"`
	Hidden  bool   `json:"hidden,omitempty"`
}

func buildVersionRCTypes(cfg *config.Config) []versionRCType {
	var types []versionRCType
	for _, name := range cfg.TypeNames() {
		t := cfg.Types[name]
		entry := versionRCType{Type: name}
		if t.ChangelogGroup != nil {
			entry.Section = *t.ChangelogGroup
		} else {
			entry.Hidden = true
		}
		types = append(types, entry)
	}
	return types
}

func freshVersionRC(types []versionRCType) ([]byte, error) {
	doc := map[string]any{
		"types": types,
	}
	return marshalJSON(doc)
}

func mergeVersionRC(existing []byte, types []versionRCType) ([]byte, error) {
	var doc map[string]any
	if err := json.Unmarshal(existing, &doc); err != nil {
		return nil, fmt.Errorf("parsing existing .versionrc.json: %w", err)
	}
	doc["types"] = types
	return marshalJSON(doc)
}
