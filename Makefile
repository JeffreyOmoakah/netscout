.PHONY: build run clean test fmt vet install

# Binary name
BINARY_NAME=netscout
BUILD_DIR=bin

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/netscout
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application with example arguments
run: build
	@echo "Running $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME) -t 127.0.0.1 -p 80,443,8080 -v

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	@go vet ./...

# Install to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install ./cmd/netscout
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/netscout
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/netscout
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/netscout
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/netscout
	@echo "Build complete for all platforms"

# Development mode - build and run with verbose
dev: build
	@./$(BUILD_DIR)/$(BINARY_NAME) -t 127.0.0.1 -p 22,80,443,3389,8080 -v -w 50

# Help command
help:
	@echo "Available commands:"
	@echo "  make build      - Build the application"
	@echo "  make run        - Build and run with example args"
	@echo "  make dev        - Build and run in development mode"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make test       - Run tests"
	@echo "  make fmt        - Format code"
	@echo "  make vet        - Run go vet"
	@echo "  make install    - Install to GOPATH/bin"
	@echo "  make build-all  - Build for multiple platforms"


