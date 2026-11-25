-- Sample data for local development and demos. DEV ONLY.
--
-- Load it with `make seed` after the migrations (000001 and 000002). It is not a migration:
-- golang-migrate ignores the seed/ folder, so `migrate up` never runs it.
-- Every INSERT is an upsert on the fixed id, so running it again on a database that already 
-- has an older version of this data updates those rows to the
-- current demo state (statuses, dates, amounts) and adds the new ones. Rows created later
-- through the app are left alone. For a completely clean demo, recreate the database volume
-- and run `make seed` again.
--
-- Ids 1-99 are the original demo set; ids from 101 are additional demo rows. The range in
-- between is left free so seeded ids never collide with rows created through the app.
--
-- All 15 users share the dev password Rentzy123!. pgcrypto stores it only as a
-- bcrypt hash.
--
-- The data is a small tool rental shop. It is written as a snapshot of Friday 2026-11-06 and
-- moved to "today" when it is loaded (see pg_temp.t below), so there are always rentals
-- out, overdue ones, reservations ahead, and history reaching back about five months.
-- Timestamps are Budapest local time with their UTC offset (+01 winter, +02 summer).

SET client_encoding = 'UTF8'; -- names and addresses contain Hungarian letters

BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto; -- crypt() and gen_salt() for the password hashes

-- The timestamps below are written as of the snapshot day 2026-11-06. pg_temp.t() moves each
-- one forward by the number of days since then (Budapest calendar days), so the data always
-- looks current on the day the seed is loaded. The local time of day is kept across DST changes.
CREATE FUNCTION pg_temp.t(ts text) RETURNS timestamptz
LANGUAGE sql STABLE AS $$
    SELECT ((ts::timestamptz AT TIME ZONE 'Europe/Budapest')
            + ((now() AT TIME ZONE 'Europe/Budapest')::date - DATE '2026-11-06') * INTERVAL '1 day')
           AT TIME ZONE 'Europe/Budapest'
$$;

-- users: 1-2 ADMIN, 3-5 STAFF, 6-15 CUSTOMER.
-- Each CUSTOMER account has the same email as one of the customers below.
INSERT INTO users (id, username, email, password_hash, role, created_at) VALUES
    (1,  'nemeth.andras',  'nemeth.andras@example.com',  crypt('Rentzy123!', gen_salt('bf', 10)), 'ADMIN',    pg_temp.t('2025-11-03 09:00:00+01')),
    (2,  'olah.krisztina', 'olah.krisztina@example.com', crypt('Rentzy123!', gen_salt('bf', 10)), 'ADMIN',    pg_temp.t('2025-11-03 09:10:00+01')),
    (3,  'toth.gabor',     'toth.gabor@example.com',     crypt('Rentzy123!', gen_salt('bf', 10)), 'STAFF',    pg_temp.t('2025-11-04 08:30:00+01')),
    (4,  'papp.judit',     'papp.judit@example.com',     crypt('Rentzy123!', gen_salt('bf', 10)), 'STAFF',    pg_temp.t('2025-11-04 08:45:00+01')),
    (5,  'fekete.laszlo',  'fekete.laszlo@example.com',  crypt('Rentzy123!', gen_salt('bf', 10)), 'STAFF',    pg_temp.t('2026-03-02 08:00:00+01')),
    (6,  'kovacs.anna',    'kovacs.anna@example.com',    crypt('Rentzy123!', gen_salt('bf', 10)), 'CUSTOMER', pg_temp.t('2025-11-18 10:15:00+01')),
    (7,  'nagy.peter',     'nagy.peter@example.com',     crypt('Rentzy123!', gen_salt('bf', 10)), 'CUSTOMER', pg_temp.t('2025-12-03 14:30:00+01')),
    (8,  'szabo.eszter',   'szabo.eszter@example.com',   crypt('Rentzy123!', gen_salt('bf', 10)), 'CUSTOMER', pg_temp.t('2026-01-12 09:40:00+01')),
    (9,  'horvath.bence',  'horvath.bence@example.com',  crypt('Rentzy123!', gen_salt('bf', 10)), 'CUSTOMER', pg_temp.t('2026-01-27 16:05:00+01')),
    (10, 'varga.zsofia',   'varga.zsofia@example.com',   crypt('Rentzy123!', gen_salt('bf', 10)), 'CUSTOMER', pg_temp.t('2026-02-09 11:20:00+01')),
    (11, 'kiss.daniel',    'kiss.daniel@example.com',    crypt('Rentzy123!', gen_salt('bf', 10)), 'CUSTOMER', pg_temp.t('2026-02-23 13:50:00+01')),
    (12, 'molnar.reka',    'molnar.reka@example.com',    crypt('Rentzy123!', gen_salt('bf', 10)), 'CUSTOMER', pg_temp.t('2026-03-10 10:00:00+01')),
    (13, 'farkas.tamas',   'farkas.tamas@example.com',   crypt('Rentzy123!', gen_salt('bf', 10)), 'CUSTOMER', pg_temp.t('2026-03-16 15:35:00+01')),
    (14, 'balogh.katalin', 'balogh.katalin@example.com', crypt('Rentzy123!', gen_salt('bf', 10)), 'CUSTOMER', pg_temp.t('2026-06-15 10:05:00+02')),
    (15, 'takacs.marton',  'takacs.marton@example.com',  crypt('Rentzy123!', gen_salt('bf', 10)), 'CUSTOMER', pg_temp.t('2026-09-02 18:30:00+02'))
ON CONFLICT (id) DO UPDATE SET
        username = EXCLUDED.username,
        email = EXCLUDED.email,
        role = EXCLUDED.role,
        created_at = EXCLUDED.created_at;

INSERT INTO categories (id, name) VALUES
    (1, 'Fúrás és bontás'),
    (2, 'Fűrészelés és vágás'),
    (3, 'Csiszolás'),
    (4, 'Kert és szabadtér'),
    (5, 'Takarítás')
ON CONFLICT (id) DO UPDATE SET
        name = EXCLUDED.name;

