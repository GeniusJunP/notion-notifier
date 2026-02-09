package server

import (
	"net/http"

	"notion-notifier/internal/config"
)

func (s *Server) basicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg, _ := s.cfg.Get()
		if !cfg.Security.BasicAuth.Enabled {
			next.ServeHTTP(w, r)
			return
		}
		username, password, ok := r.BasicAuth()
		if !ok || username != cfg.Security.BasicAuth.Username || password != cfg.Security.BasicAuth.Password {
			w.Header().Set("WWW-Authenticate", `Basic realm="notion-notifier"`)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func SanitizeConfig(cfg config.Config) config.Config {
	return cfg
}
