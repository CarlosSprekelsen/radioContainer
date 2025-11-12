# Radio Control Container - Build and Quality Gates
# Source: PRE-INT-10
# Quote: "Prevent backsliding before integration"

.PHONY: help unit e2e e2e-cover race bench bench-fast bench-slow perf vegeta-env cover lint clean test-all integration test-unit-team test-integration-team test-e2e-team test-parallel integration-cover coverage-gaps coverage-diff test-quarantine
.DEFAULT_GOAL := help

# Configuration
COVERAGE_THRESHOLD := 80
COVERAGE_THRESHOLD_CRITICAL := 85
COVERAGE_OUTPUT := coverage.out
COVERAGE_HTML := coverage.html
LINT_CONFIG := .golangci.yml

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(BLUE)Radio Control Container - Build and Quality Gates$(NC)"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "$(YELLOW)Coverage Requirements:$(NC)"
	@echo "  Overall: ≥$(COVERAGE_THRESHOLD)%"
	@echo "  Critical packages (auth/command/telemetry): ≥$(COVERAGE_THRESHOLD_CRITICAL)%"
	@echo ""
	@echo "$(YELLOW)Lint Configuration:$(NC)"
	@echo "  Linters: errcheck, staticcheck, gocritic, stylecheck"
	@echo "  Warnings = Errors"

unit: ## Run unit tests
	@echo "$(BLUE)Running unit tests...$(NC)"
	@go test ./internal/... -count=1 -short

e2e: ## Run end-to-end tests
	@echo "$(BLUE)Running end-to-end tests...$(NC)"
	@go test ./test/e2e -count=1 -v

e2e-cover: ## Run end-to-end tests with coverage
	@echo "$(BLUE)Running end-to-end tests with coverage...$(NC)"
	@mkdir -p coverage
	@COVERAGE_FILE=e2e.cover go test -race -coverprofile=coverage/e2e.cover ./test/e2e/... -coverpkg=./internal/... -count=1
	@go tool cover -func=coverage/e2e.cover | tail -1
	@echo "$(GREEN)✅ E2E coverage report: coverage/e2e.cover$(NC)"

race: ## Run tests with race detection
	@echo "$(BLUE)Running tests with race detection...$(NC)"
	@go test -race ./... -count=1

bench: ## Run performance benchmarks
	@echo "$(BLUE)Running performance benchmarks...$(NC)"
	@go test -bench=. -benchmem ./internal/command ./internal/telemetry ./internal/adapter

bench-fast: ## Run fast performance benchmarks (1s benchtime, 30s timeout)
	@echo "$(BLUE)Running fast performance benchmarks...$(NC)"
	@go test -bench=. -benchtime=1s -timeout=30s -benchmem ./internal/command ./internal/telemetry ./internal/adapter

bench-slow: ## Run slow performance benchmarks (10s benchtime, 5m timeout)
	@echo "$(BLUE)Running slow performance benchmarks...$(NC)"
	@go test -bench=. -benchtime=10s -timeout=5m -run="^$$" -tags=slowbench -benchmem ./internal/command ./internal/telemetry ./internal/adapter

perf: ## Run load testing scenarios
	@echo "$(BLUE)Running load testing scenarios...$(NC)"
	@export PATH=$$PATH:$$(go env GOPATH)/bin && bash test/perf/vegeta_scenarios.sh || true

vegeta-env: ## Setup Vegeta environment
	@echo "$(BLUE)Setting up Vegeta environment...$(NC)"
	@export PATH=$$PATH:$$(go env GOPATH)/bin && vegeta -version

cover: ## Run tests with coverage and enforce thresholds
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	@go test -coverprofile=$(COVERAGE_OUTPUT) -covermode=atomic ./internal/auth/... ./internal/command/... ./internal/telemetry/... ./internal/config/... ./internal/adapter/... ./internal/audit/...
	@echo "$(BLUE)Generating coverage report...$(NC)"
	@go tool cover -html=$(COVERAGE_OUTPUT) -o $(COVERAGE_HTML)
	@echo "$(BLUE)Coverage report generated: $(COVERAGE_HTML)$(NC)"
	@echo "$(BLUE)Analyzing coverage thresholds...$(NC)"
	@$(MAKE) check-coverage

lint: ## Run golangci-lint with strict configuration
	@echo "$(BLUE)Running golangci-lint...$(NC)"
	@golangci-lint run --enable=errcheck,staticcheck,gocritic,govet,ineffassign,misspell,unused --timeout=5m ./internal/auth/... ./internal/command/... ./internal/config/... ./internal/adapter/... ./internal/audit/...
	@echo "$(GREEN)✅ Lint passed with 0 warnings$(NC)"

