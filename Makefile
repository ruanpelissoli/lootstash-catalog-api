# LootStash Catalog API Makefile
# ===============================
# Manages catalog data imports for different games

.PHONY: help build clean deps migrate-d2 import-d2 verify-d2 run

# Default target
help:
	@echo "LootStash Catalog API"
	@echo "====================="
	@echo ""
	@echo "Available commands:"
	@echo ""
	@echo "  Build:"
	@echo "    make build          - Build the catalog CLI binary"
	@echo "    make deps           - Download Go dependencies"
	@echo "    make clean          - Remove build artifacts"
	@echo ""
	@echo "  Diablo II:"
	@echo "    make migrate-d2     - Run D2 database migrations"
	@echo "    make import-d2      - Import D2 catalog data"
	@echo "    make verify-d2      - Verify D2 data integrity (no duplicates)"
	@echo "    make setup-d2       - Run migrations and import (full setup)"
	@echo ""
	@echo "  Development:"
	@echo "    make run CMD=...    - Run arbitrary command"
	@echo ""
	@echo "Environment variables:"
	@echo "  DATABASE_URL          - PostgreSQL connection string"
	@echo "                          Default: postgres://postgres:postgres@localhost:54322/postgres"
	@echo "  REDIS_URL             - Redis connection string"
	@echo "                          Default: localhost:6379"

# Build configuration
BINARY_NAME=lootstash-catalog
BUILD_DIR=./bin

# Database defaults (can be overridden via environment variables)
DATABASE_URL ?= postgres://postgres:postgres@localhost:54322/postgres
REDIS_URL ?= localhost:6379

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies downloaded"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@echo "Clean complete"

# ==================
# Diablo II Commands
# ==================

# Run D2 migrations
migrate-d2: build
	@echo "Running Diablo II migrations..."
	@$(BUILD_DIR)/$(BINARY_NAME) migrate d2 --database-url="$(DATABASE_URL)"

# Import D2 catalog data
import-d2: build
	@echo "Importing Diablo II catalog data..."
	@$(BUILD_DIR)/$(BINARY_NAME) import d2 --database-url="$(DATABASE_URL)" --redis-url="$(REDIS_URL)"

# Verify D2 data integrity
verify-d2: build
	@echo "Verifying Diablo II catalog data..."
	@$(BUILD_DIR)/$(BINARY_NAME) verify d2 --database-url="$(DATABASE_URL)"

# Full D2 setup (migrations + import)
setup-d2: migrate-d2 import-d2
	@echo "Diablo II setup complete!"

# ==================
# Development
# ==================

# Run arbitrary command
run: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(CMD)

# Development run (without building)
dev:
	@go run . $(CMD)

# Run tests
test:
	@go test ./...

# Run tests with coverage
test-coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Lint code
lint:
	@golangci-lint run

# Format code
fmt:
	@go fmt ./...

# ==================
# Docker Support
# ==================

# Build Docker image
docker-build:
	@docker build -t lootstash-catalog-api .

# Run in Docker
docker-run:
	@docker run --rm \
		-e DATABASE_URL="$(DATABASE_URL)" \
		-e REDIS_URL="$(REDIS_URL)" \
		lootstash-catalog-api $(CMD)

# ==================
# Database Utilities
# ==================

# Reset D2 catalog (drop and recreate schema)
reset-d2:
	@echo "WARNING: This will drop the entire d2 schema!"
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ]
	@psql "$(DATABASE_URL)" -c "DROP SCHEMA IF EXISTS d2 CASCADE;"
	@$(MAKE) setup-d2
