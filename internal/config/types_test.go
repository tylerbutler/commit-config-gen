package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "commit-types.json")
	data := []byte(`{
  "description": "test config",
  "types": {
    "feat": {"description": "A new feature", "changelog_group": "Added", "bump": "minor"},
    "fix": {"description": "A bug fix", "changelog_group": "Fixed", "bump": "patch"},
    "chore": {"description": "Chores"}
  },
  "excluded_scopes": ["release"],
  "commitlint_rules": {
    "body-max-line-length": [0, "always", 200]
  }
}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Description != "test config" {
		t.Errorf("expected description 'test config', got %q", cfg.Description)
	}
	if len(cfg.Types) != 3 {
		t.Errorf("expected 3 types, got %d", len(cfg.Types))
	}
	if cfg.Types["feat"].Description != "A new feature" {
		t.Errorf("unexpected feat description: %q", cfg.Types["feat"].Description)
	}
	if cfg.Types["feat"].Bump != "minor" {
		t.Errorf("expected feat bump 'minor', got %q", cfg.Types["feat"].Bump)
	}
	if len(cfg.ExcludedScopes) != 1 || cfg.ExcludedScopes[0] != "release" {
		t.Errorf("unexpected excluded_scopes: %v", cfg.ExcludedScopes)
	}
	if len(cfg.CommitlintRules) != 1 {
		t.Errorf("expected 1 commitlint rule, got %d", len(cfg.CommitlintRules))
	}
}

func TestLoadChangelogGroupNullability(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "commit-types.json")
	data := []byte(`{
  "types": {
    "feat": {"description": "feature", "changelog_group": "Added"},
    "chore": {"description": "chore", "changelog_group": null},
    "ci": {"description": "ci"}
  }
}`)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Types["feat"].ChangelogGroup == nil {
		t.Error("feat changelog_group should not be nil")
	}
	if *cfg.Types["feat"].ChangelogGroup != "Added" {
		t.Errorf("expected feat changelog_group 'Added', got %q", *cfg.Types["feat"].ChangelogGroup)
	}
	if cfg.Types["chore"].ChangelogGroup != nil {
		t.Error("chore changelog_group should be nil (explicit null)")
	}
	if cfg.Types["ci"].ChangelogGroup != nil {
		t.Error("ci changelog_group should be nil (omitted)")
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/commit-types.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte(`{not valid json}`), 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadEmptyTypes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(path, []byte(`{"types": {}}`), 0o644); err != nil {
		t.Fatalf("writing test file: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Types) != 0 {
		t.Errorf("expected 0 types, got %d", len(cfg.Types))
	}
}

func TestVisibleTypes(t *testing.T) {
	added := "Added"
	fixed := "Fixed"
	cfg := &Config{
		Types: map[string]CommitType{
			"feat":  {Description: "feature", ChangelogGroup: &added},
			"fix":   {Description: "bug fix", ChangelogGroup: &fixed},
			"chore": {Description: "chore"},
			"ci":    {Description: "ci"},
		},
	}

	visible := cfg.VisibleTypes()
	if len(visible) != 2 {
		t.Errorf("expected 2 visible types, got %d", len(visible))
	}
	if _, ok := visible["feat"]; !ok {
		t.Error("feat should be visible")
	}
	if _, ok := visible["fix"]; !ok {
		t.Error("fix should be visible")
	}
	if _, ok := visible["chore"]; ok {
		t.Error("chore should not be visible")
	}
}

func TestVisibleTypesNone(t *testing.T) {
	cfg := &Config{
		Types: map[string]CommitType{
			"chore": {Description: "chore"},
			"ci":    {Description: "ci"},
		},
	}

	visible := cfg.VisibleTypes()
	if len(visible) != 0 {
		t.Errorf("expected 0 visible types, got %d", len(visible))
	}
}

func TestVisibleTypesAll(t *testing.T) {
	added := "Added"
	fixed := "Fixed"
	cfg := &Config{
		Types: map[string]CommitType{
			"feat": {Description: "feature", ChangelogGroup: &added},
			"fix":  {Description: "bug fix", ChangelogGroup: &fixed},
		},
	}

	visible := cfg.VisibleTypes()
	if len(visible) != 2 {
		t.Errorf("expected 2 visible types, got %d", len(visible))
	}
}

func TestTypeNamesPredefinedOrder(t *testing.T) {
	added := "Added"
	cfg := &Config{
		Types: map[string]CommitType{
			"chore":    {Description: "chore"},
			"feat":     {Description: "feature", ChangelogGroup: &added},
			"fix":      {Description: "fix"},
			"docs":     {Description: "docs"},
			"refactor": {Description: "refactor"},
		},
	}

	names := cfg.TypeNames()
	if len(names) != 5 {
		t.Fatalf("expected 5 names, got %d", len(names))
	}

	// Verify predefined order: feat, fix, refactor, docs, chore
	expected := []string{"feat", "fix", "refactor", "docs", "chore"}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("position %d: expected %q, got %q", i, expected[i], name)
		}
	}
}

func TestTypeNamesUnknownTypesAppended(t *testing.T) {
	cfg := &Config{
		Types: map[string]CommitType{
			"feat":   {Description: "feature"},
			"custom": {Description: "custom type"},
		},
	}

	names := cfg.TypeNames()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}

	// feat comes first (predefined order), custom is appended
	if names[0] != "feat" {
		t.Errorf("expected feat first, got %q", names[0])
	}
	if names[1] != "custom" {
		t.Errorf("expected custom second, got %q", names[1])
	}
}

func TestTypeNamesEmpty(t *testing.T) {
	cfg := &Config{
		Types: map[string]CommitType{},
	}

	names := cfg.TypeNames()
	if len(names) != 0 {
		t.Errorf("expected 0 names, got %d", len(names))
	}
}

func TestTypeNamesAllUnknown(t *testing.T) {
	cfg := &Config{
		Types: map[string]CommitType{
			"alpha":  {Description: "alpha"},
			"beta":   {Description: "beta"},
			"gamma":  {Description: "gamma"},
		},
	}

	names := cfg.TypeNames()
	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}

	// All three should be present (order among unknowns is map iteration order, not guaranteed)
	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	for _, expected := range []string{"alpha", "beta", "gamma"} {
		if !found[expected] {
			t.Errorf("missing expected type %q", expected)
		}
	}
}

func TestTypeNamesFullPredefinedOrder(t *testing.T) {
	// All predefined types present — verify full ordering
	cfg := &Config{
		Types: map[string]CommitType{
			"revert":      {Description: "revert"},
			"data":        {Description: "data"},
			"ci":          {Description: "ci"},
			"chore":       {Description: "chore"},
			"build":       {Description: "build"},
			"test":        {Description: "test"},
			"style":       {Description: "style"},
			"docs":        {Description: "docs"},
			"refactor":    {Description: "refactor"},
			"perf":        {Description: "perf"},
			"improvement": {Description: "improvement"},
			"fix":         {Description: "fix"},
			"feat":        {Description: "feat"},
		},
	}

	names := cfg.TypeNames()
	expected := []string{"feat", "fix", "improvement", "perf", "refactor", "docs", "style", "test", "build", "ci", "chore", "revert", "data"}

	if len(names) != len(expected) {
		t.Fatalf("expected %d names, got %d", len(expected), len(names))
	}
	for i, name := range names {
		if name != expected[i] {
			t.Errorf("position %d: expected %q, got %q", i, expected[i], name)
		}
	}
}
