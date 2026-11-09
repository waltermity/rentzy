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
| Node.js 20.19+ or 22.12+ (frontend) | https://nodejs.org | `brew install node` | `node -v` |

## Getting Started

### Quick start (one command)

From a fresh clone, at the repo root:

```bash
# 1. Once: create your local env file (git-ignored) and fill it in:
#    POSTGRES_PASSWORD, the same password inside DATABASE_URL,
#    and JWT_SECRET (at least 32 chars: openssl rand -hex 32)
cp .env.example .env          # Windows PowerShell: Copy-Item .env.example .env

# 2. Every time: start the database + API
make dev

# 3. In a second terminal: start the frontend
make web
```

`make dev` starts Postgres, waits until it is ready, applies the migrations, loads
the sample data **only if the database is empty** (your own data is never
overwritten), and runs the API in the foreground (Ctrl+C stops the API; Postgres
keeps running, `make docker-down` stops it).

`make web` installs the frontend dependencies on first run and starts Vite. It
forwards `/api` to the API on port 8080, so start `make dev` first.

Then open:

- App (Vue frontend): http://localhost:5173
- API docs (Swagger UI): http://localhost:8080/docs
- Health check: `make health` → `{"data":{"status":"ok"}}`

Sample users (password `Rentzy123!`): `nemeth.andras` (ADMIN), `toth.gabor` (STAFF),
`kovacs.anna` (CUSTOMER). Visitors can browse the catalog without logging in. In Swagger, log in with `POST /api/auth/login`, click
**Authorize** and paste `Bearer <token>`.

### Step by step

The same as `make dev`, one step at a time:

```bash
make docker-up       # start Postgres
docker compose ps    # wait until STATUS is "Up (healthy)"
make migrate-up      # apply migrations
make seed            # load sample data (dev only)
make run             # start the API
```

# Project Layout

```
rentzy/
│
├── dev.sh
├── docker-compose.yml
├── .env                    # Local secrets (git-ignored)
├── .env.example
├── Makefile                # Dev entry point (make dev)
├── README.md
├── .gitignore
│
├── backend/
│   ├── Dockerfile
│   ├── .swaggo                        # swag type overrides (decimal -> number)
│   ├── go.mod
│   ├── go.sum
│   │
│   ├── cmd/
│   │   └── rentzy/
│   │       └── main.go                # config, wiring, graceful shutdown
│   │
│   ├── docs/                          # generated Swagger (make swagger)
│   │
│   ├── internal/
│   │   ├── auth/
│   │   │   └── password.go            # bcrypt hashing
│   │   │
│   │   ├── db/
│   │   │   └── db.go                  # pgx connection pool
│   │   │
│   │   ├── domain/                    # entities, statuses, pure business rules
│   │   │   ├── tool.go
│   │   │   ├── customer.go
│   │   │   ├── user.go
│   │   │   ├── reservation.go
│   │   │   ├── rental.go              # fee calculation
│   │   │   ├── maintenance.go         # maintenance-due check
│   │   │   ├── report.go
│   │   │   └── errors.go
│   │   │
│   │   ├── repository/                # storage interfaces
│   │   │   ├── repository.go          # TxManager
│   │   │   ├── tool.go                # Category/ToolType/ToolItem repositories
│   │   │   ├── customer.go            # CustomerRepository
│   │   │   ├── user.go                # UserRepository
│   │   │   ├── reservation.go         # ReservationRepository
│   │   │   ├── rental.go              # RentalRepository
│   │   │   ├── maintenance.go         # MaintenanceRepository
│   │   │   ├── report.go              # ReportRepository
│   │   │   └── postgres/              # PostgreSQL implementations
│   │   │       ├── db.go              # shared DB struct / tx helpers
│   │   │       ├── tool.go
│   │   │       ├── customer.go
│   │   │       ├── user.go
│   │   │       ├── reservation.go
│   │   │       ├── rental.go
│   │   │       ├── maintenance.go
│   │   │       └── report.go
│   │   │
│   │   ├── service/                   # business logic (uses repository interfaces)
│   │   │   ├── service.go             # shared helpers
│   │   │   ├── status.go              # OVERDUE / EXPIRED background job
│   │   │   ├── auth.go                # login, register, JWT
│   │   │   ├── user.go
│   │   │   ├── customer.go
│   │   │   ├── tool.go
│   │   │   ├── reservation.go
│   │   │   ├── rental.go
│   │   │   ├── maintenance.go
│   │   │   └── report.go
│   │   │
│   │   └── http/
│   │       ├── router.go              # routes + RBAC groups
│   │       ├── respond/
│   │       │   └── respond.go         # JSON envelope, error mapping
│   │       ├── middleware/
│   │       │   ├── auth.go            # JWT check
│   │       │   ├── role.go            # role check
│   │       │   └── logging.go
│   │       └── handler/
│   │           ├── params.go          # request parsing helpers
│   │           ├── health_handler.go
│   │           ├── auth_handler.go
│   │           ├── user_handler.go
│   │           ├── customer_handler.go
│   │           ├── tool_handler.go
│   │           ├── reservation_handler.go
│   │           ├── rental_handler.go
│   │           ├── maintenance_handler.go
│   │           └── report_handler.go
│   │
│   └── migrations/
│       ├── 000001_init.up.sql
│       ├── 000001_init.down.sql
│       ├── 000002_maintenance_and_user_link.up.sql
│       ├── 000002_maintenance_and_user_link.down.sql
│       └── seed/
│           └── seed_sample_data.sql   # dev sample data (make seed)
│
├── frontend/                          # Vue 3 + TypeScript + Pinia + PrimeVue 4
│   ├── package.json
│   ├── vite.config.ts                 # dev server, /api proxy to :8080
│   ├── index.html
│   └── src/
│       ├── main.ts                    # app setup, PrimeVue theme, session restore
│       ├── App.vue
│       ├── style.css                  # global layout tokens (light + dark)
│       ├── api/
│       │   ├── client.ts              # axios instance, JWT header, error handling
│       │   ├── index.ts               # one function per API endpoint
│       │   └── types.ts               # mirrors the Go JSON types
│       ├── stores/                    # Pinia
│       │   ├── auth.ts                # JWT, user, role
│       │   ├── tools.ts               # catalog cache
│       │   ├── reservations.ts        # customer's reservations and rentals
│       │   └── rentals.ts             # staff handover / return
│       ├── router/index.ts            # routes + role guards
│       ├── components/                # header, status tag, maintenance dialog
│       ├── i18n/hu.ts                 # Hungarian enum names, PrimeVue locale, error texts
│       ├── utils/                     # formatting, toasts, status colours
│       └── views/
│           ├── CatalogView.vue        # public
│           ├── ToolDetailView.vue     # public, availability calendar + reserve
│           ├── LoginView.vue
│           ├── RegisterView.vue
│           ├── ProfileView.vue        # any logged-in user
│           ├── NotFoundView.vue
│           ├── staff/                 # Rentals, Customers, Maintenance
│           └── admin/                 # Tools, Users, Reports
│
└── docs/
    ├── architecture.md
    ├── database.md
    ├── frontend.md
    └── design/                        # Mermaid diagrams
```
## Common Commands

