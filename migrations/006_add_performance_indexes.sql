-- Up
-- Performance indexes for Marimo ERP database
-- This migration adds optimized indexes to improve query performance

-- ====================
-- Users table indexes
-- ====================

-- Composite index for tenant filtering and email lookup
CREATE INDEX IF NOT EXISTS idx_users_tenant_email
ON users(tenant_id, email);

-- Index for tenant filtering and role-based queries
CREATE INDEX IF NOT EXISTS idx_users_tenant_role
ON users(tenant_id, role);

-- Index for tenant filtering with created_at for sorting
CREATE INDEX IF NOT EXISTS idx_users_tenant_created
ON users(tenant_id, created_at DESC);

-- Index for status filtering
CREATE INDEX IF NOT EXISTS idx_users_status
ON users(status)
WHERE status IS NOT NULL;

-- Partial index for active users only (reduces index size)
CREATE INDEX IF NOT EXISTS idx_users_active
ON users(tenant_id, email)
WHERE status = 'active';

-- ====================
-- Tenants table indexes
-- ====================

-- Index for slug lookup (unique already handles this, but explicit for clarity)
CREATE INDEX IF NOT EXISTS idx_tenants_slug
ON tenants(slug);

-- Index for custom domain lookup
CREATE INDEX IF NOT EXISTS idx_tenants_domain
ON tenants(domain)
WHERE domain IS NOT NULL;

-- Index for status filtering
CREATE INDEX IF NOT EXISTS idx_tenants_status
ON tenants(status);

-- Index for subscription queries
CREATE INDEX IF NOT EXISTS idx_tenants_subscription_status
ON tenants((subscription->>'status'));

-- Index for trial expiration queries
CREATE INDEX IF NOT EXISTS idx_tenants_trial_ends
ON tenants(trial_ends_at)
WHERE trial_ends_at IS NOT NULL AND status = 'trial';

-- ====================
-- Analytics tables indexes
-- ====================

-- Index for analytics queries by tenant
CREATE INDEX IF NOT EXISTS idx_analytics_queries_tenant
ON analytics_queries(tenant_id, created_at DESC);

-- Index for analytics queries by creator
CREATE INDEX IF NOT EXISTS idx_analytics_queries_creator
ON analytics_queries(created_by, created_at DESC);

-- Index for dashboards by tenant
CREATE INDEX IF NOT EXISTS idx_dashboards_tenant
ON dashboards(tenant_id, created_at DESC);

-- Index for scheduled reports
CREATE INDEX IF NOT EXISTS idx_scheduled_reports_next_run
ON scheduled_reports(next_run_at)
WHERE enabled = true AND next_run_at IS NOT NULL;

-- ====================
-- Webhooks tables indexes
-- ====================

-- Index for webhooks by tenant and status
CREATE INDEX IF NOT EXISTS idx_webhooks_tenant_status
ON webhooks(tenant_id, status);

-- Index for webhook deliveries by webhook_id and status
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_webhook_status
ON webhook_deliveries(webhook_id, status, created_at DESC);

-- Index for pending deliveries (for retry processing)
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_pending
ON webhook_deliveries(next_retry_at)
WHERE status = 'pending' AND next_retry_at IS NOT NULL;

-- Index for failed deliveries
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_failed
ON webhook_deliveries(webhook_id, created_at DESC)
WHERE status = 'failed';

-- Index for webhook logs cleanup
CREATE INDEX IF NOT EXISTS idx_webhook_logs_created
ON webhook_logs(created_at)
WHERE created_at < NOW() - INTERVAL '30 days';

-- ====================
-- Integrations tables indexes
-- ====================

-- Index for integrations by tenant and provider
CREATE INDEX IF NOT EXISTS idx_integrations_tenant_provider
ON integrations(tenant_id, provider);

-- Index for integration logs
CREATE INDEX IF NOT EXISTS idx_integration_logs_integration_created
ON integration_logs(integration_id, created_at DESC);

-- Index for Stripe customers by tenant
CREATE INDEX IF NOT EXISTS idx_stripe_customers_tenant
ON stripe_customers(tenant_id);

-- Index for Stripe customers by customer_id
CREATE INDEX IF NOT EXISTS idx_stripe_customers_customer_id
ON stripe_customers(stripe_customer_id);

-- Index for Stripe subscriptions by customer
CREATE INDEX IF NOT EXISTS idx_stripe_subscriptions_customer
ON stripe_subscriptions(stripe_customer_id, status);

-- Index for SendGrid contacts by tenant
CREATE INDEX IF NOT EXISTS idx_sendgrid_contacts_tenant_email
ON sendgrid_contacts(tenant_id, email);

-- ====================
-- Usage tracking indexes
-- ====================

-- Index for usage metrics by tenant and date
CREATE INDEX IF NOT EXISTS idx_usage_metrics_tenant_date
ON usage_metrics(tenant_id, date DESC);

