# Makefile for algonius-supervisor

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOGET=$(GOCMD) get
GOTIDY=$(GOCMD) mod tidy
GOLIST=$(GOCMD) list

# Binary name and output directory
BINARY_NAME=algonius-supervisor
BINARY_UNIX=algonius-supervisor-unix
BUILD_DIR=build

# Directories
CMD_DIR=./cmd/supervisor
TEST_DIR=./tests

# Environment variables for testing
TEST_FLAGS?=-v
TEST_PATTERN?=.

# Version and build information
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_TIME?=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT?=$(shell git rev-parse HEAD)

# Build flags for version information
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(GIT_COMMIT) -X main.date=$(BUILD_TIME)"

.PHONY: all build clean test test-unit test-integration install deps tidy vet fmt fmt-check lint check help

# Default target - show help
all: help

# Build the project
build: $(BUILD_DIR)/$(BINARY_NAME)

# Create build directory
$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

# Build binary for current platform
$(BUILD_DIR)/$(BINARY_NAME): $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(LDFLAGS) $(CMD_DIR)
	@echo "Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

# Build with error handling (to work around current build issues)
build-safe:
	@echo "Building $(BINARY_NAME) version $(VERSION) (safe mode)..."
	@mkdir -p $(BUILD_DIR)
	@echo "Note: Building may fail due to A2A integration issues. For core functionality only:"
	@$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)_core -buildv=false ./cmd/supervisor 2>/dev/null || \
	 (echo "Build failed with A2A integration issues. This is known and will be fixed in next version."; \
	 echo "For now, please run 'go build ./cmd/supervisor' directly to see detailed errors.")

# Cross-compile for different platforms
build-linux: $(BUILD_DIR)
	@echo "Building for Linux..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_UNIX) $(LDFLAGS) $(CMD_DIR)

build-windows: $(BUILD_DIR)
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME).exe $(LDFLAGS) $(CMD_DIR)

build-darwin: $(BUILD_DIR)
	@echo "Building for macOS..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin $(LDFLAGS) $(CMD_DIR)

# Build for all platforms
build-all: build-linux build-windows build-darwin
	@echo "All builds completed in $(BUILD_DIR)/"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN) -cache -modcache -testcache
	rm -rf $(BUILD_DIR)
	@echo "Clean completed"

# Run all tests
test: test-unit test-integration

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	$(GOTEST) $(TEST_FLAGS) ./tests/unit/... -run "$(TEST_PATTERN)"

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) $(TEST_FLAGS) ./tests/integration/... -run "$(TEST_PATTERN)"

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out -coverpkg=./... $(TEST_FLAGS) ./tests/...
	$(GOTEST) -coverprofile=coverage.out -coverpkg=./... $(TEST_FLAGS) ./...
	@echo "Coverage report generated in coverage.out"
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "HTML coverage report generated in coverage.html"

# Run specific test package
test-package:
	@echo "Running tests for package: $(PKG)"
	$(GOTEST) $(TEST_FLAGS) ./$(PKG)/... -run "$(TEST_PATTERN)"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) verify
	@echo "Dependencies installed"

# Update and tidy dependencies
tidy: 
	@echo "Tidying dependencies..."
	$(GOTIDY)
	@echo "Dependencies tidied"

# Install the binary globally
install: build
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install $(CMD_DIR)
	@echo "Installation completed"

# Run go vet for static analysis
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...
	@echo "Vet completed"

# Run go fmt to format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@echo "Formatting completed"

# Check if code is formatted
fmt-check:
	@echo "Checking code formatting..."
	@gofmt -l . 2>&1 | grep -E '\.go$$' > /dev/null && echo "Code is not formatted. Run 'make fmt' or 'gofmt -s -w .'" && exit 1 || echo "Code is properly formatted"

# Check for outdated dependencies
deps-outdated:
	@echo "Checking for outdated dependencies..."
	$(GOCMD) list -u -m all

# Run all checks
check: vet fmt-check
	@echo "All checks passed"

# Help target to show available commands
help:
	@echo ''
	@echo 'Usage:'
	@echo '  make build           - Build the project'
	@echo '  make build-safe      - Build the project (with error handling)'
	@echo '  make build-all       - Build for all platforms'
	@echo '  make build-linux     - Build for Linux'
	@echo '  make build-windows   - Build for Windows'
	@echo '  make build-darwin    - Build for macOS'
	@echo '  make test            - Run all tests'
	@echo '  make test-unit       - Run unit tests only'
	@echo '  make test-integration - Run integration tests only'
	@echo '  make test-coverage   - Run tests with coverage report'
	@echo '  make test-package PKG=path/to/package - Run tests for a specific package'
	@echo '  make install         - Install the binary globally'
	@echo '  make clean           - Remove build artifacts'
	@echo '  make deps            - Install dependencies'
	@echo '  make tidy            - Tidy dependencies'
	@echo '  make vet             - Run go vet'
	@echo '  make fmt             - Format code'
	@echo '  make fmt-check       - Check if code is formatted'
	@echo '  make check           - Run all checks'
	@echo '  make deps-outdated   - Check for outdated dependencies'
	@echo '  make help            - Show this help'
	@echo ''

# Run the application in development mode
run-dev:
	@echo "Running development server..."
	$(GOCMD) run $(CMD_DIR)/main.go

# Build and run
build-run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Generate test coverage and open in browser
test-coverage-html: test-coverage
	@echo "Opening coverage report in browser..."
	$(GOCMD) tool cover -html=coverage.out