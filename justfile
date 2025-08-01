# Build variables
VERSION := `git describe --tags --always --dirty`
COMMIT := `git rev-parse --short HEAD`
DATE := `date -u +"%Y-%m-%dT%H:%M:%SZ"`

# Binary name
BINARY_NAME := "gitclean"

# Build flags
LDFLAGS := "-s -w -X main.version=" + VERSION + " -X main.commit=" + COMMIT + " -X main.date=" + DATE

# Default recipe
default: test build

# Build the binary
build:
    go build -ldflags "{{LDFLAGS}}" -o {{BINARY_NAME}} -v

# Clean build artifacts
clean:
    go clean
    rm -f {{BINARY_NAME}}
    rm -rf dist/

# Run tests
test:
    go test -v ./...

# Test the release process locally (dry run)
release-test:
    goreleaser release --snapshot --clean

# Install dependencies
deps:
    go mod download
    go mod tidy

# Install goreleaser if not present
install-goreleaser:
    #!/usr/bin/env bash
    if ! which goreleaser > /dev/null; then
        echo "Installing goreleaser..."
        go install github.com/goreleaser/goreleaser@latest
    fi

# Run locally built binary
run: build
    ./{{BINARY_NAME}}

# Install binary to /usr/local/bin
install: build
    cp {{BINARY_NAME}} /usr/local/bin/

# Cross compile for different platforms
build-all:
    #!/usr/bin/env bash
    mkdir -p dist
    # macOS
    GOOS=darwin GOARCH=amd64 go build -ldflags "{{LDFLAGS}}" -o dist/{{BINARY_NAME}}-darwin-amd64
    GOOS=darwin GOARCH=arm64 go build -ldflags "{{LDFLAGS}}" -o dist/{{BINARY_NAME}}-darwin-arm64
    # Linux
    GOOS=linux GOARCH=amd64 go build -ldflags "{{LDFLAGS}}" -o dist/{{BINARY_NAME}}-linux-amd64
    GOOS=linux GOARCH=arm64 go build -ldflags "{{LDFLAGS}}" -o dist/{{BINARY_NAME}}-linux-arm64
    # Windows
    GOOS=windows GOARCH=amd64 go build -ldflags "{{LDFLAGS}}" -o dist/{{BINARY_NAME}}-windows-amd64.exe

# Check GoReleaser configuration
check:
    goreleaser check

# Build a single target for testing
build-snapshot:
    goreleaser build --snapshot --clean --single-target

# Show available commands
help:
    @just --list

# Run with arguments (e.g., just run-with --help)
run-with *args: build
    ./{{BINARY_NAME}} {{args}}

# Format Go code
fmt:
    go fmt ./...

# Run linting
lint:
    golangci-lint run

# Run security scan
security:
    gosec ./...

# Generate release notes
changelog:
    git log --oneline --decorate --graph

# Tag a new version (e.g., just tag v1.0.0)
tag version:
    git tag {{version}}
    git push origin {{version}}
