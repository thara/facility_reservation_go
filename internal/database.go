package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thara/facility_reservation_go/internal/db"
)

// DatabaseService manages database connections and provides query interface
type DatabaseService struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewDatabaseService creates a new database service with connection pool
func NewDatabaseService(ctx context.Context, databaseURL string) (*DatabaseService, error) {
	// Configure connection pool
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Set connection pool settings
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = time.Minute * 30

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DatabaseService{
		pool:    pool,
		queries: db.New(pool),
	}, nil
}

// Queries returns the sqlc-generated query interface
func (ds *DatabaseService) Queries() *db.Queries {
	return ds.queries
}

// Pool returns the underlying connection pool for transactions
func (ds *DatabaseService) Pool() *pgxpool.Pool {
	return ds.pool
}

// Close closes the database connection pool
func (ds *DatabaseService) Close() {
	ds.pool.Close()
}

// HealthCheck verifies database connectivity
func (ds *DatabaseService) HealthCheck(ctx context.Context) error {
	return ds.pool.Ping(ctx)
}
