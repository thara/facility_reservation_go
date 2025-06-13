package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/thara/facility_reservation_go/internal"
	"github.com/thara/facility_reservation_go/oas"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func run() error {
	svc := &internal.Service{}

	srv, err := oas.NewServer(svc)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	if err := http.ListenAndServe(":8080", srv); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}
