#!/bin/bash

# Check if schema.sql is up-to-date with the current database state
# This script can be used locally and in CI to ensure schema.sql matches the database

set -e

DATABASE_URL="${DATABASE_URL:-postgres://postgres:postgres@localhost:5432/facility_reservation_dev?sslmode=disable}"

echo "Checking if schema.sql is up-to-date..."

# Generate current schema to a temporary file
temp_schema=$(mktemp)

# Check if we're in CI (GitHub Actions) or local environment
if [ "$CI" = "true" ]; then
    # In CI, use pg_dump directly and filter out version comments
    pg_dump --schema-only --no-owner --no-privileges "$DATABASE_URL" | \
        grep -v -E "^-- Dumped (from database version|by pg_dump version)" > "$temp_schema"
else
    # Locally, use Docker Compose and filter out version comments
    docker compose exec -u postgres postgres pg_dump \
        --schema-only \
        --no-owner \
        --no-privileges \
        facility_reservation_dev | \
        grep -v -E "^-- Dumped (from database version|by pg_dump version)" > "$temp_schema"
fi

# Compare with existing schema.sql
if ! diff -q _db/schema.sql "$temp_schema" > /dev/null; then
    echo "❌ ERROR: schema.sql is not up-to-date!"
    echo ""
    echo "The current database schema differs from _db/schema.sql"
    echo "Please run 'make schema-generate' to update it."
    echo ""
    echo "Differences:"
    diff _db/schema.sql "$temp_schema" || true
    rm "$temp_schema"
    exit 1
fi

rm "$temp_schema"
echo "✅ schema.sql is up-to-date"