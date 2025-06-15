package internal

import (
	"github.com/thara/facility_reservation_go/api"
)

// Service implements the facility reservation API handlers by embedding the generated handler interface.
type Service struct {
	api.UnimplementedHandler
	db DatabaseService
}

// NewService creates a new service with database dependency.
func NewService(db DatabaseService) *Service {
	return &Service{
		UnimplementedHandler: api.UnimplementedHandler{},
		db:                   db,
	}
}
