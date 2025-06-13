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
	"github.com/thara/facility_reservation_go/oas"
)

var addr string

func init() {
	flag.StringVar(&addr, "addr", ":8080", "HTTP server address")
	flag.Parse()

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
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	svc := &internal.Service{}

	handler, err := oas.NewServer(svc)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		slog.Info("starting server", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	slog.Info("server exited")
	return nil
}
