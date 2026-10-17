# Frontend Architecture

## Overview

The frontend is a single-page application built to manage the Rentzy tool rental system. It provides distinct interfaces for customers, staff, and administrators. 

The application is built with **Vue 3** and relies on **Pinia** for state management. It communicates with the Go backend via REST API and uses the **PrimeVue** component library for a responsive, consistent UI.

## Tech Stack

- **Framework** — Vue 3
- **State Management** — Pinia
- **Component Library** — PrimeVue 
- **Routing** — Vue Router
- **Communication** — Fetch / Axios

## State Management (Pinia Stores)

- `useAuthStore` — Manages the JWT, user profile, and RBAC roles (`admin`, `staff`, `customer`).
- `useToolStore` — Caches catalog data and inventory.
- `useReservationStore` — Manages the current customer's reservations.
- `useRentalStore` — Handles operational data (rentals) for staff actions (handovers, returns).

## Views & Routes

Access to routes is enforced by role-based guards in Vue Router, matching the backend's RBAC.

### 1. Public Views
- **`/login`**
  - Authentication interface.
  - Exchanges credentials for a JWT, which is then stored and attached to subsequent API requests.

### 2. Customer Views
- **`/catalog`** 
  - Tool browsing interface with search and category filters.
  - Uses CSS Grid with PrimeVue `<Card>` components to display tool types.
- **`/catalog/:id`**
  - Detailed view of a specific tool type.
  - Integrates a PrimeVue `<Calendar>` for date selection. The calendar visually disables dates where the tool is unavailable, allowing users to create reservations.
- **`/profile`**
  - Customer's personal dashboard.
  - Displays active/past reservations and rentals. Allows cancellation of pending reservations.

### 3. Staff Views
- **`/staff/rentals`**
  - Operational dashboard for daily tasks (lists active and upcoming rentals).
  - **Handover Flow**: Submits handover details to the API. Staff records the initial condition of the tool and confirms the deposit payment.
  - **Return Flow**: Submits return details to the API. Staff records the final condition of the tool. The interface then displays the calculated final fees returned by the server.

### 4. Administrator Views
- **`/admin/tools`**
  - Management interface for tool types and individual tool items.
  - Allows adding new tools, updating pricing, and moving physical items into maintenance status.
- **`/admin/users`**
  - Account management for customers and staff (role assignments, profile edits).
- **`/admin/reports`**
  - Statistical dashboard displaying tool utilization metrics, revenue generated per tool, and system-generated procurement suggestions.