All run from the **repo root**:

| Command | What it does |
|---------|--------------|
| `make help` | List all targets |
| `make dev` | **One command**: Postgres up, migrate, seed if empty, run the API |
| `make run` | Start the API (foreground) |
| `make build` | Build binary to `./bin/rentzy` |
| `make vet` | Run `go vet` |
| `make tidy` | Tidy `go.mod` / `go.sum` |
| `make docker-up` | Start Postgres in background |
| `make docker-down` | Stop Postgres (keeps data volume) |
| `make migrate-up` | Apply migrations (golang-migrate in Docker, needs `make docker-up`) |
| `make migrate-down` | Roll back the last migration |
| `make seed` | Load dev sample data (after `migrate-up`) |
| `make web` | Start the frontend dev server on :5173 (needs `make dev` running) |
| `make web-build` | Typecheck and build the frontend to `frontend/dist` |
| `make swagger` | Regenerate the API docs served at `/docs` |

## Backend configuration (`.env`)

| Variable | Required | Default | Meaning |
|----------|----------|---------|---------|
| `DATABASE_URL` | yes | | PostgreSQL connection string |
| `JWT_SECRET` | yes | | HS256 signing key, at least 32 characters (`openssl rand -hex 32`) |
| `PORT` | no | `8080` | HTTP port |
| `JWT_TTL` | no | `24h` | Token lifetime |
| `RESERVATION_GRACE_HOURS` | no | `24` | An open reservation not handed over this long after its start becomes `EXPIRED` |
| `STATUS_REFRESH_INTERVAL` | no | `1m` | How often `ACTIVE`→`OVERDUE` and reservation expiry run |

## Backend business rules

- **Reservations**: customers reserve a *tool type*; the backend picks a free item
  (row locks prevent double booking). Staff/admin can name an exact item. New
  reservations are `CONFIRMED`. Items in `MAINTENANCE`, `BROKEN` or `RETIRED` are
  never booked. A late (overdue) rental blocks its item until it is returned.
- **Handover** needs an open reservation and an `AVAILABLE` item. Deposit = tool type
  deposit; expected end = reservation end.
- **Fees** on return: base = daily rate × started planned days (min. 1, charged in full
  on early return); late fee = daily rate × started days late. A `DAMAGED` return
  makes the item `BROKEN`, otherwise it becomes `AVAILABLE`.
- **Maintenance due**: any threshold on the tool type (`maintenance_every_rentals`,
  `_days`, `_hours`) reached since the last completed maintenance (or acquisition).
  Operating hours are recorded at return.
- **Procurement suggestion**: tool types whose non-retired items' utilization in the
  window (default 90 days) is ≥ threshold (default 0.7).