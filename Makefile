.PHONY: build test run clean deps lint help

GEN_DB_URI="postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@postgres:5432/$(POSTGRES_DB)_gen?sslmode=disable"

# Default target
.DEFAULT_GOAL := help

# Build variables
BINARY_NAME=ohshift
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Go variables
GO=go
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)

help: ## Show this help message
	@echo "OhShift - Slack Incident Management Bot"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download and tidy dependencies
	$(GO) mod download
	$(GO) mod tidy

build: deps ## Build the application
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) main.go

build-linux: ## Build for Linux
	GOOS=linux GOARCH=amd64 $(MAKE) build

build-macos: ## Build for macOS
	GOOS=darwin GOARCH=amd64 $(MAKE) build

build-windows: ## Build for Windows
	GOOS=windows GOARCH=amd64 $(MAKE) build

build-all: build-linux build-macos build-windows ## Build for all platforms

run: build ## Build and run the application
	@echo "Running $(BINARY_NAME)..."
	@echo "Make sure to set required environment variables:"
	@echo "  SLACK_BOT_TOKEN=xoxb-your-token"
	@echo "  SLACK_SIGNING_SECRET=your-secret"
	@echo ""
	./$(BUILD_DIR)/$(BINARY_NAME)

dev: ## Run in development mode
	@echo "Running in development mode..."
	@echo "Make sure to set required environment variables:"
	@echo "  SLACK_BOT_TOKEN=xoxb-your-token"
	@echo "  SLACK_SIGNING_SECRET=your-secret"
	@echo ""
	$(GO) run main.go

test: ## Run tests
	$(GO) test -v ./...

test-coverage: ## Run tests with coverage
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "golangci-lint not found."; \
		echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		echo "Or visit: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

install: build ## Install the binary to /usr/local/bin
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "Installed $(BINARY_NAME) to /usr/local/bin/"

docker-build: ## Build Docker image
	docker build -t ohshift:$(VERSION) .
	docker tag ohshift:$(VERSION) ohshift:latest

docker-run: ## Run with Docker
	docker run --rm -p 8080:8080 \
		-e SLACK_BOT_TOKEN \
		-e SLACK_SIGNING_SECRET \
		-e NOTIFICATIONS_CHANNEL \
		ohshift:latest

fmt: ## Format code
	$(GO) fmt ./...

vet: ## Run go vet
	$(GO) vet ./...

check: fmt vet lint test ## Run all checks

release: clean build-all ## Create release builds
	@echo "Creating release builds..."
	@mkdir -p release
	cp $(BUILD_DIR)/ohshift-linux release/ohshift-linux-amd64
	cp $(BUILD_DIR)/ohshift-macos release/ohshift-darwin-amd64
	cp $(BUILD_DIR)/ohshift.exe release/ohshift-windows-amd64.exe
	@echo "Release builds created in release/ directory" 


gen-database: | deps
	@PGPASSWORD=${POSTGRES_PASSWORD} psql -h postgres -U ${POSTGRES_USER} -w -e -c "select version()" postgres
	@PGPASSWORD=${POSTGRES_PASSWORD} psql -h postgres -U ${POSTGRES_USER} -w -e -c "drop database if exists ${POSTGRES_DB}_gen" postgres
	@PGPASSWORD=${POSTGRES_PASSWORD} psql -h postgres -U ${POSTGRES_USER} -w -e -c "create database ${POSTGRES_DB}_gen" postgres
	@DB_URI=${GEN_DB_URI} go run main.go migrate up

gen-models:
	$(MAKE) gen-database
	bobgen-psql
	mermerd -c "postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@postgres:5432/${POSTGRES_DB}_gen?sslmode=disable" -e -s 'public' --ignoreTables 'goose_db_version' --useAllTables -o docs/gen_models_erd.md

local-db:
	@echo "Wiping and initializing local database..."
	@PGPASSWORD=${POSTGRES_PASSWORD} psql -U ${POSTGRES_USER} -h postgres ${POSTGRES_DB} -f .devcontainer/scripts/local_wipe.sql
	@PGPASSWORD=${POSTGRES_PASSWORD} psql -U ${POSTGRES_USER} -h postgres ${POSTGRES_DB} -f .devcontainer/scripts/local_init.sql