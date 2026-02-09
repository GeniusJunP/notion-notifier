package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"notion-notifier/internal/calendar"
	"notion-notifier/internal/config"
	"notion-notifier/internal/db"
	"notion-notifier/internal/notion"
	"notion-notifier/internal/retry"
	"notion-notifier/internal/scheduler"
	tpl "notion-notifier/internal/template"
	"notion-notifier/internal/webhook"
)

type App struct {
	cfg       *config.Manager
	repo      *db.Repository
	scheduler *scheduler.Scheduler
	server    *http.Server
}

func New(cfgPath, envPath, dbPath string) (*App, error) {
	if cfgPath == "" || envPath == "" || dbPath == "" {
		return nil, errors.New("config, env, and db paths are required")
	}
	manager, err := config.NewManager(cfgPath, envPath)
	if err != nil {
		return nil, err
	}
	repo, err := db.Open(dbPath)
	if err != nil {
		return nil, err
	}
	cfg, env := manager.Get()
	retryCfg := retry.Config{}
	httpClient := &http.Client{Timeout: 20 * time.Second}

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

	srv, err := NewServer(manager, repo, sched)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	httpSrv := &http.Server{
		Addr:    ":8080",
		Handler: srv.Routes(),
	}

	return &App{
		cfg:       manager,
		repo:      repo,
		scheduler: sched,
		server:    httpSrv,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	a.scheduler.Start(ctx)
	go func() {
		fmt.Printf("Starting dashboard server on %s\n", a.server.Addr)
		fmt.Printf("Dashboard URL: http://localhost%s\n", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Printf("HTTP server error: %v\n", err)
		}
	}()
	<-ctx.Done()
	return nil
}

func (a *App) Close() error {
	a.scheduler.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = a.server.Shutdown(ctx)
	return a.repo.Close()
}
