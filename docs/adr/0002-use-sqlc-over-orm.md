---
layout: default
parent: Architecture Decision Records
nav_order: 2
title: "2. Use sqlc over ORM"
---

# 2. Use sqlc over ORM

Date: 2025-06-14

## Status

Accepted

## Context

Database access in Go applications typically follows one of several patterns:

1. **Traditional ORMs** (GORM, Ent): Object-relational mapping with automatic query generation
2. **Query Builders** (Squirrel): Programmatic SQL construction with some type safety
3. **Raw SQL with manual mapping**: Direct SQL with manual scanning to structs
4. **Code generators** (sqlc): Write SQL, generate type-safe Go code

Our application needs:
- Complex queries for facility filtering and user management
- Strong type safety to prevent runtime database errors
- Good performance for API endpoints
- Clear visibility into what SQL is being executed
- Easy testing with predictable query patterns

ORMs like GORM are popular because they:
- Provide automatic CRUD operations
- Handle relationships and eager loading
- Include migration capabilities
- Offer a familiar Active Record pattern

However, our team values:
- Explicit control over SQL queries
- Compile-time type safety
- Performance predictability
- Clear mapping between business operations and database queries

## Decision

We will use **sqlc** for database access instead of a traditional ORM.

Implementation approach:
- Write explicit SQL queries in `_db/query_*.sql` files
- Use sqlc to generate type-safe Go functions
- Combine with Atlas for schema management
- Maintain clear separation between SQL (data access) and Go (business logic)

Query organization:
- Group related queries by domain (`query_users.sql`, `query_facilities.sql`)
- Use sqlc annotations for type safety (`-- name: GetUserByID :one`)
- Explicit parameter binding and result scanning

## Decision

**Rationale:**
- **Performance**: Direct SQL control allows optimization for specific use cases
- **Type Safety**: sqlc generates compile-time safe database interfaces
- **Clarity**: Business logic can see exactly what SQL operations are performed
- **Simplicity**: No ORM magic or hidden query generation
- **Debugging**: Easy to understand and debug actual SQL being executed

## Consequences

**Positive:**
- Complete control over SQL queries and performance characteristics
- Compile-time type safety prevents common database errors
- Generated code is readable and follows Go conventions
- No impedance mismatch between object model and relational model
- Easy to optimize queries for specific performance requirements
- Clear separation between data access (SQL) and business logic (Go)

**Negative:**
- More verbose than ORM for simple CRUD operations
- Manual SQL writing requires more database knowledge from developers
- No automatic relationship handling or eager loading
- Schema changes require manual query updates
- Less abstraction means more code to maintain

**Tradeoffs accepted:**
- Verbosity in exchange for explicit control and performance
- Manual query maintenance in exchange for compile-time safety
- Learning curve for SQL in exchange for better debugging capability

**Implementation requirements:**
- Establish consistent SQL query patterns and naming conventions
- Document common query patterns for team reference
- Include SQL knowledge requirements in developer onboarding
- Use Atlas for schema migrations to complement sqlc query generation

**Monitoring:**
- Track query performance to validate performance benefits
- Monitor for SQL injection vulnerabilities (mitigated by parameterized queries)
- Ensure generated code quality meets team standards