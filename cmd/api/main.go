package main

import (
	"log/slog"
	"os"
	"questionify/internal/server"
)

var Version = "1.0.0"

func main() {
	server := server.NewServer()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Starting server", "version", Version, "port", server.Addr)

	server.ListenAndServe()
}
