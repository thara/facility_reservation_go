package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	t.Run("creates service with nil database", func(t *testing.T) {
		// Test that NewService can handle nil database for unit testing
		svc := NewService(nil)
		require.NotNil(t, svc)
		assert.Nil(t, svc.db)
	})

	t.Run("creates service with mock database", func(t *testing.T) {
		// Create a mock database service for testing
		mockDB := &DatabaseService{}
		svc := NewService(mockDB)
		require.NotNil(t, svc)
		assert.Equal(t, mockDB, svc.db)
	})
}

func TestService_Structure(t *testing.T) {
	t.Run("service embeds UnimplementedHandler", func(t *testing.T) {
		svc := NewService(nil)
		require.NotNil(t, svc)
		
		// Verify that the service has the embedded UnimplementedHandler
		// This ensures it satisfies the ogen-generated interface
		assert.NotNil(t, svc.UnimplementedHandler)
	})
}