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
	"notion-notifier/internal/discord"
	"notion-notifier/internal/notion"
	"notion-notifier/internal/retry"
	"notion-notifier/internal/scheduler"
	tpl "notion-notifier/internal/template"
)

type App struct {
	cfg       *config.Manager
	repo      *db.Repository
	scheduler *scheduler.Scheduler
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
	discordClient := discord.New(httpClient, retryCfg)

	var calendarClient *calendar.Client
	if cfg.CalendarSync.Enabled && env.Google.CalendarID != "" && env.Google.ServiceAccountKey != "" {
		calendarClient, err = calendar.NewClient(context.Background(), env.Google.CalendarID, env.Google.ServiceAccountKey)
		if err != nil {
			return nil, fmt.Errorf("calendar client: %w", err)
		}
	}
	renderer := tpl.New()
	sched := scheduler.New(manager, repo, notionClient, discordClient, calendarClient, renderer)
	return &App{
		cfg:       manager,
		repo:      repo,
		scheduler: sched,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	a.scheduler.Start(ctx)
	<-ctx.Done()
	return nil
}

func (a *App) Close() error {
	a.scheduler.Stop()
	return a.repo.Close()
}
