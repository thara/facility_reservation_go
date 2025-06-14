# Phase 1 Database Design: Booking Engine Schema

## Overview

This document details the database schema design for the core booking functionality. The design prioritizes **conflict prevention**, **performance**, and **data integrity** while maintaining simplicity.

## Core Design Principles

1. **Conflict Prevention at Database Level**: Use constraints and locking to prevent booking conflicts
2. **Performance First**: Design indexes and queries for sub-200ms response times
3. **Audit Everything**: Track all booking changes for debugging and analytics
4. **Time Zone Awareness**: Handle multi-timezone organizations correctly
5. **Extensibility**: Design for future features (recurring bookings, approval workflows)

## Schema Design

### Booking Entity

```sql
-- Main booking table
CREATE TABLE bookings (
    -- Primary identification
    id SERIAL PRIMARY KEY,
    
    -- Foreign key relationships
    facility_id INTEGER NOT NULL REFERENCES facilities(id) ON DELETE RESTRICT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    
    -- Time management (always stored in UTC)
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Booking content
    title VARCHAR(255) NOT NULL,
    description TEXT,
    attendee_count INTEGER NOT NULL DEFAULT 1,
    external_meeting_url VARCHAR(500), -- Zoom, Teams, etc.
    
    -- Status management
    status booking_status NOT NULL DEFAULT 'confirmed',
    cancelled_at TIMESTAMP WITH TIME ZONE,
    cancelled_by INTEGER REFERENCES users(id),
    cancellation_reason TEXT,
    
    -- Audit fields
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by INTEGER REFERENCES users(id), -- for proxy bookings
    
    -- Business rules as database constraints
    CONSTRAINT booking_time_valid 
        CHECK (end_time > start_time),
    
    CONSTRAINT booking_time_granularity 
        CHECK (
            EXTRACT(minute FROM start_time) % 15 = 0 
            AND EXTRACT(minute FROM end_time) % 15 = 0
        ),
    
    CONSTRAINT booking_duration_reasonable 
        CHECK (end_time <= start_time + INTERVAL '24 hours'),
    
    CONSTRAINT booking_not_in_past 
        CHECK (start_time >= DATE_TRUNC('day', NOW() AT TIME ZONE 'UTC')),
    
    CONSTRAINT booking_attendee_count_valid 
        CHECK (attendee_count > 0 AND attendee_count <= 1000),
    
    CONSTRAINT booking_cancellation_consistency 
        CHECK (
            (status = 'cancelled' AND cancelled_at IS NOT NULL AND cancelled_by IS NOT NULL)
            OR (status != 'cancelled' AND cancelled_at IS NULL AND cancelled_by IS NULL)
        )
);

-- Booking status enum
CREATE TYPE booking_status AS ENUM (
    'confirmed',    -- Normal active booking
    'cancelled',    -- User or admin cancelled
    'pending'       -- Future: approval workflow
);
```

### Conflict Prevention Strategy

```sql
-- Prevent overlapping bookings for the same facility
-- This is the critical constraint that prevents double-booking
CREATE UNIQUE INDEX idx_bookings_no_conflicts 
ON bookings (facility_id, start_time, end_time) 
WHERE status IN ('confirmed', 'pending');

-- Alternative approach using exclusion constraint (PostgreSQL specific)
-- More flexible but potentially slower
-- ALTER TABLE bookings ADD CONSTRAINT booking_no_overlap 
-- EXCLUDE USING GIST (
--     facility_id WITH =,
--     tsrange(start_time, end_time) WITH &&
-- ) WHERE (status IN ('confirmed', 'pending'));
```

### Performance Indexes

```sql
-- Primary availability lookup - most critical for performance
CREATE INDEX idx_bookings_facility_time_range 
ON bookings USING GIST (facility_id, tsrange(start_time, end_time))
WHERE status = 'confirmed';

-- User booking management
CREATE INDEX idx_bookings_user_active 
ON bookings (user_id, start_time DESC) 
WHERE status = 'confirmed';

-- Admin oversight and reporting
CREATE INDEX idx_bookings_created_at 
ON bookings (created_at DESC);

-- Status-based queries
CREATE INDEX idx_bookings_status_time 
ON bookings (status, start_time)
WHERE status != 'cancelled';

-- Facility utilization reporting
CREATE INDEX idx_bookings_facility_date 
ON bookings (facility_id, DATE(start_time AT TIME ZONE 'UTC'))
WHERE status = 'confirmed';
```

### Audit and History Tracking

