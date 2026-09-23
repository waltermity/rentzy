#!/usr/bin/env bash
set -euo pipefail

# ---------- config ----------
DB_URL="${DATABASE_URL:-postgres://rentzy:rentzy@localhost:5432/rentzy?sslmode=disable}"
BACKEND_DIR="backend"

# ---------- helpers ----------
log()  { printf "\033[1;34m[dev]\033[0m %s\n" "$*"; }
fail() { printf "\033[1;31m[dev]\033[0m %s\n" "$*" >&2; exit 1; }

require() {
    command -v "$1" >/dev/null 2>&1 || fail "'$1' not found in PATH"
}

wait_for_db() {
    log "waiting for PostgreSQL to accept connections..."
    local retries=30
    until docker compose exec -T db pg_isready -U rentzy -d rentzy >/dev/null 2>&1; do
        retries=$((retries - 1))
        if [ "$retries" -le 0 ]; then
            fail "PostgreSQL did not become ready in time"
        fi
        sleep 1
    done
    log "PostgreSQL is ready"
}

# ---------- commands ----------
cmd_up() {
    require docker
    require migrate

    log "starting PostgreSQL..."
    docker compose up -d

    wait_for_db

    log "applying migrations..."
    ( cd "$BACKEND_DIR" && migrate -path migrations -database "$DB_URL" up )

    log "starting backend..."
    ( cd "$BACKEND_DIR" && go run ./cmd/rentzy )
}

cmd_down() {
    log "stopping PostgreSQL..."
    docker compose down
}

cmd_migrate_up() {
    require migrate
    ( cd "$BACKEND_DIR" && migrate -path migrations -database "$DB_URL" up )
}

cmd_migrate_down() {
    require migrate
    ( cd "$BACKEND_DIR" && migrate -path migrations -database "$DB_URL" down 1 )
}

cmd_migrate_create() {
    require migrate
    [ $# -ge 1 ] || fail "usage: ./dev.sh migrate-create <name>"
    ( cd "$BACKEND_DIR" && migrate create -ext sql -dir migrations -seq "$1" )
}

cmd_psql() {
    docker compose exec db psql -U rentzy -d rentzy
}

cmd_reset() {
    log "WARNING: this will destroy the database and re-apply migrations"
    read -r -p "Are you sure? [y/N] " ans
    [[ "$ans" =~ ^[Yy]$ ]] || { log "aborted"; exit 0; }

    docker compose down -v
    cmd_up
}

cmd_help() {
    cat <<EOF
Usage: ./dev.sh <command>

Commands:
  up               Start PostgreSQL, apply migrations, run the backend
  down             Stop PostgreSQL
  migrate-up       Apply pending migrations
  migrate-down     Roll back the last migration
  migrate-create   Create a new migration (usage: ./dev.sh migrate-create <name>)
  psql             Open a psql shell inside the DB container
  reset            Destroy the DB volume and re-run everything
  help             Show this message
EOF
}

# ---------- dispatch ----------
case "${1:-help}" in
    up)              cmd_up ;;
    down)            cmd_down ;;
    migrate-up)      cmd_migrate_up ;;
    migrate-down)    cmd_migrate_down ;;
    migrate-create)  shift; cmd_migrate_create "$@" ;;
    psql)            cmd_psql ;;
    reset)           cmd_reset ;;
    help|-h|--help)  cmd_help ;;
    *)               fail "unknown command: $1 (try ./dev.sh help)" ;;
esac