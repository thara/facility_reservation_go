# Phase 1 Technical Specification: Core Booking Engine

## Executive Summary

Phase 1 transforms our facility management foundation into a working booking system. The core challenge is **preventing booking conflicts in real-time** while maintaining the <60 second booking experience users demand.

**Key Technical Goals:**
- Zero booking conflicts at confirmation time
- Sub-200ms API response times for availability queries
- Database-level concurrency control
- Horizontal scalability for 50+ concurrent users

## Critical Technical Decisions

### Decision 1: Conflict Prevention Strategy

**Problem**: Two users booking the same time slot simultaneously  
**Solution**: Database-level pessimistic locking with optimistic UX

```sql
-- Atomic booking creation with conflict check
BEGIN;
SELECT facility_id FROM bookings 
WHERE facility_id = $1 
  AND (start_time, end_time) OVERLAPS ($2, $3)
FOR UPDATE NOWAIT;

-- If no conflicts, insert booking
INSERT INTO bookings (facility_id, user_id, start_time, end_time, title)
VALUES ($1, $4, $2, $3, $5);
COMMIT;
```

**Rationale**: 
- Database constraints prevent conflicts at the lowest level
- `FOR UPDATE NOWAIT` fails fast for concurrent attempts
- Application handles conflict gracefully with alternative suggestions

### Decision 2: Time Slot Granularity

**Choice**: 15-minute minimum booking slots  
**Rationale**:
- Balances flexibility with system performance
- Matches common meeting patterns (30min, 1hr, 1.5hr)
- Simplifies availability calculations and indexing
- Prevents "booking spam" with tiny reservations

**Implementation**:
```sql
-- Constraint to enforce 15-minute boundaries
ALTER TABLE bookings ADD CONSTRAINT booking_time_granularity 
CHECK (
  EXTRACT(minute FROM start_time) % 15 = 0 
  AND EXTRACT(minute FROM end_time) % 15 = 0
);
```

### Decision 3: Availability Calculation

**Choice**: Real-time calculation with intelligent caching  
**Alternative Considered**: Pre-computed availability tables  
**Rationale**:
- Real-time ensures accuracy without complex cache invalidation
- PostgreSQL interval operations are efficient with proper indexing
- Caching at application level for frequently accessed time ranges

**Performance Strategy**:
```sql
-- Optimized availability query
CREATE INDEX idx_bookings_facility_time ON bookings 
USING GIST (facility_id, tsrange(start_time, end_time));

-- Availability check query
SELECT NOT EXISTS (
  SELECT 1 FROM bookings 
  WHERE facility_id = $1 
    AND tsrange(start_time, end_time) && tsrange($2, $3)
);
```

### Decision 4: Real-time Updates

**Choice**: Server-Sent Events (SSE) for availability updates  
**Alternative Considered**: WebSockets, polling  
**Rationale**:
- SSE is simpler than WebSockets for one-way updates
- Better than polling for real-time user experience
- Falls back gracefully to polling if SSE unavailable

## Database Schema Design

### User Management Schema

```sql
-- Simple user table for token-based authentication
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    is_staff BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- API tokens for authentication
CREATE TABLE user_tokens (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Performance indexes for user management
CREATE INDEX idx_users_username ON users (username);
CREATE INDEX idx_user_tokens_token ON user_tokens (token);
CREATE INDEX idx_user_tokens_user ON user_tokens (user_id);
```

### Core Booking Entity

```sql
CREATE TABLE bookings (
    id SERIAL PRIMARY KEY,
    facility_id INTEGER NOT NULL REFERENCES facilities(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    
    -- Time management
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Booking details
    title VARCHAR(255) NOT NULL,
    description TEXT,
    attendee_count INTEGER DEFAULT 1,
    
    -- Status and metadata
    status booking_status DEFAULT 'confirmed',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Business constraints
    CONSTRAINT booking_time_valid CHECK (end_time > start_time),
    CONSTRAINT booking_time_granularity CHECK (
        EXTRACT(minute FROM start_time) % 15 = 0 
        AND EXTRACT(minute FROM end_time) % 15 = 0
    ),
    CONSTRAINT booking_duration_reasonable CHECK (
        end_time <= start_time + INTERVAL '8 hours'
    )
);

CREATE TYPE booking_status AS ENUM ('confirmed', 'cancelled', 'pending');
```

### Critical Indexes for Performance