```sql
-- Comprehensive audit log for all booking changes
CREATE TABLE booking_audit_log (
    id SERIAL PRIMARY KEY,
    booking_id INTEGER NOT NULL, -- Don't use foreign key - keep history even if booking deleted
    
    -- What changed
    action audit_action NOT NULL,
    old_values JSONB,
    new_values JSONB,
    
    -- Who and when
    changed_by INTEGER REFERENCES users(id),
    changed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    -- Request context
    user_agent TEXT,
    ip_address INET,
    request_id UUID -- for request tracing
);

CREATE TYPE audit_action AS ENUM (
    'created',
    'updated', 
    'cancelled',
    'admin_cancelled'
);

-- Index for audit queries
CREATE INDEX idx_booking_audit_booking_id ON booking_audit_log (booking_id, changed_at DESC);
CREATE INDEX idx_booking_audit_user ON booking_audit_log (changed_by, changed_at DESC);
CREATE INDEX idx_booking_audit_action ON booking_audit_log (action, changed_at DESC);
```

### Automatic Timestamp Updates

```sql
-- Trigger to automatically update updated_at (explicit, not using database triggers per ADR-003)
-- Instead, we'll handle this in application code with explicit updated_at = NOW()
-- This comment serves as a reminder of the architectural decision
```

## Critical Queries and Performance

### Availability Search Query

```sql
-- Core availability check - must be fast
WITH requested_slots AS (
    SELECT 
        generate_series(
            $2::timestamp with time zone,  -- start_time
            $3::timestamp with time zone - INTERVAL '15 minutes', -- end_time - slot duration
            INTERVAL '15 minutes'
        ) AS slot_start
),
available_slots AS (
    SELECT 
        rs.slot_start,
        rs.slot_start + INTERVAL '15 minutes' AS slot_end
    FROM requested_slots rs
    WHERE NOT EXISTS (
        SELECT 1 FROM bookings b
        WHERE b.facility_id = $1  -- facility_id
        AND b.status = 'confirmed'
        AND tsrange(b.start_time, b.end_time) && 
            tsrange(rs.slot_start, rs.slot_start + INTERVAL '15 minutes')
    )
)
SELECT 
    slot_start,
    slot_end,
    EXTRACT(EPOCH FROM (slot_end - slot_start))/60 AS duration_minutes
FROM available_slots
ORDER BY slot_start;
```

**Expected Performance**: <50ms for 1-day search, <100ms for 1-week search

### Booking Creation with Conflict Check

```sql
-- Atomic booking creation - prevents race conditions
WITH conflict_check AS (
    SELECT COUNT(*) as conflicts
    FROM bookings 
    WHERE facility_id = $1
    AND status IN ('confirmed', 'pending')
    AND tsrange(start_time, end_time) && tsrange($3, $4) -- start_time, end_time overlap
),
booking_insert AS (
    INSERT INTO bookings (facility_id, user_id, start_time, end_time, title, description, attendee_count)
    SELECT $1, $2, $3, $4, $5, $6, $7
    WHERE (SELECT conflicts FROM conflict_check) = 0
    RETURNING *
)
SELECT 
    bi.*,
    CASE WHEN bi.id IS NULL THEN (SELECT conflicts FROM conflict_check) ELSE 0 END as conflict_count
FROM booking_insert bi
FULL OUTER JOIN conflict_check cc ON true;
```

**Expected Performance**: <100ms including conflict detection

### User Booking Management

```sql
-- Get user's upcoming bookings
SELECT 
    b.id,
    b.start_time,
    b.end_time,
    b.title,
    b.status,
    f.name as facility_name,
    f.location as facility_location
FROM bookings b
JOIN facilities f ON b.facility_id = f.id
WHERE b.user_id = $1
AND b.status = 'confirmed'
AND b.start_time >= NOW()
ORDER BY b.start_time
LIMIT 50;
```

**Expected Performance**: <50ms

### Facility Utilization Report

```sql
-- Daily utilization report for admins
WITH facility_hours AS (
    SELECT 
        f.id,
        f.name,
        -- Assume 8 hours per day available (8 AM - 6 PM)
        8 * 60 AS available_minutes_per_day
    FROM facilities f
    WHERE f.is_active = true
),
daily_usage AS (
    SELECT 
        b.facility_id,
        DATE(b.start_time AT TIME ZONE 'UTC') as booking_date,
        SUM(EXTRACT(EPOCH FROM (b.end_time - b.start_time))/60) as booked_minutes
    FROM bookings b
    WHERE b.status = 'confirmed'
    AND b.start_time >= $1  -- start_date
    AND b.start_time <= $2  -- end_date
    GROUP BY b.facility_id, DATE(b.start_time AT TIME ZONE 'UTC')
)
SELECT 
    fh.name as facility_name,
    du.booking_date,
    du.booked_minutes,
    fh.available_minutes_per_day,
    ROUND((du.booked_minutes::numeric / fh.available_minutes_per_day) * 100, 2) as utilization_percentage
FROM facility_hours fh
LEFT JOIN daily_usage du ON fh.id = du.facility_id
ORDER BY du.booking_date DESC, fh.name;
```

