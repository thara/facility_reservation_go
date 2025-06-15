package internal_test

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thara/facility_reservation_go/internal"
)

func NewTestDatabaseService(ctx context.Context, databaseURL string) *internal.DatabaseService {
	ds, err := internal.NewDatabaseServiceWithStrategy(ctx, databaseURL, func(pool *pgxpool.Pool) internal.DatabaseStrategy {
		return &testDatabaseStrategy{pool: pool}
	})
	if err != nil {
		panic("failed to create test database service: " + err.Error())
	}
	return ds
}

type testDatabaseStrategy struct {
	pool *pgxpool.Pool
}

func (s *testDatabaseStrategy) DB() internal.RDBMS {
	return internal.NewRDBWithStrategy(s.pool, &testTransactionStrategy{})
}

func (s *testDatabaseStrategy) Close() {
	fmt.Printf("Closing test database connection pool\n")
	s.pool.Close()
}

type testTransactionStrategy struct{}

func (s *testTransactionStrategy) Rollback(ctx context.Context, tx pgx.Tx) error {
	return nil // No-op for testing
}

func (s *testTransactionStrategy) Commit(ctx context.Context, tx pgx.Tx) error {
	return nil // No-op for testing
}
