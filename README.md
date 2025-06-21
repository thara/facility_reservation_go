# Facility Reservation API

A Go-based facility reservation system built with a contract-first approach using modern tooling for type-safe development.

## Architecture

This project uses a **contract-first** architecture with two key boundaries that drive code generation:

- **Data Contract**: Database schema (`_db/schema.sql`) defines the data model
- **API Contract**: TypeSpec specification (`spec/main.tsp`) defines the interface

The architecture uses the following key technologies:

- **Atlas**: Database schema-as-code for migrations and schema management
- **sqlc**: Type-safe Go code generation from SQL queries
- **ogen**: HTTP server generation from TypeSpec API specifications
- **PostgreSQL**: Primary database with pgx driver for connection pooling

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+
- Ports 5432 and 5433 available on localhost

### Setup

1. Install development dependencies:
   ```bash
   make dev-deps
   ```

2. Start the database:
   ```bash
   make db-up
   ```

3. Setup database schema:
   ```bash
   make db-setup
   ```

4. Build and test:
   ```bash
   make build_dev
   ```

5. Run the server:
   ```bash
   go run cmd/api-server/main.go
   ```

## API Endpoints

The API provides three main endpoint groups:

- `/api/v1/admin/users/` - User management (admin only)
- `/api/v1/facilities/` - Facility CRUD operations
- `/api/v1/me/` - Current user profile

## Development Workflow

1. **Schema Changes**: Modify `_db/schema.sql`
2. **Apply Schema**: `make atlas-apply`
3. **Update Queries**: Edit `_db/query_*.sql`
4. **Generate Code**: `make sqlc-generate`
5. **API Changes**: Modify `spec/main.tsp`
6. **Update Server**: `make ogen`
7. **Implement Logic**: Update `internal/api_service.go`
8. **Test**: `make build_dev`

## Key Commands

### Database Operations
```bash
make db-up          # Start PostgreSQL
make db-down        # Stop database
make db-setup       # Setup schema
make atlas-apply    # Apply schema changes
make sqlc-generate  # Generate Go code from SQL
```

### Development
```bash
make build_dev      # Full development build pipeline
make test-short     # Unit tests only
make test-integration # Integration tests with database
make fmt           # Format code
make lint          # Run linter
```

### API Generation
```bash
make ogen          # Generate Go server from TypeSpec
make tsp           # Generate TypeSpec schema only
```

## Database Configuration

- **Development**: `postgres://postgres:postgres@localhost:5432/facility_reservation_db?sslmode=disable`
- **Test**: `postgres://postgres:postgres@localhost:5433/facility_reservation_db?sslmode=disable`
- **Custom**: Set `DATABASE_URL` environment variable or use `-database-url` flag

## Project Structure


```
_db/                    # Database schema and queries
├── schema.sql         # Database schema (source of truth)
└── query_*.sql        # SQL queries for CRUD operations
api/                   # Auto-generated HTTP server code
internal/
├── db/               # Auto-generated database code
├── api_service.go        # Business logic implementation
└── db_service.go       # Database service
spec/
└── main.tsp          # TypeSpec API specification
cmd/api-server/       # Server entry point
```


## Important Notes

- **Never edit** files in `api/` or `internal/db/` directories (auto-generated)
- **Database schema** in `_db/schema.sql` is the single source of truth
- **All code changes** must pass `make build_dev` before being considered complete
- **No database functions** - all business logic stays in the application layer