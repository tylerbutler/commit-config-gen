# justfile for commit-config-gen

# Default recipe: show available commands
default:
    @just --list

# Aliases
alias b := build
alias t := test

# Build the binary
build:
    go build -o commit-config-gen .

# Run tests
test:
    go test ./...

# Format code
format:
    go fmt ./...

# Lint code (requires golangci-lint)
lint:
    golangci-lint run

# Clean build artifacts
clean:
    rm -rf commit-config-gen commit-config-gen.exe dist/

# Install locally
install:
    go install .

# Download dependencies
deps:
    go mod download

# Update dependencies
deps-update:
    go get -u ./...
    go mod tidy

# --- Changie changelog management ---

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

# --- GoReleaser ---

# Check goreleaser config
release-check:
    goreleaser check

# Build a snapshot release (no publish)
release-snapshot:
    CHANGIE_CHANGELOG="$(changie latest)" goreleaser release --snapshot --clean

# Run a full release (requires GITHUB_TOKEN)
release:
    CHANGIE_CHANGELOG="$(changie latest)" goreleaser release --clean
