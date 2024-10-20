package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"questionify/internal/server"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	shutdownError := make(chan error)
	srv := server.NewServer(logger, shutdownError)

	logger.Info("Starting server", "port", srv.Addr)

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}

	err = <-shutdownError
	if err != nil {
		logger.Error("server error", "error", err)
		os.Exit(1)
	}

	logger.Info("server shutdown")
	os.Exit(0)
}
