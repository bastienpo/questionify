package server

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func addRoutes(router *httprouter.Router, logger *slog.Logger) http.Handler {
	router.Handler(http.MethodGet, "/v1/healthcheck", healthGETHandler())

	standard := alice.New(
		recoverPanicMiddleware(logger),
		enableCORS(),
		logRequestMiddleware(logger),
	)

	return standard.Then(router)
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
