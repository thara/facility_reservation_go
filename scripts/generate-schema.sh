#!/bin/bash

# Generate schema.sql from current database state
# This script uses Docker to run pg_dump against the development database

set -e

echo "Generating schema.sql from database..."

# Generate schema from development database and filter out version comments
docker compose exec -u postgres postgres pg_dump \
    --schema-only \
    --no-owner \
    --no-privileges \
    facility_reservation_dev | \
    grep -v -E "^-- Dumped (from database version|by pg_dump version)" > _db/schema.sql

echo "Schema generated successfully at _db/schema.sql"