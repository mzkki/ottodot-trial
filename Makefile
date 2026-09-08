.PHONY: build run wire migrate-up migrate-down migrate-create setup clean help

APP_NAME    := ottodot-trial
APP_PORT    ?= 8080
DB_PORT     ?= 5432
DB_USER     ?= ottodot
DB_PASS     ?= ottodot123
DB_NAME     ?= db_ottodot_trial
DB_TEST_NAME?= db_ottodot_trial_test
DB_HOST     ?= localhost

# ─── Help ──────────────────────────────────────────────────────────────────
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ─── Local Development ─────────────────────────────────────────────────────
build: ## Build the binary
	go build -o bin/$(APP_NAME) ./cmd/api/

run: ## Run the application
	go run ./cmd/api/

clean: ## Clean build artifacts
	rm -rf bin/ tmp/

# ─── Wire DI ───────────────────────────────────────────────────────────────
wire: ## Generate Wire DI code
	cd internal/bootstrap && wire

# ─── Swagger / OpenAPI ─────────────────────────────────────────────────────
swag: ## Generate Swagger / OpenAPI documentation
	swag init -g cmd/api/main.go -o ./docs --parseDependency --parseInternal

# ─── Database ──────────────────────────────────────────────────────────────
migrate-up: ## Run pending migrations on system database
	migrate -database "postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" \
		-path migrations up

migrate-down: ## Rollback last migration on system database
	migrate -database "postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" \
		-path migrations down 1

migrate-test-up: ## Run pending migrations on test database
	migrate -database "postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_TEST_NAME)?sslmode=disable" \
		-path migrations up

migrate-test-down: ## Rollback last migration on test database
	migrate -database "postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_TEST_NAME)?sslmode=disable" \
		-path migrations down 1

db-fresh: ## Drop all tables and re-run all migrations & seed on system DB
	@./scripts/db-fresh.sh system

db-fresh-test: ## Drop all tables and re-run all migrations & seed on test DB
	@./scripts/db-fresh.sh test

db-fresh-all: ## Drop all tables and re-run all migrations & seed on BOTH databases
	@./scripts/db-fresh.sh all

migrate-create: ## Create a new migration (NAME=xxx)
	@if [ -z "$(NAME)" ]; then \
		echo "Usage: make migrate-create NAME=add_some_table"; \
		exit 1; \
	fi
	migrate create -ext sql -dir migrations -seq $(NAME)

# Auto-detect Podman vs Docker Compose
DOCKER_BIN  := $(shell which podman 2>/dev/null || which docker 2>/dev/null)
COMPOSE_BIN := $(shell if which podman >/dev/null 2>&1; then echo "podman compose"; elif which docker-compose >/dev/null 2>&1; then echo "docker-compose"; else echo "docker compose"; fi)

# ─── Container Management ──────────────────────────────────────────────────
docker-up: ## Start PostgreSQL with Podman / Docker Compose
	$(COMPOSE_BIN) up -d

docker-down: ## Stop PostgreSQL
	$(COMPOSE_BIN) down

docker-logs: ## View PostgreSQL logs
	$(COMPOSE_BIN) logs -f

podman-up: docker-up ## Alias for docker-up using podman
podman-down: docker-down ## Alias for docker-down using podman

# ─── Tests ─────────────────────────────────────────────────────────────────
test: ## Run tests with coverage against dedicated test database
	@APP_ENV=test DB_NAME=$(DB_TEST_NAME) go test -v -coverprofile=coverage.out ./...
	@echo ""
	@echo "=== Core Business Logic Coverage (internal/usecase) ==="
	@go tool cover -func=coverage.out | grep 'internal/usecase'
	@echo "---------------------------------------------------------"
	@go tool cover -func=coverage.out | grep 'total:'
	@echo ""
	@echo "Tip: Run 'go tool cover -html=coverage.out' to view the visual HTML coverage in browser."

# ─── Convenience ───────────────────────────────────────────────────────────
setup: ## Full setup: docker-up + migrate + run
	$(MAKE) docker-up
	sleep 2
	$(MAKE) migrate-up
	$(MAKE) run
