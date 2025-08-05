package middleware

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/nuric/go-api-template/utils"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

// ---------------------------
// Zerolog based middleware for logging HTTP requests
func ZeroLoggerMetrics(next http.Handler) http.Handler {
	handler := hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Str("remote_ip", ClientIP(r)).
			Msg("")
	})(next)
	handler = hlog.NewHandler(log.Logger)(handler)
	return handler
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().Interface("error", err).Msg("panic recovered")
				log.Error().Str("stack", string(debug.Stack())).Msg("stack trace")
				w.WriteHeader(http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func APIKey(next http.Handler, key string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != key {
			utils.Encode(w, http.StatusProxyAuthRequired, map[string]string{"error": "forbidden"})
			return
		}
		next.ServeHTTP(w, r)
	})
}
