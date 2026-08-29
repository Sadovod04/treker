COMPOSE := docker compose
DATABASE_URL ?= postgres://tracker_user:tracker_pass@localhost:5432/sports_tracker?sslmode=disable

.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

## --- docker ---
.PHONY: up
up: ## Build and start the whole stack
	$(COMPOSE) up --build

.PHONY: down
down: ## Stop the stack and drop volumes
	$(COMPOSE) down -v

.PHONY: logs
logs: ## Tail all container logs
	$(COMPOSE) logs -f

.PHONY: infra
infra: ## Start only Postgres + Redis (for local backend dev)
	$(COMPOSE) up -d postgres redis

## --- backend ---
.PHONY: backend-run
backend-run: ## Run the Go backend locally against DATABASE_URL
	cd backend && DB_HOST=localhost go run ./cmd/server

.PHONY: backend-test
backend-test: ## Run Go unit tests
	cd backend && go test ./...

.PHONY: backend-tidy
backend-tidy: ## go mod tidy + gofmt
	cd backend && go mod tidy && gofmt -w cmd internal migrations

## --- frontend ---
.PHONY: frontend-install
frontend-install: ## Install npm dependencies
	cd frontend && npm install

.PHONY: frontend-dev
frontend-dev: ## Start the Vite dev server (proxies /api and /ws to :8080)
	cd frontend && npm run dev

## --- demo ---
.PHONY: seed-samples
seed-samples: ## POST a synthetic tracker batch to a running backend
	./scripts/send_demo_batch.sh
