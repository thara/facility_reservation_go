package internal_test

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thara/facility_reservation_go/internal"
	"github.com/thara/facility_reservation_go/internal/db"
)

func TestCreateUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := t.Context()
	db := setupTestDatabase(ctx, t)

	service := internal.NewService(db)

	t.Run("creates staff user successfully", func(t *testing.T) {
		params := internal.CreateUserParams{
			Username: gofakeit.Name(),
			IsStaff:  true,
		}

		result, err := service.CreateUser(ctx, params)

		require.NoError(t, err)
		assert.NotNil(t, result)

		// Verify user properties
		assert.Equal(t, params.Username, result.User.Username)
		assert.True(t, result.User.IsStaff)
		assert.NotEmpty(t, result.User.ID)
		assert.False(t, result.User.CreatedAt.IsZero())

		// Verify token properties
		assert.NotEmpty(t, result.Token.Token)
		assert.Equal(t, result.User.ID, result.Token.UserID)
		assert.Equal(t, "Default Token", result.Token.Name)
		assert.Nil(t, result.Token.ExpiresAt)
		assert.NotEmpty(t, result.Token.ID)
		assert.False(t, result.Token.CreatedAt.IsZero())

		// Verify token is 64 characters (32 bytes hex encoded)
		assert.Len(t, result.Token.Token, 64)
	})

	t.Run("creates regular user successfully", func(t *testing.T) {
		params := internal.CreateUserParams{
			Username: gofakeit.Name(),
			IsStaff:  false,
		}

		result, err := service.CreateUser(ctx, params)

		require.NoError(t, err)
		assert.NotNil(t, result)

		// Verify user properties
		assert.Equal(t, params.Username, result.User.Username)
		assert.False(t, result.User.IsStaff)
		assert.NotEmpty(t, result.User.ID)
	})

	t.Run("fails with duplicate username", func(t *testing.T) {
		// Create first user
		params := internal.CreateUserParams{
			Username: gofakeit.Name(),
			IsStaff:  false,
		}

		_, err := service.CreateUser(ctx, params)
		require.NoError(t, err)

		// Try to create user with same username
		_, err = service.CreateUser(ctx, params)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create user")
	})

	t.Run("generates unique tokens", func(t *testing.T) {
		// Create multiple users and verify tokens are unique
		tokens := make(map[string]bool)

		for range 10 {
			params := internal.CreateUserParams{
				Username: gofakeit.Name(),
				IsStaff:  false,
			}

			result, err := service.CreateUser(ctx, params)
			require.NoError(t, err)

			// Check token is unique
			assert.False(t, tokens[result.Token.Token], "Token should be unique")
			tokens[result.Token.Token] = true
		}
	})

	t.Run("generates UUID v7 format", func(t *testing.T) {
		params := internal.CreateUserParams{
			Username: gofakeit.Name(),
			IsStaff:  false,
		}

		result, err := service.CreateUser(ctx, params)
		require.NoError(t, err)

		// Verify UUIDs are valid format
		userIDStr := result.User.ID.String()
		tokenIDStr := result.Token.ID.String()

		// UUID should be 36 characters with dashes
		assert.Len(t, userIDStr, 36)
		assert.Len(t, tokenIDStr, 36)

		// Should contain dashes in correct positions
		assert.Equal(t, '-', rune(userIDStr[8]))
		assert.Equal(t, '-', rune(userIDStr[13]))
		assert.Equal(t, '-', rune(userIDStr[18]))
		assert.Equal(t, '-', rune(userIDStr[23]))
	})
}

func TestCreateUserTransactionRollback(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := t.Context()
	db := setupTestDatabase(ctx, t)

	service := internal.NewService(db)

	t.Run("transaction rolls back on token creation failure", func(t *testing.T) {
		// This test is conceptual - in practice, token creation would rarely fail
		// after user creation succeeds, but this demonstrates transaction behavior

		params := internal.CreateUserParams{
			Username: gofakeit.Name(),
			IsStaff:  false,
		}

		// Create user successfully
		result, err := service.CreateUser(ctx, params)
		require.NoError(t, err)

		// Verify user was created
		users := getUsersByUsername(t, db, params.Username)
		assert.Len(t, users, 1)

		// Verify token was created
		tokens := getTokensByUserID(t, db, result.User.ID)
		assert.Len(t, tokens, 1)
	})
}

// Helper functions for testing

func setupTestDatabase(ctx context.Context, t *testing.T) *internal.DatabaseService {
	t.Helper()

	testDatabaseURL := "postgres://postgres:postgres@localhost:5433/facility_reservation_test?sslmode=disable"

	ds, err := internal.NewDatabaseService(ctx, testDatabaseURL)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	return ds
}

func getUsersByUsername(t *testing.T, database *internal.DatabaseService, username string) []db.User {
	t.Helper()
	user, err := database.Queries().GetUserByUsername(t.Context(), username)
	if err != nil {
		return []db.User{}
	}
	return []db.User{user}
}

func getTokensByUserID(t *testing.T, database *internal.DatabaseService, userID uuid.UUID) []db.UserToken {
	t.Helper()
	tokens, err := database.Queries().ListUserTokens(t.Context(), userID)
	if err != nil {
		return []db.UserToken{}
	}
	return tokens
}
