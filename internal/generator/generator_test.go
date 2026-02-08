package generator

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/tylerbutler/commit-config-gen/internal/config"
	"gopkg.in/yaml.v3"
)

func testConfig() *config.Config {
	feat := "Features"
	fix := "Bug Fixes"
	return &config.Config{
		Types: map[string]config.CommitType{
			"feat":  {Description: "A new feature", ChangelogGroup: &feat, Bump: "minor"},
			"fix":   {Description: "A bug fix", ChangelogGroup: &fix, Bump: "patch"},
			"chore": {Description: "Other changes"},
		},
		ExcludedScopes:  []string{"deps"},
		CommitlintRules: map[string]config.CommitlintRule{"header-max-length": {2, "always", 100}},
	}
}

// --- Registry tests ---

func TestRegistryAll(t *testing.T) {
	gens := All()
	if len(gens) != 7 {
		t.Errorf("expected 7 registered generators, got %d", len(gens))
	}
}

func TestRegistrySorted(t *testing.T) {
	names := Names()
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Errorf("generators not sorted: %s before %s", names[i-1], names[i])
		}
	}
}

func TestRegistryGet(t *testing.T) {
	g, err := Get("cliff")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g.Name() != "cliff" {
		t.Errorf("expected cliff, got %s", g.Name())
	}

	_, err = Get("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent generator")
	}
}

// --- Per-generator tests ---

