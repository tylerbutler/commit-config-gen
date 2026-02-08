package generator

import (
	"encoding/json"
	"fmt"

	"github.com/tylerbutler/commit-config-gen/internal/config"
)

func init() {
	Register(&SemanticReleaseGenerator{})
}

// SemanticReleaseGenerator generates .releaserc.json for semantic-release.
type SemanticReleaseGenerator struct{}

func (g *SemanticReleaseGenerator) Name() string     { return "semantic-release" }
func (g *SemanticReleaseGenerator) FileName() string { return ".releaserc.json" }

func (g *SemanticReleaseGenerator) Generate(cfg *config.Config, existing []byte) ([]byte, error) {
	releaseRules := buildReleaseRules(cfg)
	presetTypes := buildPresetTypes(cfg)

	if existing != nil {
		return mergeSemanticRelease(existing, releaseRules, presetTypes)
	}
	return freshSemanticRelease(releaseRules, presetTypes)
}

type releaseRule struct {
	Type    string `json:"type"`
	Release string `json:"release,omitempty"`
}

type presetType struct {
	Type    string `json:"type"`
	Section string `json:"section,omitempty"`
	Hidden  bool   `json:"hidden,omitempty"`
}

func buildReleaseRules(cfg *config.Config) []releaseRule {
	var rules []releaseRule
	for _, name := range cfg.TypeNames() {
		t := cfg.Types[name]
		rule := releaseRule{Type: name}
		if t.Bump != "" && t.Bump != "none" {
			rule.Release = t.Bump
		} else if t.Bump == "none" {
			rule.Release = ""
		} else {
			// Default: feat->minor, fix->patch, perf->patch
			switch name {
			case "feat":
				rule.Release = "minor"
			case "fix", "perf":
				rule.Release = "patch"
			default:
				continue // skip types without explicit bump and no default
			}
		}
		rules = append(rules, rule)
	}
	return rules
}

func buildPresetTypes(cfg *config.Config) []presetType {
	var types []presetType
	for _, name := range cfg.TypeNames() {
		t := cfg.Types[name]
		entry := presetType{Type: name}
		if t.ChangelogGroup != nil {
			entry.Section = *t.ChangelogGroup
		} else {
			entry.Hidden = true
		}
		types = append(types, entry)
	}
	return types
}

func freshSemanticRelease(releaseRules []releaseRule, presetTypes []presetType) ([]byte, error) {
	plugins := []any{
		[]any{
			"@semantic-release/commit-analyzer",
			map[string]any{
				"releaseRules": releaseRules,
				"presetConfig": map[string]any{
					"types": presetTypes,
				},
			},
		},
		[]any{
			"@semantic-release/release-notes-generator",
			map[string]any{
				"presetConfig": map[string]any{
					"types": presetTypes,
				},
			},
		},
		"@semantic-release/changelog",
		"@semantic-release/npm",
		"@semantic-release/github",
	}

	doc := map[string]any{
		"branches": []string{"main"},
		"plugins":  plugins,
	}
	return marshalJSON(doc)
}

func mergeSemanticRelease(existing []byte, releaseRules []releaseRule, presetTypes []presetType) ([]byte, error) {
	var doc map[string]any
	if err := json.Unmarshal(existing, &doc); err != nil {
		return nil, fmt.Errorf("parsing existing .releaserc.json: %w", err)
	}

	plugins, ok := doc["plugins"].([]any)
	if !ok {
		// No plugins array, generate fresh structure within existing doc
		return freshSemanticRelease(releaseRules, presetTypes)
	}

	for i, plugin := range plugins {
		arr, ok := plugin.([]any)
		if !ok || len(arr) < 2 {
			continue
		}
		name, ok := arr[0].(string)
		if !ok {
			continue
		}
		pluginCfg, ok := arr[1].(map[string]any)
		if !ok {
			continue
		}

		switch name {
		case "@semantic-release/commit-analyzer":
			pluginCfg["releaseRules"] = releaseRules
			pc, ok := pluginCfg["presetConfig"].(map[string]any)
			if !ok {
				pc = map[string]any{}
				pluginCfg["presetConfig"] = pc
			}
			pc["types"] = presetTypes
			arr[1] = pluginCfg
			plugins[i] = arr

		case "@semantic-release/release-notes-generator":
			pc, ok := pluginCfg["presetConfig"].(map[string]any)
			if !ok {
				pc = map[string]any{}
				pluginCfg["presetConfig"] = pc
			}
			pc["types"] = presetTypes
			arr[1] = pluginCfg
			plugins[i] = arr
		}
	}

	doc["plugins"] = plugins
	return marshalJSON(doc)
}
