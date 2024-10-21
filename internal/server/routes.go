package server

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func addRoutes(router *httprouter.Router, logger *slog.Logger) http.Handler {
	router.MethodNotAllowed = methodNotAllowedHandler(logger)
	router.NotFound = notFoundHandler(logger)

	router.Handler(http.MethodGet, "/v1/healthcheck", healthGETHandler())

	standard := alice.New(
		recoverPanicMiddleware(logger),
		enableCORSMiddleware(),
		logRequestMiddleware(logger),
	)

	return standard.Then(router)
}

func notFoundHandler(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := "the requested resource could not be found"
		errorResponse(logger, w, r, http.StatusNotFound, msg)
	})
}

func methodNotAllowedHandler(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
		errorResponse(logger, w, r, http.StatusMethodNotAllowed, msg)
	})
}

func healthGETHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := map[string]string{"status": "available", "environment": os.Getenv("ENVIRONMENT")}

		err := writeJSON(w, http.StatusOK, data, nil)
		if err != nil {
			http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
		}
	})
}
