# Project Layout

```
rentzy/
├── backend/
│   ├── cmd/rentzy/main.go          # entry point
│   ├── internal/
│   │   ├── db/                     # DB connection (pgx)
│   │   ├── domain/                 # entities, enums, errors
│   │   ├── service/                # business logic + SQL
│   │   └── http/                   # router, handlers, respond helpers
│   ├── migrations/                 # SQL schema migrations
│   ├── go.mod
│   └── Dockerfile
├── frontend/
├── docs/
├── dev.sh                          # dev workflow script
└── docker-compose.yml              # deployment for local dev
```