-- Index for usage metrics by metric type
CREATE INDEX IF NOT EXISTS idx_usage_metrics_type_date
ON usage_metrics(metric_type, date DESC);

-- Index for user activities by user and date
CREATE INDEX IF NOT EXISTS idx_user_activities_user_created
ON user_activities(user_id, created_at DESC);

-- Index for user activities by tenant and date
CREATE INDEX IF NOT EXISTS idx_user_activities_tenant_created
ON user_activities(tenant_id, created_at DESC);

-- Index for API usage logs cleanup
CREATE INDEX IF NOT EXISTS idx_api_usage_logs_created
ON api_usage_logs(created_at)
WHERE created_at < NOW() - INTERVAL '90 days';

-- ====================
-- Full-text search indexes (if needed)
-- ====================

-- GIN index for full-text search on user names
CREATE INDEX IF NOT EXISTS idx_users_name_gin
ON users USING gin(to_tsvector('english', name));

-- GIN index for full-text search on tenant names
CREATE INDEX IF NOT EXISTS idx_tenants_name_gin
ON tenants USING gin(to_tsvector('english', name));

-- ====================
-- JSONB indexes for better performance
-- ====================

-- Index for tenant settings
CREATE INDEX IF NOT EXISTS idx_tenants_settings_gin
ON tenants USING gin(settings);

-- Index for tenant subscription
CREATE INDEX IF NOT EXISTS idx_tenants_subscription_gin
ON tenants USING gin(subscription);

-- Index for webhook headers
CREATE INDEX IF NOT EXISTS idx_webhooks_headers_gin
ON webhooks USING gin(headers);

-- ====================
-- Covering indexes (include columns)
-- ====================

-- Covering index for user list queries (PostgreSQL 11+)
CREATE INDEX IF NOT EXISTS idx_users_tenant_id_email_name
ON users(tenant_id, id)
INCLUDE (email, name, role, created_at);

-- ====================
-- Statistics update
-- ====================

-- Update statistics for query planner
ANALYZE users;
ANALYZE tenants;
ANALYZE analytics_queries;
ANALYZE dashboards;
ANALYZE webhooks;
ANALYZE webhook_deliveries;
ANALYZE integrations;
ANALYZE usage_metrics;

-- ====================
-- Index maintenance
-- ====================

-- VACUUM and ANALYZE after index creation
VACUUM ANALYZE;

-- Down
-- Remove all performance indexes

-- Users indexes
DROP INDEX IF EXISTS idx_users_tenant_email;
DROP INDEX IF EXISTS idx_users_tenant_role;
DROP INDEX IF EXISTS idx_users_tenant_created;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_active;
DROP INDEX IF EXISTS idx_users_name_gin;
DROP INDEX IF EXISTS idx_users_tenant_id_email_name;

-- Tenants indexes
DROP INDEX IF EXISTS idx_tenants_slug;
DROP INDEX IF EXISTS idx_tenants_domain;
DROP INDEX IF EXISTS idx_tenants_status;
DROP INDEX IF EXISTS idx_tenants_subscription_status;
DROP INDEX IF EXISTS idx_tenants_trial_ends;
DROP INDEX IF EXISTS idx_tenants_name_gin;
DROP INDEX IF EXISTS idx_tenants_settings_gin;
DROP INDEX IF EXISTS idx_tenants_subscription_gin;

-- Analytics indexes
DROP INDEX IF EXISTS idx_analytics_queries_tenant;
DROP INDEX IF EXISTS idx_analytics_queries_creator;
DROP INDEX IF EXISTS idx_dashboards_tenant;
DROP INDEX IF EXISTS idx_scheduled_reports_next_run;

-- Webhooks indexes
DROP INDEX IF EXISTS idx_webhooks_tenant_status;
DROP INDEX IF EXISTS idx_webhook_deliveries_webhook_status;
DROP INDEX IF EXISTS idx_webhook_deliveries_pending;
DROP INDEX IF EXISTS idx_webhook_deliveries_failed;
DROP INDEX IF EXISTS idx_webhook_logs_created;
DROP INDEX IF EXISTS idx_webhooks_headers_gin;

-- Integrations indexes
DROP INDEX IF EXISTS idx_integrations_tenant_provider;
DROP INDEX IF EXISTS idx_integration_logs_integration_created;
DROP INDEX IF EXISTS idx_stripe_customers_tenant;
DROP INDEX IF EXISTS idx_stripe_customers_customer_id;
DROP INDEX IF EXISTS idx_stripe_subscriptions_customer;
DROP INDEX IF EXISTS idx_sendgrid_contacts_tenant_email;

-- Usage tracking indexes
DROP INDEX IF EXISTS idx_usage_metrics_tenant_date;
DROP INDEX IF EXISTS idx_usage_metrics_type_date;
DROP INDEX IF EXISTS idx_user_activities_user_created;
DROP INDEX IF EXISTS idx_user_activities_tenant_created;
DROP INDEX IF EXISTS idx_api_usage_logs_created;

-- Update statistics
VACUUM ANALYZE;
