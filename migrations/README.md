# Database Migrations

SQL migrations for Marimo ERP system.

## Migration Files

| File | Description | Tables Created |
|------|-------------|----------------|
| `001_create_tenants.sql` | Multi-tenancy support | `tenants` |
| `002_create_analytics.sql` | Analytics and reporting | `analytics_queries`, `dashboards`, `scheduled_reports` |
| `003_create_webhooks.sql` | Webhook system | `webhooks`, `webhook_events`, `webhook_deliveries`, `webhook_logs` |
| `004_create_integrations.sql` | Third-party integrations | `integrations`, `integration_logs`, `stripe_customers`, `stripe_subscriptions`, `sendgrid_contacts` |
| `005_create_usage_tracking.sql` | Usage metrics and billing | `usage_metrics`, `user_activities`, `api_usage_logs` |

## Running Migrations

### Manual Execution

```bash
# Run all migrations in order
psql -U postgres -d marimo_erp -f migrations/001_create_tenants.sql
psql -U postgres -d marimo_erp -f migrations/002_create_analytics.sql
psql -U postgres -d marimo_erp -f migrations/003_create_webhooks.sql
psql -U postgres -d marimo_erp -f migrations/004_create_integrations.sql
psql -U postgres -d marimo_erp -f migrations/005_create_usage_tracking.sql
```

### Using Migration Script

```bash
./scripts/run-migrations.sh
```

### Docker

```bash
docker-compose exec postgres psql -U postgres -d marimo_erp -f /migrations/001_create_tenants.sql
```

## Migration Details

### 001: Tenants Table

Creates multi-tenancy foundation:
- Tenant identification (slug, custom domain)
- Subscription management (plan, status, Stripe IDs)
- Settings (max users, storage, features)
- Trial and suspension tracking
- Adds `tenant_id` to existing tables

**After running**: Update all tables to include `tenant_id` column.

### 002: Analytics Tables

Analytics and reporting infrastructure:
- `analytics_queries`: Saved custom queries
- `dashboards`: Custom dashboard configurations
- `scheduled_reports`: Automated report generation

### 003: Webhooks Tables

Webhook delivery system:
- `webhooks`: Endpoint configurations
- `webhook_events`: Events to dispatch
- `webhook_deliveries`: Delivery attempts and status
- `webhook_logs`: Detailed request/response logs

**Features**:
- Retry logic with exponential backoff
- HMAC signature verification
- Automatic cleanup of old logs (30 days)

### 004: Integrations Tables

Third-party integration management:
- `integrations`: Provider configurations
- `integration_logs`: API call logging
- `stripe_customers`: Stripe customer records
- `stripe_subscriptions`: Subscription tracking
- `sendgrid_contacts`: Email contact management

### 005: Usage Tracking Tables

Tenant usage monitoring:
- `usage_metrics`: Aggregated metrics (API calls, storage, etc.)
- `user_activities`: Detailed activity logs
- `api_usage_logs`: API request tracking

**Functions**:
- `aggregate_usage_metrics()`: Calculate aggregates
- `check_tenant_limit()`: Verify usage limits
- Automatic cleanup of old data

## Maintenance Functions

### Cleanup Old Data

```sql
-- Clean webhook logs (> 30 days)
SELECT cleanup_webhook_logs();

-- Clean integration logs (> 90 days)
SELECT cleanup_integration_logs();

-- Clean activity logs (> 180 days)
SELECT cleanup_activity_logs();
```

### Check Tenant Limits

```sql
SELECT * FROM check_tenant_limit(
    'tenant-uuid'::UUID,
    'users_count',
    50
);
```

### Aggregate Usage

```sql
SELECT * FROM aggregate_usage_metrics(
    'tenant-uuid'::UUID,
    'api_calls',
    '2024-01-01'::TIMESTAMP,
    '2024-01-31'::TIMESTAMP
);
```

## Indexes

All tables include appropriate indexes for:
- Tenant filtering (`tenant_id`)
- Time-based queries (`created_at`, `updated_at`)
- Status filtering
- Foreign key relationships
- Full-text search (where applicable)

## Triggers

Automatic `updated_at` timestamp updates for all relevant tables.

## Best Practices

1. **Always backup** before running migrations
2. **Test in staging** environment first
3. **Run migrations in transaction** (most are idempotent)
4. **Monitor performance** after adding indexes
5. **Schedule cleanup functions** as cron jobs

## Rollback

To rollback migrations, drop tables in reverse order:

```sql
DROP TABLE IF EXISTS api_usage_logs CASCADE;
DROP TABLE IF EXISTS user_activities CASCADE;
DROP TABLE IF EXISTS usage_metrics CASCADE;

DROP TABLE IF EXISTS sendgrid_contacts CASCADE;
DROP TABLE IF EXISTS stripe_subscriptions CASCADE;
DROP TABLE IF EXISTS stripe_customers CASCADE;
DROP TABLE IF EXISTS integration_logs CASCADE;
DROP TABLE IF EXISTS integrations CASCADE;

DROP TABLE IF EXISTS webhook_logs CASCADE;
DROP TABLE IF EXISTS webhook_deliveries CASCADE;
DROP TABLE IF EXISTS webhook_events CASCADE;
DROP TABLE IF EXISTS webhooks CASCADE;

DROP TABLE IF EXISTS scheduled_reports CASCADE;
DROP TABLE IF EXISTS dashboards CASCADE;
DROP TABLE IF EXISTS analytics_queries CASCADE;

DROP TABLE IF EXISTS tenants CASCADE;
```

## Monitoring

### Table Sizes

```sql
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

### Index Usage

```sql
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;
```

## Troubleshooting

**Migration fails with "relation already exists"**:
- Migrations are idempotent (use `IF NOT EXISTS`)
- Safe to re-run

**Performance issues after migration**:
- Run `ANALYZE` on new tables
- Check query plans with `EXPLAIN ANALYZE`
- Verify indexes are being used

**Foreign key constraint violations**:
- Ensure dependent data exists
- Check `tenant_id` is set correctly
- Verify cascading deletes are configured

## Support

For issues or questions:
- Check logs: `docker-compose logs postgres`
- Verify permissions: `\du` in psql
- Check table structure: `\d+ table_name`
