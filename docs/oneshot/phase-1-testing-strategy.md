# Phase 1 Testing Strategy: Core Booking Engine

## Overview

Testing the booking engine requires validating **concurrent operations**, **data consistency**, and **user experience** under realistic conditions. The core challenge is ensuring zero booking conflicts while maintaining performance targets.

## Testing Philosophy

1. **Conflict Prevention First**: Test concurrent booking scenarios extensively
2. **Performance Under Load**: Validate targets with realistic user simulation
3. **User Experience Validation**: Measure actual booking completion times
4. **Edge Case Coverage**: Test unusual but possible scenarios
5. **Regression Prevention**: Comprehensive test suite for ongoing development

## Test Categories

### Unit Testing

**Database Layer Tests**
```go
// Test conflict detection at database level
func TestBookingConflictDetection(t *testing.T) {
    tests := []struct {
        name           string
        existingStart  time.Time
        existingEnd    time.Time
        newStart       time.Time
        newEnd         time.Time
        expectConflict bool
    }{
        {
            name:           "exact_overlap",
            existingStart:  parseTime("2025-06-15T10:00:00Z"),
            existingEnd:    parseTime("2025-06-15T11:00:00Z"),
            newStart:       parseTime("2025-06-15T10:00:00Z"),
            newEnd:         parseTime("2025-06-15T11:00:00Z"),
            expectConflict: true,
        },
        {
            name:           "partial_overlap_start",
            existingStart:  parseTime("2025-06-15T10:00:00Z"),
            existingEnd:    parseTime("2025-06-15T11:00:00Z"),
            newStart:       parseTime("2025-06-15T10:30:00Z"),
            newEnd:         parseTime("2025-06-15T11:30:00Z"),
            expectConflict: true,
        },
        {
            name:           "adjacent_no_conflict",
            existingStart:  parseTime("2025-06-15T10:00:00Z"),
            existingEnd:    parseTime("2025-06-15T11:00:00Z"),
            newStart:       parseTime("2025-06-15T11:00:00Z"),
            newEnd:         parseTime("2025-06-15T12:00:00Z"),
            expectConflict: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Create existing booking
            existing := createBooking(t, facilityID, userID, tt.existingStart, tt.existingEnd)
            
            // Attempt new booking
            result, err := bookingService.CreateBooking(ctx, CreateBookingParams{
                FacilityID: facilityID,
                UserID:     userID,
                StartTime:  tt.newStart,
                EndTime:    tt.newEnd,
                Title:      "Test Meeting",
            })

            if tt.expectConflict {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), "conflict")
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
            }
        })
    }
}
```

**Business Logic Tests**
```go
// Test availability calculation performance
func TestAvailabilityCalculationPerformance(t *testing.T) {
    // Setup: Create 100 existing bookings across next week
    setupTestBookings(t, 100)
    
    start := time.Now()
    
    // Search for availability across one week
    availability, err := bookingService.SearchAvailability(ctx, SearchParams{
        FacilityID: facilityID,
        StartTime:  time.Now(),
        EndTime:    time.Now().Add(7 * 24 * time.Hour),
        Duration:   time.Hour,
    })
    
    elapsed := time.Since(start)
    
    assert.NoError(t, err)
    assert.Less(t, elapsed, 100*time.Millisecond, "Availability search too slow")
    assert.NotEmpty(t, availability)
}
```

