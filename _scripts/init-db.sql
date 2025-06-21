-- Initialize development database
-- This script runs when the PostgreSQL container starts for the first time

-- Create main database if it doesn't exist
SELECT 'CREATE DATABASE facility_reservation_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'facility_reservation_db')\gexec

-- Enable UUID extension for the database
\c facility_reservation_db;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";