**Expected Performance**: <200ms for 30-day reports

## Data Migration Strategy

### Migration Steps

1. **Create new tables** without constraints
2. **Add indexes** for performance
3. **Add constraints** after data validation
4. **Create audit logging** infrastructure
5. **Test conflict prevention** thoroughly

### Sample Data for Testing

```sql
-- Insert sample bookings for testing
INSERT INTO bookings (facility_id, user_id, start_time, end_time, title, attendee_count) VALUES
(1, 1, '2025-06-15 09:00:00+00', '2025-06-15 10:00:00+00', 'Team Standup', 5),
(1, 2, '2025-06-15 10:30:00+00', '2025-06-15 11:30:00+00', 'Client Call', 3),
(2, 1, '2025-06-15 14:00:00+00', '2025-06-15 15:00:00+00', 'Project Review', 8),
-- Add more test data for various scenarios
(1, 3, '2025-06-16 09:15:00+00', '2025-06-16 10:15:00+00', 'One-on-One', 2);
```

## Monitoring and Alerting

### Key Metrics to Track

**Conflict Detection:**
```sql
-- Monitor booking conflicts (should be near zero)
SELECT 
    DATE(created_at) as date,
    COUNT(*) as total_bookings,
    COUNT(*) FILTER (WHERE id NOT IN (
        SELECT DISTINCT b1.id 
        FROM bookings b1, bookings b2 
        WHERE b1.id != b2.id 
        AND b1.facility_id = b2.facility_id
        AND b1.status = 'confirmed' 
        AND b2.status = 'confirmed'
        AND tsrange(b1.start_time, b1.end_time) && tsrange(b2.start_time, b2.end_time)
    )) as conflict_free_bookings
FROM bookings 
WHERE created_at >= NOW() - INTERVAL '7 days'
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

**Performance Monitoring:**
- Index usage statistics
- Query execution times
- Connection pool utilization
- Lock wait times

### Database Health Checks

```sql
-- Check for database health issues
SELECT 
    'booking_conflicts' as check_name,
    CASE WHEN COUNT(*) = 0 THEN 'PASS' ELSE 'FAIL' END as status,
    COUNT(*) as conflict_count
FROM (
    SELECT b1.id 
    FROM bookings b1, bookings b2 
    WHERE b1.id != b2.id 
    AND b1.facility_id = b2.facility_id
    AND b1.status = 'confirmed' 
    AND b2.status = 'confirmed'
    AND tsrange(b1.start_time, b1.end_time) && tsrange(b2.start_time, b2.end_time)
) conflicts

UNION ALL

SELECT 
    'index_usage' as check_name,
    CASE WHEN idx_scan > seq_scan THEN 'PASS' ELSE 'WARN' END as status,
    idx_scan - seq_scan as index_preference
FROM pg_stat_user_tables 
WHERE relname = 'bookings';
```

## Security Considerations

### Data Protection

- **PII Handling**: Booking titles and descriptions may contain sensitive information
- **Access Control**: Users can only modify their own bookings (except admins)
- **Audit Logging**: All changes tracked with user attribution
- **Data Retention**: Define retention policies for cancelled bookings

### Query Security

```sql
-- Example of secure booking query with user authorization
SELECT b.* 
FROM bookings b
WHERE b.id = $1 
AND (
    b.user_id = $2  -- User can access their own bookings
    OR EXISTS (     -- Or user is admin
        SELECT 1 FROM users u 
        WHERE u.id = $2 AND u.is_staff = true
    )
);
```

## Future Considerations

### Extensibility Design

The schema is designed to support future enhancements:

1. **Recurring Bookings**: Add `recurring_pattern` JSONB column
2. **Approval Workflows**: Expand `booking_status` enum
3. **Resource Requirements**: Add `resources` JSONB for equipment needs
4. **Multi-tenant**: Add `organization_id` for multi-tenant deployment
5. **Time Zone Support**: Add `timezone` field for display purposes

### Scaling Considerations

- **Partitioning**: Partition by date when booking volume grows
- **Read Replicas**: Separate reporting queries from transactional load
- **Archiving**: Move old bookings to archive tables
- **Caching**: Add Redis layer for high-frequency availability queries

---

This database design provides a solid foundation for the Phase 1 booking engine while maintaining flexibility for future enhancements.