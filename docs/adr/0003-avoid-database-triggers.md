# 3. Avoid Database Triggers

Date: 2025-06-14

## Status

Accepted

## Context

Database triggers are a common pattern for handling cross-cutting concerns like:
- Automatic timestamp updates (`updated_at` fields)
- Audit logging and change tracking
- Data validation and business rule enforcement
- Automatic calculations and derived fields

Many teams use triggers because they:
- Ensure consistency regardless of how data is modified
- Centralize business rules at the database level
- Reduce application code complexity
- Provide guarantees even for direct database modifications

However, triggers create implicit behavior that can be problematic:
- Hidden side effects that are not visible in application code
- Difficult to debug when triggers have complex logic
- Hard to test trigger behavior in isolation
- Performance impact that's not obvious from application code
- Complex interactions between multiple triggers

Our team values explicit, traceable behavior where:
- All business logic is visible in the application codebase
- Data changes can be tracked through application logs
- Testing can control and verify all side effects
- Debugging follows clear execution paths

The specific case that prompted this decision:
- We had triggers for automatic `updated_at` timestamp updates
- This created hidden behavior where timestamps changed without explicit application control
- Testing and debugging timestamp behavior was unnecessarily complex

## Decision

We will **avoid database triggers** and implement equivalent functionality explicitly in application code.

For timestamp management specifically:
- Remove automatic `updated_at` triggers
- Add explicit `updated_at = NOW()` to all UPDATE queries
- Make timestamp updates visible in application code and logs

General policy:
- Database constraints for data integrity only (foreign keys, check constraints, etc.)
- All business logic and side effects in application code
- Explicit over implicit behavior in all cases

## Consequences

**Positive:**
- **Debuggability**: All data changes are traceable in application code and logs
- **Testability**: Full control over side effects in unit and integration tests
- **Code Clarity**: No hidden database behavior affecting application state
- **Maintainability**: All business logic is visible and searchable in the codebase
- **Performance Transparency**: No hidden database operations affecting query performance

**Negative:**
- **Verbosity**: Must explicitly handle concerns that triggers would handle automatically
- **Consistency Risk**: Application code must remember to handle these concerns
- **Duplication**: Similar logic may need to be repeated across multiple operations
- **Direct Database Changes**: No protection if data is modified outside the application

**Implementation Requirements:**
- Update all UPDATE queries to explicitly set `updated_at = NOW()`
- Establish code review practices to ensure explicit handling of cross-cutting concerns
- Document patterns for handling common concerns (timestamps, audit logging, etc.)
- Consider application-level abstractions for repeated patterns

**Accepted Risks:**
- Developers might forget to update timestamps in new queries (mitigated by code review)
- Direct database modifications won't have automatic timestamp updates (acceptable tradeoff)
- More verbose SQL queries (acceptable for clarity benefits)

**Monitoring:**
- Ensure all UPDATE operations include timestamp updates through code review
- Consider adding tests that verify timestamp behavior
- Document the timestamp update pattern for new developers