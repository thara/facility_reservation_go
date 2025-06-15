package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	maxConnIdleTimeMinutes = 30
)

// DatabaseService manages database connections and provides query interface.
type DatabaseService struct {
	pool     *pgxpool.Pool
	strategy DatabaseStrategy
}

// NewDatabaseService creates a new database service with connection pool.
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
	config.MaxConnIdleTime = time.Minute * maxConnIdleTimeMinutes

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
		pool: pool,
		strategy: &defaultDatabaseStrategy{
			pool: pool,
		},
	}, nil
}

func (ds *DatabaseService) DB() RDBMS {
	return ds.strategy.DB()
}

// Pool returns the underlying connection pool for transactions.
func (ds *DatabaseService) Pool() *pgxpool.Pool {
	return ds.pool
}

// Close closes the database connection pool.
func (ds *DatabaseService) Close() {
	fmt.Println("Closing database connection pool")
	ds.strategy.Close()
}

// HealthCheck verifies database connectivity.
func (ds *DatabaseService) HealthCheck(ctx context.Context) error {
	if err := ds.pool.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}

type DatabaseStrategy interface {
	DB() RDBMS
	Close()
}

type defaultDatabaseStrategy struct {
	pool *pgxpool.Pool
}

func (s *defaultDatabaseStrategy) DB() RDBMS {
	return NewRDB(s.pool)
}

func (s *defaultDatabaseStrategy) Close() {
	s.pool.Close()
}

func newDatabaseServiceWithStrategy(ctx context.Context, databaseURL string, f func(*pgxpool.Pool) DatabaseStrategy) (*DatabaseService, error) {
	ds, err := NewDatabaseService(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	ds.strategy = f(ds.pool)
	return ds, nil
}