```sql
-- Primary availability lookup
CREATE INDEX idx_bookings_facility_time ON bookings 
USING GIST (facility_id, tsrange(start_time, end_time));

-- User booking management
CREATE INDEX idx_bookings_user_time ON bookings (user_id, start_time);

-- Admin oversight
CREATE INDEX idx_bookings_status_time ON bookings (status, created_at);

-- Conflict prevention (unique constraint would be ideal but complex with time ranges)
CREATE UNIQUE INDEX idx_bookings_no_conflicts ON bookings 
(facility_id, start_time, end_time) 
WHERE status = 'confirmed';
```

### Audit and History

```sql
-- Track all booking changes for debugging and analytics
CREATE TABLE booking_audit (
    id SERIAL PRIMARY KEY,
    booking_id INTEGER REFERENCES bookings(id),
    action audit_action NOT NULL,
    old_values JSONB,
    new_values JSONB,
    changed_by INTEGER REFERENCES users(id),
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TYPE audit_action AS ENUM ('created', 'updated', 'cancelled');
```

## API Design Updates

### New TypeSpec Endpoints

```typescript
// Availability search
@route("/api/v1/availability/search")
@get
op availability_search(
  @query facility_id?: integer,
  @query start_time: utcDateTime,
  @query end_time: utcDateTime,
  @query duration?: integer  // minutes
): AvailabilityResult[] | ErrorResponse;

// Create booking
@route("/api/v1/bookings/")
@post  
op bookings_create(
  @body booking: CreateBookingRequest
): BookingDetail | ConflictResponse | ErrorResponse;

// Manage bookings
@route("/api/v1/bookings/{id}/")
@get
op bookings_retrieve(@path id: integer): BookingDetail | ErrorResponse;

@route("/api/v1/bookings/{id}/")
@patch
op bookings_update(
  @path id: integer,
  @body updates: UpdateBookingRequest
): BookingDetail | ConflictResponse | ErrorResponse;

@route("/api/v1/bookings/{id}/")
@delete
op bookings_cancel(@path id: integer): NoContentResponse | ErrorResponse;

// User's bookings
@route("/api/v1/me/bookings/")
@get
op my_bookings_list(
  @query start_date?: string,
  @query end_date?: string
): BookingSummary[] | ErrorResponse;
```

### New Data Models

```typescript
model BookingDetail {
  @visibility(Lifecycle.Read) id: integer;
  facility: FacilityBasic;
  user: UserBasic;
  start_time: utcDateTime;
  end_time: utcDateTime;
  title: string;
  description?: string;
  attendee_count: integer;
  status: BookingStatus;
  @visibility(Lifecycle.Read) created_at: utcDateTime;
  @visibility(Lifecycle.Read) updated_at: utcDateTime;
}

model CreateBookingRequest {
  facility_id: integer;
  start_time: utcDateTime;
  end_time: utcDateTime;
  title: string;
  description?: string;
  attendee_count?: integer;
}

model AvailabilityResult {
  facility: FacilityBasic;
  available_slots: TimeSlot[];
  next_available?: utcDateTime;
}

model TimeSlot {
  start_time: utcDateTime;
  end_time: utcDateTime;
  duration_minutes: integer;
}

enum BookingStatus {
  confirmed: "confirmed",
  cancelled: "cancelled", 
  pending: "pending"
}

// Conflict response for better UX
model ConflictResponse {
  @header("content-type") contentType: "application/problem+json";
  type: "booking_conflict";
  title: "Booking Conflict";
  status: 409;
  detail: string;
  conflicting_booking?: BookingBasic;
  suggested_alternatives?: TimeSlot[];
}
```

## Implementation Strategy

### Week 1-2: Database Foundation

**Day 1-3: Schema Implementation**
- Create booking tables with constraints
- Implement audit logging triggers
- Create performance indexes
- Database migration scripts

**Day 4-7: Conflict Prevention Core**
- Implement atomic booking creation
- Test concurrent booking scenarios
- Performance testing with simulated load
- Conflict detection and resolution logic

**Day 8-10: SQL Query Development**
- Availability search queries
- Booking CRUD operations
- User booking history queries
- Admin oversight queries

### Week 3-4: API Implementation

**Day 11-14: TypeSpec and Code Generation**
- Update TypeSpec with booking endpoints
- Generate Go handlers with ogen
- Implement business logic layer
- Error handling and validation

**Day 15-18: Availability Engine**
- Real-time availability calculation
- Caching strategy implementation
- Performance optimization
- Alternative suggestion algorithm

**Day 19-21: Booking Management**
- Create/update/cancel booking logic
- User permission checking
- Audit logging integration
- Conflict resolution workflows

### Week 5-6: User Interface

**Day 22-25: Availability Search UI**
- Calendar/grid view of availability
- Facility filtering and search
- Real-time availability updates
- Mobile-responsive design

