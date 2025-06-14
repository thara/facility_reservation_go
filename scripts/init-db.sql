-- Initialize development database
-- This script runs when the PostgreSQL container starts for the first time

-- Create test database if it doesn't exist
SELECT 'CREATE DATABASE facility_reservation_test'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'facility_reservation_test')\gexec

-- Enable UUID extension for both databases
\c facility_reservation_dev;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\c facility_reservation_test;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";