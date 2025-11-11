package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"service", "method", "endpoint", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "endpoint", "status"},
	)

	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"service", "method", "endpoint"},
	)

	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"service", "method", "endpoint"},
	)

	// Business metrics
	activeUsers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users_total",
			Help: "Number of currently active users",
		},
	)

	authAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"service", "result"},
	)

	tokensIssued = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tokens_issued_total",
			Help: "Total number of tokens issued",
		},
		[]string{"service", "token_type"},
	)

	// Database metrics
	dbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"service", "operation", "status"},
	)

	dbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "operation"},
	)
)

// responseWriter wraps http.ResponseWriter to capture status code and size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// PrometheusMiddleware creates a middleware that records Prometheus metrics
func PrometheusMiddleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			// Record request size
			httpRequestSize.WithLabelValues(serviceName, r.Method, r.URL.Path).Observe(float64(r.ContentLength))

			// Process request
			next.ServeHTTP(wrapped, r)

			// Record metrics
			duration := time.Since(start).Seconds()
			status := strconv.Itoa(wrapped.statusCode)

			httpRequestsTotal.WithLabelValues(serviceName, r.Method, r.URL.Path, status).Inc()
			httpRequestDuration.WithLabelValues(serviceName, r.Method, r.URL.Path, status).Observe(duration)
			httpResponseSize.WithLabelValues(serviceName, r.Method, r.URL.Path).Observe(float64(wrapped.size))
		})
	}
}

// RecordAuthAttempt records an authentication attempt
func RecordAuthAttempt(service, result string) {
	authAttempts.WithLabelValues(service, result).Inc()
}

// RecordTokenIssued records a token issuance
func RecordTokenIssued(service, tokenType string) {
	tokensIssued.WithLabelValues(service, tokenType).Inc()
}

// RecordDBQuery records a database query
func RecordDBQuery(service, operation, status string, duration time.Duration) {
	dbQueriesTotal.WithLabelValues(service, operation, status).Inc()
	dbQueryDuration.WithLabelValues(service, operation).Observe(duration.Seconds())
}

// SetActiveUsers sets the number of active users
func SetActiveUsers(count float64) {
	activeUsers.Set(count)
}
