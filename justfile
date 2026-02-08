# Commit config generator - Go project for generating commit configuration

# === ALIASES ===
alias b := build
alias t := test
alias f := format
alias l := lint
alias c := clean

# Default recipe
default:
    @just --list

# === STANDARD RECIPES ===

# Compile the project
build:
    go build -o commit-config-gen .

# Run tests
test:
    go test ./...

# Format code
format:
    gofumpt -w .

# Run linter
lint:
    golangci-lint run

# Remove build artifacts
clean:
    rm -rf commit-config-gen commit-config-gen.exe dist/

# Full validation workflow
ci: format lint test build

alias pr := ci

# === DEPENDENCIES ===

# Install locally
install:
    go install .

# Download dependencies
deps:
    go mod download

# === CONFIG GENERATION ===

# Generate commit configs (changie, commitlint) from commit-types.json
config-gen: build
    ./commit-config-gen generate -g changie -g commitlint

# Check commit configs are in sync with commit-types.json
config-check: build
    ./commit-config-gen check -g changie -g commitlint

# === CHANGIE ===

# Create a new changelog entry
change:
    changie new

# Show pending (unreleased) changelog entries
change-list:
    changie list

# Batch unreleased changes into a version
change-batch version:
    changie batch {{version}}

# Merge all versioned changelogs into CHANGELOG.md
change-merge:
    changie merge

# Show the next auto-determined version
change-next:
    changie next auto

# === GORELEASER ===

# Check goreleaser config
release-check:
    goreleaser check

# Build a snapshot release (no publish)
release-snapshot:
    CHANGIE_CHANGELOG="$(changie latest)" goreleaser release --snapshot --clean

# Run a full release (requires GITHUB_TOKEN)
release:
    CHANGIE_CHANGELOG="$(changie latest)" goreleaser release --clean
