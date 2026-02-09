package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"notion-notifier/internal/app"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to config.yaml")
	envPath := flag.String("env", "env.yaml", "path to env.yaml")
	dbPath := flag.String("db", "data.db", "path to sqlite db")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := app.New(*cfgPath, *envPath, *dbPath)
	if err != nil {
		log.Fatalf("init failed: %v", err)
	}
	defer func() {
		if err := application.Close(); err != nil {
			log.Printf("shutdown error: %v", err)
		}
	}()

	if err := application.Start(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server failed: %v", err)
	}
}
