package internal

import (
	"context"
	"errors"
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

// TransactionFunc defines the function signature for database transactions.
type TransactionFunc func(context.Context, *Transaction) error

// DBService defines the contract for database operations.
type DBService interface {
	Queries() db.Querier
	Close()
	HealthCheck(ctx context.Context) error
	Transaction(ctx context.Context, fn TransactionFunc) error
}

// Transaction wraps db.Queries to indicate transaction usage.
type Transaction struct {
	db.Querier
}

// PgxDBService implements DatabaseInterface using pgx.
type PgxDBService struct {
	pool    *pgxpool.Pool
	queries *db.Queries
}

// NewDBService creates a new database service with connection pool.
func NewDBService(
	ctx context.Context,
	databaseURL string,
) (DBService, error) {
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

	return &PgxDBService{
		pool:    pool,
		queries: db.New(pool),
	}, nil
}

// Queries returns the sqlc-generated query interface.
func (ds *PgxDBService) Queries() db.Querier { //nolint:ireturn // returns interface to encupsulate implementation details
	return ds.queries
}

// Pool returns the underlying connection pool for transactions.
func (ds *PgxDBService) Pool() *pgxpool.Pool {
	return ds.pool
}

// Close closes the database connection pool.
func (ds *PgxDBService) Close() {
	ds.pool.Close()
}

// Transaction executes a function within a database transaction.
func (ds *PgxDBService) Transaction(ctx context.Context, fn TransactionFunc) error {
	tx, err := ds.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	committed := false
	defer func() {
		if committed {
			return
		}
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			slog.ErrorContext(ctx, "Failed to rollback transaction", "error", rollbackErr)
		}
	}()

	q := &Transaction{db.New(tx)}
	if err := fn(ctx, q); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true

	return nil
}

// HealthCheck verifies database connectivity.
func (ds *PgxDBService) HealthCheck(ctx context.Context) error {
	if err := ds.pool.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	return nil
}
