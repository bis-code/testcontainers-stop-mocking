.PHONY: dev infra-up infra-up-detached infra-down test unit-test integration-test

## --- Infrastructure ---

infra-up: ## Start infrastructure (foreground)
	docker compose up

infra-up-detached: ## Start infrastructure (detached)
	docker compose up -d

infra-down: ## Stop infrastructure and remove volumes
	docker compose down -v

## --- Application ---

dev: infra-up-detached ## Run the app locally (starts infra if needed)
	go run ./cmd

## --- Tests ---

test: unit-test integration-test ## Run all tests

unit-test: ## Run unit tests only (no Docker required)
	go test -v ./...

integration-test: ## Run integration tests (requires Docker)
	go test -v -tags=integration ./...
