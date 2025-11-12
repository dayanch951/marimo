package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	// HTTP metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPResponseSize    *prometheus.HistogramVec

	// Database metrics
	DBQueriesTotal    *prometheus.CounterVec
	DBQueryDuration   *prometheus.HistogramVec
	DBConnectionsOpen prometheus.Gauge
	DBConnectionsIdle prometheus.Gauge

	// Analytics metrics
	AnalyticsQueriesTotal    *prometheus.CounterVec
	AnalyticsQueryDuration   *prometheus.HistogramVec
	AnalyticsResultSize      *prometheus.HistogramVec
	AnalyticsCacheHits       *prometheus.CounterVec
	AnalyticsCacheMisses     *prometheus.CounterVec

	// Webhook metrics
	WebhooksDispatched       *prometheus.CounterVec
	WebhookDeliveriesTotal   *prometheus.CounterVec
	WebhookDeliveryDuration  *prometheus.HistogramVec
	WebhookRetries           *prometheus.CounterVec
	WebhookFailures          *prometheus.CounterVec

	// Integration metrics
	IntegrationCallsTotal    *prometheus.CounterVec
	IntegrationCallDuration  *prometheus.HistogramVec
	IntegrationErrors        *prometheus.CounterVec

	// Tenant metrics
	TenantsTotal             prometheus.Gauge
	TenantsActive            prometheus.Gauge
	TenantUsageMetrics       *prometheus.GaugeVec

	// API Gateway metrics
	APIRateLimitExceeded     *prometheus.CounterVec
	APICircuitBreakerOpen    *prometheus.GaugeVec

	// WebSocket metrics
	WebSocketConnectionsActive prometheus.Gauge
	WebSocketMessagesTotal     *prometheus.CounterVec
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics() *Metrics {
	return &Metrics{
		// HTTP metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		HTTPResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: []float64{100, 1000, 10000, 100000, 1000000},
			},
			[]string{"method", "endpoint"},
		),

		// Database metrics
		DBQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "db_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table", "status"},
		),
		DBQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "db_query_duration_seconds",
				Help:    "Database query duration in seconds",
				Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 5},
			},
			[]string{"operation", "table"},
		),
		DBConnectionsOpen: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_open",
				Help: "Number of open database connections",
			},
		),
		DBConnectionsIdle: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "db_connections_idle",
				Help: "Number of idle database connections",
			},
		),

		// Analytics metrics
		AnalyticsQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "analytics_queries_total",
				Help: "Total number of analytics queries executed",
			},
			[]string{"tenant_id", "query_type", "status"},
		),
		AnalyticsQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "analytics_query_duration_seconds",
				Help:    "Analytics query execution duration in seconds",
				Buckets: []float64{0.1, 0.5, 1, 5, 10, 30},
			},
			[]string{"tenant_id", "query_type"},
		),
		AnalyticsResultSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "analytics_result_size_rows",
				Help:    "Number of rows in analytics query result",
				Buckets: []float64{10, 100, 1000, 10000, 100000},
			},
			[]string{"tenant_id", "query_type"},
		),
		AnalyticsCacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "analytics_cache_hits_total",
				Help: "Total number of analytics cache hits",
			},
			[]string{"tenant_id"},
		),
		AnalyticsCacheMisses: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "analytics_cache_misses_total",
				Help: "Total number of analytics cache misses",
			},
			[]string{"tenant_id"},
		),

		// Webhook metrics
		WebhooksDispatched: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "webhooks_dispatched_total",
				Help: "Total number of webhooks dispatched",
			},
			[]string{"tenant_id", "event_type"},
		),
		WebhookDeliveriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "webhook_deliveries_total",
				Help: "Total number of webhook delivery attempts",
			},
			[]string{"tenant_id", "webhook_id", "status"},
		),
		WebhookDeliveryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "webhook_delivery_duration_seconds",
				Help:    "Webhook delivery duration in seconds",
				Buckets: []float64{0.1, 0.5, 1, 5, 10, 30},
			},
			[]string{"tenant_id", "webhook_id"},
		),
		WebhookRetries: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "webhook_retries_total",
				Help: "Total number of webhook retry attempts",
			},
			[]string{"tenant_id", "webhook_id", "attempt"},
		),
		WebhookFailures: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "webhook_failures_total",
				Help: "Total number of webhook failures",
			},
			[]string{"tenant_id", "webhook_id", "error_type"},
		),

		// Integration metrics
		IntegrationCallsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "integration_calls_total",
				Help: "Total number of third-party integration API calls",
			},
			[]string{"tenant_id", "provider", "action", "status"},
		),
		IntegrationCallDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "integration_call_duration_seconds",
				Help:    "Integration API call duration in seconds",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
			},
			[]string{"tenant_id", "provider", "action"},
		),
		IntegrationErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "integration_errors_total",
				Help: "Total number of integration errors",
			},
			[]string{"tenant_id", "provider", "error_type"},
		),

		// Tenant metrics
		TenantsTotal: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "tenants_total",
				Help: "Total number of tenants",
			},
		),
		TenantsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "tenants_active",
				Help: "Number of active tenants",
			},
		),
		TenantUsageMetrics: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "tenant_usage",
				Help: "Tenant usage metrics",
			},
			[]string{"tenant_id", "metric_type"},
		),

		// API Gateway metrics
		APIRateLimitExceeded: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "api_rate_limit_exceeded_total",
				Help: "Total number of rate limit exceeded events",
			},
			[]string{"tenant_id", "endpoint"},
		),
		APICircuitBreakerOpen: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "api_circuit_breaker_open",
				Help: "Circuit breaker state (1=open, 0=closed)",
			},
			[]string{"service"},
		),

		// WebSocket metrics
		WebSocketConnectionsActive: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "websocket_connections_active",
				Help: "Number of active WebSocket connections",
			},
		),
		WebSocketMessagesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "websocket_messages_total",
				Help: "Total number of WebSocket messages",
			},
			[]string{"tenant_id", "type", "direction"},
		),
	}
}
