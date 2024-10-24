package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"questionify/internal/data"
	"questionify/internal/validator"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func addRoutes(router *httprouter.Router, logger *slog.Logger, modelStore *data.ModelStore) http.Handler {
	router.MethodNotAllowed = methodNotAllowed(logger)
	router.NotFound = notFound(logger)

	// Healthcheck
	router.Handler(http.MethodGet, "/v1/healthcheck", healthCheckGet())

	// Users
	router.Handler(http.MethodPost, "/v1/users", registerUserPost(logger, modelStore))
	router.Handler(http.MethodPost, "/v1/tokens/authentication", createAuthenticationToken(logger, modelStore))

	standard := alice.New(
		recoverPanic(logger),
		enableCORS(),
		logRequest(logger),
	)

	return standard.Then(router)
}

func notFound(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := "the requested resource could not be found"
		errorResponse(logger, w, r, http.StatusNotFound, msg)
	})
}

func methodNotAllowed(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := fmt.Sprintf("the %s method is not supported for this resource", r.Method)
		errorResponse(logger, w, r, http.StatusMethodNotAllowed, msg)
	})
}

func healthCheckGet() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		env := os.Getenv("ENVIRONMENT")
		if env == "" {
			env = "development"
		}

		data := map[string]string{"status": "available", "environment": env}

		err := writeJSON(w, http.StatusOK, data, nil)
		if err != nil {
			http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
		}
	})
}

func registerUserPost(logger *slog.Logger, modelStore *data.ModelStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var input struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		err := readJSON(w, r, &input)
		if err != nil {
			errorResponse(logger, w, r, http.StatusBadRequest, err.Error())
			return
		}

		user := &data.User{
			Name:  input.Name,
			Email: input.Email,
		}

		err = user.Password.Set(input.Password)
		if err != nil {
			errorResponse(logger, w, r, http.StatusBadRequest, err.Error())
			return
		}

		v := validator.New()
		if data.ValidateUser(v, user); !v.Valid() {
			validationErrorResponse(logger, w, r, v.Errors)
			return
		}

		err = modelStore.Users.Insert(user)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrDuplicateEmail):
				v.AddError("email", "a user with this email address already exists")
				validationErrorResponse(logger, w, r, v.Errors)
			default:
				serverErrorResponse(logger, w, r, err)
			}
			return
		}

		err = writeJSON(w, http.StatusCreated, user, nil)
		if err != nil {
			serverErrorResponse(logger, w, r, err)
		}
	})
}

func createAuthenticationToken(logger *slog.Logger, modelStore *data.ModelStore) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		err := readJSON(w, r, &input)
		if err != nil {
			errorResponse(logger, w, r, http.StatusBadRequest, err.Error())
			return
		}

		v := validator.New()
		data.ValidateEmail(v, input.Email)
		data.ValidatePasswordPlaintext(v, input.Password)

		if !v.Valid() {
			validationErrorResponse(logger, w, r, v.Errors)
			return
		}

		user, err := modelStore.Users.GetByEmail(input.Email)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				v.AddError("email", "no matching email address found")
				validationErrorResponse(logger, w, r, v.Errors)
			default:
				serverErrorResponse(logger, w, r, err)
			}
			return
		}

		match, err := user.Password.Matches(input.Password)
		if err != nil {
			serverErrorResponse(logger, w, r, err)
			return
		}

		if !match {
			v.AddError("password", "invalid password")
			validationErrorResponse(logger, w, r, v.Errors)
			return
		}

		// If the password matches, generate a new token and send it back to the client in a JSON response.
		token, err := modelStore.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
		if err != nil {
			serverErrorResponse(logger, w, r, err)
			return
		}

		err = writeJSON(w, http.StatusCreated, token, nil)
		if err != nil {
			serverErrorResponse(logger, w, r, err)
		}
	})
}
