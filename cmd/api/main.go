package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	database "questionify/internal/data"
	"questionify/internal/server"
	"syscall"
	"time"
)

func gracefulShutdown(ctx context.Context, srv *http.Server, done chan struct{}, logger *slog.Logger) {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Wait for the signal
	<-ctx.Done()

	logger.Info("Shutting down server")

	// allow for requests to complete
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", "error", err)
	}

	logger.Info("Server shutdown completed")

	close(done)
}

func run(ctx context.Context, w io.Writer, args []string) error {
	logger := slog.New(slog.NewTextHandler(w, nil))

	// Connect to the database
	db, err := database.New(logger)
	if err != nil {
		logger.Error("failed to create database", "error", err)
		return fmt.Errorf("failed to create database: %s", err)
	}

	defer db.Close()

	// Create the HTTP server
	srv := server.NewServer(logger)

	done := make(chan struct{})
	go gracefulShutdown(ctx, srv, done, logger)

	logger.Info("Starting server", "port", srv.Addr)

	// Start the HTTP server and listen for requests
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("http server error: %s", err)
	}

	// Wait for the shutdown signal
	<-done
	return nil
}

func main() {
	if err := run(context.Background(), os.Stdout, os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
