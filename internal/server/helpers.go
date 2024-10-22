package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"strings"
)

type envelope map[string]any

func readJSON[T any](w http.ResponseWriter, r *http.Request, dst *T) error {
	max_bytes := 1_048_576 // Limit the request body size to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, int64(max_bytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	// Check that the body only contains a single JSON value
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func writeJSON[T any](w http.ResponseWriter, status int, data T, headers http.Header) error {
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Insert the headers and write the JSON response
	maps.Insert(w.Header(), maps.All(headers))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func logError(logger *slog.Logger, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	logger.Error(err.Error(), "method", method, "uri", uri)
}

func errorResponse(logger *slog.Logger, w http.ResponseWriter, r *http.Request, status int, message any) {
	data := envelope{"error": message}

	err := writeJSON(w, status, data, nil)
	if err != nil {
		logError(logger, r, err)
		w.WriteHeader(500)
	}
}

func serverErrorResponse(logger *slog.Logger, w http.ResponseWriter, r *http.Request, err error) {
	logError(logger, r, err)
	msg := "the server encountered a problem and could not process your request"
	errorResponse(logger, w, r, http.StatusInternalServerError, msg)
}

func validationErrorResponse(logger *slog.Logger, w http.ResponseWriter, r *http.Request, errors map[string]string) {

	errorResponse(logger, w, r, http.StatusBadRequest, errors)
}
