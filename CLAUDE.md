# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Code Quality Enforcement

**MANDATORY**: After making ANY code edits, you MUST run:
```bash
make build_dev
```

This command performs the complete development build pipeline:
- Clean build artifacts
- Format code (`make fmt`)
- Run linter checks (`make lint`) 
- Regenerate SQL code (`make sqlc-generate`)
- Regenerate API code (`make ogen`)
- Run all tests (`make test_all`)
- Build the binary

**This is not optional** - all code changes must pass the full build pipeline before being considered complete.

## Database Policy

**DO NOT define any custom functions and procedures in RDB.**

All business logic, UUID generation, and data processing must be handled in the application layer, not in the database.

## Documentation Policy

**DO NOT include production code in documentation files.**

Documentation should reference code locations and patterns, but not duplicate actual implementation. This prevents maintenance burden and keeps docs focused on concepts rather than implementation details.

## Essential Commands

### API Schema Generation
```bash
# Generate TypeSpec schema and Go server code
make ogen

# Generate TypeSpec schema only
make tsp
```

### Local Database Setup
```bash
# Start PostgreSQL with Docker Compose
make db-up

# Stop database
make db-down

# View database logs
make db-logs

# Clean database (removes all data)
make db-clean

# Setup database schema and generate Go code
make db-setup
```

### Database Operations
```bash
# Apply schema changes to database
make atlas-apply

# Generate Go code from SQL queries
make sqlc-generate

# Check current database schema status
make atlas-status

# Preview schema changes
make atlas-diff
```

### Testing
```bash
# Run unit tests (skip integration tests)
make test-short

# Run all tests (including skipped integration tests)
make test

# Run integration tests with test database
make test-integration
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
- **`_db/schema.sql`**: Database schema (source of truth) used by both Atlas and sqlc
- **`_db/query_*.sql`**: SQL queries for CRUD operations, compiled by sqlc to Go code
- **`api/`**: Auto-generated Go server code (handlers, schemas, validators) - DO NOT EDIT manually
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

1. **Database Schema**: Modify `_db/schema.sql` (declarative schema)
2. **Apply Schema**: Run `make atlas-apply` to update database
3. **SQL Queries**: Add/modify queries in `_db/query_*.sql`
4. **Generate Code**: Run `make sqlc-generate` to regenerate database code
5. **API Changes**: Modify `spec/main.tsp` if needed
6. **Server Code**: Run `make ogen` to regenerate HTTP handlers
7. **Business Logic**: Implement handlers in `internal/service.go` using generated database code
8. **Testing**: Start server with `go run cmd/api-server/main.go`

### Database Connection

The application uses PostgreSQL with Docker Compose for local development:
- **Development DB**: `postgres://postgres:postgres@localhost:5432/facility_reservation_dev?sslmode=disable`
- **Test DB**: `postgres://postgres:postgres@localhost:5433/facility_reservation_test?sslmode=disable`
- **Environment**: Set `DATABASE_URL` environment variable to override
- **CLI Flag**: Use `-database-url` command line argument

### Prerequisites
- Docker and Docker Compose installed
- Ports 5432 and 5433 available on localhost

## Unit Test Strategy

### Test Organization

Tests are organized using **external test packages** to enforce proper encapsulation:
- **Package Name**: Use `package internal_test` instead of `package internal`
- **Import Required**: Must import `"github.com/thara/facility_reservation_go/internal"` 
- **Access Control**: Can only access exported functions, types, and fields

### Testing Principles

1. **Test Behavior, Not Implementation**
   - Focus on public API and expected outcomes
   - Avoid testing internal state or implementation details
   - Test what the code does, not how it does it

2. **No Unexported Field Access**
   - Never access unexported struct fields (e.g., `service.db`, `db.pool`)
   - Use public methods to verify behavior instead
   - If you need to test internal state, consider exposing it through public methods

3. **Context Usage**
   - Use `t.Context()` instead of `context.Background()` in tests
   - Provides proper timeout handling and cancellation
   - Integrates with Go's testing framework

4. **Database Testing**
   - **Unit Tests**: Use `make test-short` - skip database integration
   - **Integration Tests**: Use `make test-integration` - requires real database
   - **Test Database**: Separate PostgreSQL instance on port 5433
   - **Schema Application**: Automatically applied before integration tests

### Testing Guidelines

- **DO NOT USE any unexported fields in tests**
- **Test only through exported functions and methods**
- **Use external test packages to enforce proper encapsulation**
- **Mock Dependencies**: Use nil or mock implementations for unit tests
- **Test Coverage**: Focus on business logic and error handling
- **Integration Tests**: Test with real database for database-dependent functionality
- **Concurrent Safety**: Use `for range 10` syntax for concurrent test loops
- **Error Wrapping**: Verify error messages contain expected context

### Important Notes

- **Schema as Code**: `_db/schema.sql` is the single source of truth for database structure
- **Type Safety**: sqlc generates type-safe Go structs and functions from SQL queries
- **Auto-generated Code**: Never edit files in `api/` or `internal/db/` directories
- **Atlas Migration**: Atlas automatically calculates and applies schema changes
- **Connection Pooling**: Uses pgx connection pool with configurable limits
- **Graceful Shutdown**: Database connections are properly closed on server shutdown
