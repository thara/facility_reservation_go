package middlewares

import (
	"log/slog"
	"net/http"
)

// RecoveryMiddleware recovers from panics in HTTP handlers and returns 500 Internal Server Error.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				slog.ErrorContext(r.Context(), "HTTP handler panic",
					"method", r.Method,
					"path", r.URL.Path,
					"panic", err,
					"remote_addr", r.RemoteAddr,
				)

				// Return 500 Internal Server Error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
