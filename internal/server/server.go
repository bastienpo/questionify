package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

func NewServer(logger *slog.Logger) *http.Server {
	port, err := strconv.Atoi(os.Getenv("PORT"))

	if err != nil {
		logger.Error("failed to parse port", "error", err)
		return nil
	}

	router := httprouter.New()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      addRoutes(router, logger),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	return srv
}
