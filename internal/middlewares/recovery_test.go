package middlewares_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thara/facility_reservation_go/internal/middlewares"
)

func TestRecoveryMiddleware_Panic(t *testing.T) {
	// Handler that panics when invoked
	panicHandler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("unexpected error")
	})

	handler := middlewares.RecoveryMiddleware(panicHandler)

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()

	// Execute the request; the middleware should recover the panic
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal Server Error")
}
