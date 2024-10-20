package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
)

type config struct {
	port int
}

type application struct {
	logger *slog.Logger
}

var version = "1.0.0"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		logger.Error(err.Error())
		return
	}

	fmt.Printf("Starting server on port: %d\n", port)
}
