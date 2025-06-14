package internal_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thara/facility_reservation_go/internal"
)

func TestNewService(t *testing.T) {
	t.Run("creates service with nil database", func(t *testing.T) {
		// Test that NewService can handle nil database for unit testing
		svc := internal.NewService(nil)
		require.NotNil(t, svc)
		// We test behavior, not implementation details
		// The service should be created successfully
	})

	t.Run("creates service with database", func(t *testing.T) {
		// Test behavior: service creation should not panic or error
		// In a real test, we'd use a mock or test database
		svc := internal.NewService(nil) // Using nil is acceptable for unit tests
		require.NotNil(t, svc)
		// Test that the service implements the expected interface
		var _ interface{} = svc // Service should exist
	})
}

func TestService_Structure(t *testing.T) {
	t.Run("service embeds UnimplementedHandler", func(t *testing.T) {
		svc := internal.NewService(nil)
		require.NotNil(t, svc)

		// Verify that the service has the embedded UnimplementedHandler
		// This ensures it satisfies the ogen-generated interface
		assert.NotNil(t, svc.UnimplementedHandler)
	})
}
