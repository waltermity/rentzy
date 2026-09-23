# Database Design

## Overview

The system uses **PostgreSQL** as its single data store.

Design goals:

- **Integrity** — foreign keys and `CHECK` constraints enforce relationships
  and valid states at the database level.
- **Correctness under concurrency** — reservation and rental overlap checks
  rely on transactions and row locks.
- **Monetary precision** — all money uses `NUMERIC(12,2)`, never `FLOAT`.


## Entity-Relationship Diagram

```mermaid
erDiagram
    CATEGORIES   ||--o{ TOOL_TYPES   : "classifies"
    TOOL_TYPES   ||--o{ TOOL_ITEMS   : "instantiates"
    CUSTOMERS    ||--o{ RESERVATIONS : "makes"
    CUSTOMERS    ||--o{ RENTALS      : "makes"
    TOOL_ITEMS   ||--o{ RESERVATIONS : "reserved as"
    TOOL_ITEMS   ||--o{ RENTALS      : "rented as"
    TOOL_ITEMS   ||--o{ MAINTENANCE  : "serviced by"
    RESERVATIONS ||--o| RENTALS      : "may become"

    CATEGORIES {
        bigserial id PK
        text      name UK
    }
    TOOL_TYPES {
        bigserial    id PK
        text         name
        bigint       category_id FK
        text         manufacturer
        text         model
        numeric      purchase_price
        numeric      daily_rate
        numeric      deposit
    }
    TOOL_ITEMS {
        bigserial    id PK
        bigint       tool_type_id FK
        text         status
        text         current_condition
        timestamptz  acquired_at
    }
    CUSTOMERS {
        bigserial    id PK
        text         name
        text         email UK
        text         phone
        text         address
        timestamptz  created_at
    }
    RESERVATIONS {
        bigserial    id PK
        bigint       customer_id FK
        bigint       tool_item_id FK
        timestamptz  start_date
        timestamptz  end_date
        text         status
        timestamptz  created_at
    }
    RENTALS {
        bigserial    id PK
        bigint       customer_id FK
        bigint       tool_item_id FK
        bigint       reservation_id FK
        timestamptz  start_date
        timestamptz  expected_end_date
        timestamptz  actual_end_date
        numeric      deposit_paid
        text         condition_at_handover
        text         condition_at_return
        numeric      base_fee
        numeric      late_fee
        numeric      total_fee
        text         status
    }
    MAINTENANCE {
        bigserial    id PK
        bigint       tool_item_id FK
        text         type
        text         description
        timestamptz  started_at
        timestamptz  completed_at
        numeric      cost
        text         performed_by
    }