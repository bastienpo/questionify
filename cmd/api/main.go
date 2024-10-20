package main

import (
	"log/slog"
	"os"
	"questionify/internal/server"
)

var version = "1.0.0"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	server := server.NewServer(logger)

	logger.Info("Starting server", "port", server.Addr)

	server.ListenAndServe()
}
