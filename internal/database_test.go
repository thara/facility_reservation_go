package internal

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDatabaseService(t *testing.T) {
	ctx := context.Background()

	t.Run("successful connection", func(t *testing.T) {
		// Use test database URL - requires PostgreSQL running
		databaseURL := getTestDatabaseURL()
		if databaseURL == "" {
			t.Skip("TEST_DATABASE_URL not set, skipping integration test")
		}

		ds, err := NewDatabaseService(ctx, databaseURL)
		require.NoError(t, err)
		require.NotNil(t, ds)
		defer ds.Close()

		// Verify queries interface is available
		assert.NotNil(t, ds.Queries())

		// Verify pool is available
		assert.NotNil(t, ds.Pool())
	})

	t.Run("invalid database URL", func(t *testing.T) {
		invalidURL := "invalid://url"

		ds, err := NewDatabaseService(ctx, invalidURL)
		require.Error(t, err)
		assert.Nil(t, ds)
		assert.Contains(t, err.Error(), "failed to parse database URL")
	})

	t.Run("connection failure", func(t *testing.T) {
		// Use non-existent database
		badURL := "postgres://user:pass@nonexistent:5432/db"

		ds, err := NewDatabaseService(ctx, badURL)
		require.Error(t, err)
		assert.Nil(t, ds)
		assert.Contains(t, err.Error(), "failed to ping database")
	})
}

func TestDatabaseService_HealthCheck(t *testing.T) {
	ctx := context.Background()
	databaseURL := getTestDatabaseURL()
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	ds, err := NewDatabaseService(ctx, databaseURL)
	require.NoError(t, err)
	defer ds.Close()

	t.Run("successful health check", func(t *testing.T) {
		err := ds.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	t.Run("health check with timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
		defer cancel()

		// Give context time to expire
		time.Sleep(1 * time.Millisecond)

		err := ds.HealthCheck(ctx)
		assert.Error(t, err)
	})
}

func TestDatabaseService_Close(t *testing.T) {
	ctx := context.Background()
	databaseURL := getTestDatabaseURL()
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	ds, err := NewDatabaseService(ctx, databaseURL)
	require.NoError(t, err)

	// Verify connection works before closing
	err = ds.HealthCheck(ctx)
	require.NoError(t, err)

	// Close the service
	ds.Close()

	// Verify connection no longer works after closing
	err = ds.HealthCheck(ctx)
	assert.Error(t, err)
}

func TestDatabaseService_ConnectionPoolConfiguration(t *testing.T) {
	ctx := context.Background()
	databaseURL := getTestDatabaseURL()
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	ds, err := NewDatabaseService(ctx, databaseURL)
	require.NoError(t, err)
	defer ds.Close()

	pool := ds.Pool()
	require.NotNil(t, pool)

	// Test that pool configuration is applied correctly
	config := pool.Config()
	assert.Equal(t, int32(25), config.MaxConns)
	assert.Equal(t, int32(5), config.MinConns)
	assert.Equal(t, time.Hour, config.MaxConnLifetime)
	assert.Equal(t, 30*time.Minute, config.MaxConnIdleTime)
}

// getTestDatabaseURL returns the test database URL from environment
// or empty string if not set
func getTestDatabaseURL() string {
	// Check for test-specific database URL
	// In CI/CD or local testing, set TEST_DATABASE_URL
	// Example: postgres://postgres:postgres@localhost:5433/facility_reservation_test?sslmode=disable
	return os.Getenv("TEST_DATABASE_URL")
}

// TestDatabaseService_Integration runs integration tests if database is available
func TestDatabaseService_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	databaseURL := getTestDatabaseURL()
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	t.Run("multiple connections", func(t *testing.T) {
		// Test that we can create multiple database services
		ds1, err := NewDatabaseService(ctx, databaseURL)
		require.NoError(t, err)
		defer ds1.Close()

		ds2, err := NewDatabaseService(ctx, databaseURL)
		require.NoError(t, err)
		defer ds2.Close()

		// Both should be able to ping successfully
		assert.NoError(t, ds1.HealthCheck(ctx))
		assert.NoError(t, ds2.HealthCheck(ctx))
	})

	t.Run("concurrent health checks", func(t *testing.T) {
		ds, err := NewDatabaseService(ctx, databaseURL)
		require.NoError(t, err)
		defer ds.Close()

		// Run multiple health checks concurrently
		done := make(chan error, 10)
		for i := 0; i < 10; i++ {
			go func() {
				done <- ds.HealthCheck(ctx)
			}()
		}

		// All should succeed
		for i := 0; i < 10; i++ {
			err := <-done
			assert.NoError(t, err)
		}
	})
}
