.PHONY: all build clean run install test deps

# Binary name
BINARY=ps-ide

# Build directory
BUILD_DIR=.

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)'"

all: deps build

build:
	@echo "Building $(BINARY)..."
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/ps-ide
	@echo "Build complete: $(BUILD_DIR)/$(BINARY)"

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -f $(BUILD_DIR)/$(BINARY)
	@echo "Clean complete"

run: build
	@echo "Running $(BINARY)..."
	./$(BUILD_DIR)/$(BINARY)

install:
	@echo "Installing system dependencies..."
	@bash install-deps.sh

deps:
	@echo "Downloading Go dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -cover ./...

fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

vet:
	@echo "Vetting code..."
	$(GOCMD) vet ./...

lint:
	@echo "Linting code..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run --timeout=5m

lint-fix:
	@echo "Linting code with auto-fix..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run --fix --timeout=5m

pre-publish:
	@echo "Running pre-publish checks..."
	@./lint.sh

ci: lint test build
	@echo "CI checks complete"

# Cross-compilation targets
build-linux:
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY)-linux ./cmd/ps-ide

build-windows:
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY).exe ./cmd/ps-ide

build-all: build-linux build-windows

help:
	@echo "Available targets:"
	@echo "  all           - Download deps and build"
	@echo "  build         - Build the binary"
	@echo "  clean         - Remove build artifacts"
	@echo "  run           - Build and run the application"
	@echo "  install       - Install system dependencies"
	@echo "  deps          - Download Go dependencies"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  fmt           - Format code"
	@echo "  vet           - Vet code"
	@echo "  lint          - Run golangci-lint"
	@echo "  lint-fix      - Run golangci-lint with auto-fix"
	@echo "  pre-publish   - Run all pre-publish checks (lint.sh)"
	@echo "  ci            - Run CI checks (lint + test + build)"
	@echo "  build-linux   - Cross-compile for Linux"
	@echo "  build-windows - Cross-compile for Windows"
	@echo "  build-all     - Build for all platforms"
	@echo "  help          - Show this help"