**Time Boundary Tests**
```go
// Test time granularity constraints
func TestTimeGranularityConstraints(t *testing.T) {
    tests := []struct {
        name        string
        startTime   string
        endTime     string
        expectError bool
    }{
        {
            name:        "valid_15min_boundary",
            startTime:   "2025-06-15T10:00:00Z",
            endTime:     "2025-06-15T10:15:00Z",
            expectError: false,
        },
        {
            name:        "invalid_minute_boundary",
            startTime:   "2025-06-15T10:07:00Z",
            endTime:     "2025-06-15T10:22:00Z",
            expectError: true,
        },
        {
            name:        "valid_hour_boundary",
            startTime:   "2025-06-15T10:00:00Z",
            endTime:     "2025-06-15T11:00:00Z",
            expectError: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            startTime := parseTime(tt.startTime)
            endTime := parseTime(tt.endTime)
            
            _, err := bookingService.CreateBooking(ctx, CreateBookingParams{
                FacilityID: facilityID,
                UserID:     userID,
                StartTime:  startTime,
                EndTime:    endTime,
                Title:      "Test Meeting",
            })

            if tt.expectError {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Integration Testing

**Concurrent Booking Tests**
```go
// Test race conditions with multiple simultaneous bookings
func TestConcurrentBookingAttempts(t *testing.T) {
    const numConcurrentUsers = 20
    const targetFacilityID = 1
    
    startTime := parseTime("2025-06-15T10:00:00Z")
    endTime := parseTime("2025-06-15T11:00:00Z")
    
    // Channel to collect results
    results := make(chan BookingResult, numConcurrentUsers)
    
    // Start all booking attempts simultaneously
    var wg sync.WaitGroup
    for i := 0; i < numConcurrentUsers; i++ {
        wg.Add(1)
        go func(userID int) {
            defer wg.Done()
            
            booking, err := bookingService.CreateBooking(ctx, CreateBookingParams{
                FacilityID: targetFacilityID,
                UserID:     userID,
                StartTime:  startTime,
                EndTime:    endTime,
                Title:      fmt.Sprintf("User %d Meeting", userID),
            })
            
            results <- BookingResult{
                UserID:  userID,
                Booking: booking,
                Error:   err,
            }
        }(i + 1)
    }
    
    wg.Wait()
    close(results)
    
    // Analyze results
    var successes, conflicts int
    for result := range results {
        if result.Error == nil {
            successes++
        } else if strings.Contains(result.Error.Error(), "conflict") {
            conflicts++
        } else {
            t.Errorf("Unexpected error: %v", result.Error)
        }
    }
    
    // Exactly one booking should succeed, rest should be conflicts
    assert.Equal(t, 1, successes, "Expected exactly one successful booking")
    assert.Equal(t, numConcurrentUsers-1, conflicts, "Expected all others to be conflicts")
    
    // Verify no actual conflicts in database
    bookings := getBookingsForTimeSlot(t, targetFacilityID, startTime, endTime)
    assert.Len(t, bookings, 1, "Database should contain exactly one booking")
}
```

**API Integration Tests**
```go
// Test complete booking flow via HTTP API
func TestBookingFlowHTTP(t *testing.T) {
    // 1. Search for availability
    availabilityResp := httpGet(t, "/api/v1/availability/search", map[string]string{
        "facility_id": "1",
        "start_time":  "2025-06-15T09:00:00Z",
        "end_time":    "2025-06-15T17:00:00Z",
        "duration":    "60",
    })
    
    require.Equal(t, http.StatusOK, availabilityResp.StatusCode)
    
    var availability []AvailabilitySlot
    json.Unmarshal(availabilityResp.Body, &availability)
    require.NotEmpty(t, availability)
    
    // 2. Create booking for first available slot
    bookingData := CreateBookingRequest{
        FacilityID:    1,
        StartTime:     availability[0].StartTime,
        EndTime:       availability[0].EndTime,
        Title:         "Test Meeting",
        AttendeeCount: 5,
    }
    
    start := time.Now()
    bookingResp := httpPost(t, "/api/v1/bookings/", bookingData)
    bookingDuration := time.Since(start)
    
    // Verify booking creation performance
    assert.Less(t, bookingDuration, time.Second, "Booking creation too slow")
    assert.Equal(t, http.StatusCreated, bookingResp.StatusCode)
    
    var booking BookingDetail
    json.Unmarshal(bookingResp.Body, &booking)
    
    // 3. Verify booking appears in user's bookings
    userBookingsResp := httpGet(t, "/api/v1/me/bookings/", nil)
    require.Equal(t, http.StatusOK, userBookingsResp.StatusCode)
    
    var userBookings []BookingSummary
    json.Unmarshal(userBookingsResp.Body, &userBookings)
    
    found := false
    for _, b := range userBookings {
        if b.ID == booking.ID {
            found = true
            break
        }
    }
    assert.True(t, found, "Booking should appear in user's booking list")
    
    // 4. Verify slot is no longer available
    availabilityResp2 := httpGet(t, "/api/v1/availability/search", map[string]string{
        "facility_id": "1",
        "start_time":  availability[0].StartTime.Format(time.RFC3339),
        "end_time":    availability[0].EndTime.Format(time.RFC3339),
    })
    
    var newAvailability []AvailabilitySlot
    json.Unmarshal(availabilityResp2.Body, &newAvailability)
    
    // Should not include the booked slot
    for _, slot := range newAvailability {
        assert.False(t, 
            slot.StartTime.Equal(availability[0].StartTime) && 
            slot.EndTime.Equal(availability[0].EndTime),
            "Booked slot should not appear in availability")
    }
}
```

### Load Testing

**Performance Test Suite**
```go
// Load test for concurrent availability searches
func TestAvailabilitySearchLoad(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping load test in short mode")
    }
    
    const (
        concurrentUsers = 50
        requestsPerUser = 10
        maxResponseTime = 200 * time.Millisecond
    )
    
    // Setup test data
    setupTestFacilities(t, 10)
    setupTestBookings(t, 200)
    
    results := make(chan LoadTestResult, concurrentUsers*requestsPerUser)
    
    var wg sync.WaitGroup
    for user := 0; user < concurrentUsers; user++ {
        wg.Add(1)
        go func(userID int) {
            defer wg.Done()
            
            for req := 0; req < requestsPerUser; req++ {
                start := time.Now()
                
                _, err := bookingService.SearchAvailability(ctx, SearchParams{
                    StartTime: time.Now().Add(time.Duration(req) * time.Hour),
                    EndTime:   time.Now().Add(time.Duration(req+8) * time.Hour),
                    Duration:  time.Hour,
                })
                
                elapsed := time.Since(start)
                
                results <- LoadTestResult{
                    UserID:       userID,
                    RequestID:    req,
                    Duration:     elapsed,
                    Error:        err,
                }
            }
        }(user)
    }
    
    wg.Wait()
    close(results)
    
    // Analyze performance
    var durations []time.Duration
    var errors int
    
    for result := range results {
        if result.Error != nil {
            errors++
            t.Logf("Error: %v", result.Error)
        } else {
            durations = append(durations, result.Duration)
        }
    }
    
    // Calculate percentiles
    sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
    
    p50 := durations[len(durations)/2]
    p95 := durations[int(float64(len(durations))*0.95)]
    p99 := durations[int(float64(len(durations))*0.99)]
    
    t.Logf("Performance Results:")
    t.Logf("  Total requests: %d", len(durations))
    t.Logf("  Errors: %d", errors)
    t.Logf("  P50: %v", p50)
    t.Logf("  P95: %v", p95)
    t.Logf("  P99: %v", p99)
    
    // Assert performance targets
    assert.Less(t, errors, 5, "Too many errors under load")
    assert.Less(t, p95, maxResponseTime, "95th percentile response time too high")
    assert.Less(t, p99, 500*time.Millisecond, "99th percentile response time too high")
}
```

### End-to-End Testing

**User Workflow Tests**
```javascript
// Cypress test for complete booking workflow
describe('Booking Workflow', () => {
    beforeEach(() => {
        cy.login('emma@company.com');
        cy.visit('/');
    });

    it('completes booking in under 60 seconds', () => {
        const startTime = Date.now();
        
        // Search for availability
        cy.get('[data-cy=duration-1hr]').click();
        cy.get('[data-cy=search-button]').click();
        
        // Wait for results and select first available slot
        cy.get('[data-cy=available-slot]').first().click();
        
        // Fill booking details
        cy.get('[data-cy=booking-title]').type('Client Strategy Meeting');
        cy.get('[data-cy=attendee-count]').select('5');
        
        // Confirm booking
        cy.get('[data-cy=confirm-booking]').click();
        
        // Verify success
        cy.get('[data-cy=booking-success]').should('be.visible');
        
        const endTime = Date.now();
        const duration = (endTime - startTime) / 1000;
        
        expect(duration).to.be.lessThan(60);
    });

    it('handles booking conflicts gracefully', () => {
        // Pre-book a room via API
        cy.createBooking({
            facility_id: 1,
            start_time: '2025-06-15T10:00:00Z',
            end_time: '2025-06-15T11:00:00Z',
            title: 'Existing Meeting'
        });
        
        // Try to book the same slot via UI
        cy.selectTimeSlot('2025-06-15T10:00:00Z', '2025-06-15T11:00:00Z');
        cy.get('[data-cy=booking-title]').type('Conflicting Meeting');
        cy.get('[data-cy=confirm-booking]').click();
        
        // Should show conflict resolution
        cy.get('[data-cy=conflict-modal]').should('be.visible');
        cy.get('[data-cy=suggested-alternatives]').should('not.be.empty');
        
        // Select alternative
        cy.get('[data-cy=alternative-slot]').first().click();
        cy.get('[data-cy=confirm-alternative]').click();
        
        // Should succeed with alternative
        cy.get('[data-cy=booking-success]').should('be.visible');
    });
});
```

**Mobile Responsiveness Tests**
```javascript
describe('Mobile Booking Experience', () => {
    beforeEach(() => {
        cy.viewport('iphone-x');
        cy.login('emma@company.com');
        cy.visit('/');
    });

    it('works efficiently on mobile', () => {
        // Test touch interactions
        cy.get('[data-cy=duration-1hr]').tap();
        cy.get('[data-cy=facility-card]').first().tap();
        
        // Test swipe navigation
        cy.get('[data-cy=calendar-view]').swipe('left');
        cy.get('[data-cy=next-day]').should('be.visible');
        
        // Verify mobile-specific UI elements
        cy.get('[data-cy=mobile-time-picker]').should('be.visible');
        cy.get('[data-cy=touch-friendly-buttons]').should('have.class', 'large');
    });
});
```

## Performance Testing

### Baseline Performance Targets

| Metric | Target | Measurement Method |
|--------|--------|--------------------|
| Availability Search | <100ms (P95) | Load testing with 50 concurrent users |
| Booking Creation | <200ms (P95) | API response time measurement |
| Conflict Detection | <50ms | Database query performance |
| UI Booking Flow | <60 seconds | End-to-end automation timing |

### Load Testing Scenarios

**Scenario 1: Peak Usage Simulation**
```bash
# Artillery.js load test configuration
config:
  target: 'http://localhost:8080'
  phases:
    - duration: 300  # 5 minutes ramp-up
      arrivalRate: 1
      rampTo: 50
    - duration: 600  # 10 minutes sustained load
      arrivalRate: 50
    - duration: 300  # 5 minutes ramp-down
      arrivalRate: 50
      rampTo: 1

