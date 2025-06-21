package internal_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thara/facility_reservation_go/internal"
	"github.com/thara/facility_reservation_go/internal/db"
)

func TestNewDBService(t *testing.T) {
	ctx := t.Context()

	t.Run("successful connection", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping integration test in short mode")
		}

		ds := setupTestDatabase(ctx, t)

		// Verify queries interface is available
		assert.NotNil(t, ds.Queries())

		// Verify pool is available (cast to concrete type for testing)
		if pgxService, ok := ds.(*internal.PgxDBService); ok {
			assert.NotNil(t, pgxService.Pool())
		}
	})

	t.Run("invalid database URL", func(t *testing.T) {
		invalidURL := "invalid://url"

		ds, err := internal.NewDBService(ctx, invalidURL)
		require.Error(t, err)
		assert.Nil(t, ds)
		assert.Contains(t, err.Error(), "failed to parse database URL")
	})

	t.Run("connection failure", func(t *testing.T) {
		// Use non-existent database
		badURL := "postgres://user:pass@nonexistent:5432/db"

		ds, err := internal.NewDBService(ctx, badURL)
		require.Error(t, err)
		assert.Nil(t, ds)
		assert.Contains(t, err.Error(), "failed to ping database")
	})
}

func TestDatabaseService_HealthCheck(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := t.Context()
	ds := setupTestDatabase(ctx, t)

	t.Run("successful health check", func(t *testing.T) {
		err := ds.HealthCheck(ctx)
		assert.NoError(t, err)
	})

	t.Run("health check with timeout", func(t *testing.T) {
		timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
		defer cancel()

		// Give context time to expire
		time.Sleep(1 * time.Millisecond)

		err := ds.HealthCheck(timeoutCtx)
		assert.Error(t, err)
	})
}

func TestDatabaseService_Close(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := t.Context()
	ds := setupTestDatabase(ctx, t)

	// Verify connection works before closing
	err := ds.HealthCheck(ctx)
	require.NoError(t, err)

	// Close the service
	ds.Close()

	// Verify connection no longer works after closing
	err = ds.HealthCheck(ctx)
	assert.Error(t, err)
}

func TestDatabaseService_ConnectionPoolConfiguration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := t.Context()
	ds := setupTestDatabase(ctx, t)

	// Cast to concrete type to access pool configuration
	pgxService, ok := ds.(*internal.PgxDBService)
	require.True(t, ok, "Expected PgxDBService implementation")

	pool := pgxService.Pool()
	require.NotNil(t, pool)

	// Test that pool configuration is applied correctly
	config := pool.Config()
	assert.Equal(t, int32(25), config.MaxConns)
	assert.Equal(t, int32(5), config.MinConns)
	assert.Equal(t, time.Hour, config.MaxConnLifetime)
	assert.Equal(t, 30*time.Minute, config.MaxConnIdleTime)
}

// TestDatabaseService_Integration runs integration tests if database is available.
func TestDatabaseService_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := t.Context()

	t.Run("concurrent health checks", func(t *testing.T) {
		ds := setupTestDatabase(ctx, t)

		// Run multiple health checks concurrently
		done := make(chan error, 10)
		for range 10 {
			go func() {
				done <- ds.HealthCheck(ctx)
			}()
		}

		// All should succeed
		for range 10 {
			err := <-done
			assert.NoError(t, err)
		}
	})
}

func TestDatabaseService_Transaction_Rollback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := t.Context()
	ds := setupTestDatabase(ctx, t)
	store := internal.NewDataStore(ds)

	userID := uuid.Must(uuid.NewV7())
	username := gofakeit.Name()
	intentionalErr := errors.New("intentional failure")

	err := ds.Transaction(ctx, func(ctx context.Context, tx *internal.Transaction) error {
		_, err := tx.CreateUser(ctx, db.CreateUserParams{
			ID:       userID,
			Username: username,
			IsStaff:  false,
		})
		require.NoError(t, err)

		return intentionalErr
	})

	require.ErrorIs(t, err, intentionalErr)

	_, err = store.GetUserByID(ctx, userID)
	require.Error(t, err)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}
