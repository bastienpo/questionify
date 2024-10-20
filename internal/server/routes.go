package server

import (
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

func (s *Server) addRoutes() http.Handler {
	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", healthGetHandler)

	return router
}

func healthGetHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{"status": "available", "environment": os.Getenv("ENVIRONMENT")}

	err := writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}
}
