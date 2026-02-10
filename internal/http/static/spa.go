package static

import (
	"io/fs"
	"net/http"
	"strings"
)

// NewSPAHandler returns an http.Handler that serves an SPA from the given
// filesystem. Non-API paths that don't match a real file are served index.html
// so that client-side routing works correctly.
func NewSPAHandler(distFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(distFS))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// API routes are not handled here.
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Try to serve the actual file first.
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		if _, err := fs.Stat(distFS, path); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fallback to index.html for SPA client-side routing.
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
