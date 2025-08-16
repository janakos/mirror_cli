.PHONY: build clean install test proto deps help

# Build variables
BINARY_NAME=mirror_cli
BUILD_DIR=build
PROTO_DIR=proto
PROTO_GEN_DIR=$(PROTO_DIR)/gen

# Go build flags
LDFLAGS=-ldflags="-s -w"
BUILD_FLAGS=-trimpath $(LDFLAGS)

# Default target
all: build

# Help target
help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Install dependencies
deps: ## Install Go dependencies
	go mod download
	go mod tidy

# Install buf for protobuf generation
install-buf: ## Install buf CLI tool
	@which buf > /dev/null || (echo "Installing buf..." && \
		curl -sSL "https://github.com/bufbuild/buf/releases/latest/download/buf-$$(uname -s)-$$(uname -m)" \
		-o "/usr/local/bin/buf" && chmod +x "/usr/local/bin/buf")

# Generate protobuf files
proto: ## Generate protobuf Go files
	@echo "Generating protobuf files..."
	@mkdir -p $(PROTO_GEN_DIR)
	protoc --go_out=$(PROTO_GEN_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(PROTO_GEN_DIR) --go-grpc_opt=paths=source_relative \
		-I $(PROTO_DIR) $(PROTO_DIR)/*.proto

# Build the binary
build: proto deps ## Build the CLI binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# Build for multiple platforms
build-all: proto deps ## Build for multiple platforms
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 .
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 .
	
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 .
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe .

# Install the binary to $GOPATH/bin
install: build ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	go install $(BUILD_FLAGS) .

# Run tests
test: ## Run tests
	go test -v ./...

# Run tests with coverage
test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Lint the code
lint: ## Lint the code
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2)
	golangci-lint run

# Format the code
fmt: ## Format the code
	go fmt ./...
	goimports -w .

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -rf $(PROTO_GEN_DIR)
	rm -f coverage.out coverage.html

# Run the CLI (for development)
run: build ## Run the CLI with example args
	./$(BUILD_DIR)/$(BINARY_NAME) --help

# Quick development cycle
dev: clean build run ## Clean, build and run for development

# Release preparation
release: clean test lint build-all ## Prepare release (clean, test, lint, build all platforms)
	@echo "Release artifacts created in $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

# Example usage targets
example-config: build ## Run example: initialize config
	./$(BUILD_DIR)/$(BINARY_NAME) config init

example-list-peers: build ## Run example: list peers
	./$(BUILD_DIR)/$(BINARY_NAME) peer list

example-list-mirrors: build ## Run example: list mirrors
	./$(BUILD_DIR)/$(BINARY_NAME) mirror list
