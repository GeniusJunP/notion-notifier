// Package app wires together all dependencies and manages the application lifecycle.
// It does NOT contain HTTP handlers, templates, or business logic.
package app

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io/fs"
	"math/big"
	"net"
	"net/http"
	"time"

	"notion-notifier/internal/config"
	"notion-notifier/internal/db"
	httpapi "notion-notifier/internal/http/api"
	"notion-notifier/internal/http/middleware"
	"notion-notifier/internal/http/static"
	"notion-notifier/internal/logging"
	"notion-notifier/internal/notion"
	"notion-notifier/internal/retry"
	"notion-notifier/internal/scheduler"
	"notion-notifier/internal/webhook"
	"notion-notifier/web"
)

// App is the top-level application container.
type App struct {
	cfg       *config.Manager
	repo      *db.Repository
	scheduler *scheduler.Scheduler
	server    *http.Server
	tls       tlsRuntime
}

type tlsRuntime struct {
	certFile string
	keyFile  string
	selfSign bool
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
	env := manager.Env()
	httpClient := &http.Client{Timeout: 20 * time.Second}
	retryCfg := retry.Config{}

	notionClient := notion.New(httpClient, env.Notion.APIKey, retryCfg)
	webhookClient := webhook.New(httpClient, retryCfg)

	sched := scheduler.New(manager, repo, notionClient, webhookClient, nil)

	// HTTP Router
	handler := buildRouter(manager, repo, sched)
	addr, tlsCfg, err := resolveServerRuntime(env)
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	httpSrv := &http.Server{
		Addr:    addr,
		Handler: handler,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}
	if tlsCfg.selfSign {
		cert, err := generateSelfSignedCert()
		if err != nil {
			return nil, fmt.Errorf("self-signed cert: %w", err)
		}
		httpSrv.TLSConfig.Certificates = []tls.Certificate{cert}
	}

	return &App{
		cfg:       manager,
		repo:      repo,
		scheduler: sched,
		server:    httpSrv,
		tls:       tlsCfg,
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

func resolveServerRuntime(env config.Env) (string, tlsRuntime, error) {
	port := env.Server.Port
	if port == 0 {
		port = 18080
	}
	if port < 1 || port > 65535 {
		return "", tlsRuntime{}, fmt.Errorf("server.port must be between 1 and 65535: %d", port)
	}
	tlsCfg := tlsRuntime{
		certFile: env.Server.TLS.CertFile,
		keyFile:  env.Server.TLS.KeyFile,
	}
	if tlsCfg.certFile == "" && tlsCfg.keyFile == "" {
		tlsCfg.selfSign = true
	} else if tlsCfg.certFile == "" || tlsCfg.keyFile == "" {
		return "", tlsRuntime{}, errors.New("set both server.tls.cert_file and server.tls.key_file, or leave both empty for self-signed")
	}
	return fmt.Sprintf(":%d", port), tlsCfg, nil
}

func generateSelfSignedCert() (tls.Certificate, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}
	serialLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialLimit)
	if err != nil {
		return tls.Certificate{}, err
	}
	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore:   time.Now().Add(-1 * time.Hour),
		NotAfter:    time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	return tls.X509KeyPair(certPEM, keyPEM)
}

// Start begins the scheduler and HTTP server. Blocks until ctx is done.
func (a *App) Start(ctx context.Context) error {
	a.scheduler.Start(ctx)

	go func() {
		logging.Info("MAIN", "starting server on %s", a.server.Addr)
		run := func() error { return a.server.ListenAndServeTLS(a.tls.certFile, a.tls.keyFile) }
		if a.tls.selfSign {
			logging.Info("MAIN", "TLS: using generated self-signed certificate for localhost")
			run = func() error { return a.server.ListenAndServeTLS("", "") }
		}
		logging.Info("MAIN", "URL: https://localhost%s", a.server.Addr)
		if err := run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logging.Error("MAIN", "HTTP server error: %v", err)
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
	if err := a.server.Shutdown(ctx); err != nil {
		logging.Error("MAIN", "http shutdown error: %v", err)
	}
	return a.repo.Close()
}
