package internal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
	"github.com/thara/facility_reservation_go/internal/db"
)

const (
	tokenSizeBytes = 32
)

// CreateUserParams holds parameters for creating a new user.
type CreateUserParams struct {
	Username string
	IsStaff  bool
}

// CreateUserResult holds the result of creating a user with token.
type CreateUserResult struct {
	User  db.User
	Token db.UserToken
}

// CreateUser creates a new user with an API token.
func (s *Service) CreateUser(ctx context.Context, params CreateUserParams) (*CreateUserResult, error) {
	var result *CreateUserResult

	err := s.db.Transaction(ctx, func(ctx context.Context, q *TxQueries) error {
		// Generate UUID v7 for user
		userID := uuid.Must(uuid.NewV7())

		// Create user
		user, err := q.CreateUser(ctx, db.CreateUserParams{
			ID:       userID,
			Username: params.Username,
			IsStaff:  params.IsStaff,
		})
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		// Generate UUID v7 for token
		tokenID := uuid.Must(uuid.NewV7())

		// Generate secure token
		token := generateToken()

		// Create token for user
		userToken, err := q.CreateToken(ctx, db.CreateTokenParams{
			ID:        tokenID,
			UserID:    user.ID,
			Token:     token,
			Name:      "Default Token",
			ExpiresAt: nil, // No expiration
		})
		if err != nil {
			return fmt.Errorf("failed to create token: %w", err)
		}

		result = &CreateUserResult{
			User:  user,
			Token: userToken,
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("transaction failed: %w", err)
	}

	return result, nil
}

// generateToken creates a secure random token.
func generateToken() string {
	bytes := make([]byte, tokenSizeBytes)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(fmt.Sprintf("failed to generate token: %v", err))
	}
	return hex.EncodeToString(bytes)
}
