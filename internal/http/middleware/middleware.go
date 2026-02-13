package middleware

import (
	"net/http"
	"time"

	"notion-notifier/internal/config"
	"notion-notifier/internal/logging"
)

// Logging wraps a handler with request logging.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		logging.Info("HTTP", "%s %s %d %s", r.Method, r.URL.Path, rec.status, time.Since(start).Truncate(time.Millisecond))
	})
}

// BasicAuth returns middleware that enforces HTTP Basic Authentication
// when enabled in env config. Credentials are sourced from env.yaml.
func BasicAuth(cfg *config.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, env := cfg.Get()
			if !env.Security.BasicAuth.Enabled {
				next.ServeHTTP(w, r)
				return
			}
			user, pass, ok := r.BasicAuth()
			if !ok || user != env.Security.BasicAuth.Username || pass != env.Security.BasicAuth.Password {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
