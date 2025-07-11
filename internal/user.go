package internal

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/thara/facility_reservation_go/internal/db"
	"github.com/thara/facility_reservation_go/internal/derrors"
)

const (
	tokenSizeBytes = 32
)

// UserTokenQuerier defines the interface for querying user tokens.
type UserTokenQuerier interface {
	GetUserByToken(ctx context.Context, token string) (db.GetUserByTokenRow, error)
}

// AuthenticatedUser represents the authenticated user information.
type AuthenticatedUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	IsStaff  bool   `json:"is_staff"`
}

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

// CreateUser creates a new user with a secure token.
// Only staff users can create new users.
func CreateUser(
	ctx context.Context,
	ds *DataStore,
	user *AuthenticatedUser,
	params CreateUserParams,
) (result *CreateUserResult, err error) {
	defer derrors.Wrap(&err, "CreateUser(ctx, ds, user, params)")
	// Validate that the authenticated user is staff
	if user == nil {
		return nil, errors.New("authenticated user is required")
	}
	if !user.IsStaff {
		return nil, errors.New("only staff users can create new users")
	}

	err = ds.Transaction(ctx, func(ctx context.Context, tx *Transaction) error {
		// Generate UUID v7 for user
		userID := uuid.Must(uuid.NewV7())

		// Create user
		user, err := tx.CreateUser(ctx, db.CreateUserParams{
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
		userToken, err := tx.CreateToken(ctx, db.CreateTokenParams{
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

// GetAuthenticatedUser validates the token and returns the authenticated user.
func GetAuthenticatedUser(
	ctx context.Context,
	querier UserTokenQuerier,
	token string,
) (user *AuthenticatedUser, err error) {
	defer derrors.Wrap(&err, "GetAuthenticatedUser(ctx, querier, token)")
	if querier == nil {
		return nil, errors.New("querier is nil")
	}

	// Get user by token from database
	userRow, err := querier.GetUserByToken(ctx, token)
	if err != nil {
		// Check if it's a "not found" error (typical for invalid tokens)
		return nil, errors.New("invalid or expired token")
	}

	// Convert database row to our AuthenticatedUser type
	user = &AuthenticatedUser{
		ID:       userRow.ID.String(),
		Username: userRow.Username,
		IsStaff:  userRow.IsStaff,
	}

	return user, nil
}
