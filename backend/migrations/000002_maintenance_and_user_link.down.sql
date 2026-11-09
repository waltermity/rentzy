ALTER TABLE users
    DROP CONSTRAINT IF EXISTS users_customer_id_role_check,
    DROP CONSTRAINT IF EXISTS users_customer_id_key,
    DROP COLUMN IF EXISTS customer_id;

ALTER TABLE rentals
    DROP CONSTRAINT IF EXISTS rentals_operating_hours_check,
    DROP COLUMN IF EXISTS operating_hours;

ALTER TABLE tool_types
    DROP CONSTRAINT IF EXISTS tool_types_maintenance_every_hours_check,
    DROP CONSTRAINT IF EXISTS tool_types_maintenance_every_days_check,
    DROP CONSTRAINT IF EXISTS tool_types_maintenance_every_rentals_check,
    DROP COLUMN IF EXISTS maintenance_every_hours,
    DROP COLUMN IF EXISTS maintenance_every_days,
    DROP COLUMN IF EXISTS maintenance_every_rentals;
