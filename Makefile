# UI MCP Server Makefile
# Spec: mcp.md

# Build configuration
BINARY_NAME := frictionless
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Go configuration
GO := go
GOFLAGS := -trimpath

# Output directories
BUILD_DIR := build
RELEASE_DIR := release
CACHE_DIR := cache

# ui-engine project location (adjust if needed)
UI_ENGINE_DIR ?= ../ui-engine

.PHONY: all build clean test lint fmt vet deps release install check help cache cache-clean cache-refresh

# Default target
all: deps cache build

# Build binary for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/frictionless
	@$(BUILD_DIR)/$(BINARY_NAME) bundle -o $(BUILD_DIR)/$(BINARY_NAME).bundled install
	@mv $(BUILD_DIR)/$(BINARY_NAME).bundled $(BUILD_DIR)/$(BINARY_NAME)

# Cache ui-engine web assets (only if cache doesn't exist)
cache: $(CACHE_DIR)/.cached

$(UI_ENGINE_DIR)/build/ui-engine-bundled:
	@cd $(UI_ENGINE_DIR); $(MAKE) bundle

$(CACHE_DIR)/.cached: $(UI_ENGINE_DIR)/build/ui-engine-bundled
	@echo "Extracting web assets from ui-engine-bundled..."
	@if [ ! -f "$(UI_ENGINE_DIR)/build/ui-engine-bundled" ]; then \
		echo "Error: ui-engine-bundled not found at $(UI_ENGINE_DIR)/build/"; \
		echo "Run 'make bundle' in ui-engine first, or set UI_ENGINE_DIR"; \
		exit 1; \
	fi
	@mkdir -p $(CACHE_DIR)/html/themes $(CACHE_DIR)/viewdefs $(CACHE_DIR)/lua
	$(UI_ENGINE_DIR)/build/ui-engine-bundled cp 'html/*' $(CACHE_DIR)/html/
	$(UI_ENGINE_DIR)/build/ui-engine-bundled cp 'viewdefs/*' $(CACHE_DIR)/viewdefs/ 2>/dev/null || true
	$(UI_ENGINE_DIR)/build/ui-engine-bundled cp 'lua/*' $(CACHE_DIR)/lua/ 2>/dev/null || true
	$(UI_ENGINE_DIR)/build/ui-engine-bundled cp 'themes/*' $(CACHE_DIR)/html/themes/ 2>/dev/null || true
	@# Copy frictionless specific viewdefs (e.g., Prompt.DEFAULT.html)
	@if [ -d "web/viewdefs" ]; then \
		cp -r web/viewdefs/* $(CACHE_DIR)/viewdefs/ 2>/dev/null || true; \
		echo "Copied frictionless viewdefs"; \
	fi
	@# Copy frictionless specific lua files
	@if [ -d "web/lua" ]; then \
		cp -r web/lua/* $(CACHE_DIR)/lua/ 2>/dev/null || true; \
		echo "Copied frictionless lua files"; \
	fi
	@# Copy agents for bundling
	@if [ -d "agents" ]; then \
		mkdir -p $(CACHE_DIR)/agents; \
		cp -r agents/* $(CACHE_DIR)/agents/ 2>/dev/null || true; \
		echo "Copied agents"; \
	fi
	@# Copy html files to install/html for bundling
	@rm -f install/html/*.js
	@mkdir -p install/html
	@cp -r $(CACHE_DIR)/html/* install/html/
	@echo "Copied html files to install/html/"
	@touch $(CACHE_DIR)/.cached
	@echo "Cached web assets in $(CACHE_DIR)/"

# Force rebuild of cache
cache-refresh: cache-clean cache

# Remove cached assets
cache-clean:
	@echo "Cleaning cache..."
	@rm -rf $(CACHE_DIR)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR) $(RELEASE_DIR)
	@$(GO) clean -cache -testcache

# Run tests
test:
	@echo "Running tests..."
	CGO_ENABLED=0 $(GO) test -v ./...

# Run tests with race detector (requires CGO)
test-race:
	@echo "Running tests with race detector..."
	CGO_ENABLED=1 $(GO) test -v -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GO) test -v -coverprofile=$(BUILD_DIR)/coverage.out ./...
	$(GO) tool cover -html=$(BUILD_DIR)/coverage.out -o $(BUILD_DIR)/coverage.html
	@echo "Coverage report: $(BUILD_DIR)/coverage.html"

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./... || echo "Install golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

# Run go vet
vet:
	@echo "Running vet..."
	$(GO) vet ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GO) work sync

#	$(GO) mod download
#	$(GO) mod tidy

# Build release binaries for all platforms (with bundling)
release: build
	@echo "Building release binaries..."
	@mkdir -p $(RELEASE_DIR)
	@# Linux AMD64
	@echo "  Building linux/amd64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(RELEASE_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/frictionless
	@$(BUILD_DIR)/$(BINARY_NAME) bundle -src $(RELEASE_DIR)/$(BINARY_NAME)-linux-amd64 -o $(RELEASE_DIR)/$(BINARY_NAME)-linux-amd64.bundled install
	@mv $(RELEASE_DIR)/$(BINARY_NAME)-linux-amd64.bundled $(RELEASE_DIR)/$(BINARY_NAME)-linux-amd64
	@# Linux ARM64
	@echo "  Building linux/arm64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(RELEASE_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/frictionless
	@$(BUILD_DIR)/$(BINARY_NAME) bundle -src $(RELEASE_DIR)/$(BINARY_NAME)-linux-arm64 -o $(RELEASE_DIR)/$(BINARY_NAME)-linux-arm64.bundled install
	@mv $(RELEASE_DIR)/$(BINARY_NAME)-linux-arm64.bundled $(RELEASE_DIR)/$(BINARY_NAME)-linux-arm64
	@# macOS AMD64
	@echo "  Building darwin/amd64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(RELEASE_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/frictionless
	@$(BUILD_DIR)/$(BINARY_NAME) bundle -src $(RELEASE_DIR)/$(BINARY_NAME)-darwin-amd64 -o $(RELEASE_DIR)/$(BINARY_NAME)-darwin-amd64.bundled install
	@mv $(RELEASE_DIR)/$(BINARY_NAME)-darwin-amd64.bundled $(RELEASE_DIR)/$(BINARY_NAME)-darwin-amd64
	@# macOS ARM64 (Apple Silicon)
	@echo "  Building darwin/arm64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(RELEASE_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/frictionless
	@$(BUILD_DIR)/$(BINARY_NAME) bundle -src $(RELEASE_DIR)/$(BINARY_NAME)-darwin-arm64 -o $(RELEASE_DIR)/$(BINARY_NAME)-darwin-arm64.bundled install
	@mv $(RELEASE_DIR)/$(BINARY_NAME)-darwin-arm64.bundled $(RELEASE_DIR)/$(BINARY_NAME)-darwin-arm64
	@# Windows AMD64
	@echo "  Building windows/amd64..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) $(LDFLAGS) -o $(RELEASE_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/frictionless
	@$(BUILD_DIR)/$(BINARY_NAME) bundle -src $(RELEASE_DIR)/$(BINARY_NAME)-windows-amd64.exe -o $(RELEASE_DIR)/$(BINARY_NAME)-windows-amd64.bundled.exe install
	@mv $(RELEASE_DIR)/$(BINARY_NAME)-windows-amd64.bundled.exe $(RELEASE_DIR)/$(BINARY_NAME)-windows-amd64.exe
	@echo "Release binaries in $(RELEASE_DIR)/"
	@ls -la $(RELEASE_DIR)/

# Create release archives
release-archives: release
	@echo "Creating release archives..."
	@cd $(RELEASE_DIR) && \
		for f in $(BINARY_NAME)-*; do \
			if [ -f "$$f" ] && ! echo "$$f" | grep -q '\.\(tar\.gz\|zip\)$$'; then \
				if echo "$$f" | grep -q "windows"; then \
					zip -q "$${f%.exe}.zip" "$$f" && echo "  Created $${f%.exe}.zip"; \
				else \
					tar -czf "$$f.tar.gz" "$$f" && echo "  Created $$f.tar.gz"; \
				fi; \
			fi; \
		done
	@echo "Release archives created"

# Run the MCP server (stdio mode for AI assistants)
run: build
	@echo "Starting MCP server..."
	$(BUILD_DIR)/$(BINARY_NAME)

# Install to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	CGO_ENABLED=0 $(GO) install $(GOFLAGS) $(LDFLAGS) ./cmd/frictionless

# Check build requirements
check:
	@echo "Checking requirements..."
	@command -v go >/dev/null 2>&1 || { echo "Go is required but not installed."; exit 1; }
	@echo "Go version: $$(go version)"
	@echo "All requirements met"

# Show help
help:
	@echo "UI MCP Server Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build Targets:"
	@echo "  all             Build everything (deps, cache, binary)"
	@echo "  build           Build binary for current platform"
	@echo "  cache           Extract web assets from ui-engine-bundled"
	@echo "  cache-refresh   Force rebuild of cached assets"
	@echo ""
	@echo "Release Targets:"
	@echo "  release         Build binaries for all platforms"
	@echo "  release-archives Create release archives (tar.gz, zip)"
	@echo ""
	@echo "Development Targets:"
	@echo "  run             Run MCP server (stdio mode)"
	@echo "  test            Run tests"
	@echo "  test-race       Run tests with race detector (CGO)"
	@echo "  test-coverage   Run tests with coverage report"
	@echo ""
	@echo "Maintenance Targets:"
	@echo "  clean           Remove build artifacts (keeps cache)"
	@echo "  cache-clean     Remove cached ui-engine assets"
	@echo "  deps            Install Go dependencies"
	@echo "  lint            Run linter"
	@echo "  fmt             Format Go code"
	@echo "  vet             Run go vet"
	@echo "  install         Install to GOPATH/bin"
	@echo "  check           Check build requirements"
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION         Version string (default: git describe)"
	@echo "  UI_ENGINE_DIR   Path to ui-engine project (default: ../ui-engine)"
