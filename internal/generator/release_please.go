package generator

import (
	"encoding/json"
	"fmt"

	"github.com/tylerbutler/commit-config-gen/internal/config"
)

func init() {
	Register(&ReleasePleaseGenerator{})
}

// ReleasePleaseGenerator generates release-please-config.json.
type ReleasePleaseGenerator struct{}

func (g *ReleasePleaseGenerator) Name() string     { return "release-please" }
func (g *ReleasePleaseGenerator) FileName() string { return "release-please-config.json" }

func (g *ReleasePleaseGenerator) Generate(cfg *config.Config, existing []byte) ([]byte, error) {
	sections := buildChangelogSections(cfg)

	if existing != nil {
		return mergeReleasePlease(existing, sections)
	}
	return freshReleasePlease(sections)
}

type changelogSection struct {
	Type    string `json:"type"`
	Section string `json:"section,omitempty"`
	Hidden  bool   `json:"hidden,omitempty"`
}

func buildChangelogSections(cfg *config.Config) []changelogSection {
	var sections []changelogSection
	for _, name := range cfg.TypeNames() {
		t := cfg.Types[name]
		entry := changelogSection{Type: name}
		if t.ChangelogGroup != nil {
			entry.Section = *t.ChangelogGroup
		} else {
			entry.Hidden = true
		}
		sections = append(sections, entry)
	}
	return sections
}

func freshReleasePlease(sections []changelogSection) ([]byte, error) {
	doc := map[string]any{
		"packages": map[string]any{
			".": map[string]any{
				"changelog-sections": sections,
			},
		},
	}
	return marshalJSON(doc)
}

func mergeReleasePlease(existing []byte, sections []changelogSection) ([]byte, error) {
	var doc map[string]any
	if err := json.Unmarshal(existing, &doc); err != nil {
		return nil, fmt.Errorf("parsing existing release-please-config.json: %w", err)
	}

	packages, ok := doc["packages"].(map[string]any)
	if !ok {
		packages = map[string]any{}
		doc["packages"] = packages
	}

	root, ok := packages["."].(map[string]any)
	if !ok {
		root = map[string]any{}
		packages["."] = root
	}

	root["changelog-sections"] = sections
	return marshalJSON(doc)
}
