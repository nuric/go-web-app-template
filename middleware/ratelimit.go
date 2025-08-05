package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/time/rate"
)

func ClientIP(r *http.Request) string {
	// Assume a trusted proxy is in front the application like a load balancer
	// or reverse proxy. If you don't have such a proxy, you cannot trust the
	// X-Forwarded-For header and should use r.RemoteAddr directly.
	var ip string
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ip = strings.TrimSpace(strings.SplitN(xff, ",", 2)[0])
	} else {
		var err error
		ip, _, err = net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}
	}
	// Parse and normalise the IP address
	netIP := net.ParseIP(ip)
	if netIP == nil {
		log.Warn().Str("ip", ip).Msg("failed to parse IP address")
		return "unknown"
	}
	return netIP.String()
}

// A client holds the rate limiter and the last seen time for a given IP.
type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter stores the rate limiters for each client.
type RateLimiter struct {
	mu      sync.Mutex
	clients map[string]*client
	rate    rate.Limit
	burst   int
	expiry  time.Duration
}

// NewRateLimiter creates a new rate limiter with a background cleanup goroutine.
func NewRateLimiter(eventsPerSecond float64, burst int, expiry time.Duration) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*client),
		rate:    rate.Limit(eventsPerSecond),
		burst:   burst,
		expiry:  expiry,
	}

	// Start a background goroutine to run cleanup periodically.
	go rl.backgroundCleanup()

	return rl
}

func (rl *RateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := ClientIP(r)

		rl.mu.Lock()
		c, exists := rl.clients[ip]
		if !exists {
			// Create a new limiter for the client.
			c = &client{
				limiter: rate.NewLimiter(rl.rate, rl.burst),
			}
			rl.clients[ip] = c
		}
		// Update the last seen time on every request.
		c.lastSeen = time.Now()
		rl.mu.Unlock()

		if !c.limiter.Allow() {
			log.Warn().Str("ip", ip).Msg("rate limit exceeded")
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// backgroundCleanup runs periodically to remove expired clients.
func (rl *RateLimiter) backgroundCleanup() {
	ticker := time.NewTicker(rl.expiry)
	defer ticker.Stop()

	log.Debug().Msg("Starting background cleanup for rate limiter")
	for range ticker.C {
		rl.mu.Lock()
		log.Debug().Int("clients", len(rl.clients)).Msg("Running background cleanup for rate limiter")
		for ip, c := range rl.clients {
			// If the client hasn't been seen in the expiry window, delete it.
			if time.Since(c.lastSeen) > rl.expiry {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}
