// Package app wires together all dependencies and manages the application lifecycle.
// It does NOT contain HTTP handlers, templates, or business logic.
package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"notion-notifier/internal/calendar"
	"notion-notifier/internal/config"
	"notion-notifier/internal/db"
	httpapi "notion-notifier/internal/http/api"
	"notion-notifier/internal/http/middleware"
	"notion-notifier/internal/http/static"
	"notion-notifier/internal/notion"
	"notion-notifier/internal/retry"
	"notion-notifier/internal/scheduler"
	tpl "notion-notifier/internal/template"
	"notion-notifier/internal/webhook"
	"notion-notifier/web"
)

// App is the top-level application container.
type App struct {
	cfg       *config.Manager
	repo      *db.Repository
	scheduler *scheduler.Scheduler
	server    *http.Server
}

// New creates a fully-wired App ready to Start.
func New(cfgPath, envPath, dbPath string) (*App, error) {
	if cfgPath == "" || envPath == "" || dbPath == "" {
		return nil, errors.New("config, env, and db paths are required")
	}

	// Config
	manager, err := config.NewManager(cfgPath, envPath)
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	// Database
	repo, err := db.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("db: %w", err)
	}

	// External clients
	cfg, env := manager.Get()
	httpClient := &http.Client{Timeout: 20 * time.Second}
	retryCfg := retry.Config{}

	notionClient := notion.New(httpClient, env.Notion.APIKey, retryCfg)
	webhookClient := webhook.New(httpClient, retryCfg)

	var calendarClient *calendar.Client
	if cfg.CalendarSync.Enabled && env.Google.CalendarID != "" && env.Google.ServiceAccountKey != "" {
		calendarClient, err = calendar.NewClient(context.Background(), env.Google.CalendarID, env.Google.ServiceAccountKey)
		if err != nil {
			return nil, fmt.Errorf("calendar client: %w", err)
		}
	}

	renderer := tpl.New()
	sched := scheduler.New(manager, repo, notionClient, webhookClient, calendarClient, renderer)

	// HTTP Router
	handler := buildRouter(manager, repo, sched)

	httpSrv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	return &App{
		cfg:       manager,
		repo:      repo,
		scheduler: sched,
		server:    httpSrv,
	}, nil
}

// buildRouter creates the http.Handler with middleware and routes.
func buildRouter(cfg *config.Manager, repo *db.Repository, sched *scheduler.Scheduler) http.Handler {
	mux := http.NewServeMux()

	// API routes
	apiHandler := httpapi.NewHandler(cfg, repo, sched)
	apiHandler.Register(mux)

	// SPA static files
	distFS, err := fs.Sub(web.DistFS, "dist")
	if err != nil {
		// Should never happen with embedded FS.
		panic("failed to sub web/dist: " + err.Error())
	}
	spaHandler := static.NewSPAHandler(distFS)
	mux.Handle("/", spaHandler)

	// Middleware chain: Logging → BasicAuth → Router
	var handler http.Handler = mux
	handler = middleware.BasicAuth(cfg)(handler)
	handler = middleware.Logging(handler)
	return handler
}

// Start begins the scheduler and HTTP server. Blocks until ctx is done.
func (a *App) Start(ctx context.Context) error {
	a.scheduler.Start(ctx)

	go func() {
		fmt.Printf("Starting server on %s\n", a.server.Addr)
		fmt.Printf("URL: http://localhost%s\n", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()

	<-ctx.Done()
	return nil
}

// Close gracefully shuts down the scheduler, HTTP server, and database.
func (a *App) Close() error {
	a.scheduler.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = a.server.Shutdown(ctx)
	return a.repo.Close()
}
