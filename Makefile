.PHONY: help dev env-check db-wait seed-if-empty run build test lint vet tidy docker-up docker-down health migrate-up migrate-down seed swagger

help:
	@echo Available targets:
	@echo   make help        List all targets
	@echo   make dev         ONE COMMAND: start postgres, migrate, seed if empty, run the API
	@echo   make run         Start the API server
	@echo   make build       Build binary to ./bin/rentzy
	@echo   make vet         Run go vet
	@echo   make tidy        Tidy go.mod
	@echo   make docker-up   Start postgres in background
	@echo   make docker-down Stop postgres
	@echo   make health      Curl /healthz
	@echo   make migrate-up  Apply all migrations, golang-migrate in Docker
	@echo   make migrate-down Roll back the last migration
	@echo   make seed        Load dev sample data
	@echo   make swagger     Regenerate Swagger docs from annotations

run:
	cd backend && go run ./cmd/rentzy

build:
	cd backend && go build -o ../bin/rentzy ./cmd/rentzy

vet:
	cd backend && go vet ./...

tidy:
	cd backend && go mod tidy

docker-up:
	docker compose up -d db

docker-down:
	docker compose down

health:
	curl -i http://localhost:8080/healthz

swagger:
	cd backend && go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/rentzy/main.go -o docs --overridesFile .swaggo

# Migrations run with the migrate/migrate Docker image on the compose network,
# so nothing has to be installed locally. Values come from .env.
-include .env
export

MIGRATE = docker run --rm -v "$(CURDIR)/backend/migrations:/migrations" --network rentzy_default \
	migrate/migrate:v4.18.3 -path=/migrations \
	-database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@db:5432/$(POSTGRES_DB)?sslmode=disable"

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down 1

seed:
	docker compose exec -T db psql -v ON_ERROR_STOP=1 -U $(POSTGRES_USER) -d $(POSTGRES_DB) < backend/migrations/seed/seed_sample_data.sql

# ---------- one-command dev start ----------

# make dev: Postgres up -> wait until ready -> migrations -> sample data (only
# into an empty database, so your own data is never overwritten) -> API.
dev: env-check docker-up db-wait migrate-up seed-if-empty run

env-check:
	@test -f .env || { echo ".env not found: run 'cp .env.example .env' and fill it in"; exit 1; }

db-wait:
	@echo "waiting for PostgreSQL..."
	@for i in $$(seq 1 30); do \
		docker compose exec -T db pg_isready -U $(POSTGRES_USER) -d $(POSTGRES_DB) >/dev/null 2>&1 && { echo "PostgreSQL is ready"; exit 0; }; \
		sleep 1; \
	done; echo "PostgreSQL did not become ready in 30s"; exit 1

seed-if-empty:
	@n=$$(docker compose exec -T db psql -tA -U $(POSTGRES_USER) -d $(POSTGRES_DB) -c "SELECT COUNT(*) FROM users") \
		|| { echo "could not count users (did migrate-up run?)"; exit 1; }; \
	if [ "$$n" = "0" ]; then echo "empty database: loading sample data"; $(MAKE) --no-print-directory seed; \
	else echo "database has $$n users: skipping sample data"; fi