-- Prices in HUF.
INSERT INTO tool_types (id, name, category_id, manufacturer, model, purchase_price, daily_rate, deposit) VALUES
    (1,  'Fúrókalapács (SDS-plus)',       1, 'Bosch',   'GBH 2-28 F',        120000.00, 4500.00, 20000.00),
    (2,  'Bontókalapács (SDS-max)',       1, 'Makita',  'HM1317C',           395000.00, 9000.00, 60000.00),
    (3,  'Akkus ütvefúró-csavarozó 18 V', 1, 'Bosch',   'GSB 18V-55',         89000.00, 3000.00, 15000.00),
    (4,  'Kézi körfűrész 190 mm',         2, 'Makita',  'HS7601',             52000.00, 3500.00, 15000.00),
    (5,  'Szúrófűrész',                   2, 'Bosch',   'GST 90 BE',          54000.00, 2500.00, 10000.00),
    (6,  'Asztali körfűrész 210 mm',      2, 'DeWalt',  'DWE7485',           229000.00, 7500.00, 40000.00),
    (7,  'Sarokcsiszoló 230 mm',          3, 'Makita',  'GA9020',             46000.00, 3000.00, 10000.00),
    (8,  'Excentercsiszoló 150 mm',       3, 'Festool', 'ETS 150/3 EQ',      189000.00, 5000.00, 30000.00),
    (9,  'Láncfűrész 35 cm',              4, 'Stihl',   'MS 180',            119000.00, 6500.00, 30000.00),
    (10, 'Fűkasza',                       4, 'Stihl',   'FS 55',              99000.00, 5000.00, 25000.00),
    (11, 'Magasnyomású mosó',             5, 'Kärcher', 'K 5 Power Control', 124000.00, 5500.00, 25000.00),
    (12, 'Szőnyeg- és kárpittisztító',    5, 'Kärcher', 'Puzzi 10/1',        285000.00, 8000.00, 40000.00)
ON CONFLICT (id) DO UPDATE SET
        name = EXCLUDED.name,
        category_id = EXCLUDED.category_id,
        manufacturer = EXCLUDED.manufacturer,
        model = EXCLUDED.model,
        purchase_price = EXCLUDED.purchase_price,
        daily_rate = EXCLUDED.daily_rate,
        deposit = EXCLUDED.deposit;

-- Status on 2026-11-06:
--   RENTED       1, 4, 8, 10, 11, 16, 18, 20   each has an ACTIVE or OVERDUE rental
--   MAINTENANCE  5, 15, 19              each has an open maintenance record
--   BROKEN       9, 13                  returned damaged
--   RETIRED      3
--   AVAILABLE    all others
INSERT INTO tool_items (id, tool_type_id, status, current_condition, acquired_at) VALUES
    (1,  1,  'RENTED',      'GOOD',    pg_temp.t('2024-02-12 10:00:00+01')),
    (2,  1,  'AVAILABLE',   'GOOD',    pg_temp.t('2024-02-12 10:00:00+01')),
    (3,  1,  'RETIRED',     'POOR',    pg_temp.t('2022-01-17 10:00:00+01')),
    (4,  2,  'RENTED',      'GOOD',    pg_temp.t('2024-11-05 10:00:00+01')),
    (5,  2,  'MAINTENANCE', 'FAIR',    pg_temp.t('2023-03-06 10:00:00+01')),
    (6,  3,  'AVAILABLE',   'GOOD',    pg_temp.t('2025-02-11 10:00:00+01')),
    (7,  3,  'AVAILABLE',   'NEW',     pg_temp.t('2026-10-30 10:00:00+01')),
    (8,  4,  'RENTED',      'GOOD',    pg_temp.t('2024-12-02 10:00:00+01')),
    (9,  4,  'BROKEN',      'DAMAGED', pg_temp.t('2023-11-20 10:00:00+01')),
    (10, 5,  'RENTED',      'FAIR',    pg_temp.t('2023-12-04 10:00:00+01')),
    (11, 6,  'RENTED',      'GOOD',    pg_temp.t('2025-03-10 10:00:00+01')),
    (12, 7,  'AVAILABLE',   'GOOD',    pg_temp.t('2025-01-20 10:00:00+01')),
    (13, 7,  'BROKEN',      'DAMAGED', pg_temp.t('2023-02-27 10:00:00+01')),
    (14, 8,  'AVAILABLE',   'GOOD',    pg_temp.t('2025-03-18 10:00:00+01')),
    (15, 9,  'MAINTENANCE', 'FAIR',    pg_temp.t('2024-03-04 10:00:00+01')),
    (16, 9,  'RENTED',      'GOOD',    pg_temp.t('2025-12-01 10:00:00+01')),
    (17, 10, 'AVAILABLE',   'GOOD',    pg_temp.t('2026-02-16 10:00:00+01')),
    (18, 11, 'RENTED',      'GOOD',    pg_temp.t('2024-01-22 10:00:00+01')),
    (19, 11, 'MAINTENANCE', 'DAMAGED', pg_temp.t('2023-01-09 10:00:00+01')),
    (20, 12, 'RENTED',      'GOOD',    pg_temp.t('2026-03-02 10:00:00+01'))
ON CONFLICT (id) DO UPDATE SET
        tool_type_id = EXCLUDED.tool_type_id,
        status = EXCLUDED.status,
        current_condition = EXCLUDED.current_condition,
        acquired_at = EXCLUDED.acquired_at;

