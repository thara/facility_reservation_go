package internal

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	// Wait strategy constants for testcontainers.
	logOccurrences    = 2
	startupTimeoutSec = 60
)

// TestPostgresContainer wraps a PostgreSQL testcontainer with helper methods.
type TestPostgresContainer struct {
	container *postgres.PostgresContainer
	dbURL     string
}

// NewTestPostgresContainer creates and starts a PostgreSQL container for testing.
func NewTestPostgresContainer(ctx context.Context, t *testing.T) *TestPostgresContainer {
	t.Helper()

	// Create PostgreSQL container
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(logOccurrences).
				WithStartupTimeout(startupTimeoutSec*time.Second),
		),
	)
	require.NoError(t, err, "Failed to start PostgreSQL container")

	// Get connection string
	dbURL, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "Failed to get connection string")

	tc := &TestPostgresContainer{
		container: container,
		dbURL:     dbURL,
	}

	// Setup cleanup
	t.Cleanup(func() {
		if err := tc.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	})

	return tc
}

// DatabaseURL returns the connection string for the test database.
func (tc *TestPostgresContainer) DatabaseURL() string {
	return tc.dbURL
}

// CreateDatabaseService creates a new DatabaseService connected to the test container.
func (tc *TestPostgresContainer) CreateDatabaseService(ctx context.Context, t *testing.T) *DatabaseService {
	t.Helper()

	ds, err := NewDatabaseService(ctx, tc.dbURL)
	require.NoError(t, err, "Failed to create database service")

	return ds
}

// ApplySchema applies the database schema to the test container
// ApplySchema applies the database schema to the test container.
// This reads and executes the schema from _db/schema.sql.
func (tc *TestPostgresContainer) ApplySchema(ctx context.Context, t *testing.T) {
	t.Helper()

	ds := tc.CreateDatabaseService(ctx, t)
	defer ds.Close()

	// Read the actual schema file (relative to project root)
	schemaContent, err := os.ReadFile("../_db/schema.sql")
	require.NoError(t, err, "Failed to read schema file")

	// Execute the schema
	_, err = ds.Pool().Exec(ctx, string(schemaContent))
	require.NoError(t, err, "Failed to execute schema")
}

// Terminate stops and removes the container.
func (tc *TestPostgresContainer) Terminate(ctx context.Context) error {
	if tc.container != nil {
		if err := tc.container.Terminate(ctx); err != nil {
			return fmt.Errorf("failed to terminate container: %w", err)
		}
	}
	return nil
}

// SetupTestDatabase creates a test database with schema applied
// SetupTestDatabase creates a test database with schema applied.
// Falls back to external database if testcontainers fails.
func SetupTestDatabase(ctx context.Context, t *testing.T) *DatabaseService {
	t.Helper()

	// Try to use testcontainers first with panic recovery
	var container *TestPostgresContainer
	var err error

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Testcontainers panicked: %v", r)
				err = fmt.Errorf("testcontainers panic: %v", r)
			}
		}()
		container, err = tryNewTestPostgresContainer(ctx, t)
	}()

	if err != nil {
		t.Logf("Failed to create testcontainer, falling back to external database: %v", err)
		return setupExternalTestDatabase(ctx, t)
	}

	container.ApplySchema(ctx, t)
	return container.CreateDatabaseService(ctx, t)
}

// tryNewTestPostgresContainer attempts to create a testcontainer, returns error if fails.
func tryNewTestPostgresContainer(ctx context.Context, t *testing.T) (*TestPostgresContainer, error) {
	t.Helper()

	// Create PostgreSQL container
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(logOccurrences).
				WithStartupTimeout(startupTimeoutSec*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres container: %w", err)
	}

	// Get connection string
	dbURL, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		if termErr := container.Terminate(ctx); termErr != nil {
			t.Logf("Failed to terminate container after connection error: %v", termErr)
		}
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	tc := &TestPostgresContainer{
		container: container,
		dbURL:     dbURL,
	}

	// Setup cleanup
	t.Cleanup(func() {
		if err := tc.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	})

	return tc, nil
}

// setupExternalTestDatabase sets up database using external PostgreSQL instance.
func setupExternalTestDatabase(ctx context.Context, t *testing.T) *DatabaseService {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5433/facility_reservation_test?sslmode=disable"
	}

	ds, err := NewDatabaseService(ctx, databaseURL)
	if err != nil {
		t.Skipf("Failed to connect to external test database: %v", err)
	}

	// Setup cleanup to close database
	t.Cleanup(ds.Close)

	// Apply schema and cleanup data for external database
	applySchemaToExternalDB(ctx, t, ds)
	cleanupExternalTestData(ctx, t, ds)

	return ds
}

// applySchemaToExternalDB applies schema to external test database.
func applySchemaToExternalDB(ctx context.Context, t *testing.T, ds *DatabaseService) {
	t.Helper()

	// Read the actual schema file (relative to project root)
	schemaContent, err := os.ReadFile("../_db/schema.sql")
	require.NoError(t, err, "Failed to read schema file")

	// Execute the schema (IF NOT EXISTS clauses handle existing tables/indexes)
	_, err = ds.Pool().Exec(ctx, string(schemaContent))
	require.NoError(t, err, "Failed to execute schema on external database")
}

// cleanupExternalTestData cleans existing test data from external database.
func cleanupExternalTestData(ctx context.Context, t *testing.T, ds *DatabaseService) {
	t.Helper()

	// Clean up test data in dependency order (foreign keys)
	cleanupQueries := []string{
		"DELETE FROM user_tokens",
		"DELETE FROM users",
		"DELETE FROM facilities",
	}

	for _, query := range cleanupQueries {
		_, err := ds.Pool().Exec(ctx, query)
		// Ignore errors if tables don't exist or are already empty
		if err != nil {
			t.Logf("Cleanup query failed (ignoring): %s - %v", query, err)
		}
	}
}
