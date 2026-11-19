-- Maintenance thresholds per tool type. Any threshold that is set can make an
-- item due: N rentals, N days or N operating hours since its last completed
-- maintenance (or since acquisition, if it was never serviced). NULL = not used.
ALTER TABLE tool_types
    ADD COLUMN maintenance_every_rentals INTEGER,
    ADD COLUMN maintenance_every_days    INTEGER,
    ADD COLUMN maintenance_every_hours   NUMERIC(10,2),
    ADD CONSTRAINT tool_types_maintenance_every_rentals_check
        CHECK (maintenance_every_rentals IS NULL OR maintenance_every_rentals > 0),
    ADD CONSTRAINT tool_types_maintenance_every_days_check
        CHECK (maintenance_every_days IS NULL OR maintenance_every_days > 0),
    ADD CONSTRAINT tool_types_maintenance_every_hours_check
        CHECK (maintenance_every_hours IS NULL OR maintenance_every_hours > 0);

-- Operating hours the tool ran during a rental, recorded at return.
ALTER TABLE rentals
    ADD COLUMN operating_hours NUMERIC(10,2),
    ADD CONSTRAINT rentals_operating_hours_check
        CHECK (operating_hours IS NULL OR operating_hours >= 0);

-- A CUSTOMER login belongs to exactly one customer record.
ALTER TABLE users
    ADD COLUMN customer_id BIGINT REFERENCES customers(id),
    ADD CONSTRAINT users_customer_id_key UNIQUE (customer_id),
    ADD CONSTRAINT users_customer_id_role_check
        CHECK (customer_id IS NULL OR role = 'CUSTOMER');

-- Link existing CUSTOMER accounts by email.
UPDATE users u
SET customer_id = c.id
FROM customers c
WHERE u.role = 'CUSTOMER' AND u.email = c.email;