INSERT INTO customers (id, name, email, phone, address, created_at) VALUES
    (1,  'Kovács Anna',    'kovacs.anna@example.com',    '+45671', '1052 Budapest, Petőfi Sándor utca 8.', pg_temp.t('2025-11-18 10:15:00+01')),
    (2,  'Nagy Péter',     'nagy.peter@example.com',     '+45672', '2030 Érd, Fő utca 14.',                pg_temp.t('2025-12-03 14:30:00+01')),
    (3,  'Szabó Eszter',   'szabo.eszter@example.com',   '+45673', '1137 Budapest, Pozsonyi út 30.',       pg_temp.t('2026-01-12 09:40:00+01')),
    (4,  'Horváth Bence',  'horvath.bence@example.com',  '+45674', '2040 Budaörs, Kossuth Lajos utca 52.', pg_temp.t('2026-01-27 16:05:00+01')),
    (5,  'Varga Zsófia',   'varga.zsofia@example.com',   '+45675', '2000 Szentendre, Ady Endre út 3.',     pg_temp.t('2026-02-09 11:20:00+01')),
    (6,  'Kiss Dániel',    'kiss.daniel@example.com',    '+45676', '1095 Budapest, Mester utca 41.',       pg_temp.t('2026-02-23 13:50:00+01')),
    (7,  'Molnár Réka',    'molnar.reka@example.com',    '+45677', '2100 Gödöllő, Szabadság tér 6.',       pg_temp.t('2026-03-10 10:00:00+01')),
    (8,  'Farkas Tamás',   'farkas.tamas@example.com',   '+45678', '2600 Vác, Arany János utca 17.',       pg_temp.t('2026-03-16 15:35:00+01')),
    (9,  'Balogh Katalin', 'balogh.katalin@example.com', '+45679', '1118 Budapest, Rétköz utca 9.',        pg_temp.t('2026-06-15 10:05:00+02')),
    (10, 'Takács Márton',  'takacs.marton@example.com',  '+45680', '2120 Dunakeszi, Rákóczi út 25.',       pg_temp.t('2026-09-02 18:30:00+02'))
ON CONFLICT (id) DO UPDATE SET
        name = EXCLUDED.name,
        email = EXCLUDED.email,
        phone = EXCLUDED.phone,
        address = EXCLUDED.address,
        created_at = EXCLUDED.created_at;

