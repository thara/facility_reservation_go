package internal

import (
	"context"

	"github.com/thara/facility_reservation_go/api"
)

// Service implements the facility reservation API handlers by embedding the generated handler interface.
type Service struct {
	api.UnimplementedHandler
	dbService DatabaseService
}

// NewService creates a new service with database dependency.
func NewService(dbService DatabaseService) *Service {
	return &Service{
		UnimplementedHandler: api.UnimplementedHandler{},
		dbService:            dbService,
	}
}

// CreateUser creates a new user with an API token.
func (s *Service) CreateUser(ctx context.Context, params CreateUserParams) (*CreateUserResult, error) {
	ds := NewDataStore(s.dbService)
	return CreateUser(ctx, ds, params)
}
