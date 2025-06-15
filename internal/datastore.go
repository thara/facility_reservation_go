package internal

import (
	"context"

	"github.com/thara/facility_reservation_go/internal/db"
)

type DataStore struct {
	db.Querier

	dbService DatabaseService
}

func NewDataStore(ds DatabaseService) *DataStore {
	return &DataStore{
		Querier:   ds.Queries(),
		dbService: ds,
	}
}

func (ds *DataStore) Transaction(ctx context.Context, fn TransactionFunc) error {
	return ds.dbService.Transaction(ctx, fn) //nolint:wrapcheck // propagate error
}
