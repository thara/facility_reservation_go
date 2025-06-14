# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Essential Commands

### API Schema Generation
```bash
# Generate TypeSpec schema and Go server code
make ogen

# Generate TypeSpec schema only
make tsp
```

### Database Operations
```bash
# Setup database schema and generate Go code
make db-setup

# Apply schema changes to database
make atlas-apply

# Generate Go code from SQL queries
make sqlc-generate

# Check current database schema status
make atlas-status

# Preview schema changes
make atlas-diff
```

### Development Tools Installation
```bash
# Install development dependencies (Atlas, sqlc)
make dev-deps
```

### Running the Server
```bash
# Run with default database (requires PostgreSQL)
go run cmd/api-server/main.go

# Run with custom database URL
go run cmd/api-server/main.go -database-url="postgres://user:pass@host:port/dbname"

# Run on custom port
go run cmd/api-server/main.go -addr=":3000"
```

## Architecture Overview

This is a facility reservation API built with a **database-first** approach using Atlas for schema management, sqlc for type-safe database access, and ogen for HTTP server generation.

### Key Components

- **`spec/main.tsp`**: TypeSpec API specification defining all endpoints, models, and operations
- **`schema/schema.sql`**: Database schema (source of truth) used by both Atlas and sqlc
- **`queries/*.sql`**: SQL queries for CRUD operations, compiled by sqlc to Go code
- **`oas/`**: Auto-generated Go server code (handlers, schemas, validators) - DO NOT EDIT manually
- **`internal/db/`**: Auto-generated Go database code from sqlc - DO NOT EDIT manually
- **`internal/service.go`**: Business logic implementation with database integration
- **`internal/database.go`**: Database service with connection pooling
- **`cmd/api-server/main.go`**: HTTP server entry point with database initialization

### Database Tools

- **Atlas**: Database schema-as-code tool for migrations and schema management
- **sqlc**: Generates type-safe Go code from SQL queries
- **PostgreSQL**: Primary database with pgx driver for connection pooling

### API Structure

The API provides three main endpoint groups:
- `/api/v1/admin/users/` - User management (admin only)
- `/api/v1/facilities/` - Facility CRUD operations 
- `/api/v1/me/` - Current user profile

### Development Workflow

1. **Database Schema**: Modify `schema/schema.sql` (declarative schema)
2. **Apply Schema**: Run `make atlas-apply` to update database
3. **SQL Queries**: Add/modify queries in `queries/*.sql`
4. **Generate Code**: Run `make sqlc-generate` to regenerate database code
5. **API Changes**: Modify `spec/main.tsp` if needed
6. **Server Code**: Run `make ogen` to regenerate HTTP handlers
7. **Business Logic**: Implement handlers in `internal/service.go` using generated database code
8. **Testing**: Start server with `go run cmd/api-server/main.go`

### Database Connection

The application requires PostgreSQL and uses these connection methods:
- **Default**: `postgres://postgres:postgres@localhost:5432/facility_reservation_dev?sslmode=disable`
- **Environment**: Set `DATABASE_URL` environment variable
- **CLI Flag**: Use `-database-url` command line argument

### Important Notes

- **Schema as Code**: `schema/schema.sql` is the single source of truth for database structure
- **Type Safety**: sqlc generates type-safe Go structs and functions from SQL queries
- **Auto-generated Code**: Never edit files in `oas/` or `internal/db/` directories
- **Atlas Migration**: Atlas automatically calculates and applies schema changes
- **Connection Pooling**: Uses pgx connection pool with configurable limits
- **Graceful Shutdown**: Database connections are properly closed on server shutdown
