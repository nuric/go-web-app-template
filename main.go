package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/lmittmann/tint"

	"github.com/caarlos0/env/v11"
	"github.com/glebarez/sqlite"
	"github.com/gorilla/sessions"
	"github.com/nuric/go-api-template/controllers"
	"github.com/nuric/go-api-template/email"
	"github.com/nuric/go-api-template/middleware"
	"github.com/nuric/go-api-template/models"
	"github.com/nuric/go-api-template/storage"
	"github.com/nuric/go-api-template/utils"
	"gorm.io/gorm"
)

type Config struct {
	PrettyLogOutput bool   `env:"PRETTY_LOG_OUTPUT" envDefault:"true"`
	Debug           bool   `env:"DEBUG" envDefault:"true"`
	Port            int    `env:"PORT" envDefault:"8080"`
	DBUrl           string `env:"DB_URL" envDefault:"data.db"`
	SessionSecret   string `env:"SESSION_SECRET" envDefault:"32-character-long-secret-key-abc"`
	CSRFSecret      string `env:"CSRF_SECRET" envDefault:"32-character-long-csrf-secret-key-xyz"`
	DataFolder      string `env:"DATA_FOLDER" envDefault:"data"`
}

func main() {
	// ---------------------------
	// Setup config
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		slog.Error("Failed to parse environment variables", "error", err)
		os.Exit(1)
	}
	// ---------------------------
	// Setup logging
	// UNIX Time is faster and smaller than most timestamps
	logLevel := slog.LevelInfo
	if cfg.Debug {
		logLevel = slog.LevelDebug
	}
	if cfg.PrettyLogOutput {
		// Set global logger with custom options
		slog.SetDefault(slog.New(
			tint.NewHandler(os.Stdout, &tint.Options{
				Level:      logLevel,
				TimeFormat: time.Kitchen,
			}),
		))
	} else {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	}
	// ---------------------------
	// Setup database connection
	db, err := gorm.Open(sqlite.Open(cfg.DBUrl), &gorm.Config{})
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Token{}); err != nil {
		slog.Error("Failed to auto-migrate database", "error", err)
		os.Exit(1)
	}
	// ---------------------------
	// Our routes
	ss := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		utils.Encode(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	config := controllers.Config{
		Mux:        mux,
		Database:   db,
		Session:    ss,
		Emailer:    email.LogEmailer{},
		Storer:     storage.OsStorer{Path: cfg.DataFolder},
		CSRFSecret: cfg.CSRFSecret,
		Debug:      cfg.Debug,
	}
	handler := controllers.Setup(config)
	// Middleware
	handler = middleware.NewRateLimiter(7, 14, 15*time.Minute).Limit(handler)
	handler = middleware.Recover(handler)
	// ---------------------------
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: handler,
	}
	go func() {
		slog.Info("HTTPAPI.Serve", "httpAddr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Failed to listen on port 8080", "error", err)
			os.Exit(1)
		}
	}()
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	sig := <-quit
	slog.Info("Shutting down server...", "signal", sig.String())
	// The default kubernetes grace period is 30 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}
	cancel()
	slog.Info("Server stopped")
}