-- Every rental was created from one of these reservations, which then became CONVERTED:
--   3 -> rental 13, 4 -> 15, 5 -> 21, 6 -> 16, 8 -> 23,
--   16-27 -> rentals 1-12, 28 -> 14, 29-32 -> 17-20, 33 -> 22, 34 -> 24, 35 -> 25.
--   101-108 -> rentals 101-108, 109 -> 109, 110 -> 110, 117-120 -> rentals 111-114.
-- CONFIRMED ones are still ahead. New reservations are created CONFIRMED, so PENDING is not used.
INSERT INTO reservations (id, customer_id, tool_item_id, start_date, end_date, status, created_at) VALUES
    (1,  6,  17, pg_temp.t('2026-10-17 08:00:00+02'), pg_temp.t('2026-10-19 08:00:00+02'), 'CANCELLED', pg_temp.t('2026-10-11 19:45:00+02')),
    (2,  9,  14, pg_temp.t('2026-10-17 09:00:00+02'), pg_temp.t('2026-10-19 09:00:00+02'), 'EXPIRED',   pg_temp.t('2026-10-12 11:20:00+02')),
    (3,  3,  18, pg_temp.t('2026-10-23 15:00:00+02'), pg_temp.t('2026-10-24 15:00:00+02'), 'CONVERTED', pg_temp.t('2026-10-19 10:05:00+02')),
    (4,  1,  10, pg_temp.t('2026-10-27 08:30:00+01'), pg_temp.t('2026-10-30 08:30:00+01'), 'CONVERTED', pg_temp.t('2026-10-20 16:40:00+02')),
    (5,  2,  11, pg_temp.t('2026-11-02 08:00:00+01'), pg_temp.t('2026-11-09 08:00:00+01'), 'CONVERTED', pg_temp.t('2026-10-21 09:15:00+02')),
    (6,  4,  8,  pg_temp.t('2026-10-29 13:00:00+01'), pg_temp.t('2026-10-31 13:00:00+01'), 'CONVERTED', pg_temp.t('2026-10-26 12:30:00+01')),
    (7,  10, 6,  pg_temp.t('2026-10-30 10:00:00+01'), pg_temp.t('2026-10-31 10:00:00+01'), 'EXPIRED',   pg_temp.t('2026-10-27 18:10:00+01')),
    (8,  3,  1,  pg_temp.t('2026-11-04 09:00:00+01'), pg_temp.t('2026-11-09 09:00:00+01'), 'CONVERTED', pg_temp.t('2026-10-28 14:00:00+01')),
    (9,  2,  7,  pg_temp.t('2026-11-13 08:00:00+01'), pg_temp.t('2026-11-14 08:00:00+01'), 'CANCELLED', pg_temp.t('2026-11-01 20:30:00+01')),
    (10, 9,  7,  pg_temp.t('2026-11-09 08:00:00+01'), pg_temp.t('2026-11-11 08:00:00+01'), 'CONFIRMED', pg_temp.t('2026-11-02 10:00:00+01')),
    (11, 4,  2,  pg_temp.t('2026-11-12 08:00:00+01'), pg_temp.t('2026-11-14 08:00:00+01'), 'CONFIRMED', pg_temp.t('2026-11-03 15:20:00+01')),
    (12, 5,  14, pg_temp.t('2026-11-14 08:00:00+01'), pg_temp.t('2026-11-16 08:00:00+01'), 'CONFIRMED', pg_temp.t('2026-11-04 09:45:00+01')),
    (13, 10, 1,  pg_temp.t('2026-11-16 08:00:00+01'), pg_temp.t('2026-11-18 08:00:00+01'), 'CONFIRMED', pg_temp.t('2026-11-05 17:05:00+01')),
    (14, 1,  12, pg_temp.t('2026-11-20 14:00:00+01'), pg_temp.t('2026-11-23 08:00:00+01'), 'CONFIRMED', pg_temp.t('2026-11-05 21:10:00+01')),
    (15, 8,  6,  pg_temp.t('2026-11-23 09:00:00+01'), pg_temp.t('2026-11-25 09:00:00+01'), 'CONFIRMED', pg_temp.t('2026-11-06 07:30:00+01')),
    (16, 5,  3,  pg_temp.t('2026-09-28 09:00:00+02'), pg_temp.t('2026-09-30 09:00:00+02'), 'CONVERTED', pg_temp.t('2026-09-25 18:20:00+02')),
    (17, 2,  1,  pg_temp.t('2026-10-01 08:30:00+02'), pg_temp.t('2026-10-03 08:30:00+02'), 'CONVERTED', pg_temp.t('2026-09-29 12:10:00+02')),
    (18, 4,  2,  pg_temp.t('2026-10-05 08:00:00+02'), pg_temp.t('2026-10-07 08:00:00+02'), 'CONVERTED', pg_temp.t('2026-10-02 20:45:00+02')),
    (19, 3,  14, pg_temp.t('2026-10-08 09:00:00+02'), pg_temp.t('2026-10-10 09:00:00+02'), 'CONVERTED', pg_temp.t('2026-10-06 10:30:00+02')),
    (20, 8,  17, pg_temp.t('2026-10-09 14:00:00+02'), pg_temp.t('2026-10-12 08:00:00+02'), 'CONVERTED', pg_temp.t('2026-10-09 13:40:00+02')),
    (21, 6,  4,  pg_temp.t('2026-10-10 08:00:00+02'), pg_temp.t('2026-10-12 08:00:00+02'), 'CONVERTED', pg_temp.t('2026-10-07 17:55:00+02')),
    (22, 1,  20, pg_temp.t('2026-10-10 09:00:00+02'), pg_temp.t('2026-10-12 09:00:00+02'), 'CONVERTED', pg_temp.t('2026-10-08 11:25:00+02')),
    (23, 7,  13, pg_temp.t('2026-10-13 08:00:00+02'), pg_temp.t('2026-10-14 08:00:00+02'), 'CONVERTED', pg_temp.t('2026-10-12 19:00:00+02')),
    (24, 9,  8,  pg_temp.t('2026-10-15 10:00:00+02'), pg_temp.t('2026-10-17 10:00:00+02'), 'CONVERTED', pg_temp.t('2026-10-13 09:15:00+02')),
    (25, 10, 11, pg_temp.t('2026-10-16 09:00:00+02'), pg_temp.t('2026-10-19 09:00:00+02'), 'CONVERTED', pg_temp.t('2026-10-14 16:40:00+02')),
    (26, 2,  12, pg_temp.t('2026-10-20 10:00:00+02'), pg_temp.t('2026-10-21 10:00:00+02'), 'CONVERTED', pg_temp.t('2026-10-18 21:05:00+02')),
    (27, 5,  6,  pg_temp.t('2026-10-22 10:00:00+02'), pg_temp.t('2026-10-23 10:00:00+02'), 'CONVERTED', pg_temp.t('2026-10-22 09:45:00+02')),
    (28, 8,  5,  pg_temp.t('2026-10-26 08:00:00+01'), pg_temp.t('2026-10-28 08:00:00+01'), 'CONVERTED', pg_temp.t('2026-10-23 14:20:00+02')),
    (29, 6,  16, pg_temp.t('2026-10-30 09:00:00+01'), pg_temp.t('2026-11-02 09:00:00+01'), 'CONVERTED', pg_temp.t('2026-10-28 10:50:00+01')),
    (30, 9,  19, pg_temp.t('2026-10-30 10:00:00+01'), pg_temp.t('2026-11-02 10:00:00+01'), 'CONVERTED', pg_temp.t('2026-10-29 15:30:00+01')),
    (31, 10, 15, pg_temp.t('2026-10-31 08:00:00+01'), pg_temp.t('2026-11-02 08:00:00+01'), 'CONVERTED', pg_temp.t('2026-10-27 19:25:00+01')),
    (32, 7,  20, pg_temp.t('2026-10-31 09:00:00+01'), pg_temp.t('2026-11-03 09:00:00+01'), 'CONVERTED', pg_temp.t('2026-10-29 08:40:00+01')),
    (33, 5,  9,  pg_temp.t('2026-11-03 09:00:00+01'), pg_temp.t('2026-11-05 09:00:00+01'), 'CONVERTED', pg_temp.t('2026-11-01 17:15:00+01')),
    (34, 1,  4,  pg_temp.t('2026-11-05 14:00:00+01'), pg_temp.t('2026-11-07 12:00:00+01'), 'CONVERTED', pg_temp.t('2026-11-05 13:35:00+01')),
    (35, 8,  18, pg_temp.t('2026-11-06 08:00:00+01'), pg_temp.t('2026-11-09 08:00:00+01'), 'CONVERTED', pg_temp.t('2026-11-04 20:10:00+01')),
    (101, 7,  9,  pg_temp.t('2026-09-04 09:00:00+02'), pg_temp.t('2026-09-06 09:00:00+02'), 'CONVERTED', pg_temp.t('2026-09-02 10:20:00+02')),
    (102, 3,  2,  pg_temp.t('2026-09-10 08:00:00+02'), pg_temp.t('2026-09-13 08:00:00+02'), 'CONVERTED', pg_temp.t('2026-09-07 15:30:00+02')),
    (103, 8,  11, pg_temp.t('2026-09-12 08:00:00+02'), pg_temp.t('2026-09-14 08:00:00+02'), 'CONVERTED', pg_temp.t('2026-09-10 09:00:00+02')),
    (104, 4,  16, pg_temp.t('2026-09-18 08:30:00+02'), pg_temp.t('2026-09-21 08:30:00+02'), 'CONVERTED', pg_temp.t('2026-09-15 18:45:00+02')),
    (105, 9,  19, pg_temp.t('2026-09-19 09:00:00+02'), pg_temp.t('2026-09-20 09:00:00+02'), 'CONVERTED', pg_temp.t('2026-09-18 12:15:00+02')),
    (106, 6,  6,  pg_temp.t('2026-09-22 10:00:00+02'), pg_temp.t('2026-09-25 10:00:00+02'), 'CONVERTED', pg_temp.t('2026-09-19 20:00:00+02')),
    (107, 2,  20, pg_temp.t('2026-09-24 09:00:00+02'), pg_temp.t('2026-09-26 09:00:00+02'), 'CONVERTED', pg_temp.t('2026-09-22 11:40:00+02')),
    (108, 10, 4,  pg_temp.t('2026-09-05 08:00:00+02'), pg_temp.t('2026-09-07 08:00:00+02'), 'CONVERTED', pg_temp.t('2026-09-03 09:05:00+02')),
    (109, 4,  8,  pg_temp.t('2026-11-05 10:00:00+01'), pg_temp.t('2026-11-08 10:00:00+01'), 'CONVERTED', pg_temp.t('2026-11-03 14:00:00+01')),
    (110, 9,  10, pg_temp.t('2026-11-02 09:00:00+01'), pg_temp.t('2026-11-04 09:00:00+01'), 'CONVERTED', pg_temp.t('2026-10-30 11:00:00+01')),
    (111, 7,  12, pg_temp.t('2026-11-12 09:00:00+01'), pg_temp.t('2026-11-14 09:00:00+01'), 'CONFIRMED', pg_temp.t('2026-11-05 12:00:00+01')),
    (112, 3,  4,  pg_temp.t('2026-11-10 08:00:00+01'), pg_temp.t('2026-11-17 08:00:00+01'), 'CONFIRMED', pg_temp.t('2026-11-06 09:15:00+01')),
    (113, 2,  2,  pg_temp.t('2026-11-14 09:00:00+01'), pg_temp.t('2026-11-16 08:00:00+01'), 'CONFIRMED', pg_temp.t('2026-11-05 18:30:00+01')),
    (114, 5,  17, pg_temp.t('2026-11-13 08:00:00+01'), pg_temp.t('2026-11-15 08:00:00+01'), 'CONFIRMED', pg_temp.t('2026-11-04 08:30:00+01')),
    (115, 5,  14, pg_temp.t('2026-11-18 08:00:00+01'), pg_temp.t('2026-11-20 08:00:00+01'), 'CANCELLED', pg_temp.t('2026-11-03 19:00:00+01')),
    (116, 10, 17, pg_temp.t('2026-11-02 08:00:00+01'), pg_temp.t('2026-11-03 08:00:00+01'), 'EXPIRED',   pg_temp.t('2026-10-30 12:00:00+01')),
    (117, 1,  1,  pg_temp.t('2026-07-08 08:30:00+02'), pg_temp.t('2026-07-11 08:30:00+02'), 'CONVERTED', pg_temp.t('2026-07-03 10:00:00+02')),
    (118, 8,  11, pg_temp.t('2026-07-20 09:00:00+02'), pg_temp.t('2026-07-23 09:00:00+02'), 'CONVERTED', pg_temp.t('2026-07-16 14:20:00+02')),
    (119, 4,  18, pg_temp.t('2026-08-05 10:00:00+02'), pg_temp.t('2026-08-08 10:00:00+02'), 'CONVERTED', pg_temp.t('2026-08-01 17:30:00+02')),
    (120, 3,  9,  pg_temp.t('2026-08-14 08:00:00+02'), pg_temp.t('2026-08-17 08:00:00+02'), 'CONVERTED', pg_temp.t('2026-08-10 09:45:00+02'))
