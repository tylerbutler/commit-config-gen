package generator

import (
	"bytes"
	"fmt"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/tylerbutler/commit-config-gen/internal/config"
)

func init() {
	Register(&CliffGenerator{})
}

// CliffGenerator generates cliff.toml for git-cliff.
type CliffGenerator struct{}

func (g *CliffGenerator) Name() string     { return "cliff" }
func (g *CliffGenerator) FileName() string { return "cliff.toml" }

func (g *CliffGenerator) Generate(cfg *config.Config, existing []byte) ([]byte, error) {
	parsers := buildCommitParsers(cfg)

	if existing != nil {
		return mergeCliff(existing, parsers)
	}
	return freshCliff(cfg, parsers)
}

type commitParser struct {
	Message string `toml:"message"`
	Group   string `toml:"group,omitempty"`
	Skip    bool   `toml:"skip,omitempty"`
}

func buildCommitParsers(cfg *config.Config) []commitParser {
	var parsers []commitParser

	for _, scope := range cfg.ExcludedScopes {
		parsers = append(parsers, commitParser{
			Message: fmt.Sprintf(`^[a-z]+\(%s\)`, scope),
			Skip:    true,
		})
	}

	for _, name := range cfg.TypeNames() {
		t := cfg.Types[name]
		if t.ChangelogGroup != nil {
			parsers = append(parsers, commitParser{
				Message: "^" + name,
				Group:   *t.ChangelogGroup,
			})
		}
	}

	parsers = append(parsers, commitParser{
		Message: ".*",
		Group:   "_ignored",
	})

	return parsers
}

func mergeCliff(existing []byte, parsers []commitParser) ([]byte, error) {
	var doc map[string]any
	if err := toml.Unmarshal(existing, &doc); err != nil {
		return nil, fmt.Errorf("parsing existing cliff.toml: %w", err)
	}

	git, ok := doc["git"].(map[string]any)
	if !ok {
		git = map[string]any{}
		doc["git"] = git
	}

	// Convert parsers to []any for TOML serialization
	var parserMaps []any
	for _, p := range parsers {
		m := map[string]any{"message": p.Message}
		if p.Group != "" {
			m["group"] = p.Group
		}
		if p.Skip {
			m["skip"] = p.Skip
		}
		parserMaps = append(parserMaps, m)
	}
	git["commit_parsers"] = parserMaps

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	enc.SetIndentTables(true)
	if err := enc.Encode(doc); err != nil {
		return nil, fmt.Errorf("encoding cliff.toml: %w", err)
	}
	return buf.Bytes(), nil
}

const cliffHeader = `# git-cliff config
# Auto-generated from commit-types.json - do not edit directly
# Run: commit-config-gen generate

[changelog]
header = """# Changelog

All notable changes to this project will be documented in this file.
"""
body = """
{% set visible_commits = commits | filter(attribute="group", value="_ignored") | length %}\
{% set total_commits = commits | length %}\
{% set has_visible_commits = visible_commits != total_commits %}\
{% if version or has_visible_commits %}\
## {% if version %}[{{ version | trim_start_matches(pat="v") }}] - {{ timestamp | date(format="%Y-%m-%d") }}{% else %}[unreleased]{% endif %}
{% if has_visible_commits %}\
{% for group, group_commits in commits | group_by(attribute="group") %}\
{% if group != "_ignored" %}

### {{ group | upper_first }}
{% for commit in group_commits %}
- {{ commit.message | upper_first }}
{% endfor %}
{% endif %}\
{% endfor %}\
{% else %}
No notable changes in this release.
{% endif %}\
{% endif %}\
"""
trim = false

[git]
conventional_commits = true
filter_unconventional = true
tag_pattern = "v[0-9].*"
commit_parsers = [
`

func freshCliff(cfg *config.Config, parsers []commitParser) ([]byte, error) {
	var sb strings.Builder

	sb.WriteString(cliffHeader)

	for _, p := range parsers {
		if p.Skip {
			sb.WriteString(fmt.Sprintf(`    { message = '%s', skip = true },`, p.Message))
		} else {
			sb.WriteString(fmt.Sprintf(`    { message = '%s', group = '%s' },`, p.Message, p.Group))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("]\n")

	return []byte(sb.String()), nil
}