scenarios:
  - name: "Booking workflow"
    weight: 70
    flow:
      - get:
          url: "/api/v1/availability/search"
          qs:
            facility_id: "{{ $randomInt(1, 10) }}"
            start_time: "{{ $randomFutureTime() }}"
            duration: "{{ $randomChoice([30, 60, 120]) }}"
      - post:
          url: "/api/v1/bookings/"
          json:
            facility_id: "{{ facility_id }}"
            start_time: "{{ selected_start_time }}"
            end_time: "{{ selected_end_time }}"
            title: "Load Test Meeting {{ $randomString() }}"
            
  - name: "Browse availability"
    weight: 30
    flow:
      - get:
          url: "/api/v1/availability/search"
          qs:
            start_time: "{{ $randomFutureTime() }}"
            end_time: "{{ $randomFutureTime(24) }}"
```

## Test Data Management

### Test Database Setup

```sql
-- Create test data for consistent testing
INSERT INTO facilities (name, location, capacity, is_active) VALUES
('Conference Room A', 'Floor 1', 8, true),
('Conference Room B', 'Floor 1', 12, true),
('Meeting Room C', 'Floor 2', 6, true),
('Training Room', 'Floor 3', 20, true),
('Phone Booth 1', 'Floor 1', 1, true),
('Phone Booth 2', 'Floor 1', 1, true),
('Auditorium', 'Ground Floor', 100, true),
('Kitchen', 'Floor 1', 15, true),
('Rooftop Terrace', 'Roof', 50, true),
('Disabled Room', 'Floor 2', 8, false); -- for testing inactive facilities

