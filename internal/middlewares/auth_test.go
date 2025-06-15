package middlewares_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/thara/facility_reservation_go/internal"
	"github.com/thara/facility_reservation_go/internal/db"
	"github.com/thara/facility_reservation_go/internal/middlewares"
)

// mockUserTokenQuerier implements internal.UserTokenQuerier for testing.
type mockUserTokenQuerier struct {
	getUserByTokenFunc func(ctx context.Context, token string) (db.GetUserByTokenRow, error)
}

func (m *mockUserTokenQuerier) GetUserByToken(ctx context.Context, token string) (db.GetUserByTokenRow, error) {
	if m.getUserByTokenFunc != nil {
		return m.getUserByTokenFunc(ctx, token)
	}
	return db.GetUserByTokenRow{
		ID:       uuid.UUID{},
		Username: "",
		IsStaff:  false,
	}, nil
}

func TestAuthMiddleware(t *testing.T) {
	testUserID := uuid.New()
	validToken := "valid-token-123"

	t.Run("successful authentication", func(t *testing.T) {
		// Setup mock querier
		mockQuerier := &mockUserTokenQuerier{
			getUserByTokenFunc: func(_ context.Context, token string) (db.GetUserByTokenRow, error) {
				assert.Equal(t, validToken, token)
				return db.GetUserByTokenRow{
					ID:       testUserID,
					Username: "testuser",
					IsStaff:  true,
				}, nil
			},
		}

		// Create test handler that checks for authenticated user
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := middlewares.GetUserFromContext(r.Context())
			assert.True(t, ok, "user should be in context")
			assert.Equal(t, testUserID.String(), user.ID)
			assert.Equal(t, "testuser", user.Username)
			assert.True(t, user.IsStaff)
			w.WriteHeader(http.StatusOK)
		})

		// Setup middleware
		middleware := middlewares.AuthMiddleware(mockQuerier)
		handler := middleware(nextHandler)

		// Create request with valid Bearer token
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		w := httptest.NewRecorder()

		// Execute request
		handler.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing authorization header", func(t *testing.T) {
		mockQuerier := &mockUserTokenQuerier{
			getUserByTokenFunc: nil,
		}

		nextHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			t.Error("next handler should not be called")
		})

		middleware := middlewares.AuthMiddleware(mockQuerier)
		handler := middleware(nextHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Unauthorized")
	})

	t.Run("invalid authorization header format - no Bearer prefix", func(t *testing.T) {
		mockQuerier := &mockUserTokenQuerier{
			getUserByTokenFunc: nil,
		}

		nextHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			t.Error("next handler should not be called")
		})

		middleware := middlewares.AuthMiddleware(mockQuerier)
		handler := middleware(nextHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Basic token123")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Unauthorized")
	})

	t.Run("empty token in authorization header", func(t *testing.T) {
		mockQuerier := &mockUserTokenQuerier{
			getUserByTokenFunc: nil,
		}

		nextHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			t.Error("next handler should not be called")
		})

		middleware := middlewares.AuthMiddleware(mockQuerier)
		handler := middleware(nextHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer ")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Unauthorized")
	})

	t.Run("invalid token - database error", func(t *testing.T) {
		mockQuerier := &mockUserTokenQuerier{
			getUserByTokenFunc: func(_ context.Context, _ string) (db.GetUserByTokenRow, error) {
				return db.GetUserByTokenRow{}, assert.AnError
			},
		}

		nextHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			t.Error("next handler should not be called")
		})

		middleware := middlewares.AuthMiddleware(mockQuerier)
		handler := middleware(nextHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Unauthorized")
	})

	t.Run("nil UserTokenQuerier", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
			t.Error("next handler should not be called")
		})

		middleware := middlewares.AuthMiddleware(nil)
		handler := middleware(nextHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Unauthorized")
	})
}

func TestGetUserFromContext(t *testing.T) {
	t.Run("user exists in context", func(t *testing.T) {
		expectedUser := &internal.AuthenticatedUser{
			ID:       uuid.New().String(),
			Username: "testuser",
			IsStaff:  true,
		}

		ctx := middlewares.WithUser(t.Context(), expectedUser)

		user, ok := middlewares.GetUserFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, expectedUser, user)
	})

	t.Run("user not in context", func(t *testing.T) {
		ctx := t.Context()

		user, ok := middlewares.GetUserFromContext(ctx)
		assert.False(t, ok)
		assert.Nil(t, user)
	})
}

func TestExtractBearerToken(t *testing.T) {
	t.Run("valid Bearer token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer token123")

		// We can't test extractBearerToken directly since it's not exported,
		// but we can test it through the middleware behavior
		mockQuerier := &mockUserTokenQuerier{
			getUserByTokenFunc: func(_ context.Context, token string) (db.GetUserByTokenRow, error) {
				assert.Equal(t, "token123", token)
				return db.GetUserByTokenRow{
					ID:       uuid.New(),
					Username: "test",
					IsStaff:  false,
				}, nil
			},
		}

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := middlewares.AuthMiddleware(mockQuerier)
		handler := middleware(nextHandler)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Bearer token with spaces", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "Bearer  token-with-spaces  ")

		mockQuerier := &mockUserTokenQuerier{
			getUserByTokenFunc: func(_ context.Context, token string) (db.GetUserByTokenRow, error) {
				assert.Equal(t, " token-with-spaces  ", token)
				return db.GetUserByTokenRow{
					ID:       uuid.New(),
					Username: "test",
					IsStaff:  false,
				}, nil
			},
		}

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := middlewares.AuthMiddleware(mockQuerier)
		handler := middleware(nextHandler)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
