package internal

import (
	"context"

	"github.com/thara/facility_reservation_go/internal/db"
)

// DataStore provides database operations with transaction support.
type DataStore struct {
	db.Querier

	dbService DBService
}

// NewDataStore creates a new DataStore instance with the given database service.
func NewDataStore(ds DBService) *DataStore {
	return &DataStore{
		Querier:   ds.Queries(),
		dbService: ds,
	}
}

// Transaction executes the given function within a database transaction.
func (ds *DataStore) Transaction(ctx context.Context, fn TransactionFunc) error {
	return ds.dbService.Transaction(ctx, fn) //nolint:wrapcheck // propagate error
}
