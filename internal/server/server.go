package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Server struct {
	port int
}

func NewServer() *http.Server {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatal(err)
		return nil
	}

	NewServer := &Server{
		port: port,
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      NewServer.addRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		// ErrorLog:     slog.NewLogLogger(logger, slog.LevelError),
	}

	return server
}
