package middleware

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dayanch951/marimo/shared/cache"
	"github.com/dayanch951/marimo/shared/discovery"
	"github.com/dayanch951/marimo/shared/resilience"
)

// ProxyConfig holds configuration for the reverse proxy
type ProxyConfig struct {
	ServiceRegistry *discovery.ServiceRegistry
	Cache           *cache.RedisCache
	CircuitBreakers map[string]*resilience.CircuitBreaker
	RetryPolicy     resilience.RetryPolicy
	CacheTTL        time.Duration
}

// ResilientProxy is a reverse proxy with circuit breaker, retry, and caching
type ResilientProxy struct {
	config  ProxyConfig
	mu      sync.RWMutex
	client  *http.Client
}

// NewResilientProxy creates a new resilient reverse proxy
func NewResilientProxy(config ProxyConfig) *ResilientProxy {
	if config.RetryPolicy.MaxAttempts == 0 {
		config.RetryPolicy = resilience.DefaultRetryPolicy()
	}

	if config.CacheTTL == 0 {
		config.CacheTTL = 5 * time.Minute
	}

	return &ResilientProxy{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProxyRequest proxies a request to a backend service with resilience features
func (rp *ResilientProxy) ProxyRequest(serviceName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get or create circuit breaker for this service
		cb := rp.getCircuitBreaker(serviceName)

		// Try to get from cache first (for GET requests only)
		if r.Method == http.MethodGet && rp.config.Cache != nil {
			cacheKey := fmt.Sprintf("proxy:%s:%s", serviceName, r.URL.Path)
			var cachedResponse CachedResponse

			err := rp.config.Cache.Get(cacheKey, &cachedResponse)
			if err == nil {
				log.Printf("Cache hit for %s", cacheKey)
				w.Header().Set("X-Cache", "HIT")
				w.Header().Set("Content-Type", cachedResponse.ContentType)
				w.WriteHeader(cachedResponse.StatusCode)
				w.Write(cachedResponse.Body)
				return
			}
		}

		// Execute request with circuit breaker
		err := cb.Execute(func() error {
			return rp.executeRequest(w, r, serviceName)
		})

		if err != nil {
			if err == resilience.ErrCircuitOpen {
				http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
				log.Printf("Circuit breaker open for service %s", serviceName)
				return
			}

			http.Error(w, "Service error", http.StatusBadGateway)
			log.Printf("Error proxying request to %s: %v", serviceName, err)
		}
	}
}

// executeRequest executes the actual HTTP request with retry logic
func (rp *ResilientProxy) executeRequest(w http.ResponseWriter, r *http.Request, serviceName string) error {
	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()

	var lastResp *http.Response
	var lastErr error

	// Retry logic
	err := resilience.Retry(ctx, rp.config.RetryPolicy, func() error {
		// Discover service address
		serviceURL, err := rp.config.ServiceRegistry.DiscoverService(serviceName)
		if err != nil {
			return fmt.Errorf("service discovery failed: %w", err)
		}

		// Build target URL
		targetURL := serviceURL + r.URL.Path
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}

		// Create new request
		proxyReq, err := http.NewRequestWithContext(ctx, r.Method, targetURL, r.Body)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		// Copy headers
		for key, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}

		// Add tracing headers
		proxyReq.Header.Set("X-Forwarded-For", r.RemoteAddr)
		proxyReq.Header.Set("X-Forwarded-Proto", r.URL.Scheme)
		proxyReq.Header.Set("X-Forwarded-Host", r.Host)

		// Execute request
		resp, err := rp.client.Do(proxyReq)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}

		lastResp = resp

		// Check if status code is retryable
		if resilience.IsRetryableHTTPStatus(resp.StatusCode) {
			resp.Body.Close()
			return fmt.Errorf("retryable status code: %d", resp.StatusCode)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("all retry attempts failed: %w", err)
	}

	defer lastResp.Body.Close()

	// Read response body
	body, err := io.ReadAll(lastResp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Cache successful GET responses
	if r.Method == http.MethodGet && lastResp.StatusCode == http.StatusOK && rp.config.Cache != nil {
		cacheKey := fmt.Sprintf("proxy:%s:%s", serviceName, r.URL.Path)
		cached := CachedResponse{
			StatusCode:  lastResp.StatusCode,
			Body:        body,
			ContentType: lastResp.Header.Get("Content-Type"),
		}

		if err := rp.config.Cache.Set(cacheKey, cached, rp.config.CacheTTL); err != nil {
			log.Printf("Failed to cache response: %v", err)
		}
	}

	// Copy response headers
	for key, values := range lastResp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.Header().Set("X-Cache", "MISS")
	w.WriteHeader(lastResp.StatusCode)
	w.Write(body)

	return nil
}

// getCircuitBreaker gets or creates a circuit breaker for a service
func (rp *ResilientProxy) getCircuitBreaker(serviceName string) *resilience.CircuitBreaker {
	rp.mu.RLock()
	if cb, exists := rp.config.CircuitBreakers[serviceName]; exists {
		rp.mu.RUnlock()
		return cb
	}
	rp.mu.RUnlock()

	rp.mu.Lock()
	defer rp.mu.Unlock()

	// Double-check after acquiring write lock
	if cb, exists := rp.config.CircuitBreakers[serviceName]; exists {
		return cb
	}

	// Create new circuit breaker
	cb := resilience.NewCircuitBreaker(resilience.Settings{
		Name:        serviceName,
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		Threshold:   5,
		FailureRate: 0.5,
		OnStateChange: func(name string, from, to resilience.State) {
			log.Printf("Circuit breaker %s changed state: %s -> %s", name, from, to)
		},
	})

	rp.config.CircuitBreakers[serviceName] = cb
	return cb
}

// CachedResponse represents a cached HTTP response
type CachedResponse struct {
	StatusCode  int    `json:"status_code"`
	Body        []byte `json:"body"`
	ContentType string `json:"content_type"`
}

// ServiceRouter routes requests to appropriate backend services
type ServiceRouter struct {
	proxy *ResilientProxy
	routes map[string]string // path prefix -> service name
}

// NewServiceRouter creates a new service router
func NewServiceRouter(proxy *ResilientProxy) *ServiceRouter {
	return &ServiceRouter{
		proxy: proxy,
		routes: make(map[string]string),
	}
}

// RegisterRoute registers a route mapping
func (sr *ServiceRouter) RegisterRoute(pathPrefix, serviceName string) {
	sr.routes[pathPrefix] = serviceName
	log.Printf("Registered route: %s -> %s", pathPrefix, serviceName)
}

// ServeHTTP implements http.Handler
func (sr *ServiceRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Find matching service
	for prefix, serviceName := range sr.routes {
		if strings.HasPrefix(r.URL.Path, prefix) {
			// Remove prefix from path
			r.URL.Path = strings.TrimPrefix(r.URL.Path, prefix)
			if !strings.HasPrefix(r.URL.Path, "/") {
				r.URL.Path = "/" + r.URL.Path
			}

			// Proxy the request
			sr.proxy.ProxyRequest(serviceName)(w, r)
			return
		}
	}

	// No matching route
	http.NotFound(w, r)
}