ON CONFLICT (id) DO UPDATE SET
        customer_id = EXCLUDED.customer_id,
        tool_item_id = EXCLUDED.tool_item_id,
        start_date = EXCLUDED.start_date,
        end_date = EXCLUDED.end_date,
        status = EXCLUDED.status,
        created_at = EXCLUDED.created_at;

-- Fees are daily_rate x days; a late return costs one extra day (rentals 3 and 22).
-- ACTIVE and OVERDUE rentals have no return data or fees yet.
-- CANCELLED rentals (9, 11) came back unused the same day, free of charge.
-- staff_id is the STAFF user (3-5) who handed the tool over.
INSERT INTO rentals (id, customer_id, tool_item_id, reservation_id, staff_id, start_date, expected_end_date,
                     actual_end_date, deposit_paid, condition_at_handover, condition_at_return,
                     base_fee, late_fee, total_fee, status) VALUES
    (1,  5,  3,  16, 3, pg_temp.t('2026-09-28 09:00:00+02'), pg_temp.t('2026-09-30 09:00:00+02'), pg_temp.t('2026-09-30 08:50:00+02'), 20000.00, 'FAIR', 'POOR',    9000.00,  0.00,    9000.00,  'RETURNED'),
    (2,  2,  1,  17, 4, pg_temp.t('2026-10-01 08:30:00+02'), pg_temp.t('2026-10-03 08:30:00+02'), pg_temp.t('2026-10-03 08:15:00+02'), 20000.00, 'GOOD', 'GOOD',    9000.00,  0.00,    9000.00,  'RETURNED'),
    (3,  4,  2,  18, 3, pg_temp.t('2026-10-05 08:00:00+02'), pg_temp.t('2026-10-07 08:00:00+02'), pg_temp.t('2026-10-07 16:30:00+02'), 20000.00, 'GOOD', 'GOOD',    9000.00,  4500.00, 13500.00, 'RETURNED'),
    (4,  3,  14, 19, 5, pg_temp.t('2026-10-08 09:00:00+02'), pg_temp.t('2026-10-10 09:00:00+02'), pg_temp.t('2026-10-10 08:30:00+02'), 30000.00, 'GOOD', 'GOOD',    10000.00, 0.00,    10000.00, 'RETURNED'),
    (5,  8,  17, 20, 4, pg_temp.t('2026-10-09 14:00:00+02'), pg_temp.t('2026-10-12 08:00:00+02'), pg_temp.t('2026-10-12 07:45:00+02'), 25000.00, 'GOOD', 'GOOD',    15000.00, 0.00,    15000.00, 'RETURNED'),
    (6,  6,  4,  21, 3, pg_temp.t('2026-10-10 08:00:00+02'), pg_temp.t('2026-10-12 08:00:00+02'), pg_temp.t('2026-10-12 07:40:00+02'), 60000.00, 'GOOD', 'GOOD',    18000.00, 0.00,    18000.00, 'RETURNED'),
    (7,  1,  20, 22, 5, pg_temp.t('2026-10-10 09:00:00+02'), pg_temp.t('2026-10-12 09:00:00+02'), pg_temp.t('2026-10-12 08:45:00+02'), 40000.00, 'GOOD', 'GOOD',    16000.00, 0.00,    16000.00, 'RETURNED'),
    (8,  7,  13, 23, 4, pg_temp.t('2026-10-13 08:00:00+02'), pg_temp.t('2026-10-14 08:00:00+02'), pg_temp.t('2026-10-14 07:55:00+02'), 10000.00, 'GOOD', 'DAMAGED', 3000.00,  0.00,    3000.00,  'RETURNED'),
    (9,  9,  8,  24, 3, pg_temp.t('2026-10-15 10:00:00+02'), pg_temp.t('2026-10-17 10:00:00+02'), pg_temp.t('2026-10-15 10:40:00+02'), 15000.00, 'GOOD', 'GOOD',    0.00,     0.00,    0.00,     'CANCELLED'),
    (10, 10, 11, 25, 5, pg_temp.t('2026-10-16 09:00:00+02'), pg_temp.t('2026-10-19 09:00:00+02'), pg_temp.t('2026-10-19 08:30:00+02'), 40000.00, 'GOOD', 'GOOD',    22500.00, 0.00,    22500.00, 'RETURNED'),
    (11, 2,  12, 26, 4, pg_temp.t('2026-10-20 10:00:00+02'), pg_temp.t('2026-10-21 10:00:00+02'), pg_temp.t('2026-10-20 10:20:00+02'), 10000.00, 'GOOD', 'GOOD',    0.00,     0.00,    0.00,     'CANCELLED'),
    (12, 5,  6,  27, 3, pg_temp.t('2026-10-22 10:00:00+02'), pg_temp.t('2026-10-23 10:00:00+02'), pg_temp.t('2026-10-23 09:30:00+02'), 15000.00, 'GOOD', 'FAIR',    3000.00,  0.00,    3000.00,  'RETURNED'),
    (13, 3,  18, 3,  5, pg_temp.t('2026-10-23 15:00:00+02'), pg_temp.t('2026-10-24 15:00:00+02'), pg_temp.t('2026-10-24 12:00:00+02'), 25000.00, 'GOOD', 'GOOD',    5500.00,  0.00,    5500.00,  'RETURNED'),
    (14, 8,  5,  28, 4, pg_temp.t('2026-10-26 08:00:00+01'), pg_temp.t('2026-10-28 08:00:00+01'), pg_temp.t('2026-10-28 07:50:00+01'), 60000.00, 'FAIR', 'FAIR',    18000.00, 0.00,    18000.00, 'RETURNED'),
    (15, 1,  10, 4,  3, pg_temp.t('2026-10-27 08:30:00+01'), pg_temp.t('2026-10-30 08:30:00+01'), pg_temp.t('2026-10-30 08:00:00+01'), 10000.00, 'FAIR', 'FAIR',    7500.00,  0.00,    7500.00,  'RETURNED'),
    (16, 4,  8,  6,  5, pg_temp.t('2026-10-29 13:00:00+01'), pg_temp.t('2026-10-31 13:00:00+01'), pg_temp.t('2026-10-31 12:00:00+01'), 15000.00, 'GOOD', 'GOOD',    7000.00,  0.00,    7000.00,  'RETURNED'),
    (17, 6,  16, 29, 4, pg_temp.t('2026-10-30 09:00:00+01'), pg_temp.t('2026-11-02 09:00:00+01'), NULL,                     30000.00, 'GOOD', NULL,      NULL,     NULL,    NULL,     'OVERDUE'),
    (18, 9,  19, 30, 3, pg_temp.t('2026-10-30 10:00:00+01'), pg_temp.t('2026-11-02 10:00:00+01'), pg_temp.t('2026-11-02 09:30:00+01'), 25000.00, 'GOOD', 'DAMAGED', 16500.00, 0.00,    16500.00, 'RETURNED'),
    (19, 10, 15, 31, 5, pg_temp.t('2026-10-31 08:00:00+01'), pg_temp.t('2026-11-02 08:00:00+01'), pg_temp.t('2026-11-02 08:00:00+01'), 30000.00, 'FAIR', 'FAIR',    13000.00, 0.00,    13000.00, 'RETURNED'),
    (20, 7,  20, 32, 4, pg_temp.t('2026-10-31 09:00:00+01'), pg_temp.t('2026-11-03 09:00:00+01'), NULL,                     40000.00, 'GOOD', NULL,      NULL,     NULL,    NULL,     'OVERDUE'),
    (21, 2,  11, 5,  3, pg_temp.t('2026-11-02 08:00:00+01'), pg_temp.t('2026-11-09 08:00:00+01'), NULL,                     40000.00, 'GOOD', NULL,      NULL,     NULL,    NULL,     'ACTIVE'),
    (22, 5,  9,  33, 5, pg_temp.t('2026-11-03 09:00:00+01'), pg_temp.t('2026-11-05 09:00:00+01'), pg_temp.t('2026-11-05 17:00:00+01'), 15000.00, 'GOOD', 'DAMAGED', 7000.00,  3500.00, 10500.00, 'RETURNED'),
    (23, 3,  1,  8,  4, pg_temp.t('2026-11-04 09:00:00+01'), pg_temp.t('2026-11-09 09:00:00+01'), NULL,                     20000.00, 'GOOD', NULL,      NULL,     NULL,    NULL,     'ACTIVE'),
    (24, 1,  4,  34, 3, pg_temp.t('2026-11-05 14:00:00+01'), pg_temp.t('2026-11-07 12:00:00+01'), NULL,                     60000.00, 'GOOD', NULL,      NULL,     NULL,    NULL,     'ACTIVE'),
    (25, 8,  18, 35, 5, pg_temp.t('2026-11-06 08:00:00+01'), pg_temp.t('2026-11-09 08:00:00+01'), NULL,                     25000.00, 'GOOD', NULL,      NULL,     NULL,    NULL,     'ACTIVE'),
    (101, 7,  9,  101, 4, pg_temp.t('2026-09-04 09:00:00+02'), pg_temp.t('2026-09-06 09:00:00+02'), pg_temp.t('2026-09-06 08:45:00+02'), 15000.00, 'GOOD', 'GOOD',    7000.00,  0.00,    7000.00,  'RETURNED'),
    (102, 3,  2,  102, 4, pg_temp.t('2026-09-10 08:00:00+02'), pg_temp.t('2026-09-13 08:00:00+02'), pg_temp.t('2026-09-13 07:50:00+02'), 20000.00, 'GOOD', 'GOOD',    13500.00, 0.00,    13500.00, 'RETURNED'),
    (103, 8,  11, 103, 3, pg_temp.t('2026-09-12 08:00:00+02'), pg_temp.t('2026-09-14 08:00:00+02'), pg_temp.t('2026-09-14 17:30:00+02'), 40000.00, 'GOOD', 'GOOD',    15000.00, 7500.00, 22500.00, 'RETURNED'),
    (104, 4,  16, 104, 5, pg_temp.t('2026-09-18 08:30:00+02'), pg_temp.t('2026-09-21 08:30:00+02'), pg_temp.t('2026-09-21 08:10:00+02'), 30000.00, 'NEW',  'GOOD',    19500.00, 0.00,    19500.00, 'RETURNED'),
    (105, 9,  19, 105, 4, pg_temp.t('2026-09-19 09:00:00+02'), pg_temp.t('2026-09-20 09:00:00+02'), pg_temp.t('2026-09-20 08:50:00+02'), 25000.00, 'GOOD', 'FAIR',    5500.00,  0.00,    5500.00,  'RETURNED'),
    (106, 6,  6,  106, 3, pg_temp.t('2026-09-22 10:00:00+02'), pg_temp.t('2026-09-25 10:00:00+02'), pg_temp.t('2026-09-25 09:40:00+02'), 15000.00, 'GOOD', 'FAIR',    9000.00,  0.00,    9000.00,  'RETURNED'),
    (107, 2,  20, 107, 5, pg_temp.t('2026-09-24 09:00:00+02'), pg_temp.t('2026-09-26 09:00:00+02'), pg_temp.t('2026-09-26 08:30:00+02'), 40000.00, 'NEW',  'GOOD',    16000.00, 0.00,    16000.00, 'RETURNED'),
    (108, 10, 4,  108, 4, pg_temp.t('2026-09-05 08:00:00+02'), pg_temp.t('2026-09-07 08:00:00+02'), pg_temp.t('2026-09-07 07:45:00+02'), 60000.00, 'GOOD', 'GOOD',    18000.00, 0.00,    18000.00, 'RETURNED'),
    (109, 4,  8,  109, 3, pg_temp.t('2026-11-05 10:00:00+01'), pg_temp.t('2026-11-08 10:00:00+01'), NULL,                     15000.00, 'GOOD', NULL,      NULL,     NULL,    NULL,     'ACTIVE'),
    (110, 9,  10, 110, 4, pg_temp.t('2026-11-02 09:00:00+01'), pg_temp.t('2026-11-04 09:00:00+01'), NULL,                     10000.00, 'FAIR', NULL,      NULL,     NULL,    NULL,     'OVERDUE'),
    (111, 1,  1,  117, 3, pg_temp.t('2026-07-08 08:30:00+02'), pg_temp.t('2026-07-11 08:30:00+02'), pg_temp.t('2026-07-11 08:10:00+02'), 20000.00, 'GOOD', 'GOOD',    13500.00, 0.00,    13500.00, 'RETURNED'),
    (112, 8,  11, 118, 4, pg_temp.t('2026-07-20 09:00:00+02'), pg_temp.t('2026-07-23 09:00:00+02'), pg_temp.t('2026-07-23 18:00:00+02'), 40000.00, 'GOOD', 'GOOD',    22500.00, 7500.00, 30000.00, 'RETURNED'),
    (113, 4,  18, 119, 5, pg_temp.t('2026-08-05 10:00:00+02'), pg_temp.t('2026-08-08 10:00:00+02'), pg_temp.t('2026-08-08 09:30:00+02'), 25000.00, 'GOOD', 'GOOD',    16500.00, 0.00,    16500.00, 'RETURNED'),
    (114, 3,  9,  120, 3, pg_temp.t('2026-08-14 08:00:00+02'), pg_temp.t('2026-08-17 08:00:00+02'), pg_temp.t('2026-08-17 07:50:00+02'), 15000.00, 'GOOD', 'FAIR',    10500.00, 0.00,    10500.00, 'RETURNED')
