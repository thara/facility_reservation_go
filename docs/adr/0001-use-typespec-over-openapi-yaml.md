# 1. Use TypeSpec over OpenAPI YAML

Date: 2025-06-14

## Status

Accepted

## Context

API-first development requires a specification format to define HTTP endpoints, request/response schemas, and validation rules. The two main approaches are:

1. **Hand-written OpenAPI YAML**: Direct authoring of OpenAPI v3 specifications
2. **TypeSpec**: Microsoft's type-safe API specification language that compiles to OpenAPI

Most teams choose hand-written YAML because:
- It's the "standard" approach that most developers know
- Direct control over the generated OpenAPI specification
- No additional compilation step required
- Extensive tooling ecosystem

However, we face specific challenges:
- Complex nested object schemas with strict validation requirements
- Need for type safety across API specification and implementation
- Desire for reusable type definitions and consistent patterns
- Team preference for compile-time error detection over runtime discovery

The facility reservation API has moderate complexity with user management, facility CRUD operations, and admin-specific endpoints that require careful schema design.

## Decision

We will use **TypeSpec** as our API specification language instead of hand-written OpenAPI YAML.

Rationale:
- **Type Safety**: TypeSpec provides compile-time validation of API specifications
- **Reusability**: Common types (like `EmailString`, `ProblemDetails`) can be defined once and reused
- **Better Developer Experience**: IDE support with autocomplete and error checking
- **Consistency**: Enforces consistent patterns across all API endpoints
- **Maintainability**: Changes to common types propagate automatically to all usage sites

We will use the TypeSpec → OpenAPI → ogen pipeline:
1. Define APIs in `spec/main.tsp`
2. Compile to OpenAPI with `tsp compile`
3. Generate Go server code with `ogen`

## Consequences

**Positive:**
- Compile-time validation catches specification errors early
- Strong typing prevents common API design mistakes
- Reusable components reduce duplication and ensure consistency
- IDE support improves specification authoring experience
- Generated OpenAPI is still available for external client generation

**Negative:**
- Additional learning curve for team members unfamiliar with TypeSpec
- Extra compilation step in the build process (`make tsp`)
- Smaller ecosystem compared to direct OpenAPI tooling
- Debugging requires understanding both TypeSpec and generated OpenAPI
- TypeSpec syntax is less familiar than YAML to most developers

**Risks:**
- TypeSpec is relatively new and evolving (though backed by Microsoft)
- Generated OpenAPI might not match exactly what we would write by hand
- Migration away from TypeSpec would require rewriting all specifications

**Mitigation:**
- Keep generated OpenAPI files in version control for inspection
- Document TypeSpec patterns and conventions in design docs
- Ensure the build pipeline clearly indicates TypeSpec compilation failures