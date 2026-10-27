.PHONY: run build test lint vet tidy docker-up docker-down health

help:
	@echo Available targets:
	@echo   make help        List all targets
	@echo   make run         Start the API server
	@echo   make build       Build binary to ./bin/rentzy
	@echo   make vet         Run go vet
	@echo   make tidy        Tidy go.mod
	@echo   make docker-up   Start postgres in background
	@echo   make docker-down Stop postgres
	@echo   make health      Curl /healthz
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
	cd backend && swag init -g cmd/rentzy/main.go -o docs