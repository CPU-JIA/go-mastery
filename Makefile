# Go Mastery Project - Comprehensive Makefile
# Supports development workflow, CI/CD pipeline, and quality gates

# Configuration
BINARY_NAME=go-mastery
BINARY_UNIX=go-mastery_unix
BUILD_DIR=dist
COVERAGE_DIR=coverage
COVERAGE_FILE=$(COVERAGE_DIR)/coverage.out
COVERAGE_HTML=$(COVERAGE_DIR)/coverage.html
BENCHMARK_FILE=benchmarks.txt

# Go configuration
GO=go
GOFLAGS=-v
GOBUILD=$(GO) build $(GOFLAGS)
GOTEST=$(GO) test $(GOFLAGS)
GOCLEAN=$(GO) clean
GOGET=$(GO) get
GOMOD=$(GO) mod

# External tools
GOFMT=gofmt
GOIMPORTS=goimports
GOLINT=golint
GOVET=$(GO) vet
STATICCHECK=staticcheck
GOSEC=gosec
GOVULNCHECK=govulncheck

# Coverage threshold (percentage)
COVERAGE_THRESHOLD=75

# Build flags
BUILD_FLAGS=-ldflags="-s -w"
RACE_FLAG=-race

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[0;33m
BLUE=\033[0;34m
PURPLE=\033[0;35m
CYAN=\033[0;36m
NC=\033[0m # No Color

# Default target
.DEFAULT_GOAL := help

# Help target
.PHONY: help
help: ## Display this help message
	@echo "$(BLUE)Go Mastery Project - Make Commands$(NC)"
	@echo "=================================="
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "$(CYAN)%-20s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Setup and Installation
.PHONY: install-tools
install-tools: ## Install required development tools
	@echo "$(YELLOW)Installing development tools...$(NC)"
	$(GOGET) honnef.co/go/tools/cmd/staticcheck@latest
	$(GOGET) github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	$(GOGET) golang.org/x/vuln/cmd/govulncheck@latest
	$(GOGET) golang.org/x/tools/cmd/goimports@latest
	@echo "$(GREEN)Development tools installed successfully!$(NC)"

.PHONY: setup
setup: install-tools deps ## Setup development environment
	@echo "$(GREEN)Development environment setup complete!$(NC)"

# Dependencies
.PHONY: deps
deps: ## Download and verify dependencies
	@echo "$(YELLOW)Downloading dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) verify
	@echo "$(GREEN)Dependencies downloaded and verified!$(NC)"

.PHONY: deps-update
deps-update: ## Update dependencies to latest versions
	@echo "$(YELLOW)Updating dependencies...$(NC)"
	$(GOGET) -u all
	$(GOMOD) tidy
	@echo "$(GREEN)Dependencies updated!$(NC)"

.PHONY: deps-clean
deps-clean: ## Clean up dependency cache
	@echo "$(YELLOW)Cleaning dependency cache...$(NC)"
	$(GO) clean -modcache

# Build targets
.PHONY: build
build: ## Build the application
	@echo "$(YELLOW)Building application...$(NC)"
	$(GOBUILD) ./...
	@echo "$(GREEN)Build completed successfully!$(NC)"

.PHONY: build-all
build-all: clean fmt vet build ## Build with full quality checks
	@echo "$(GREEN)Full build completed successfully!$(NC)"

.PHONY: build-release
build-release: clean ## Build release binaries for multiple platforms
	@echo "$(YELLOW)Building release binaries...$(NC)"
	@mkdir -p $(BUILD_DIR)

	@echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./...

	@echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./...

	@echo "Building for macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./...

	@echo "Building for macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./...

	@echo "$(GREEN)Release binaries built successfully!$(NC)"
	@ls -la $(BUILD_DIR)/

# Testing
.PHONY: test
test: ## Run all tests
	@echo "$(YELLOW)Running tests...$(NC)"
	$(GOTEST) ./...
	@echo "$(GREEN)All tests passed!$(NC)"

.PHONY: test-race
test-race: ## Run tests with race detection
	@echo "$(YELLOW)Running tests with race detection...$(NC)"
	$(GOTEST) $(RACE_FLAG) ./...
	@echo "$(GREEN)Race tests passed!$(NC)"

.PHONY: test-short
test-short: ## Run short tests only
	@echo "$(YELLOW)Running short tests...$(NC)"
	$(GOTEST) -short ./...

.PHONY: test-verbose
test-verbose: ## Run tests with verbose output
	@echo "$(YELLOW)Running tests with verbose output...$(NC)"
	$(GOTEST) -v ./...

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(YELLOW)Running integration tests...$(NC)"
	$(GOTEST) -tags=integration ./...

# Coverage
.PHONY: coverage
coverage: ## Generate test coverage report
	@echo "$(YELLOW)Generating coverage report...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) $(RACE_FLAG) -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "$(GREEN)Coverage report generated: $(COVERAGE_HTML)$(NC)"

.PHONY: coverage-check
coverage-check: coverage ## Check if coverage meets threshold
	@echo "$(YELLOW)Checking coverage threshold...$(NC)"
	@COVERAGE=$$($(GO) tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo "Current coverage: $${COVERAGE}%"; \
	if [ $$(echo "$${COVERAGE} < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "$(RED)Coverage $${COVERAGE}% is below threshold $(COVERAGE_THRESHOLD)%$(NC)"; \
		exit 1; \
	else \
		echo "$(GREEN)Coverage $${COVERAGE}% meets threshold $(COVERAGE_THRESHOLD)%$(NC)"; \
	fi

.PHONY: coverage-open
coverage-open: coverage ## Open coverage report in browser
	@if command -v xdg-open > /dev/null; then \
		xdg-open $(COVERAGE_HTML); \
	elif command -v open > /dev/null; then \
		open $(COVERAGE_HTML); \
	else \
		echo "Please open $(COVERAGE_HTML) in your browser"; \
	fi

# Code quality
.PHONY: fmt
fmt: ## Format code
	@echo "$(YELLOW)Formatting code...$(NC)"
	$(GOFMT) -s -w .
	@if command -v $(GOIMPORTS) > /dev/null; then \
		$(GOIMPORTS) -w .; \
	fi
	@echo "$(GREEN)Code formatted successfully!$(NC)"

.PHONY: fmt-check
fmt-check: ## Check if code is properly formatted
	@echo "$(YELLOW)Checking code formatting...$(NC)"
	@UNFORMATTED=$$($(GOFMT) -s -l .); \
	if [ -n "$${UNFORMATTED}" ]; then \
		echo "$(RED)The following files are not properly formatted:$(NC)"; \
		echo "$${UNFORMATTED}"; \
		echo "$(YELLOW)Run 'make fmt' to fix formatting issues.$(NC)"; \
		exit 1; \
	else \
		echo "$(GREEN)All files are properly formatted!$(NC)"; \
	fi

.PHONY: lint
lint: ## Run various linters
	@echo "$(YELLOW)Running linters...$(NC)"
	@if command -v $(STATICCHECK) > /dev/null; then \
		$(STATICCHECK) ./...; \
	else \
		echo "$(RED)staticcheck not found. Install with: go install honnef.co/go/tools/cmd/staticcheck@latest$(NC)"; \
	fi
	@echo "$(GREEN)Linting completed!$(NC)"

.PHONY: vet
vet: ## Run go vet
	@echo "$(YELLOW)Running go vet...$(NC)"
	$(GOVET) ./...
	@echo "$(GREEN)Go vet completed successfully!$(NC)"

# Security
.PHONY: security
security: ## Run security analysis
	@echo "$(YELLOW)Running security analysis...$(NC)"
	@if command -v $(GOSEC) > /dev/null; then \
		$(GOSEC) ./...; \
	else \
		echo "$(RED)gosec not found. Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest$(NC)"; \
	fi
	@echo "$(GREEN)Security analysis completed!$(NC)"

.PHONY: vuln-check
vuln-check: ## Check for known vulnerabilities
	@echo "$(YELLOW)Checking for vulnerabilities...$(NC)"
	@if command -v $(GOVULNCHECK) > /dev/null; then \
		$(GOVULNCHECK) ./...; \
	else \
		echo "$(RED)govulncheck not found. Install with: go install golang.org/x/vuln/cmd/govulncheck@latest$(NC)"; \
	fi
	@echo "$(GREEN)Vulnerability check completed!$(NC)"

# Benchmarks
.PHONY: bench
bench: ## Run benchmarks
	@echo "$(YELLOW)Running benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem -count=3 ./... | tee $(BENCHMARK_FILE)
	@echo "$(GREEN)Benchmarks completed! Results saved to $(BENCHMARK_FILE)$(NC)"

.PHONY: bench-compare
bench-compare: ## Compare benchmark results
	@if [ ! -f $(BENCHMARK_FILE) ]; then \
		echo "$(RED)No benchmark file found. Run 'make bench' first.$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Benchmark Results:$(NC)"
	@cat $(BENCHMARK_FILE)

# Quality gates (comprehensive)
.PHONY: quality-check
quality-check: fmt-check vet lint security vuln-check test-race coverage-check ## Run all quality checks
	@echo "$(GREEN)✅ All quality checks passed!$(NC)"

.PHONY: ci
ci: quality-check bench ## Run CI pipeline locally
	@echo "$(GREEN)🎉 CI pipeline completed successfully!$(NC)"

# Cleanup
.PHONY: clean
clean: ## Clean build artifacts and cache
	@echo "$(YELLOW)Cleaning up...$(NC)"
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -f $(BENCHMARK_FILE)
	@rm -f *.out *.html *.json
	@echo "$(GREEN)Cleanup completed!$(NC)"

.PHONY: clean-all
clean-all: clean deps-clean ## Clean everything including dependency cache
	@echo "$(GREEN)Complete cleanup finished!$(NC)"

# Docker support
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(YELLOW)Building Docker image...$(NC)"
	@if [ -f Dockerfile ]; then \
		docker build -t $(BINARY_NAME):latest .; \
		echo "$(GREEN)Docker image built successfully!$(NC)"; \
	else \
		echo "$(RED)Dockerfile not found!$(NC)"; \
		exit 1; \
	fi

.PHONY: docker-run
docker-run: docker-build ## Run application in Docker container
	@echo "$(YELLOW)Running Docker container...$(NC)"
	docker run --rm -p 8080:8080 $(BINARY_NAME):latest

# Development helpers
.PHONY: dev-setup
dev-setup: setup ## Setup development environment with pre-commit hooks
	@echo "$(YELLOW)Setting up pre-commit hooks...$(NC)"
	@if command -v pre-commit > /dev/null; then \
		pre-commit install; \
		pre-commit install --hook-type commit-msg; \
		echo "$(GREEN)Pre-commit hooks installed!$(NC)"; \
	else \
		echo "$(YELLOW)pre-commit not found. Install with: pip install pre-commit$(NC)"; \
	fi

.PHONY: dev-check
dev-check: ## Quick development checks before commit
	@echo "$(YELLOW)Running quick development checks...$(NC)"
	@make fmt-check vet test-short
	@echo "$(GREEN)Development checks passed!$(NC)"

.PHONY: pre-commit
pre-commit: dev-check ## Run pre-commit checks manually
	@if command -v pre-commit > /dev/null; then \
		pre-commit run --all-files; \
	else \
		echo "$(YELLOW)pre-commit not installed, running make quality-check instead$(NC)"; \
		make quality-check; \
	fi

# Project info
.PHONY: info
info: ## Display project information
	@echo "$(BLUE)Go Mastery Project Information$(NC)"
	@echo "=============================="
	@echo "Go version: $$($(GO) version)"
	@echo "Project modules: $$(find . -name 'go.mod' | wc -l)"
	@echo "Go files: $$(find . -name '*.go' | grep -v vendor | wc -l)"
	@echo "Test files: $$(find . -name '*_test.go' | grep -v vendor | wc -l)"
	@echo "Build directory: $(BUILD_DIR)"
	@echo "Coverage directory: $(COVERAGE_DIR)"
	@echo "Coverage threshold: $(COVERAGE_THRESHOLD)%"

# Maintenance
.PHONY: mod-tidy
mod-tidy: ## Tidy go modules
	@echo "$(YELLOW)Tidying go modules...$(NC)"
	$(GOMOD) tidy
	@echo "$(GREEN)Go modules tidied!$(NC)"

.PHONY: upgrade
upgrade: deps-update mod-tidy ## Upgrade all dependencies
	@echo "$(GREEN)All dependencies upgraded!$(NC)"

# Quick commands for common workflows
.PHONY: quick-test
quick-test: fmt vet test-short ## Quick test for development
	@echo "$(GREEN)Quick tests passed!$(NC)"

.PHONY: full-check
full-check: ci ## Alias for ci - full comprehensive check
	@echo "$(GREEN)Full check completed!$(NC)"

# Logging migration helpers
.PHONY: logging-check
logging-check: ## Check logging migration status
	@echo "$(YELLOW)Checking logging migration status...$(NC)"
	@if [ -f migrate_logging.sh ]; then \
		echo "fmt.Print calls remaining: $$(grep -r "fmt\.Print" --include="*.go" . | wc -l)"; \
		echo "Structured logging calls: $$(grep -r "logger\." --include="*.go" . | wc -l)"; \
	fi

# Help for specific categories
.PHONY: help-build
help-build: ## Show build-related commands
	@echo "$(BLUE)Build Commands:$(NC)"
	@echo "  build         - Build the application"
	@echo "  build-all     - Build with quality checks"
	@echo "  build-release - Build release binaries"

.PHONY: help-test
help-test: ## Show test-related commands
	@echo "$(BLUE)Test Commands:$(NC)"
	@echo "  test          - Run all tests"
	@echo "  test-race     - Run tests with race detection"
	@echo "  test-short    - Run short tests only"
	@echo "  coverage      - Generate coverage report"

.PHONY: help-quality
help-quality: ## Show quality-related commands
	@echo "$(BLUE)Quality Commands:$(NC)"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linters"
	@echo "  vet           - Run go vet"
	@echo "  security      - Run security analysis"
	@echo "  quality-check - Run all quality checks"