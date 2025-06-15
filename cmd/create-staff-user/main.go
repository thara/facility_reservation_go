package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/thara/facility_reservation_go/internal"
)

var (
	username    string
	databaseURL string
)

func init() {
	flag.StringVar(&username, "username", "", "Username for the staff user (required)")
	flag.StringVar(&databaseURL, "database-url", "", "Database connection URL")
	flag.Parse()

	// Set default database URL if not provided
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
		if databaseURL == "" {
			databaseURL = "postgres://postgres:postgres@localhost:5432/facility_reservation_dev?sslmode=disable"
		}
	}

	// Configure logging
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
	ctx := context.Background()

	if err := run(ctx); err != nil {
		slog.ErrorContext(ctx, "command failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Validate required username
	if username == "" {
		return errors.New("username is required. Use -username flag")
	}

	// Initialize database
	db, err := internal.NewDBService(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	slog.InfoContext(ctx, "database connection established", "url", databaseURL)

	// Create datastore
	ds := internal.NewDataStore(db)

	// Create staff user
	params := internal.CreateUserParams{
		Username: username,
		IsStaff:  true,
	}

	result, err := internal.CreateUser(ctx, ds, params)
	if err != nil {
		return fmt.Errorf("failed to create staff user: %w", err)
	}

	// Output success information using structured logging
	slog.InfoContext(ctx, "staff user created successfully",
		"user_id", result.User.ID,
		"username", result.User.Username,
		"is_staff", result.User.IsStaff,
		"created_at", result.User.CreatedAt.Format("2006-01-02 15:04:05"),
		"token_id", result.Token.ID,
		"token", result.Token.Token)

	return nil
}
