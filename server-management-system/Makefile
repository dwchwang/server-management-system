# VCS-SMS Makefile
# =============================================

.PHONY: help dev-up dev-down infra-up infra-down build test coverage seed logs clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ── Infrastructure ──
infra-up: ## Start infrastructure containers (Postgres, Redis, ES, Kafka, TCP Sim)
	docker-compose -f docker-compose.dev.yml up -d

infra-down: ## Stop infrastructure containers
	docker-compose -f docker-compose.dev.yml down

dev-up: ## Start infrastructure only (for local development)
	@echo "Starting infrastructure..."
	docker-compose -f docker-compose.dev.yml up -d
	@echo "Waiting for services to be healthy..."
	@sleep 10
	@echo "Infrastructure is ready."

dev-down: ## Stop dev environment
	docker-compose -f docker-compose.dev.yml down

# ── Build ──
build: ## Build all service Docker images
	docker-compose build

# ── Test ──
test: ## Run all tests with coverage
	@echo "=== Testing shared ==="
	cd shared && go test ./... -coverprofile=coverage.out -covermode=atomic -v
	@echo "=== Testing auth-service ==="
	cd auth-service && go test ./... -coverprofile=coverage.out -covermode=atomic -v
	@echo "=== Testing server-service ==="
	cd server-service && go test ./... -coverprofile=coverage.out -covermode=atomic -v
	@echo "=== Testing api-gateway ==="
	cd api-gateway && go test ./... -coverprofile=coverage.out -covermode=atomic -v

coverage: ## Generate HTML coverage reports
	@for svc in shared auth-service server-service api-gateway; do \
		echo ">>> $$svc coverage:"; \
		cd $$svc && go tool cover -func=coverage.out | tail -1 && cd ..; \
		go tool cover -html=$$svc/coverage.out -o $$svc/coverage.html; \
	done
	@echo "Coverage reports generated."

# ── Mock Generation ──
mocks: ## Generate mocks with mockery
	@echo "Generating mocks..."
	@go install github.com/vektra/mockery/v2@latest 2>/dev/null || true
	mockery
	@echo "Mocks generated."

# ── Deps ──
deps: ## Install all dependencies
	cd shared && go mod tidy
	cd auth-service && go mod tidy
	cd server-service && go mod tidy
	cd api-gateway && go mod tidy

# ── Run ──
seed: ## Seed 10.000 servers into PostgreSQL
	docker exec -it vcs-sms-postgres psql -U vcs_admin -d vcs_sms -f /docker-entrypoint-initdb.d/seed_10k_servers.sql

logs: ## Tail all service logs
	docker-compose logs -f

logs-gateway: ## Tail gateway logs
	docker-compose logs -f api-gateway

logs-monitor: ## Tail monitor logs
	docker-compose logs -f monitor-service

# ── Clean ──
clean: ## Clean all containers, volumes, and build artifacts
	docker-compose down -v
	rm -rf logs/* uploads/*
	@echo "Cleaned."
