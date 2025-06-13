// Package internal provides the business logic implementation for the facility reservation API.
package internal

import "github.com/thara/facility_reservation_go/oas"

// Service implements the facility reservation API handlers by embedding the generated handler interface.
type Service struct {
	oas.UnimplementedHandler
}
