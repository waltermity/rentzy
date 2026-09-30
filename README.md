# Project Layout

```
rentzy/
│
├── dev.sh
├── docker-compose.yml
├── README.md
├── .gitignore
│
├── backend/
│   ├── Dockerfile
│   ├── .env.example
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