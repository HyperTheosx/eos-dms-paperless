COMPOSE := docker compose --env-file .env -f deploy/docker-compose.yml
API_DIR := apps/api

.PHONY: help dev down down_v logs ps lint test run-api tidy

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

dev: ## Start all infrastructure containers
	$(COMPOSE) up -d

down: ## Stop and remove infrastructure containers
	$(COMPOSE) down

down_v: ## Stop and remove infrastructure containers with deleting volumes
	$(COMPOSE) down -v

logs: ## Follow infrastructure logs
	$(COMPOSE) logs -f --tail=100

ps: ## Show container status
	$(COMPOSE) ps

lint: ## Run golangci-lint for api
	cd $(API_DIR) && golangci-lint run ./...

test: ## Run unit tests for api
	cd $(API_DIR) && go test -race ./...

run-api: ## Run api locally against .env defaults
	cd $(API_DIR) && go run ./cmd/api

tidy: ## Sync go.mod/go.sum
	cd $(API_DIR) && go mod tidy