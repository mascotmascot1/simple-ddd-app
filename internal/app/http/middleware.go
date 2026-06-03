package http

import (
	"log"
	"net/http"
)

// withLogging returns a middleware that logs each incoming request.
// It logs request method, URI and remote address.
func WithLogging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Printf("Request: %s %s from %s", r.Method, r.RequestURI, r.RemoteAddr)
			next.ServeHTTP(w, r)
		})
	}
}
