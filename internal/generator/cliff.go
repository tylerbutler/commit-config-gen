package generator

import (
	"strings"

	"github.com/tylerbutler/commit-config-gen/internal/config"
)

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

// GenerateCliff generates cliff.toml content from config
func GenerateCliff(cfg *config.Config) (string, error) {
	var sb strings.Builder

	sb.WriteString(cliffHeader)

	// Add excluded scope rules
	for _, scope := range cfg.ExcludedScopes {
		sb.WriteString(`    { message = "^[a-z]+\\(`)
		sb.WriteString(scope)
		sb.WriteString(`\\)", skip = true },`)
		sb.WriteString("\n")
	}

	// Add visible type rules
	for _, name := range cfg.TypeNames() {
		t := cfg.Types[name]
		if t.ChangelogGroup != nil {
			sb.WriteString(`    { message = "^`)
			sb.WriteString(name)
			sb.WriteString(`", group = "`)
			sb.WriteString(*t.ChangelogGroup)
			sb.WriteString(`" },`)
			sb.WriteString("\n")
		}
	}

	// Add catch-all
	sb.WriteString(`    { message = ".*", group = "_ignored" },`)
	sb.WriteString("\n]\n")

	return sb.String(), nil
}
