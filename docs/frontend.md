# Frontend Architecture

## Overview

The frontend is a single-page application built to manage the Rentzy tool rental system. It provides distinct interfaces for visitors, customers, staff, and administrators.

The application is built with **Vue 3** and **TypeScript**, and relies on **Pinia** for state management. It communicates with the Go backend via REST API and uses the **PrimeVue** component library for a responsive, consistent UI. The UI is in **Hungarian** (texts, dates, calendar, error messages). Amounts are HUF, shown as `4 500 Ft`.

## Tech Stack

- **Framework** — Vue 3 (`<script setup>`, TypeScript)
- **Build tool** — Vite
- **State Management** — Pinia
- **Component Library** — PrimeVue 4.5 with the Aura theme (light and dark, follows the OS setting) and PrimeIcons 7
- **Routing** — Vue Router
- **Communication** — Axios

> **Why PrimeVue 4 and not 5:** from version 5, PrimeVue, PrimeIcons and `@primeuix/themes` are no longer MIT licensed. They need a PrimeUI license key, and without one they show an "Invalid PrimeUI License" banner. PrimeVue 4.5.5, `@primeuix/themes` 2.0.3 and `primeicons` 7.0.0 are MIT and pinned to exact versions in `package.json`. Do not upgrade them to the next major without a license decision.

## Language and formatting

- All UI texts are written in Hungarian in the components. Shared texts live in `src/i18n/hu.ts`: the Hungarian names of the backend's enum values (statuses, conditions, roles, maintenance types, using the specification's terms: *függőben, visszaigazolt, törölt, lejárt; aktív, visszavett, késett*), the PrimeVue locale (calendar, password meter, empty lists), and the API error translations.
- The backend stays English. `translateError()` in `src/i18n/hu.ts` maps the API error code and the known backend messages to Hungarian; an unknown message falls back to a Hungarian text by error code. When a new backend error message is added, add its pattern there.
- Dates use `hu-HU` (`2026. nov. 9. 08:00`), the week starts on Monday. Money is formatted by `money()` in `src/utils/format.ts` as whole forints with a space as thousands separator (`4 500 Ft`, also for 4-digit amounts).

## API Communication

- All requests go to `/api/...` on the frontend's own origin. In development, Vite proxies `/api` to the Go API at `http://localhost:8080` (override with the `VITE_API_TARGET` env variable), so no CORS is needed.
- `src/api/client.ts` holds one axios instance. It adds `Authorization: Bearer <token>`, unwraps the API's `{"data": ...}` envelope, and turns `{"error": {code, message}}` into an `ApiError`. A 401 on an authenticated request logs the user out.
- `src/api/index.ts` has one function per backend endpoint; `src/api/types.ts` mirrors the Go JSON types.
- The JWT is stored in `localStorage` (`rentzy.token`). On page load the session is restored with `GET /api/auth/me` before the router starts, so the guards know the role.

## State Management (Pinia Stores)

- `useAuthStore` — Manages the JWT, user profile, customer record and RBAC role (`ADMIN`, `STAFF`, `CUSTOMER`); login, register, logout and the landing page per role.
- `useToolStore` — Caches catalog data (categories, tool types).
- `useReservationStore` — Manages the current customer's reservations and rentals (create, cancel).
- `useRentalStore` — Handles operational data for staff: open rentals, reservations waiting for handover, handovers, returns and cancellations.

## Layout

The header is on every page: the logo and the links the user's role may open on the left, and on the right the user's name with their role (a menu with **Profile** and **Log out**), or a **Login** button for visitors. The header spans the full window width and measures itself: the links show icon and label when they fit, only icons (with tooltips) when they don't, and collapse into a single menu button (☰) when even the icons don't fit (for example the admin's 7 links on a phone).

After logging in, users land on a page by role: customers on `/catalog`, staff on `/staff/rentals`, admins on `/admin/reports`. If they were sent to `/login` from a protected page, they go back to that page instead.

## Views & Routes

Access to routes is enforced by role-based guards in Vue Router, matching the backend's RBAC. A user without the right role is sent to their landing page; a visitor is sent to `/login`.

### 1. Public Views (no login needed)
- **`/catalog`**
  - Tool browsing interface with search and category filters.
  - Uses CSS Grid with PrimeVue `<Card>` components to display tool types, with daily rate, deposit and how many copies are available now.
- **`/catalog/:id`**
  - Detailed view of a specific tool type.
  - Integrates a PrimeVue `<DatePicker>` (the v4 name of `<Calendar>`) for picking the pick-up and return days. Days on which every copy is booked all day are disabled. After a period is picked, the page checks how many copies are free for the whole period and shows the estimated fee and deposit.
  - Visitors see a **Log in to reserve** button; customers reserve for themselves; staff and admins pick a customer and reserve for them.
- **`/login`** — Exchanges credentials for a JWT, which is stored and attached to subsequent API requests.
- **`/register`** — Customer self-registration (creates the login and the customer record), then logs in.

### 2. Logged-in Views (any role)
- **`/profile`**
  - Account details and password change for everyone.
  - Customers also see their upcoming reservations (with cancel), their rentals with fees, and past reservations.

### 3. Staff Views (STAFF and ADMIN)
- **`/staff/rentals`**
  - Operational dashboard for daily tasks: tools out now (overdue ones highlighted), reservations waiting for handover, and history.
  - **Handover Flow**: Staff records the tool's condition at handover and confirms the deposit payment (a checkbox; the button stays disabled until it is ticked). Copies in maintenance, broken or still rented are refused, and a due maintenance is shown as a warning.
  - **Return Flow**: Staff records the condition at return and optional operating hours. A fee preview is shown before submitting; afterwards the interface displays the final fees calculated by the server (base fee, late fee, total, deposit). A tool returned **Damaged** becomes **Broken**.
  - **Cancel**: closes a rental free of charge when the tool came back unused.
  - **Counter rental**: reserve a tool for a customer starting now and hand it over in one step (walk-in customers; the backend always needs a reservation).
- **`/staff/customers`** — Customer list with search, and each customer's reservations and rentals. Admins can also create and edit customers.
- **`/staff/maintenance`** — Copies with maintenance due (and why), maintenance in progress, and history. Admins can start and complete maintenance; staff can only view.
- **`/admin/reports`** — Staff see the most rented tools; admins see all reports (see below).

### 4. Administrator Views (ADMIN only)
- **`/admin/tools`**
  - Management interface for tool types, individual copies and categories.
  - Allows adding new tools, updating pricing and maintenance thresholds, adding copies, changing a copy's status (Available, Broken, Retired) or condition, viewing its maintenance history, and moving copies into maintenance.
- **`/admin/users`**
  - Account management for admins, staff and customers: create users (customer logins link an existing customer or create a new one), change email or password, switch between ADMIN and STAFF.
- **`/admin/reports`**
  - Statistical dashboard for a chosen period (30 days, 90 days, 1 year or custom): revenue, rentals and average utilization tiles; most rented tools; utilization and revenue per tool and per copy; and procurement suggestions above an adjustable utilization threshold, with the estimated payback time.
