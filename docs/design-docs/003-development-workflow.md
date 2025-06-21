---
layout: default
title: Development Workflow
parent: Design Documents
nav_order: 3
---

# 3. Development Workflow

**Author**: System Architecture Team  
**Status**: Active  
**Last Updated**: 2025-06-14

## Goals

- Document the complete development workflow from schema changes to deployment
- Provide clear guidance for common development tasks
- Establish best practices for testing and validation
- Define the feedback loop for rapid development iteration

## Non-Goals

- CI/CD pipeline configuration (separate infrastructure concern)
- Production deployment strategies
- Performance profiling and optimization workflows
- Security testing and penetration testing procedures

## Overview

The development workflow is designed around the principle of **schema-driven development**, where changes to either the API specification (TypeSpec) or database schema (SQL) trigger a cascade of code regeneration, testing, and validation.

The workflow supports several common development scenarios:
1. **Adding new API endpoints**
2. **Modifying database schema**
3. **Adding business logic**
4. **Bug fixes and maintenance**

## Detailed Design

### Complete Development Cycle

```mermaid
graph TB
    subgraph "Development Setup"
        Setup[make db-setup<br/>Initial environment]
        Deps[make dev-deps<br/>Install tools]
    end
    
    subgraph "Schema Changes"
        APIChange[Modify spec/main.tsp]
        DBChange[Modify _db/schema.sql]
        QueryChange[Modify _db/query_*.sql]
    end
    
    subgraph "Code Generation"
        Generate[make build_dev<br/>Full regeneration]
        APIGen[make ogen<br/>API code only]
        DBGen[make sqlc-generate<br/>DB code only]
    end
    
    subgraph "Implementation"
        Business[Implement business logic<br/>internal/api_service.go]
        Tests[Write/update tests<br/>*_test.go]
    end
    
    subgraph "Validation"
        UnitTest[make test-short<br/>Unit tests]
        IntegrationTest[make test-integration<br/>Integration tests]
        ManualTest[Manual API testing]
    end
    
    subgraph "Commit"
        Review[Code review]
        Commit[git commit]
    end
    
    Setup --> Deps
    Deps --> APIChange
    Deps --> DBChange
    Deps --> QueryChange
    
    APIChange --> Generate
    DBChange --> Generate
    QueryChange --> Generate
    
    Generate --> Business
    Business --> Tests
    Tests --> UnitTest
    UnitTest --> IntegrationTest
    IntegrationTest --> ManualTest
    ManualTest --> Review
    Review --> Commit
```

### Common Development Scenarios

#### Scenario 1: Adding a New API Endpoint

**Step 1: Define API Contract**
```typescript
// spec/main.tsp
@tag("facilities")
@route("/api/v1/facilities/{id}/bookings/")
@get
@summary("List facility bookings")
op facilities_bookings_list(
  @path id: integer,
  @query date?: string
): FacilityBooking[] | ErrorResponse;
```

**Step 2: Generate Code**
```bash
make ogen  # Regenerate API handlers
```

**Step 3: Implement Business Logic**
```go
// internal/api_service.go
func (s *Service) FacilitiesBookingsList(ctx context.Context, req api.FacilitiesBookingsListRequest) (api.FacilitiesBookingsListResponse, error) {
    // Implementation using generated database queries
    bookings, err := s.db.Queries().ListFacilityBookings(ctx, s.db.Pool(), req.ID)
    if err != nil {
        return &api.FacilitiesBookingsListInternalServerError{}, err
    }
    
    return &api.FacilitiesBookingsListOK{
        Data: convertBookingsToAPI(bookings),
    }, nil
}
```

**Step 4: Test and Validate**
```bash
make test-short      # Unit tests
make test-integration # Integration tests with real DB
```

#### Scenario 2: Database Schema Evolution

**Step 1: Update Schema**
```sql
-- _db/schema.sql
CREATE TABLE bookings (
    id SERIAL PRIMARY KEY,
    facility_id INTEGER NOT NULL REFERENCES facilities(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

**Step 2: Add Queries**
```sql
-- _db/query_bookings.sql
-- name: ListFacilityBookings :many
SELECT id, facility_id, user_id, start_time, end_time, created_at, updated_at
FROM bookings
WHERE facility_id = $1
ORDER BY start_time ASC;

-- name: CreateBooking :one
INSERT INTO bookings (facility_id, user_id, start_time, end_time)
VALUES ($1, $2, $3, $4)
RETURNING id, facility_id, user_id, start_time, end_time, created_at, updated_at;
```

**Step 3: Apply Schema and Generate Code**
```bash
make atlas-apply      # Apply schema to database
make sqlc-generate    # Generate new Go database code
```

**Step 4: Update API Contract (if needed)**
```typescript
// spec/main.tsp
model FacilityBooking {
  @visibility(Lifecycle.Read) id: integer;
  facility_id: integer;
  user_id: integer;
  start_time: utcDateTime;
  end_time: utcDateTime;
  @visibility(Lifecycle.Read) created_at: utcDateTime;
  @visibility(Lifecycle.Read) updated_at: utcDateTime;
}
```

**Step 5: Full Build and Test**
```bash
make build_dev  # Complete pipeline: format, lint, generate, test, build
```

#### Scenario 3: Bug Fix Development

**Step 1: Reproduce Issue**
```bash
# Start local environment
make db-up
go run cmd/api-server/main.go

