package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
)

func NewServer(logger *slog.Logger, shutdownError chan error) *http.Server {
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
	}

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		logger.Info("Shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	return srv
}
