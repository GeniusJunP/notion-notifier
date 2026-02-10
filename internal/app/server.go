package app

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"notion-notifier/internal/config"
	"notion-notifier/internal/db"
	"notion-notifier/internal/scheduler"
)

//go:embed web/templates/*.html
var templatesFS embed.FS

type Server struct {
	cfg       *config.Manager
	repo      *db.Repository
	scheduler *scheduler.Scheduler
	tmpl      *template.Template
}

func NewServer(cfg *config.Manager, repo *db.Repository, sched *scheduler.Scheduler) (*Server, error) {
	tmpl := template.New("base").Funcs(template.FuncMap{
		"json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
	})
	tmpl, err := tmpl.ParseFS(templatesFS, "web/templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Server{
		cfg:       cfg,
		repo:      repo,
		scheduler: sched,
		tmpl:      tmpl,
	}, nil
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	// Page routes
	mux.HandleFunc("/", s.requireAuth(s.handleDashboard))
	mux.HandleFunc("/notifications", s.requireAuth(s.handleNotifications))
	mux.HandleFunc("/calendar", s.requireAuth(s.handleCalendar))
	mux.HandleFunc("/settings", s.requireAuth(s.handleSettings))

	// API routes
	mux.HandleFunc("/api/sync", s.requireAuth(s.handleAPISync))
	mux.HandleFunc("/api/config", s.requireAuth(s.handleAPIConfig))
	mux.HandleFunc("/api/notifications/preview", s.requireAuth(s.handleAPIPreviewNotification))
	mux.HandleFunc("/api/notifications/manual", s.requireAuth(s.handleAPIManualNotification))
	mux.HandleFunc("/api/calendar/sync", s.requireAuth(s.handleAPICalendarSync))
	mux.HandleFunc("/api/calendar/clear", s.requireAuth(s.handleAPICalendarClear))
	mux.HandleFunc("/api/history/clear", s.requireAuth(s.handleAPIHistoryClear))

	return loggingMiddleware(mux)
}
