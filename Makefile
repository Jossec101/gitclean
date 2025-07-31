.PHONY: build clean test release-test install

# Build variables
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse --short HEAD)
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Binary name
BINARY_NAME=gitclean

all: test build

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

test:
	$(GOTEST) -v ./...

# Test the release process locally (dry run)
release-test:
	goreleaser release --snapshot --clean

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Install goreleaser if not present
install-goreleaser:
	@which goreleaser > /dev/null || (echo "Installing goreleaser..." && go install github.com/goreleaser/goreleaser@latest)

# Run locally built binary
run: build
	./$(BINARY_NAME)

# Install locally
install: build
	cp $(BINARY_NAME) /usr/local/bin/

# Cross compile for different platforms
build-all:
	# macOS
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64
	# Linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64
	# Windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  clean          - Clean build artifacts"
	@echo "  test           - Run tests"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  release-test   - Test the release process locally"
	@echo "  install-goreleaser - Install goreleaser"
	@echo "  run            - Build and run the binary"
	@echo "  install        - Install the binary to /usr/local/bin"
	@echo "  build-all      - Cross compile for all platforms"
	@echo "  help           - Show this help message"
