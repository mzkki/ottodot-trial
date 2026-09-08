#!/usr/bin/env bash
set -e

# Load .env if present
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

if [ -f "$ROOT_DIR/.env" ]; then
    # Export variables from .env ignoring comments
    export $(grep -v '^#' "$ROOT_DIR/.env" | xargs)
fi

DB_USER="${DB_USER:-ottodot}"
DB_PASS="${DB_PASS:-ottodot123}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_SYSTEM="${DB_NAME:-db_ottodot_trial}"
DB_TEST="${DB_TEST_NAME:-db_ottodot_trial_test}"
MIGRATIONS_DIR="$ROOT_DIR/migrations"

# Determine target database
TARGET="${1:-system}"

fresh_db() {
    local target_db="$1"
    local db_url="postgres://${DB_USER}:${DB_PASS}@${DB_HOST}:${DB_PORT}/${target_db}?sslmode=disable"
    echo "==> Freshening database: ${target_db}..."

    if command -v migrate >/dev/null 2>&1; then
        echo "  [1/2] Dropping all tables via golang-migrate..."
        migrate -database "$db_url" -path "$MIGRATIONS_DIR" drop -f || true
        echo "  [2/2] Running all migrations and seed data..."
        migrate -database "$db_url" -path "$MIGRATIONS_DIR" up
    else
        echo "  Notice: 'migrate' command not found, falling back to psql / container execution..."
        local reset_sql="DROP SCHEMA public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO ${DB_USER};"
        
        if command -v psql >/dev/null 2>&1; then
            PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$target_db" -c "$reset_sql"
            for f in "$MIGRATIONS_DIR"/*.up.sql; do
                if [ -f "$f" ]; then
                    echo "  Applying $f..."
                    PGPASSWORD="$DB_PASS" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$target_db" -f "$f"
                fi
            done
        else
            # Try podman or docker
            local CONTAINER_BIN=""
            if command -v podman >/dev/null 2>&1; then
                CONTAINER_BIN="podman"
            elif command -v docker >/dev/null 2>&1; then
                CONTAINER_BIN="docker"
            fi

            if [ -n "$CONTAINER_BIN" ]; then
                $CONTAINER_BIN exec -i ottodot_trial_postgres psql -U "$DB_USER" -d "$target_db" -c "$reset_sql"
                for f in "$MIGRATIONS_DIR"/*.up.sql; do
                    if [ -f "$f" ]; then
                        echo "  Applying $f..."
                        $CONTAINER_BIN exec -i ottodot_trial_postgres psql -U "$DB_USER" -d "$target_db" < "$f"
                    fi
                done
            else
                echo "Error: Neither 'migrate', 'psql', nor 'docker/podman' was found." >&2
                exit 1
            fi
        fi
    fi

    echo "==> Successfully refreshed '${target_db}' with schema and seed data!"
}

case "$TARGET" in
    system)
        fresh_db "$DB_SYSTEM"
        ;;
    test)
        fresh_db "$DB_TEST"
        ;;
    all)
        fresh_db "$DB_SYSTEM"
        fresh_db "$DB_TEST"
        ;;
    *)
        echo "Usage: $0 [system|test|all]"
        echo "  system: Fresh development database ($DB_SYSTEM) [default]"
        echo "  test:   Fresh unit/integration test database ($DB_TEST)"
        echo "  all:    Fresh both databases"
        exit 1
        ;;
esac
