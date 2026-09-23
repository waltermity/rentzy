CREATE TABLE categories (
    id   BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE tool_types (
    id             BIGSERIAL PRIMARY KEY,
    name           TEXT NOT NULL,
    category_id    BIGINT NOT NULL REFERENCES categories(id),
    manufacturer   TEXT,
    model          TEXT,
    purchase_price NUMERIC(12,2) NOT NULL,
    daily_rate     NUMERIC(12,2) NOT NULL,
    deposit        NUMERIC(12,2) NOT NULL
);

CREATE TABLE tool_items (
    id                BIGSERIAL PRIMARY KEY,
    tool_type_id      BIGINT NOT NULL REFERENCES tool_types(id),
    status            TEXT NOT NULL DEFAULT 'AVAILABLE',
    current_condition TEXT NOT NULL DEFAULT 'GOOD',
    acquired_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT tool_items_status_check
        CHECK (status IN ('AVAILABLE','RENTED','MAINTENANCE','BROKEN','RETIRED'))
);

CREATE TABLE customers (
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT NOT NULL,
    email      TEXT NOT NULL UNIQUE,
    phone      TEXT,
    address    TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE reservations (
    id           BIGSERIAL PRIMARY KEY,
    customer_id  BIGINT NOT NULL REFERENCES customers(id),
    tool_item_id BIGINT NOT NULL REFERENCES tool_items(id),
    start_date   TIMESTAMPTZ NOT NULL,
    end_date     TIMESTAMPTZ NOT NULL,
    status       TEXT NOT NULL DEFAULT 'CONFIRMED',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT reservations_status_check
        CHECK (status IN ('PENDING','CONFIRMED','CANCELLED','CONVERTED','EXPIRED')),
    CONSTRAINT reservations_dates_check
        CHECK (end_date > start_date)
);

CREATE INDEX idx_reservations_tool_item_dates
    ON reservations (tool_item_id, start_date, end_date);

CREATE TABLE rentals (
    id                    BIGSERIAL PRIMARY KEY,
    customer_id           BIGINT NOT NULL REFERENCES customers(id),
    tool_item_id          BIGINT NOT NULL REFERENCES tool_items(id),
    reservation_id        BIGINT REFERENCES reservations(id),
    start_date            TIMESTAMPTZ NOT NULL,
    expected_end_date     TIMESTAMPTZ NOT NULL,
    actual_end_date       TIMESTAMPTZ,
    deposit_paid          NUMERIC(12,2) NOT NULL,
    condition_at_handover TEXT NOT NULL,
    condition_at_return   TEXT,
    base_fee              NUMERIC(12,2),
    late_fee              NUMERIC(12,2),
    total_fee             NUMERIC(12,2),
    status                TEXT NOT NULL DEFAULT 'ACTIVE',
    CONSTRAINT rentals_status_check
        CHECK (status IN ('ACTIVE','RETURNED','OVERDUE','CANCELLED'))
);

CREATE INDEX idx_rentals_tool_item_dates
    ON rentals (tool_item_id, start_date, expected_end_date);

CREATE TABLE maintenance (
    id           BIGSERIAL PRIMARY KEY,
    tool_item_id BIGINT NOT NULL REFERENCES tool_items(id),
    type         TEXT NOT NULL,
    description  TEXT,
    started_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    cost         NUMERIC(12,2) DEFAULT 0,
    performed_by TEXT,
    CONSTRAINT maintenance_type_check
        CHECK (type IN ('SCHEDULED','REPAIR','INSPECTION'))
);