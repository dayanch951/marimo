# Observability Stack Guide

## Overview

Marimo ERP includes a complete observability stack for monitoring, tracing, and logging.

## Components

### 1. Prometheus (Metrics)
- **URL**: http://localhost:9090
- **Purpose**: Time-series metrics collection and alerting
- **Metrics Collected**:
  - HTTP request rate, latency, size
  - Authentication attempts
  - Database query performance
  - System resources (CPU, memory, goroutines)

### 2. Grafana (Visualization)
- **URL**: http://localhost:3001
- **Credentials**: admin / admin
- **Purpose**: Metrics visualization and dashboards
- **Dashboards**:
  - Overview: Request rate, response times, status codes
  - Authentication: Login attempts, token issuance
  - System: Memory, CPU, goroutines

### 3. Jaeger (Distributed Tracing)
- **URL**: http://localhost:16686
- **Purpose**: Distributed request tracing across microservices
- **Features**:
  - End-to-end request flow
  - Latency analysis
  - Dependency graphs

### 4. ELK Stack (Centralized Logging)
- **Elasticsearch**: http://localhost:9200
- **Kibana**: http://localhost:5601
- **Logstash**: tcp://localhost:5000
- **Purpose**: Centralized log aggregation and analysis

## Quick Start

### Start Monitoring Stack

```bash
# Start main services
docker-compose up -d

# Start monitoring stack
docker-compose -f docker-compose.monitoring.yml up -d

# Check status
docker-compose -f docker-compose.monitoring.yml ps
```

### Access Dashboards

1. **Grafana**: http://localhost:3001
   - Login: admin / admin
   - Navigate to Dashboards → Marimo ERP Overview

2. **Prometheus**: http://localhost:9090
   - View metrics and alerts
   - Query: `rate(http_requests_total[5m])`

3. **Jaeger**: http://localhost:16686
   - Search traces by service
   - Analyze request flows

4. **Kibana**: http://localhost:5601
   - View logs from all services
   - Create custom dashboards

## Metrics

### HTTP Metrics

```promql
# Request rate per service
rate(http_requests_total[5m])

# P95 response time
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Error rate
rate(http_requests_total{status=~"5.."}[5m])
```

### Business Metrics

```promql
# Active users
active_users_total

# Authentication success rate
rate(auth_attempts_total{result="success"}[5m]) / rate(auth_attempts_total[5m])

# Token issuance rate
rate(tokens_issued_total[5m])
```

### Database Metrics

```promql
# Query rate
rate(db_queries_total[5m])

# Query latency
histogram_quantile(0.95, rate(db_query_duration_seconds_bucket[5m]))
```

## Alerts

Prometheus monitors the following conditions:

- **HighErrorRate**: Error rate > 5% for 5 minutes
- **HighResponseTime**: P95 latency > 1s for 5 minutes
- **ServiceDown**: Service unavailable for 1 minute
- **HighMemoryUsage**: Memory usage > 90% for 5 minutes
- **DatabaseConnectionFailed**: High DB error rate
- **HighAuthFailureRate**: > 50% auth failures (possible attack)

## Health Checks

Each service exposes detailed health checks:

```bash
# Users service health
curl http://localhost:8081/health | jq

# Response format:
{
  "status": "healthy",
  "service": "users-service",
  "version": "1.0.0",
  "timestamp": "2024-01-01T12:00:00Z",
  "uptime": "2h30m15s",
  "checks": {
    "database": {
      "status": "healthy",
      "latency": "2.5ms"
    }
  },
  "system": {
    "go_version": "go1.21",
    "num_goroutines": 25,
    "mem_alloc_mb": 15.2,
    "mem_total_mb": 45.8,
    "num_cpu": 8
  }
}
```

## Tracing

Services automatically emit tracing spans to Jaeger.

### View Traces

1. Go to http://localhost:16686
2. Select service from dropdown
3. Click "Find Traces"
4. Analyze request flow and timing

### Trace Context

Traces include:
- Service name
- Operation name
- Duration
- HTTP method and endpoint
- Status code
- Error details (if any)

## Logging

### Log Format

All services use structured JSON logging:

```json
{
  "timestamp": "2024-01-01T12:00:00Z",
  "level": "INFO",
  "service": "users-service",
  "message": "User logged in successfully",
  "user_id": "123",
  "email": "user@example.com",
  "remote_addr": "192.168.1.1",
  "latency_ms": 125
}
```

### View Logs in Kibana

1. Go to http://localhost:5601
2. Navigate to Discover
3. Filter by service, level, or search terms
4. Create visualizations and dashboards

### Log Levels

- **DEBUG**: Detailed diagnostic information
- **INFO**: General informational messages
- **WARN**: Warning messages
- **ERROR**: Error messages
- **FATAL**: Critical errors

## Best Practices

### Metrics

1. **Use labels wisely**: Don't create high-cardinality labels
2. **Monitor SLIs**: Track Service Level Indicators (error rate, latency)
3. **Set up alerts**: Define thresholds for critical metrics
4. **Review dashboards**: Regularly check for anomalies

### Tracing

1. **Sample appropriately**: Use sampling for high-traffic services
2. **Add context**: Include relevant tags in spans
3. **Trace errors**: Always trace error paths
4. **Analyze bottlenecks**: Use traces to identify slow operations

### Logging

1. **Use structured logs**: Always log in JSON format
2. **Include context**: Add user IDs, request IDs, etc.
3. **Log levels**: Use appropriate levels
4. **Avoid PII**: Don't log sensitive information
5. **Log errors with stack traces**: Include full context for errors

## Troubleshooting

### High Memory Usage

```bash
# Check service memory
curl http://localhost:8081/health | jq '.system.mem_alloc_mb'

# View in Grafana
# Dashboard → System → Memory Usage
```

### Slow Requests

1. Check P95 latency in Grafana
2. Find slow traces in Jaeger
3. Analyze database query times
4. Check logs for errors

### Service Down

```bash
# Check service status
docker-compose ps

# View logs
docker-compose logs users

# Check health endpoint
curl http://localhost:8081/health
```

## Performance Tuning

### Prometheus

```yaml
# Adjust scrape interval
global:
  scrape_interval: 15s  # Increase for less load
```

### Elasticsearch

```yaml
# Adjust memory
environment:
  - "ES_JAVA_OPTS=-Xms1g -Xmx1g"  # Increase for better performance
```

### Logstash

```yaml
# Adjust batch size
output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    flush_size => 500
  }
}
```

## Cleanup

```bash
# Stop monitoring stack
docker-compose -f docker-compose.monitoring.yml down

# Remove volumes (WARNING: deletes all data)
docker-compose -f docker-compose.monitoring.yml down -v
```

## Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [Elasticsearch Documentation](https://www.elastic.co/guide/)
