package internal

import (
	"github.com/thara/facility_reservation_go/internal/api"
)

// APIService implements the facility reservation API handlers by embedding the generated handler interface.
type APIService struct {
	api.UnimplementedHandler
	dbService DBService
}

// NewAPIService creates a new service with database dependency.
func NewAPIService(dbService DBService) *APIService {
	return &APIService{
		UnimplementedHandler: api.UnimplementedHandler{},
		dbService:            dbService,
	}
}