# Test the problematic endpoint
curl -X GET "http://localhost:8080/api/v1/facilities/999"
```

**Step 2: Write Failing Test**
```go
// internal/service_test.go
func TestFacilitiesRetrieve_NotFound(t *testing.T) {
    svc := internal.NewService(nil)
    
    resp, err := svc.FacilitiesRetrieve(context.Background(), api.FacilitiesRetrieveRequest{
        ID: 999, // Non-existent facility
    })
    
    // This test should initially fail, demonstrating the bug
    assert.Error(t, err)
    assert.IsType(t, &api.FacilitiesRetrieveNotFound{}, resp)
}
```

**Step 3: Fix Implementation**
```go
// internal/api_service.go
func (s *Service) FacilitiesRetrieve(ctx context.Context, req api.FacilitiesRetrieveRequest) (api.FacilitiesRetrieveResponse, error) {
    facility, err := s.db.Queries().GetFacilityByID(ctx, s.db.Pool(), req.ID)
    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return &api.FacilitiesRetrieveNotFound{}, nil  // Fix: return proper 404
        }
        return &api.FacilitiesRetrieveInternalServerError{}, err
    }
    
    return &api.FacilitiesRetrieveOK{
        Data: convertFacilityToAPI(facility),
    }, nil
}
```

**Step 4: Validate Fix**
```bash
make test-short      # Ensure test now passes
make build_dev       # Full validation
```

### Development Environment Management

#### Initial Setup
```bash
# First-time setup
git clone <repository>
cd facility_reservation_go

# Install development dependencies
make dev-deps

# Setup database and generate initial code
make db-setup
```

#### Daily Development Loop
```bash
# Start development session
make db-up           # Start PostgreSQL

# Make changes to code/schema...

# Validate changes
make build_dev       # Full build and test

# Optional: quick iteration
make ogen           # Just regenerate API code
make sqlc-generate  # Just regenerate DB code
make test-short     # Just unit tests
```

#### Database Management
```bash
# View current schema
make atlas-status

# Apply schema changes
make atlas-apply

# Reset database (destructive)
make db-clean
make db-setup
```

### Testing Strategy Integration

#### Test-Driven Development Flow
1. **Write failing test** for new functionality
2. **Run test** to confirm it fails (`make test-short`)
3. **Implement minimal code** to make test pass
4. **Refactor** while keeping tests green
5. **Run full test suite** (`make build_dev`)

#### Integration Testing Workflow
```bash
# Dedicated test database
make db-test-up                    # Start test PostgreSQL on port 5433
make atlas-apply-test              # Apply schema to test DB

# Run integration tests
make test-integration              # Uses TEST_DATABASE_URL

# Clean up
docker-compose down postgres-test  # Stop test database
```

## Alternatives Considered

### Alternative 1: Manual Code Management
**Pros**: Complete control over generated code  
**Cons**: Consistency issues, manual synchronization, prone to errors  
**Rejected**: Benefits of code generation outweigh control trade-offs

### Alternative 2: Watch-Mode Development
**Pros**: Automatic regeneration on file changes  
**Cons**: Complex setup, potential infinite loops, resource intensive  
**Rejected**: Manual control provides better debugging experience

### Alternative 3: Separate API and Database Workflows
**Pros**: Independent development of API and database concerns  
**Cons**: Integration complexity, potential for drift between layers  
**Rejected**: Unified workflow ensures consistency and reduces cognitive load

## Testing Strategy

### Development Testing Levels
1. **Unit Tests** (`make test-short`): Fast feedback on business logic
2. **Integration Tests** (`make test-integration`): Database interaction validation
3. **Build Tests** (`make build_dev`): Complete system validation
4. **Manual Testing**: API endpoint validation during development

### Continuous Validation
```bash
# Pre-commit validation
make build_dev  # Must pass before committing

# Quick iteration during development
make test-short && echo "✅ Unit tests passed"

# Database schema validation
make atlas-apply && echo "✅ Schema applied successfully"
```

## Future Considerations

### Potential Improvements
- **Watch Mode**: File watching for automatic regeneration during development
- **Development Containers**: Docker-based development environment
- **IDE Integration**: Better integration with editors for generated code
- **Performance Profiling**: Integration with profiling tools during development

### Workflow Optimization
- **Parallel Testing**: Run unit and integration tests in parallel
- **Incremental Generation**: Only regenerate changed components
- **Caching**: Cache generated artifacts for faster iteration
- **Hot Reload**: Server restart on code changes during development

### Known Limitations
- **Build Time**: Full regeneration can be slow for large schemas
- **Learning Curve**: New developers need to understand entire workflow
- **Tool Dependencies**: Requires specific versions of external tools
- **Error Correlation**: Errors may span multiple generation steps

### Scaling Considerations
As the project grows:
- **Modular Schemas**: Split large schemas into focused modules
- **Parallel Development**: Support for multiple developers working on different APIs
- **Integration Environments**: Staging environments for testing schema changes
- **Automated Testing**: More comprehensive test automation for complex workflows