ON CONFLICT (id) DO UPDATE SET
        customer_id = EXCLUDED.customer_id,
        tool_item_id = EXCLUDED.tool_item_id,
        reservation_id = EXCLUDED.reservation_id,
        staff_id = EXCLUDED.staff_id,
        start_date = EXCLUDED.start_date,
        expected_end_date = EXCLUDED.expected_end_date,
        actual_end_date = EXCLUDED.actual_end_date,
        deposit_paid = EXCLUDED.deposit_paid,
        condition_at_handover = EXCLUDED.condition_at_handover,
        condition_at_return = EXCLUDED.condition_at_return,
        base_fee = EXCLUDED.base_fee,
        late_fee = EXCLUDED.late_fee,
        total_fee = EXCLUDED.total_fee,
        status = EXCLUDED.status;

-- Records without completed_at are still open (tool_items 5, 15 and 19).
INSERT INTO maintenance (id, tool_item_id, type, description, started_at, completed_at, cost, performed_by) VALUES
    (1, 3,  'INSPECTION', 'Erős tokmány-holtjátékkal jött vissza. A hajtómű kopott, a javítás nem gazdaságos, az eszközt kivontuk a forgalomból.', pg_temp.t('2026-10-01 09:00:00+02'), pg_temp.t('2026-10-01 10:30:00+02'), 0.00,     'toth.gabor'),
    (2, 17, 'SCHEDULED',  'Szezonális szerviz: új gyújtógyertya, légszűrő tisztítva, damilfej cserélve.',        pg_temp.t('2026-10-13 08:00:00+02'), pg_temp.t('2026-10-13 11:30:00+02'), 6500.00,  'fekete.laszlo'),
    (3, 13, 'INSPECTION', 'Füstölve jött vissza. A motor forgórésze leégett, alkatrészre vár.',               pg_temp.t('2026-10-14 10:00:00+02'), pg_temp.t('2026-10-14 10:45:00+02'), 0.00,     'papp.judit'),
    (4, 11, 'SCHEDULED',  'Új 210 mm-es keményfém fűrészlap felszerelve, párhuzamvezető és lapbeállítás ellenőrizve.',                 pg_temp.t('2026-10-19 10:00:00+02'), pg_temp.t('2026-10-19 12:00:00+02'), 12900.00, 'fekete.laszlo'),
    (5, 6,  'REPAIR',     'A tokmány terhelés alatt megcsúszott, gyorstokmány cserélve.',                                  pg_temp.t('2026-10-23 11:00:00+02'), pg_temp.t('2026-10-24 12:00:00+02'), 7900.00,  'toth.gabor'),
    (6, 19, 'REPAIR',     'Kölcsönzés után szivárog a szivattyú, a nagynyomású tömítések cseréje folyamatban (alkatrész megrendelve).',          pg_temp.t('2026-11-02 11:00:00+01'), NULL,                     0.00,     'toth.gabor'),
    (7, 15, 'SCHEDULED',  'Láncélezés, láncvezető sorjázása, légszűrő és lánckerék ellenőrzése.',                pg_temp.t('2026-11-04 09:00:00+01'), NULL,                     0.00,     'fekete.laszlo'),
    (8, 5,  'INSPECTION', 'Éves érintésvédelmi felülvizsgálat (szigetelés és védővezető folytonosság).',                    pg_temp.t('2026-11-06 08:00:00+01'), NULL,                     0.00,     'papp.judit'),
    (101,  8,  'SCHEDULED', 'Új fűrészlap felszerelve, védőburkolat és vezetősín ellenőrizve.',  pg_temp.t('2026-10-20 08:00:00+02'), pg_temp.t('2026-10-20 09:30:00+02'), 5200.00,  'papp.judit'),
    (102, 2,  'REPAIR',    'Elkopott szénkefék cseréje, a motor átvizsgálva.',                  pg_temp.t('2026-10-08 13:00:00+02'), pg_temp.t('2026-10-09 10:00:00+02'), 4800.00,  'toth.gabor'),
    (103, 20, 'SCHEDULED', 'Szűrők és tömlők cseréje, a szivattyú átöblítve.',                  pg_temp.t('2026-09-30 09:00:00+02'), pg_temp.t('2026-09-30 11:00:00+02'), 3500.00,  'fekete.laszlo'),
    (104, 11, 'REPAIR',    'Szénkefék és főkapcsoló cseréje, a fűrészasztal motorja újra terhelhető.', pg_temp.t('2026-06-25 09:00:00+02'), pg_temp.t('2026-06-26 12:00:00+02'), 18500.00, 'papp.judit'),
    (105, 9,  'SCHEDULED', 'Éves szerviz: fűrészlap csere, csapágyak és védőburkolat ellenőrzése.',    pg_temp.t('2026-07-02 08:30:00+02'), pg_temp.t('2026-07-02 11:00:00+02'), 9800.00,  'fekete.laszlo'),
    (106, 16, 'SCHEDULED', 'Új lánc és láncvezető, olajozó rendszer tisztítva, gyújtógyertya cserélve.', pg_temp.t('2026-08-03 08:00:00+02'), pg_temp.t('2026-08-03 12:00:00+02'), 14200.00, 'fekete.laszlo'),
    (107, 20, 'REPAIR',    'Szivattyú és szívócső cseréje, a kárpittisztító teljes nyomáspróbája.',   pg_temp.t('2026-08-25 09:00:00+02'), pg_temp.t('2026-08-27 15:00:00+02'), 42500.00, 'toth.gabor')
