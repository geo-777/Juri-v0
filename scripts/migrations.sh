#!/usr/bin/env bash

set -euo pipefail

# Check if golang-migrate CLI is installed
if ! command -v migrate >/dev/null 2>&1; then
    echo "Error: golang-migrate CLI is not installed."
    exit 1
fi

# Ensure DATABASE_URL is available
if [ -z "${DATABASE_URL:-}" ]; then
    echo "DATABASE_URL is not set in the environment; loading .env..."
    if [ ! -f .env ]; then
        echo "Error: .env file not found."
        exit 1
    fi
    set -a
    source .env
    set +a
    if [ -z "${DATABASE_URL:-}" ]; then
        echo "Error: DATABASE_URL is missing or empty in .env."
        exit 1
    fi
fi

command="${1:-}"
arg="${2:-}"

case "$command" in
    up)
        migrate -path migrations -database "$DATABASE_URL" up
        ;;

    down)
        count="${arg:-1}"

        read -rp "Roll back $count migration(s)? [y/N] " confirm

        if [[ "$confirm" =~ ^[Yy]$ ]]; then
            migrate -path migrations -database "$DATABASE_URL" down "$count"
        else
            echo "Rollback cancelled."
        fi
        ;;

    create)
        if [ -z "$arg" ]; then
            echo "Error: Migration name is required."
            echo "Usage: ./migrate.sh create <migration_name>"
            exit 1
        fi

        migrate create -ext sql -dir migrations -seq "$arg"
        ;;

    force)
        if [ -z "$arg" ]; then
            echo "Error: Migration version is required."
            echo "Usage: ./migrate.sh force <version>"
            exit 1
        fi

        migrate -path migrations -database "$DATABASE_URL" force "$arg"
        ;;

    *)
        cat <<EOF
Usage:
  ./migrate.sh up
      Apply all pending migrations.

  ./migrate.sh down [count]
      Roll back the last migration (default: 1).

  ./migrate.sh create <migration_name>
      Create a new sequential migration.

  ./migrate.sh force <version>
      Force the migration version (use only to recover from a dirty state).
EOF
        exit 1
        ;;
esac