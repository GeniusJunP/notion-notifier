package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kardianos/service"

	"notion-notifier/internal/app"
	"notion-notifier/internal/logging"
)

// program implements service.Interface for kardianos/service.
type program struct {
	cfgPath string
	envPath string
	dbPath  string

	app    *app.App
	cancel context.CancelFunc
	done   chan struct{}
}

func (p *program) Start(s service.Service) error {
	application, err := app.New(p.cfgPath, p.envPath, p.dbPath)
	if err != nil {
		return fmt.Errorf("init failed: %w", err)
	}
	p.app = application

	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	p.done = make(chan struct{})

	go func() {
		defer close(p.done)
		_ = p.app.Start(ctx)
	}()

	return nil
}

func (p *program) Stop(s service.Service) error {
	if p.cancel != nil {
		p.cancel()
	}
	if p.done != nil {
		<-p.done
	}
	if p.app != nil {
		return p.app.Close()
	}
	return nil
}

func main() {
	cfgDir := app.DefaultConfigDir()
	dataDir := app.DefaultDataDir()

	cfgPath := flag.String("config", filepath.Join(cfgDir, "config.yaml"), "path to config.yaml")
	envPath := flag.String("env", filepath.Join(cfgDir, "env.yaml"), "path to env.yaml")
	dbPath := flag.String("db", filepath.Join(dataDir, "data.db"), "path to sqlite db")
	flag.Parse()

	// Ensure default directories and starter config files exist.
	if err := app.EnsureDefaults(cfgDir, dataDir); err != nil {
		logging.Error("INIT", "failed to create defaults: %v", err)
	}

	svcConfig := &service.Config{
		Name:        "notion-notifier",
		DisplayName: "Notion Notifier",
		Description: "Syncs Notion database and sends webhook notifications.",
		Arguments:   []string{"-config", *cfgPath, "-env", *envPath, "-db", *dbPath},
	}

	prg := &program{
		cfgPath: *cfgPath,
		envPath: *envPath,
		dbPath:  *dbPath,
	}

	s, err := service.New(prg, svcConfig)
	if err != nil {
		logging.Fatal("MAIN", "service init failed: %v", err)
	}

	// Handle service control commands: install, start, stop, uninstall
	if cmd := flag.Arg(0); cmd != "" {
		switch cmd {
		case "install", "start", "stop", "restart", "uninstall":
			if err := service.Control(s, cmd); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("notion-notifier: %s succeeded\n", cmd)
			return
		default:
			fmt.Fprintf(os.Stderr, "Unknown command: %s\nUsage: notion-notifier [install|start|stop|restart|uninstall]\n", cmd)
			os.Exit(1)
		}
	}

	// No command → run interactively (foreground).
	if err := s.Run(); err != nil {
		logging.Fatal("MAIN", "service run failed: %v", err)
	}
}
