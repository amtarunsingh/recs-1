APP := votes_storage
BIN := bin/$(APP)
PKG := ./...

COVERAGE_MIN := 95.0
GOLANGCI_VERSION := v2.5.0
DOCKER_DEV_FILE := docker/docker-compose.dev.yml
DOCKER_PROMETHEUS_FILE := docker/docker-compose.prometheus.yml

.PHONY: help fmt fmt-check lint test test-coverage generate-mocks wire build dev-up dev-down prometheus-up prometheus-down prometheus-reload check-commit-msg-has-jira check-branch-name check-todo verify-common verify-pre-commit verify-pre-push

help: ## Show this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*?##/ {printf "  \033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

fmt: ## Format Go files
	@echo ">> Running fmt"
	go fmt $(PKG)

fmt-check: ## Check if any files need to be formatted
	@echo ">> Running fmt-check"
	@out=$$(gofmt -s -l .); \
	if [ -n "$$out" ]; then \
	  echo "gofmt needed in:"; echo "$$out"; exit 1; \
	fi

lint: ## Lint files
	@echo ">> Running linter"
	docker run --rm -v "$(PWD)":/app -w /app golangci/golangci-lint:$(GOLANGCI_VERSION) golangci-lint run ./...

test: ## Run tests
	@echo ">> Running tests"
	go test -race -count=1 $(PKG)

test-coverage: ## Run test coverage
	@echo ">> Running tests with coverage"
	mkdir -p tmp
	@COVERPKG=$$(go list ./... \
    		| grep -v '/internal/testlib' \
    		| grep -v '/internal/app/di' \
    		| grep -v '/internal/test/integration' \
    		| tr '\n' ',' | sed 's/,$$//'); \
	go test -race -count=1 -covermode=atomic -coverpkg="$$COVERPKG" -coverprofile=tmp/coverage.out ./...; \
	COV_LINE=$$(go tool cover -func=tmp/coverage.out | awk '/^total:/{print $$0}'); \
	COV_NUM=$$(echo $$COV_LINE | awk '{gsub(/%/,"",$$NF); print $$NF}'); \
	echo "\033[0m\033[1;34m>> To open HTML report: go tool cover -html=./tmp/coverage.out\033[0m"; \
	awk -v cov=$$COV_NUM -v min=$(COVERAGE_MIN) 'BEGIN { \
	  if (cov+0 < min) { printf "\033[31m>> Coverage too low: %.2f%% (required >= %.1f%%)\n\033[0m", cov, min; exit 1 } \
	  else { printf "\033[32m>> Coverage OK: %.2f%%\n\033[0m", cov } }'

generate-mocks: ## Generate mocks
	@echo ">> Generating mocks"
	go generate ./...

wire: ## Generate wire dependency inversion
	@echo ">> Preparing DI container"
	wire ./internal/app/di

build: ## Build the project
	@echo ">> Build project"
	go build -o $(BIN) ./cmd/app

dev-up: ## Spin up the docker containers for local development
	docker compose -f $(DOCKER_DEV_FILE) up --build

dev-down: ## Shut down the docker containers
	docker compose -f $(DOCKER_DEV_FILE) down

prometheus-up: ## Start Prometheus container (separate from dev stack)
	@echo ">> Starting Prometheus"
	docker compose -f $(DOCKER_PROMETHEUS_FILE) up -d

prometheus-down: ## Stop and remove Prometheus container
	@echo ">> Stopping Prometheus"
	docker compose -f $(DOCKER_PROMETHEUS_FILE) down

prometheus-reload: ## Reload Prometheus configuration
	@echo ">> Reloading Prometheus configuration"
	docker exec votes-prometheus kill -HUP 1

check-commit-msg-has-jira: ## Ensure the commit message has a Jira ticket reference
	./scripts/check-commit-msg-has-jira.sh $(COMMIT_MSG_FILE)

check-branch-name: ## Ensure the branch name follows ticket number convention
	./scripts/check-branch-name.sh

check-todo: ## Ensure there aren't any TODOs without Jira ticket references
	./scripts/check-todos.sh

verify-common: fmt lint wire check-todo ## Run common checks like formatting, linting, etc.

verify-pre-commit: verify-common ## Run pre-commit checks

verify-pre-push: verify-common test test-coverage ## Run pre-push checks