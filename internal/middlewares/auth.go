package middlewares

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/thara/facility_reservation_go/internal"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	// userContextKey is the key used to store the authenticated user in the request context.
	userContextKey contextKey = "authenticated_user"
)

// AuthMiddleware provides token-based authentication for HTTP handlers.
// It expects a Bearer token in the Authorization header and validates it against the database.
func AuthMiddleware(querier internal.UserTokenQuerier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract token from Authorization header
			token, err := extractBearerToken(r)
			if err != nil {
				slog.WarnContext(ctx, "authentication failed: invalid authorization header",
					"method", r.Method,
					"path", r.URL.Path,
					"error", err.Error(),
					"remote_addr", r.RemoteAddr,
				)

				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Validate token and get user
			user, err := authenticateToken(ctx, querier, token)
			if err != nil {
				slog.WarnContext(ctx, "authentication failed: invalid token",
					"method", r.Method,
					"path", r.URL.Path,
					"error", err.Error(),
					"remote_addr", r.RemoteAddr,
				)

				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add user to context
			ctxWithUser := withUser(ctx, user)
			requestWithUser := r.WithContext(ctxWithUser)

			slog.InfoContext(ctxWithUser, "user authenticated",
				"method", r.Method,
				"path", r.URL.Path,
				"user_id", user.ID,
				"username", user.Username,
				"is_staff", user.IsStaff,
				"remote_addr", r.RemoteAddr,
			)

			// Call the next handler with the authenticated user context
			next.ServeHTTP(w, requestWithUser)
		})
	}
}

// GetUserFromContext retrieves the authenticated user from the request context.
func GetUserFromContext(ctx context.Context) (*internal.AuthenticatedUser, bool) {
	user, ok := ctx.Value(userContextKey).(*internal.AuthenticatedUser)
	return user, ok
}

// withUser returns a new context with the authenticated user stored in it.
func withUser(ctx context.Context, user *internal.AuthenticatedUser) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// extractBearerToken extracts the Bearer token from the Authorization header.
func extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}

	// Check if it's a Bearer token
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", errors.New("authorization header must use Bearer scheme")
	}

	// Extract the token
	token := strings.TrimPrefix(authHeader, bearerPrefix)
	if token == "" {
		return "", errors.New("empty token in authorization header")
	}

	return token, nil
}

// authenticateToken validates the token and returns the authenticated user.
func authenticateToken(
	ctx context.Context,
	querier internal.UserTokenQuerier,
	token string,
) (*internal.AuthenticatedUser, error) {
	if querier == nil {
		return nil, errors.New("querier is nil")
	}

	// Use the GetAuthenticatedUser function from internal/user.go
	user, err := internal.GetAuthenticatedUser(ctx, querier, token)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}
	return user, nil
}
