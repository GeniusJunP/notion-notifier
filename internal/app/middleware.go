package app

import (
	"net/http"
	"time"

	"notion-notifier/internal/logging"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, req)
		logging.Info("HTTP", "%s %s %d %s", req.Method, req.URL.Path, rec.status, time.Since(start).Truncate(time.Millisecond))
	})
}
