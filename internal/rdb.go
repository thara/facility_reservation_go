package internal

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thara/facility_reservation_go/internal/db"
)

type RDBMS interface {
	db.Querier

	Transaction(ctx context.Context, fn func(context.Context, *TxQueries) error) error
}

type RDB struct {
	*db.Queries

	pool     *pgxpool.Pool
	strategy TransactionStrategy
}

func NewRDB(pool *pgxpool.Pool) *RDB {
	return &RDB{
		Queries:  db.New(pool),
		pool:     pool,
		strategy: &defaultTransactionStrategy{},
	}
}

func newRDBWithStrategy(pool *pgxpool.Pool, strategy TransactionStrategy) *RDB {
	return &RDB{
		Queries:  db.New(pool),
		pool:     pool,
		strategy: strategy,
	}
}

// TxQueries wraps db.Queries to indicate transaction usage.
type TxQueries struct {
	*db.Queries
}

func (d *RDB) Transaction(ctx context.Context, fn func(context.Context, *TxQueries) error) error {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if rollbackErr := d.strategy.Rollback(ctx, tx); rollbackErr != nil {
			slog.ErrorContext(ctx, "Failed to rollback transaction", "error", rollbackErr)
		}
	}()

	q := &TxQueries{db.New(tx)}
	if err := fn(ctx, q); err != nil {
		return err
	}

	if err := d.strategy.Commit(ctx, tx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

type TransactionStrategy interface {
	Rollback(ctx context.Context, tx pgx.Tx) error
	Commit(ctx context.Context, tx pgx.Tx) error
}

type defaultTransactionStrategy struct{}

func (s *defaultTransactionStrategy) Rollback(ctx context.Context, tx pgx.Tx) error {
	return tx.Rollback(ctx) //nolint:wrapcheck // Caller is responsible for handling original errors
}

func (s *defaultTransactionStrategy) Commit(ctx context.Context, tx pgx.Tx) error {
	return tx.Commit(ctx) //nolint:wrapcheck // Caller is responsible for handling original errors
}