**Day 26-29: Booking Creation Flow**
- One-click booking interface
- Booking form with validation
- Conflict resolution UI
- Success/error feedback

**Day 30-32: Booking Management UI**
- User booking dashboard
- Edit/cancel booking flows
- Admin oversight interface
- Booking history and analytics

### Week 7-8: Integration and Testing

**Day 33-36: End-to-End Testing**
- Automated booking flow tests
- Concurrent user testing
- Performance benchmarking
- Security testing

**Day 37-40: Performance Optimization**
- Database query optimization
- Caching implementation
- Load testing and tuning
- Monitoring and alerting setup

**Day 41-42: Launch Preparation**
- Documentation updates
- Deployment procedures
- Rollback plans
- User training materials

## Performance Requirements

### Database Performance Targets

| Operation | Target Time | Concurrent Users |
|-----------|-------------|------------------|
| Availability Search | <100ms | 50 |
| Booking Creation | <200ms | 20 |
| Conflict Detection | <50ms | 100 |
| User Booking List | <150ms | 50 |

### Optimization Strategies

**Database Level:**
- Partial indexes for active bookings only
- Connection pooling optimization (25 max, 10 min)
- Query plan analysis and optimization
- Automated index usage monitoring

**Application Level:**
- Redis caching for frequent availability queries
- Request deduplication for concurrent identical queries
- Async processing for non-critical operations
- Intelligent prefetching for likely user actions

**Network Level:**
- CDN for static assets
- Compression for API responses
- Keep-alive connections
- Regional deployment consideration

## Risk Mitigation

### High-Risk Areas

**Race Condition Handling**
- Risk: Simultaneous bookings creating conflicts
- Mitigation: Database-level locking with fast failure
- Testing: Automated concurrent booking tests
- Monitoring: Conflict rate metrics and alerting

**Performance Under Load**
- Risk: System slowdown with many concurrent users
- Mitigation: Comprehensive load testing and optimization
- Testing: Gradual load increase testing
- Monitoring: Response time percentiles and error rates

**Data Consistency**
- Risk: Availability cache inconsistency with bookings
- Mitigation: Event-driven cache invalidation
- Testing: Cache consistency verification tests
- Monitoring: Cache hit rates and invalidation patterns

### Medium-Risk Areas

**User Experience Complexity**
- Risk: Booking flow too complex for <60 second target
- Mitigation: User testing and flow optimization
- Testing: Automated UI flow timing tests
- Monitoring: User completion rates and drop-off points

**Database Migration Issues**
- Risk: Schema changes causing downtime
- Mitigation: Zero-downtime migration strategy
- Testing: Migration testing on production-like data
- Monitoring: Migration progress and rollback procedures

## Success Metrics and Monitoring

### Key Performance Indicators

**User Experience Metrics:**
- Average booking completion time (target: <60 seconds)
- Booking conflict rate (target: <1%)
- User adoption rate (target: 70% in 60 days)
- Booking success rate (target: >95%)

**Technical Performance Metrics:**
- API response times (95th percentile targets)
- Database query performance
- Concurrent user capacity
- System uptime (target: 99.5%)

**Business Metrics:**
- Facility utilization increase
- Admin time savings
- User satisfaction scores
- Support ticket reduction

### Monitoring Implementation

**Application Monitoring:**
- Custom metrics for booking flows
- Error tracking and alerting
- Performance dashboards
- User behavior analytics

**Infrastructure Monitoring:**
- Database performance metrics
- Connection pool usage
- Memory and CPU utilization
- Network latency and throughput

## Testing Strategy

### Unit Testing
- Business logic with mocked dependencies
- Conflict detection algorithms
- Availability calculation functions
- Data validation and constraints

### Integration Testing
- Database transaction scenarios
- API endpoint functionality
- Cache consistency verification
- Email notification delivery

### Load Testing
- Concurrent booking scenarios
- Database performance under load
- Cache effectiveness testing
- System resource utilization

### User Acceptance Testing
- Booking flow completion time
- Error handling and recovery
- Mobile and desktop compatibility
- Accessibility compliance

## Deployment Strategy

### Environment Progression
1. **Development**: Local testing with Docker Compose
2. **Staging**: Production-like environment for integration testing
3. **Production**: Gradual rollout with feature flags

### Feature Flag Strategy
- Booking functionality behind feature flags
- Gradual user group enrollment
- Quick rollback capability
- A/B testing for UI variations

### Monitoring and Rollback
- Real-time metrics monitoring
- Automated rollback triggers
- Manual rollback procedures
- Post-deployment validation tests

---

This technical specification provides the foundation for implementing Phase 1. Each section should be reviewed and approved by the technical team before implementation begins.