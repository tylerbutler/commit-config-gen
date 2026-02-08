# commit-config-gen

Generate commit-related config files from a single source of truth.

## Overview

This tool generates config files for multiple commit/changelog/release tools from a single `commit-types.json` file, ensuring consistency across your toolchain.

### Supported Tools

| Generator | Output File | Description |
|-----------|-------------|-------------|
| `cliff` | `cliff.toml` | [git-cliff](https://git-cliff.org/) changelog generator |
| `commitlint` | `.commitlintrc.json` | [commitlint](https://commitlint.js.org/) commit linting |
| `conventional-changelog` | `.versionrc.json` | [conventional-changelog](https://github.com/conventional-changelog/conventional-changelog) |
| `release-please` | `release-please-config.json` | [Release Please](https://github.com/googleapis/release-please) |
| `changie` | `.changie.yaml` | [Changie](https://changie.dev/) changelog management |
| `semantic-release` | `.releaserc.json` | [semantic-release](https://semantic-release.gitbook.io/) |
| `release-plz` | `release-plz.toml` | [release-plz](https://release-plz.imo.dev/) for Rust projects |

## Installation

```bash
go install github.com/tylerbutler/commit-config-gen@latest
```

Or build from source:

```bash
just build
```

## Usage

```bash
# Generate all config files
commit-config-gen generate

# Generate specific tools only
commit-config-gen generate -g cliff -g commitlint

# Generate with custom paths
commit-config-gen -c path/to/commit-types.json generate -o output/dir

# Preview without writing files
commit-config-gen generate --dry-run

# Check if configs are in sync
commit-config-gen check

# Check specific generators
commit-config-gen check -g cliff -g commitlint

# Check configs in a specific directory
commit-config-gen -c path/to/commit-types.json check -d path/to/configs

# List available generators
commit-config-gen list
```

### Merge Behavior

When a config file already exists, generators **merge** changes into it — only updating commit-type-related fields while preserving all other configuration. This means you can customize other settings in your config files and they won't be overwritten.

## commit-types.json Format

```json
{
  "types": {
    "feat": {
      "description": "A new feature",
      "changelog_group": "Features",
      "bump": "minor"
    },
    "fix": {
      "description": "A bug fix",
      "changelog_group": "Bug Fixes",
      "bump": "patch"
    },
    "chore": {
      "description": "Other changes",
      "changelog_group": null
    }
  },
  "excluded_scopes": ["ci", "deps"],
  "commitlint_rules": {
    "header-max-length": [2, "always", 100]
  }
}
```

### Fields

- **types**: Map of commit type to configuration
  - `description`: Human-readable description (used by commitlint)
  - `changelog_group`: Section name in changelog, or `null` to exclude from changelog
  - `bump`: Version bump level — `"major"`, `"minor"`, `"patch"`, or `"none"` (optional, used by changie, semantic-release)
- **excluded_scopes**: Scopes to skip in changelog (e.g., `fix(ci)` won't appear)
- **commitlint_rules**: Additional commitlint rules to include

### Field Usage by Generator

| Field | cliff | commitlint | conventional-changelog | release-please | changie | semantic-release | release-plz |
|-------|-------|------------|----------------------|----------------|---------|-----------------|-------------|
| `description` | | | | | | | |
| `changelog_group` | group | | section | section | label | section | group |
| `bump` | | | | | auto | release | |
| `excluded_scopes` | skip | | | | | | skip |
| `commitlint_rules` | | rules | | | | | |

## Integration

### justfile

```just
# Generate configs
generate-configs:
    commit-config-gen generate

# Check configs are in sync (for CI)
check-configs-sync:
    commit-config-gen check
```

### CI

```yaml
- name: Check config sync
  run: commit-config-gen check
```