-- Create test users
INSERT INTO users (username, email, is_staff) VALUES
('emma.manager', 'emma@company.com', false),
('marcus.admin', 'marcus@company.com', true),
('sarah.assistant', 'sarah@company.com', false),
('john.developer', 'john@company.com', false),
('admin.user', 'admin@company.com', true);

-- Create some existing bookings for conflict testing
INSERT INTO bookings (facility_id, user_id, start_time, end_time, title) VALUES
(1, 1, '2025-06-15 09:00:00+00', '2025-06-15 10:00:00+00', 'Daily Standup'),
(1, 2, '2025-06-15 14:00:00+00', '2025-06-15 15:30:00+00', 'Project Review'),
(2, 3, '2025-06-15 11:00:00+00', '2025-06-15 12:00:00+00', 'Client Call'),
(3, 1, '2025-06-16 10:00:00+00', '2025-06-16 11:00:00+00', 'One-on-One');
```

### Test Data Cleanup

```go
// Clean up test data after each test
func teardownTest(t *testing.T) {
    // Delete in reverse dependency order
    db.Exec("DELETE FROM booking_audit_log WHERE changed_at >= $1", testStartTime)
    db.Exec("DELETE FROM bookings WHERE created_at >= $1", testStartTime)
    // Facilities and users are reused across tests
}
```

## Quality Gates

### Pre-Deployment Criteria

**Functional Quality Gates:**
- [ ] Zero booking conflicts in concurrent testing (1000+ attempts)
- [ ] All API endpoints respond within performance targets
- [ ] Complete booking workflow <60 seconds (P95)
- [ ] Mobile and desktop UI tests pass
- [ ] Accessibility tests pass (WCAG 2.1 AA)

**Performance Quality Gates:**
- [ ] Load test with 50 concurrent users passes
- [ ] Database queries optimized (no full table scans)
- [ ] Memory usage stable under load
- [ ] No memory leaks detected

**Security Quality Gates:**
- [ ] SQL injection tests pass
- [ ] Authorization tests pass (users can't access others' bookings)
- [ ] Input validation comprehensive
- [ ] Audit logging captures all changes

### Monitoring and Alerting

**Production Health Checks:**
```sql
-- Automated health check queries
SELECT 
    'booking_conflicts' as check_name,
    COUNT(*) as issues
FROM (
    SELECT b1.id FROM bookings b1, bookings b2 
    WHERE b1.id < b2.id 
    AND b1.facility_id = b2.facility_id
    AND b1.status = 'confirmed' AND b2.status = 'confirmed'
    AND tsrange(b1.start_time, b1.end_time) && tsrange(b2.start_time, b2.end_time)
) conflicts;

-- Alert if > 0
```

**Performance Monitoring:**
- API response time percentiles
- Database query performance
- Error rates by endpoint
- User completion rates

---

This comprehensive testing strategy ensures Phase 1 delivers a reliable, performant, and user-friendly booking system.