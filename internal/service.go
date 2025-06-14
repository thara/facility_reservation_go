package internal

import "github.com/thara/facility_reservation_go/oas"

// Service implements the facility reservation API handlers by embedding the generated handler interface.
type Service struct {
	oas.UnimplementedHandler
}
