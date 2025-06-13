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

The `make ogen` command is the primary build command that:
1. Compiles TypeSpec definitions from `spec/main.tsp`
2. Generates OpenAPI 3.0/3.1 schemas in `spec/tsp-output/schema/`
3. Uses ogen to generate Go server code in `oas/` directory

### Running the Server
```bash
# Run the API server on port 8080
go run cmd/api-server/main.go
```

## Architecture Overview

This is a facility reservation API built with a **spec-first** approach using TypeSpec for API definition and ogen for Go code generation.

### Key Components

- **`spec/main.tsp`**: TypeSpec API specification defining all endpoints, models, and operations
- **`oas/`**: Auto-generated Go server code (handlers, schemas, validators) - DO NOT EDIT manually
- **`internal/service.go`**: Business logic implementation that embeds `oas.UnimplementedHandler`
- **`cmd/api-server/main.go`**: HTTP server entry point

### API Structure

The API provides three main endpoint groups:
- `/api/v1/admin/users/` - User management (admin only)
- `/api/v1/facilities/` - Facility CRUD operations 
- `/api/v1/me/` - Current user profile

### Development Workflow

1. **Schema Changes**: Modify `spec/main.tsp` (TypeSpec definitions)
2. **Code Generation**: Run `make ogen` to regenerate Go server code
3. **Business Logic**: Implement handlers in `internal/service.go`
4. **Testing**: Start server with `go run cmd/api-server/main.go`

### Important Notes

- All `oas/` files are generated - implement business logic only in `internal/` package
- The Service struct embeds `oas.UnimplementedHandler` to satisfy the interface
- TypeSpec compilation requires Node.js dependencies in `spec/` directory
- OpenAPI schemas are generated in both 3.0.0 and 3.1.0 formats
