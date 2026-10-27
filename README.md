# Rentzy

Tool rental administration system for a small rental business. Monorepo with a
Go REST API backend and a Vue frontend.

## Prerequisites

Install once per machine:

| Tool | Windows | macOS | Verify |
|------|---------|-------|--------|
| Go 1.22+ | https://go.dev/dl/ | `brew install go` | `go version` |
| Docker Desktop | https://docker.com | `brew install --cask docker` | `docker --version` |
| make | `choco install make` | `brew install make` | `make --version` |

## Getting Started

From a fresh clone, at the repo root:

```powershell
# 1. Create your local env file (values are git-ignored)
Copy-Item backend\.env.example .env
# edit .env and fill POSTGRES_PASSWORD + DATABASE_URL
# (DATABASE_URL must embed the same password)

# 2. Start Postgres
make docker-up

# 3. Wait until healthy
docker compose ps          # STATUS must be "Up (healthy)"

# 4. Start the API from the same root
make run

# 5. Verify — second terminal
make health
# HTTP/1.1 200 OK
# {"data":{"status":"ok"}}
```

# Project Layout

```
rentzy/
│
├── dev.sh
├── docker-compose.yml
├── .env                    # Local secrets (git-ignored)
├── .env.example           
├── Makefile                # Dev entry point
├── README.md
├── .gitignore
│
├── backend/
│   ├── Dockerfile
│   ├── README.md
│   ├── go.mod
│   ├── go.sum
│   │
│   ├── cmd/
│   │   └── rentzy/
│   │       └── main.go
│   │
│   ├── internal/
│   │   ├── db/
│   │   │   └── db.go
│   │   │
│   │   ├── domain/
│   │   │   ├── tool.go
│   │   │   ├── customer.go
│   │   │   ├── reservation.go
│   │   │   ├── rental.go
│   │   │   ├── maintenance.go
│   │   │   └── errors.go
│   │   │
│   │   ├── repository/
│   │   │   ├── repository.go          # interfaces
│   │   │   ├── tool.go                # ToolRepository interface
│   │   │   ├── customer.go            # CustomerRepository interface
│   │   │   ├── reservation.go         # ReservationRepository interface
│   │   │   ├── rental.go              # RentalRepository interface
│   │   │   ├── maintenance.go         # MaintenanceRepository interface
│   │   │   └── postgres/
│   │   │       ├── db.go              # shared DB struct / tx helpers
│   │   │       ├── tool.go            # ToolRepository impl
│   │   │       ├── customer.go        # CustomerRepository impl
│   │   │       ├── reservation.go     # ReservationRepository impl
│   │   │       ├── rental.go          # RentalRepository impl
│   │   │       └── maintenance.go     # MaintenanceRepository impl
│   │   │
│   │   ├── service/
│   │   │   ├── tool.go
│   │   │   ├── customer.go
│   │   │   ├── reservation.go
│   │   │   ├── rental.go
│   │   │   ├── maintenance.go
│   │   │   └── report.go
│   │   │
│   │   └── http/
│   │       ├── router.go
│   │       ├── respond/
│   │       │   └── respond.go
│   │       ├── middleware/
│   │       │   └── logging.go
│   │       └── handler/
│   │           ├── tool_handler.go
│   │           ├── customer_handler.go
│   │           ├── health_handler.go
│   │           ├── reservation_handler.go
│   │           ├── rental_handler.go
│   │           ├── maintenance_handler.go
│   │           └── report_handler.go
│   │
│   └── migrations/
│       ├── 000001_init.up.sql
│       └── 000001_init.down.sql
│
├── frontend/
│   └── ...
│
└── docs/
    ├── architecture.md
    ├── api.md
    ├── database.md
    └── diagrams/
```
## Common Commands

All run from the **repo root**:

| Command | What it does |
|---------|--------------|
| `make help` | List all targets |
| `make run` | Start the API (foreground) |
| `make build` | Build binary to `./bin/rentzy` |
| `make vet` | Run `go vet` |
| `make tidy` | Tidy `go.mod` / `go.sum` |
| `make docker-up` | Start Postgres in background |
| `make docker-down` | Stop Postgres (keeps data volume) |