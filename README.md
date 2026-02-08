# commit-config-gen

Generate commit-related config files from a single source of truth.

## Overview

This tool generates `cliff.toml` (for git-cliff) and `.commitlintrc.json` (for commitlint) from a single `commit-types.json` file, ensuring consistency between changelog generation and commit linting.

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
# Generate config files
commit-config-gen generate

# Generate with custom paths
commit-config-gen -c path/to/commit-types.json generate -o output/dir

# Preview without writing files
commit-config-gen generate --dry-run

# Check if configs are in sync
commit-config-gen check

# Check configs in a specific directory
commit-config-gen -c path/to/commit-types.json check -d path/to/configs
```

## commit-types.json Format

```json
{
  "types": {
    "feat": {
      "description": "A new feature",
      "changelog_group": "Features"
    },
    "fix": {
      "description": "A bug fix",
      "changelog_group": "Bug Fixes"
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
  - `changelog_group`: Section name in changelog, or `null` to exclude
- **excluded_scopes**: Scopes to skip in changelog (e.g., `fix(ci)` won't appear)
- **commitlint_rules**: Additional commitlint rules to include

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
