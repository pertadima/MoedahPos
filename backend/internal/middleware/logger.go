package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// Logger is a structured HTTP request logging middleware.
func Logger(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap ResponseWriter to capture status code.
			wrapped := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", wrapped.status).
				Dur("latency", time.Since(start)).
				Str("remote_addr", r.RemoteAddr).
				Msg("request")
		})
	}
}

// statusRecorder wraps http.ResponseWriter to capture the response status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}