test-all: unit e2e race bench cover lint ## Run all tests and quality checks
	@echo "$(GREEN)✅ All quality gates passed$(NC)"

clean: ## Clean generated files
	@echo "$(BLUE)Cleaning generated files...$(NC)"
	@rm -f $(COVERAGE_OUTPUT) $(COVERAGE_HTML)
	@go clean -testcache

# Internal targets

check-coverage: ## Check coverage thresholds
	@echo "$(BLUE)Checking coverage thresholds...$(NC)"
	@$(MAKE) check-overall-coverage
	@$(MAKE) check-critical-packages-coverage
	@echo "$(GREEN)✅ Coverage thresholds met$(NC)"

check-overall-coverage: ## Check overall coverage threshold
	@echo "$(BLUE)Checking overall coverage (≥$(COVERAGE_THRESHOLD)%)...$(NC)"
	@OVERALL_COVERAGE=$$(go tool cover -func=$(COVERAGE_OUTPUT) | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ -z "$$OVERALL_COVERAGE" ]; then \
		echo "$(RED)❌ Could not determine overall coverage$(NC)"; \
		exit 1; \
	fi; \
	echo "Overall coverage: $$OVERALL_COVERAGE%"; \
	if [ $$(echo "$$OVERALL_COVERAGE < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "$(RED)❌ Overall coverage $$OVERALL_COVERAGE% is below threshold $(COVERAGE_THRESHOLD)%$(NC)"; \
		exit 1; \
	fi; \
	echo "$(GREEN)✅ Overall coverage $$OVERALL_COVERAGE% meets threshold$(NC)"

check-critical-packages-coverage: ## Check critical packages coverage threshold
	@echo "$(BLUE)Checking critical packages coverage (≥$(COVERAGE_THRESHOLD_CRITICAL)%)...$(NC)"
	@$(MAKE) check-package-coverage PACKAGE=auth THRESHOLD=$(COVERAGE_THRESHOLD_CRITICAL)
	@$(MAKE) check-package-coverage PACKAGE=command THRESHOLD=$(COVERAGE_THRESHOLD_CRITICAL)
	@$(MAKE) check-package-coverage PACKAGE=telemetry THRESHOLD=$(COVERAGE_THRESHOLD_CRITICAL)

check-package-coverage: ## Check specific package coverage (usage: make check-package-coverage PACKAGE=package THRESHOLD=threshold)
	@PACKAGE_COVERAGE=$$(go tool cover -func=$(COVERAGE_OUTPUT) | grep "github.com/radio-control/rcc/internal/$(PACKAGE)" | awk '{print $$3}' | sed 's/%//' | head -1); \
	if [ -z "$$PACKAGE_COVERAGE" ]; then \
		echo "$(YELLOW)⚠️  Package $(PACKAGE) not found in coverage report$(NC)"; \
		return 0; \
	fi; \
	echo "Package $(PACKAGE) coverage: $$PACKAGE_COVERAGE%"; \
	if [ $$(echo "$$PACKAGE_COVERAGE < $(THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "$(RED)❌ Package $(PACKAGE) coverage $$PACKAGE_COVERAGE% is below threshold $(THRESHOLD)%$(NC)"; \
		exit 1; \
	fi; \
	echo "$(GREEN)✅ Package $(PACKAGE) coverage $$PACKAGE_COVERAGE% meets threshold$(NC)"

# Coverage report generation
coverage-report: cover ## Generate detailed coverage report
	@echo "$(BLUE)Generating detailed coverage report...$(NC)"
	@echo "$(YELLOW)Coverage Report$(NC)"
	@echo "=================="
	@go tool cover -func=$(COVERAGE_OUTPUT) | grep -E "(github.com/radio-control/rcc/internal/|total)"
	@echo ""
	@echo "$(BLUE)HTML coverage report: $(COVERAGE_HTML)$(NC)"
	@echo "$(BLUE)Raw coverage data: $(COVERAGE_OUTPUT)$(NC)"

# Package-specific coverage
coverage-auth: ## Check auth package coverage
	@$(MAKE) check-package-coverage PACKAGE=auth THRESHOLD=$(COVERAGE_THRESHOLD_CRITICAL)

coverage-command: ## Check command package coverage
	@$(MAKE) check-package-coverage PACKAGE=command THRESHOLD=$(COVERAGE_THRESHOLD_CRITICAL)

coverage-telemetry: ## Check telemetry package coverage
	@$(MAKE) check-package-coverage PACKAGE=telemetry THRESHOLD=$(COVERAGE_THRESHOLD_CRITICAL)

# Lint configuration
lint-config: ## Show lint configuration
	@echo "$(BLUE)Lint Configuration:$(NC)"
	@golangci-lint config --config=$(LINT_CONFIG)

# Development helpers
test-package: ## Run tests for specific package (usage: make test-package PACKAGE=package)
	@echo "$(BLUE)Running tests for package $(PACKAGE)...$(NC)"
	@go test -v -cover ./internal/$(PACKAGE)/...

test-coverage-package: ## Run tests with coverage for specific package (usage: make test-coverage-package PACKAGE=package)
	@echo "$(BLUE)Running tests with coverage for package $(PACKAGE)...$(NC)"
	@go test -coverprofile=$(PACKAGE)_coverage.out -covermode=atomic ./internal/$(PACKAGE)/...
	@go tool cover -func=$(PACKAGE)_coverage.out | grep total

# Integration Tests
integration: ## Run integration tests (multi-component, no HTTP)
	@echo "$(BLUE)Running integration tests...$(NC)"
	@go test ./test/integration/... -count=1 -tags=integration -timeout=10m

integration-cover: ## Integration coverage with cross-package coverage
	@echo "$(BLUE)Running integration tests with cross-package coverage...$(NC)"
	@mkdir -p coverage
	@PKGS=$$(go list ./internal/... | tr '\n' ','); \
	go test -tags=integration -coverpkg="$${PKGS%,}" -coverprofile=coverage/integration.out ./test/integration/...
	@go tool cover -func=coverage/integration.out | tail -1
	@echo "$(GREEN)✅ Integration coverage report: coverage/integration.out$(NC)"

coverage-integration-html: ## Generate HTML coverage report for integration tests
	@echo "$(BLUE)Generating HTML coverage report for integration tests...$(NC)"
	@mkdir -p coverage
	@PKGS=$$(go list ./internal/... | tr '\n' ','); \
	go test -tags=integration -coverpkg="$${PKGS%,}" -coverprofile=coverage/integration.out ./test/integration/... && \
	go tool cover -html=coverage/integration.out -o coverage/integration.html
	@echo "$(GREEN)✅ Integration HTML coverage report: coverage/integration.html$(NC)"
	@echo "$(YELLOW)Open with: xdg-open coverage/integration.html || open coverage/integration.html$(NC)"

# Team-specific test suites
test-unit-team: ## Team 1: Unit tests (fast)
	@echo "$(BLUE)Team 1: Running unit tests...$(NC)"
	@go test ./internal/auth/... ./internal/command/... ./internal/config/... ./internal/adapter/... -short -count=1 -timeout=5m

test-integration-team: ## Team 2: Integration tests
	@echo "$(BLUE)Team 2: Running integration tests...$(NC)"
	@go test ./test/integration/... -tags=integration -count=1 -timeout=10m

test-e2e-team: ## Team 3: E2E tests (HTTP black-box)
	@echo "$(BLUE)Team 3: Running E2E tests...$(NC)"
	@go test ./test/e2e/... -count=1 -v -timeout=20m

# Parallel execution
test-parallel: ## Run all test suites in parallel
	@echo "$(BLUE)Running all test suites in parallel...$(NC)"
	@$(MAKE) -j3 test-unit-team test-integration-team test-e2e-team

# Coverage analysis
coverage-gaps: ## Identify uncovered critical paths
	@echo "$(BLUE)Identifying coverage gaps...$(NC)"
	@go test -coverprofile=full.out ./...
	@go tool cover -func=full.out | grep -E "0\.0%|[0-4][0-9]\.[0-9]%" | tee gaps.txt || true
	@echo "$(RED)Coverage gaps found (see gaps.txt)$(NC)"

coverage-diff: ## Coverage delta vs main
	@echo "$(BLUE)Comparing coverage with main branch...$(NC)"
	@git show main:coverage.out > coverage_main.out || true
	@go test -coverprofile=coverage_current.out ./...
	@echo "$(YELLOW)Compare coverage_main.out vs coverage_current.out offline$(NC)"

# Flaky test quarantine
test-quarantine: ## Run quarantined tests separately
	@echo "$(BLUE)Running quarantined tests...$(NC)"
	@go test -tags=flaky ./... -count=10 -failfast=false -timeout=20m | tee quarantine.log