func TestCliffFresh(t *testing.T) {
	g := &CliffGenerator{}
	out, err := g.Generate(testConfig(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)
	if !strings.Contains(s, `message = '^feat'`) {
		t.Error("missing feat parser")
	}
	if !strings.Contains(s, `group = 'Features'`) {
		t.Error("missing Features group")
	}
	if !strings.Contains(s, `skip = true`) {
		t.Error("missing skip rule for excluded scope")
	}
	if !strings.Contains(s, `'.*'`) {
		t.Error("missing catch-all rule")
	}
}

func TestCliffMerge(t *testing.T) {
	existing := []byte(`[changelog]
header = "custom header"

[git]
conventional_commits = true
commit_parsers = [
    { message = "old", group = "Old" },
]
`)
	g := &CliffGenerator{}
	out, err := g.Generate(testConfig(), existing)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc map[string]any
	if err := toml.Unmarshal(out, &doc); err != nil {
		t.Fatalf("output not valid TOML: %v", err)
	}

	// Verify custom header is preserved
	cl, ok := doc["changelog"].(map[string]any)
	if !ok {
		t.Fatal("missing changelog section")
	}
	if cl["header"] != "custom header" {
		t.Error("custom header was not preserved")
	}

	// Verify parsers were replaced
	git := doc["git"].(map[string]any)
	parsers := git["commit_parsers"].([]any)
	found := false
	for _, p := range parsers {
		pm := p.(map[string]any)
		if pm["message"] == "^feat" {
			found = true
		}
		if pm["message"] == "old" {
			t.Error("old parser should have been replaced")
		}
	}
	if !found {
		t.Error("feat parser not found in merged output")
	}
}

func TestCliffIdempotent(t *testing.T) {
	g := &CliffGenerator{}
	cfg := testConfig()

	first, err := g.Generate(cfg, nil)
	if err != nil {
		t.Fatalf("first generate: %v", err)
	}

	second, err := g.Generate(cfg, first)
	if err != nil {
		t.Fatalf("second generate: %v", err)
	}

	third, err := g.Generate(cfg, second)
	if err != nil {
		t.Fatalf("third generate: %v", err)
	}

	if !bytes.Equal(second, third) {
		t.Error("cliff generator is not idempotent after merge")
	}
}

func TestCommitlintFresh(t *testing.T) {
	g := &CommitlintGenerator{}
	out, err := g.Generate(testConfig(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}

	extends := doc["extends"].([]any)
	if len(extends) == 0 || extends[0] != "@commitlint/config-conventional" {
		t.Error("missing extends")
	}

	rules := doc["rules"].(map[string]any)
	if _, ok := rules["type-enum"]; !ok {
		t.Error("missing type-enum rule")
	}
	if _, ok := rules["header-max-length"]; !ok {
		t.Error("missing header-max-length rule")
	}
}

func TestCommitlintMerge(t *testing.T) {
	existing := []byte(`{
  "extends": ["custom-config"],
  "rules": {
    "custom-rule": [2, "always"],
    "type-enum": [2, "always", ["old"]]
  }
}
`)
	g := &CommitlintGenerator{}
	out, err := g.Generate(testConfig(), existing)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}

	// Custom extends preserved
	extends := doc["extends"].([]any)
	if extends[0] != "custom-config" {
		t.Error("custom extends was not preserved")
	}

	rules := doc["rules"].(map[string]any)
	// Custom rule preserved
	if _, ok := rules["custom-rule"]; !ok {
		t.Error("custom-rule was not preserved")
	}
	// type-enum updated
	typeEnum := rules["type-enum"].([]any)
	types := typeEnum[2].([]any)
	if len(types) <= 1 {
		t.Error("type-enum was not updated with all types")
	}
}

func TestCommitlintIdempotent(t *testing.T) {
	g := &CommitlintGenerator{}
	cfg := testConfig()

	first, err := g.Generate(cfg, nil)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := g.Generate(cfg, first)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	third, err := g.Generate(cfg, second)
	if err != nil {
		t.Fatalf("third: %v", err)
	}
	if !bytes.Equal(second, third) {
		t.Error("commitlint generator is not idempotent after merge")
	}
}

func TestConventionalChangelogFresh(t *testing.T) {
	g := &ConventionalChangelogGenerator{}
	out, err := g.Generate(testConfig(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}

	types := doc["types"].([]any)
	if len(types) != 3 {
		t.Errorf("expected 3 types, got %d", len(types))
	}

	// Check chore is hidden
	for _, entry := range types {
		m := entry.(map[string]any)
		if m["type"] == "chore" {
			if hidden, ok := m["hidden"].(bool); !ok || !hidden {
				t.Error("chore should be hidden")
			}
		}
	}
}

func TestConventionalChangelogMerge(t *testing.T) {
	existing := []byte(`{
  "header": "Custom Header",
  "types": [{"type": "old", "section": "Old"}]
}
`)
	g := &ConventionalChangelogGenerator{}
	out, err := g.Generate(testConfig(), existing)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}

	if doc["header"] != "Custom Header" {
		t.Error("custom header not preserved")
	}

	types := doc["types"].([]any)
	if len(types) != 3 {
		t.Errorf("expected 3 types, got %d", len(types))
	}
}

func TestConventionalChangelogIdempotent(t *testing.T) {
	g := &ConventionalChangelogGenerator{}
	cfg := testConfig()

	first, err := g.Generate(cfg, nil)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := g.Generate(cfg, first)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	third, err := g.Generate(cfg, second)
	if err != nil {
		t.Fatalf("third: %v", err)
	}
	if !bytes.Equal(second, third) {
		t.Error("conventional-changelog generator is not idempotent")
	}
}

func TestReleasePleaseFresh(t *testing.T) {
	g := &ReleasePleaseGenerator{}
	out, err := g.Generate(testConfig(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}

	packages := doc["packages"].(map[string]any)
	root := packages["."].(map[string]any)
	sections := root["changelog-sections"].([]any)
	if len(sections) != 3 {
		t.Errorf("expected 3 changelog sections, got %d", len(sections))
	}
}

func TestReleasePleaseMerge(t *testing.T) {
	existing := []byte(`{
  "release-type": "node",
  "packages": {
    ".": {
      "component": "myapp",
      "changelog-sections": [{"type": "old", "section": "Old"}]
    }
  }
}
`)
	g := &ReleasePleaseGenerator{}
	out, err := g.Generate(testConfig(), existing)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}

	if doc["release-type"] != "node" {
		t.Error("release-type not preserved")
	}

	packages := doc["packages"].(map[string]any)
	root := packages["."].(map[string]any)
	if root["component"] != "myapp" {
		t.Error("component not preserved")
	}

	sections := root["changelog-sections"].([]any)
	if len(sections) != 3 {
		t.Errorf("expected 3 sections, got %d", len(sections))
	}
}

func TestReleasePleaseIdempotent(t *testing.T) {
	g := &ReleasePleaseGenerator{}
	cfg := testConfig()

	first, err := g.Generate(cfg, nil)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := g.Generate(cfg, first)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	third, err := g.Generate(cfg, second)
	if err != nil {
		t.Fatalf("third: %v", err)
	}
	if !bytes.Equal(second, third) {
		t.Error("release-please generator is not idempotent")
	}
}

func TestChangieFresh(t *testing.T) {
	g := &ChangieGenerator{}
	out, err := g.Generate(testConfig(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc map[string]any
	if err := yaml.Unmarshal(out, &doc); err != nil {
		t.Fatalf("output not valid YAML: %v", err)
	}

	kinds := doc["kinds"].([]any)
	// Only visible types (feat, fix) become kinds
	if len(kinds) != 2 {
		t.Errorf("expected 2 kinds, got %d", len(kinds))
	}
}

func TestChangieMerge(t *testing.T) {
	existing := []byte(`changesDir: .changes
headerPath: header.tpl.md
kinds:
  - label: Old
`)
	g := &ChangieGenerator{}
	out, err := g.Generate(testConfig(), existing)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc map[string]any
	if err := yaml.Unmarshal(out, &doc); err != nil {
		t.Fatalf("output not valid YAML: %v", err)
	}

	if doc["changesDir"] != ".changes" {
		t.Error("changesDir not preserved")
	}
	if doc["headerPath"] != "header.tpl.md" {
		t.Error("headerPath not preserved")
	}

	kinds := doc["kinds"].([]any)
	if len(kinds) != 2 {
		t.Errorf("expected 2 kinds, got %d", len(kinds))
	}
}

func TestChangieIdempotent(t *testing.T) {
	g := &ChangieGenerator{}
	cfg := testConfig()

	first, err := g.Generate(cfg, nil)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := g.Generate(cfg, first)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	third, err := g.Generate(cfg, second)
	if err != nil {
		t.Fatalf("third: %v", err)
	}
	if !bytes.Equal(second, third) {
		t.Error("changie generator is not idempotent")
	}
}

func TestSemanticReleaseFresh(t *testing.T) {
	g := &SemanticReleaseGenerator{}
	out, err := g.Generate(testConfig(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}

	plugins := doc["plugins"].([]any)
	if len(plugins) < 2 {
		t.Fatal("expected at least 2 plugins")
	}

	// Check commit-analyzer has releaseRules
	analyzer := plugins[0].([]any)
	if analyzer[0] != "@semantic-release/commit-analyzer" {
		t.Error("first plugin should be commit-analyzer")
	}
	analyzerCfg := analyzer[1].(map[string]any)
	if _, ok := analyzerCfg["releaseRules"]; !ok {
		t.Error("missing releaseRules")
	}
}

func TestSemanticReleaseMerge(t *testing.T) {
	existing := []byte(`{
  "branches": ["main", "next"],
  "plugins": [
    ["@semantic-release/commit-analyzer", {
      "preset": "conventionalcommits",
      "releaseRules": [{"type": "old", "release": "patch"}],
      "presetConfig": {"types": []}
    }],
    ["@semantic-release/release-notes-generator", {
      "preset": "conventionalcommits",
      "presetConfig": {"types": []}
    }],
    "@semantic-release/npm"
  ]
}
`)
	g := &SemanticReleaseGenerator{}
	out, err := g.Generate(testConfig(), existing)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("output not valid JSON: %v", err)
	}

	// Custom branches preserved
	branches := doc["branches"].([]any)
	if len(branches) != 2 {
		t.Error("branches not preserved")
	}

	// Check commit-analyzer preset preserved, releaseRules updated
	plugins := doc["plugins"].([]any)
	analyzer := plugins[0].([]any)
	analyzerCfg := analyzer[1].(map[string]any)
	if analyzerCfg["preset"] != "conventionalcommits" {
		t.Error("preset not preserved")
	}

	rules := analyzerCfg["releaseRules"].([]any)
	hasOld := false
	for _, r := range rules {
		rm := r.(map[string]any)
		if rm["type"] == "old" {
			hasOld = true
		}
	}
	if hasOld {
		t.Error("old release rule should have been replaced")
	}
}

func TestSemanticReleaseIdempotent(t *testing.T) {
	g := &SemanticReleaseGenerator{}
	cfg := testConfig()

	first, err := g.Generate(cfg, nil)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := g.Generate(cfg, first)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	third, err := g.Generate(cfg, second)
	if err != nil {
		t.Fatalf("third: %v", err)
	}
	if !bytes.Equal(second, third) {
		t.Error("semantic-release generator is not idempotent")
	}
}

func TestReleasePlzFresh(t *testing.T) {
	g := &ReleasePlzGenerator{}
	out, err := g.Generate(testConfig(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc map[string]any
	if err := toml.Unmarshal(out, &doc); err != nil {
		t.Fatalf("output not valid TOML: %v", err)
	}

	changelog := doc["changelog"].(map[string]any)
	parsers := changelog["commit_parsers"].([]any)
	if len(parsers) < 3 {
		t.Errorf("expected at least 3 parsers, got %d", len(parsers))
	}
}

func TestReleasePlzMerge(t *testing.T) {
	existing := []byte(`[workspace]
allow_dirty = true

[changelog]
commit_parsers = [
    { message = "old", group = "Old" },
]
`)
	g := &ReleasePlzGenerator{}
	out, err := g.Generate(testConfig(), existing)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var doc map[string]any
	if err := toml.Unmarshal(out, &doc); err != nil {
		t.Fatalf("output not valid TOML: %v", err)
	}

	// workspace preserved
	ws := doc["workspace"].(map[string]any)
	if ws["allow_dirty"] != true {
		t.Error("workspace.allow_dirty not preserved")
	}

	changelog := doc["changelog"].(map[string]any)
	parsers := changelog["commit_parsers"].([]any)
	hasOld := false
	for _, p := range parsers {
		pm := p.(map[string]any)
		if pm["message"] == "old" {
			hasOld = true
		}
	}
	if hasOld {
		t.Error("old parser should have been replaced")
	}
}

func TestReleasePlzIdempotent(t *testing.T) {
	g := &ReleasePlzGenerator{}
	cfg := testConfig()

	first, err := g.Generate(cfg, nil)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := g.Generate(cfg, first)
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	third, err := g.Generate(cfg, second)
	if err != nil {
		t.Fatalf("third: %v", err)
	}
	if !bytes.Equal(second, third) {
		t.Error("release-plz generator is not idempotent")
	}
}
