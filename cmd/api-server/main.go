package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thara/facility_reservation_go/internal"
	"github.com/thara/facility_reservation_go/internal/api"
	"github.com/thara/facility_reservation_go/internal/middlewares"
)

const (
	readHeaderTimeout = 30 * time.Second
	shutdownTimeout   = 30 * time.Second
)

var (
	addr        string
	databaseURL string
)

func init() {
	flag.StringVar(&addr, "addr", ":8080", "HTTP server address")
	flag.StringVar(&databaseURL, "database-url", "", "Database connection URL")
	flag.Parse()

	// Set default database URL if not provided
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
		if databaseURL == "" {
			databaseURL = "postgres://postgres:postgres@localhost:5432/facility_reservation_dev?sslmode=disable"
		}
	}

	var handler slog.Handler
	env := os.Getenv("ENV")
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, nil)
	} else {
		handler = slog.NewTextHandler(os.Stdout, nil)
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		slog.ErrorContext(ctx, "server failed", "error", err)
	}
}

func run(ctx context.Context) error {
	// Initialize database
	db, err := internal.NewDBService(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	slog.InfoContext(ctx, "database connection established", "url", databaseURL)

	// Create service with database dependency
	svc := internal.NewAPIService(db)

	handler, err := api.NewServer(svc)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Wrap handler with middleware (recovery first, then auth, then logging)
	recoveredHandler := middlewares.RecoveryMiddleware(handler)
	authHandler := middlewares.AuthMiddleware(db)(recoveredHandler)
	loggedHandler := middlewares.LoggingMiddleware(authHandler)

	server := &http.Server{
		Addr:              addr,
		Handler:           loggedHandler,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	go func() {
		slog.InfoContext(ctx, "starting server", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.ErrorContext(ctx, "failed to start server", "error", err)
		}
	}()

	<-ctx.Done()
	slog.InfoContext(ctx, "shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	slog.InfoContext(ctx, "server exited")
	return nil
}
