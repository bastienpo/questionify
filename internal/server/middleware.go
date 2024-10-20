package server

import (
	"net/http"
)

func recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.Header().Set("Connection", "close")
				// app.serverErrorResponse(w, r, fmt.Errorf("%s", rec))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// var (
		// 	ip     = r.RemoteAddr
		// 	proto  = r.Proto
		// 	method = r.Method
		// 	uri    = r.URL.RequestURI()
		// )

		// app.logger.Info("received request", "ip", ip, "proto", proto, "method", method, "uri", uri)

		next.ServeHTTP(w, r)
	})
}
