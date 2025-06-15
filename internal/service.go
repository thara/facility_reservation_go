package internal

import (
	"github.com/thara/facility_reservation_go/internal/api"
)

// Service implements the facility reservation API handlers by embedding the generated handler interface.
type Service struct {
	api.UnimplementedHandler
	dbService DBService
}

// NewService creates a new service with database dependency.
func NewService(dbService DBService) *Service {
	return &Service{
		UnimplementedHandler: api.UnimplementedHandler{},
		dbService:            dbService,
	}
}