ON CONFLICT (id) DO UPDATE SET
        tool_item_id = EXCLUDED.tool_item_id,
        type = EXCLUDED.type,
        description = EXCLUDED.description,
        started_at = EXCLUDED.started_at,
        completed_at = EXCLUDED.completed_at,
        cost = EXCLUDED.cost,
        performed_by = EXCLUDED.performed_by;

-- Needs migration 000002. Link every CUSTOMER login to its customer record by email.
UPDATE users u
SET customer_id = c.id
FROM customers c
WHERE u.role = 'CUSTOMER' AND u.email = c.email AND u.customer_id IS NULL;

-- Maintenance thresholds (needs migration 000002). NULL = that rule is not used.
UPDATE tool_types SET maintenance_every_rentals = 20, maintenance_every_days = 180 WHERE id IN (1, 2, 3);
UPDATE tool_types SET maintenance_every_rentals = 15, maintenance_every_days = 180 WHERE id IN (4, 5, 6, 7, 8);
UPDATE tool_types SET maintenance_every_rentals = 10, maintenance_every_days = 90, maintenance_every_hours = 50 WHERE id IN (9, 10);
UPDATE tool_types SET maintenance_every_rentals = 25, maintenance_every_days = 365 WHERE id IN (11, 12);

-- Move each id sequence past the ids used above, so new rows get fresh ids.
SELECT setval(pg_get_serial_sequence('users', 'id'),        (SELECT MAX(id) FROM users));
SELECT setval(pg_get_serial_sequence('categories', 'id'),   (SELECT MAX(id) FROM categories));
SELECT setval(pg_get_serial_sequence('tool_types', 'id'),   (SELECT MAX(id) FROM tool_types));
SELECT setval(pg_get_serial_sequence('tool_items', 'id'),   (SELECT MAX(id) FROM tool_items));
SELECT setval(pg_get_serial_sequence('customers', 'id'),    (SELECT MAX(id) FROM customers));
SELECT setval(pg_get_serial_sequence('reservations', 'id'), (SELECT MAX(id) FROM reservations));
SELECT setval(pg_get_serial_sequence('rentals', 'id'),      (SELECT MAX(id) FROM rentals));
SELECT setval(pg_get_serial_sequence('maintenance', 'id'),  (SELECT MAX(id) FROM maintenance));

COMMIT;
