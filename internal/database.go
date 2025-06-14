package internal

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thara/facility_reservation_go/internal/db"
)

const (
	maxConnIdleTimeMinutes = 30
)

// transactionStrategy defines how transactions should be handled.
type transactionStrategy interface {
	execute(ctx context.Context, tx pgx.Tx, fn func(context.Context, *TxQueries) error) error
}

// commitStrategy commits transactions normally.
type commitStrategy struct{}

func (s *commitStrategy) execute(ctx context.Context, tx pgx.Tx, fn func(context.Context, *TxQueries) error) error {
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			slog.ErrorContext(ctx, "Failed to rollback transaction", "error", rollbackErr)
		}
	}()

	q := &TxQueries{db.New(tx)}
	if err := fn(ctx, q); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DatabaseService manages database connections and provides query interface.
type DatabaseService struct {
	pool     *pgxpool.Pool
	queries  *db.Queries
	strategy transactionStrategy
}

// TxQueries wraps db.Queries to indicate transaction usage.
type TxQueries struct {
	*db.Queries
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
		pool:     pool,
		queries:  db.New(pool),
		strategy: &commitStrategy{},
	}, nil
}

// Queries returns the sqlc-generated query interface.
func (ds *DatabaseService) Queries() *db.Queries {
	return ds.queries
}

// Pool returns the underlying connection pool for transactions.
func (ds *DatabaseService) Pool() *pgxpool.Pool {
	return ds.pool
}

// Close closes the database connection pool.
func (ds *DatabaseService) Close() {
	// ds.pool.Close()
}

// Transaction executes a function within a database transaction.
func (ds *DatabaseService) Transaction(ctx context.Context, fn func(context.Context, *TxQueries) error) error {
	tx, err := ds.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	return ds.strategy.execute(ctx, tx, fn)
}

// HealthCheck verifies database connectivity.
func (ds *DatabaseService) HealthCheck(ctx context.Context) error {
	if err := ds.pool.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}
