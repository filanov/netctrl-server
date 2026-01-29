.PHONY: proto build run test clean lint fmt deps help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Binary output
BINARY_NAME=netctrl-server
BINARY_DIR=bin

# Buf parameters
BUF=buf

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

proto: ## Generate code from proto files
	@echo "Generating code from proto files..."
	@$(BUF) generate
	@echo "Proto generation complete"

build: proto ## Build the server binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	@$(GOBUILD) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/server
	@echo "Build complete: $(BINARY_DIR)/$(BINARY_NAME)"

run: build ## Build and run the server
	@echo "Starting server..."
	@./$(BINARY_DIR)/$(BINARY_NAME)

test: ## Run tests
	@echo "Running tests..."
	@$(GOTEST) -v ./...

test-race: ## Run tests with race detector
	@echo "Running tests with race detector..."
	@$(GOTEST) -race ./...

test-cover: ## Run tests with coverage
	@echo "Running tests with coverage..."
	@$(GOTEST) -cover ./...

coverage: ## Generate and view coverage report
	@echo "Generating coverage report..."
	@$(GOTEST) -coverprofile=coverage.out ./...
	@$(GOCMD) tool cover -html=coverage.out

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out
	@echo "Clean complete"

lint: ## Run linters
	@echo "Running linters..."
	@$(GOVET) ./...
	@echo "Lint complete"

fmt: ## Format code
	@echo "Formatting code..."
	@$(GOFMT) ./...
	@echo "Format complete"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo "Dependencies updated"

tidy: ## Tidy go.mod
	@echo "Tidying go.mod..."
	@$(GOMOD) tidy
	@echo "Tidy complete"

install-tools: ## Install development tools
	@echo "Installing tools..."
	@$(GOGET) github.com/bufbuild/buf/cmd/buf@latest
	@echo "Tools installed"

all: clean deps proto build ## Clean, download deps, generate proto, and build

.DEFAULT_GOAL := help
