#!/bin/bash
set -e

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"

SERVICES=("user" "product" "order" "payment")

for service in "${SERVICES[@]}"; do
    echo "Migrating $service..."

    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -tc \
        "SELECT 1 FROM pg_database WHERE datname = '$service'" | grep -q 1 || \
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -c \
        "CREATE DATABASE $service"

    for migration in services/$service/migrations/*.sql; do
        echo "  Applying $migration..."
        PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $service -f "$migration"
    done

    echo "  Done."
done

echo "All migrations applied successfully."
