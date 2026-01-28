.PHONY: all build run dev clean test lint fmt generate deps help server cli

# Variables
BINARY_SERVER=server
BINARY_CLI=cli
GO=go
GOFLAGS=-ldflags="-s -w"

# Default target
all: generate build

## Build commands
build: build-server build-cli ## Build all binaries

build-server: ## Build server binary
	$(GO) build $(GOFLAGS) -o $(BINARY_SERVER) ./cmd/server

build-cli: ## Build CLI binary
	$(GO) build $(GOFLAGS) -o $(BINARY_CLI) ./cmd/cli

## Run commands
run: generate ## Run the server
	$(GO) run ./cmd/server

dev: ## Run with live reload (requires air: go install github.com/air-verse/air@latest)
	air

## Code generation
generate: ## Generate templ files
	templ generate

generate-watch: ## Watch and generate templ files
	templ generate --watch

## Dependencies
deps: ## Download dependencies
	$(GO) mod download

deps-update: ## Update dependencies
	$(GO) get -u ./...
	$(GO) mod tidy

deps-tidy: ## Tidy go.mod
	$(GO) mod tidy

## Testing
test: ## Run tests
	$(GO) test -v ./...

test-coverage: ## Run tests with coverage
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

## Code quality
lint: ## Run linter (requires golangci-lint)
	golangci-lint run

fmt: ## Format code
	$(GO) fmt ./...
	templ fmt .

vet: ## Run go vet
	$(GO) vet ./...

## Database
db-reset: ## Reset database (delete users.db)
	rm -f users.db

## Cleanup
clean: ## Clean build artifacts
	rm -f $(BINARY_SERVER) $(BINARY_CLI)
	rm -f coverage.out coverage.html
	rm -f users.db

clean-all: clean ## Clean all generated files
	find . -name "*_templ.go" -delete

## Docker (if needed later)
docker-build: ## Build Docker image
	docker build -t claude-test .

docker-run: ## Run Docker container
	docker run -p 8080:8080 claude-test

## Installation
install-tools: ## Install development tools
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/air-verse/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## Help
help: ## Show this help
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
