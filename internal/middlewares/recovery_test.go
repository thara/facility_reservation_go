package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thara/facility_reservation_go/internal/middlewares"
)

func TestRecoveryMiddleware(t *testing.T) {
	t.Run("panic recovers with 500", func(t *testing.T) {
		panicHandler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic("unexpected error")
		})

		handler := middlewares.RecoveryMiddleware(panicHandler)

		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal Server Error")
	})

	t.Run("no panic passes through", func(t *testing.T) {
		okHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})

		handler := middlewares.RecoveryMiddleware(okHandler)

		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusTeapot, w.Code)
	})
}
