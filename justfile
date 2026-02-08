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

# Build for all platforms
build-all:
    GOOS=linux GOARCH=amd64 go build -o dist/commit-config-gen-linux-amd64 .
    GOOS=linux GOARCH=arm64 go build -o dist/commit-config-gen-linux-arm64 .
    GOOS=darwin GOARCH=amd64 go build -o dist/commit-config-gen-darwin-amd64 .
    GOOS=darwin GOARCH=arm64 go build -o dist/commit-config-gen-darwin-arm64 .
    GOOS=windows GOARCH=amd64 go build -o dist/commit-config-gen-windows-amd64.exe .

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
    rm -rf commit-config-gen dist/

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
