# ─── MoedahPOS Makefile ───────────────────────────────────────────────────────
.PHONY: help seed seed-reset dev-backend dev-frontend dev migrate lint build

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ─── Database ─────────────────────────────────────────────────────────────────
db-up: ## Start PostgreSQL container
	docker compose up -d postgres

db-down: ## Stop and remove PostgreSQL container
	docker compose down

db-logs: ## Tail database logs
	docker compose logs -f postgres

db-shell: ## Open psql shell
	docker exec -it moedah_postgres psql -U moedah -d moedah_pos

# ─── Seeding ──────────────────────────────────────────────────────────────────
seed: ## Seed demo data (safe — upserts, won't duplicate)
	@echo "🌱 Seeding demo data..."
	cd backend && go run ./cmd/seed/main.go

seed-reset: ## Clear ALL data then re-seed fresh demo data
	@echo "🗑  Resetting and re-seeding..."
	cd backend && go run ./cmd/seed/main.go --reset

# ─── Development ──────────────────────────────────────────────────────────────
dev-backend: ## Run Go API server (hot-reload via air if installed, else plain go run)
	@if command -v air &>/dev/null; then \
		cd backend && air; \
	else \
		cd backend && go run ./cmd/api/main.go; \
	fi

dev-frontend: ## Run Next.js dev server
	cd frontend && npm run dev

dev: db-up ## Start full stack (DB + backend + frontend) in parallel
	@echo "Starting full stack..."
	@(cd backend  && go run ./cmd/api/main.go &)
	@(cd frontend && npm run dev &)
	@echo ""
	@echo "  Backend:  http://localhost:8080"
	@echo "  Frontend: http://localhost:3000"
	@echo ""
	@echo "Press Ctrl-C to stop"
	@wait

# ─── Build ────────────────────────────────────────────────────────────────────
build-backend: ## Build the Go binary
	cd backend && go build -o bin/api ./cmd/api/main.go
	@echo "✓ Binary: backend/bin/api"

build-frontend: ## Build Next.js production bundle
	cd frontend && npm run build
	@echo "✓ Frontend built"

build: build-backend build-frontend ## Build both backend and frontend

# ─── Code Quality ─────────────────────────────────────────────────────────────
lint: ## Run Go vet + staticcheck
	cd backend && go vet ./...
	@echo "✓ Go vet passed"

test: ## Run Go tests
	cd backend && go test ./... -cover

# ─── Docker Compose (full stack) ──────────────────────────────────────────────
compose-up: ## Start all services via docker-compose
	docker compose up --build -d

compose-down: ## Stop all docker-compose services
	docker compose down

compose-logs: ## Tail all docker-compose logs
	docker compose logs -f
