package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	rate     int           // requests per minute
	burst    int           // max burst size
	cleanup  time.Duration // cleanup interval
}

// Visitor represents a rate limit visitor
type Visitor struct {
	tokens     float64
	lastSeen   time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
// rate: requests per minute
// burst: maximum burst size
func NewRateLimiter(rate, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     rate,
		burst:    burst,
		cleanup:  5 * time.Minute,
	}

	// Start cleanup goroutine
	go rl.cleanupVisitors()

	return rl
}

// getVisitor returns or creates a visitor for the given key
func (rl *RateLimiter) getVisitor(key string) *Visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[key]
	if !exists {
		v = &Visitor{
			tokens:   float64(rl.burst),
			lastSeen: time.Now(),
		}
		rl.visitors[key] = v
	}

	return v
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	v := rl.getVisitor(key)
	v.mu.Lock()
	defer v.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(v.lastSeen).Seconds()
	v.lastSeen = now

	// Refill tokens based on elapsed time
	tokensToAdd := elapsed * (float64(rl.rate) / 60.0)
	v.tokens += tokensToAdd
	if v.tokens > float64(rl.burst) {
		v.tokens = float64(rl.burst)
	}

	// Check if we have at least one token
	if v.tokens >= 1.0 {
		v.tokens -= 1.0
		return true
	}

	return false
}

// cleanupVisitors removes stale visitors
func (rl *RateLimiter) cleanupVisitors() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, v := range rl.visitors {
			v.mu.Lock()
			if now.Sub(v.lastSeen) > rl.cleanup {
				delete(rl.visitors, key)
			}
			v.mu.Unlock()
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Use IP address as the key
			key := r.RemoteAddr

			// For proxied requests, use X-Forwarded-For or X-Real-IP
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				key = xff
			} else if xri := r.Header.Get("X-Real-IP"); xri != "" {
				key = xri
			}

			if !limiter.Allow(key) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"success": false, "message": "Rate limit exceeded. Please try again later."}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// EndpointRateLimiter allows different limits for different endpoints
type EndpointRateLimiter struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
	default  *RateLimiter
}

// NewEndpointRateLimiter creates a rate limiter with per-endpoint limits
func NewEndpointRateLimiter(defaultRate, defaultBurst int) *EndpointRateLimiter {
	return &EndpointRateLimiter{
		limiters: make(map[string]*RateLimiter),
		default:  NewRateLimiter(defaultRate, defaultBurst),
	}
}

// AddEndpoint adds a specific rate limit for an endpoint
func (erl *EndpointRateLimiter) AddEndpoint(path string, rate, burst int) {
	erl.mu.Lock()
	defer erl.mu.Unlock()
	erl.limiters[path] = NewRateLimiter(rate, burst)
}

// GetLimiter returns the limiter for a given path
func (erl *EndpointRateLimiter) GetLimiter(path string) *RateLimiter {
	erl.mu.RLock()
	defer erl.mu.RUnlock()

	if limiter, exists := erl.limiters[path]; exists {
		return limiter
	}
	return erl.default
}

// Middleware creates middleware for endpoint-specific rate limiting
func (erl *EndpointRateLimiter) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limiter := erl.GetLimiter(r.URL.Path)

			// Use IP address as the key
			key := r.RemoteAddr
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				key = xff
			} else if xri := r.Header.Get("X-Real-IP"); xri != "" {
				key = xri
			}

			if !limiter.Allow(key) {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"success": false, "message": "Rate limit exceeded. Please try again later."}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
