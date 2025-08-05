package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/glebarez/sqlite"
	"github.com/gorilla/sessions"
	"github.com/nuric/go-api-template/controllers"
	"github.com/nuric/go-api-template/email"
	"github.com/nuric/go-api-template/middleware"
	"github.com/nuric/go-api-template/models"
	"github.com/nuric/go-api-template/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Config struct {
	PrettyLogOutput bool   `env:"PRETTY_LOG_OUTPUT" envDefault:"true"`
	Debug           bool   `env:"DEBUG" envDefault:"true"`
	Port            int    `env:"PORT" envDefault:"8080"`
	DBUrl           string `env:"DB_URL" envDefault:"data.db"`
	SessionSecret   string `env:"SESSION_SECRET" envDefault:"32-character-long-secret-key-abc"`
	CSRFSecret      string `env:"CSRF_SECRET" envDefault:"32-character-long-csrf-secret-key-xyz"`
}

func main() {
	// ---------------------------
	// Setup config
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatal().Err(err).Msg("Failed to parse environment variables")
	}
	// ---------------------------
	// Setup logging
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// ---------------------------
	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if cfg.PrettyLogOutput {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}
	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Interface("config", cfg).Msg("Configuration")
	}
	log.Debug().Msg("Debug mode enabled")
	// ---------------------------
	// Setup database connection
	db, err := gorm.Open(sqlite.Open(cfg.DBUrl), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	if err := db.AutoMigrate(&models.User{}, &models.Token{}); err != nil {
		log.Fatal().Err(err).Msg("Failed to auto-migrate database")
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
		Store:      ss,
		Emailer:    email.LogEmailer{},
		CSRFSecret: cfg.CSRFSecret,
		Debug:      cfg.Debug,
	}
	handler := controllers.Setup(config)
	// Middleware
	handler = middleware.NewRateLimiter(7, 14, 15*time.Minute).Limit(handler)
	handler = middleware.ZeroLoggerMetrics(handler)
	handler = middleware.Recover(handler)
	// ---------------------------
	server := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: handler,
	}
	go func() {
		log.Info().Str("httpAddr", server.Addr).Msg("HTTPAPI.Serve")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to listen on port 8080")
		}
	}()
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	sig := <-quit
	log.Info().Str("signal", sig.String()).Msg("Shutting down server...")
	// The default kubernetes grace period is 30 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}
	cancel()
	log.Info().Msg("Server stopped")

}
