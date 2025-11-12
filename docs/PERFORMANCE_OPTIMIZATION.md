# Performance Optimization Guide

Comprehensive guide to performance optimizations implemented in Marimo ERP.

## Table of Contents

1. [Database Query Optimization](#database-query-optimization)
2. [Caching Strategy](#caching-strategy)
3. [CDN Configuration](#cdn-configuration)
4. [Image Optimization](#image-optimization)
5. [Lazy Loading](#lazy-loading)
6. [Monitoring and Metrics](#monitoring-and-metrics)
7. [Best Practices](#best-practices)

## Database Query Optimization

### Connection Pool Settings

Optimized connection pool configuration for PostgreSQL:

```go
import "marimo/shared/database"

poolSettings := database.DefaultPoolSettings()
// MaxOpenConns: 25
// MaxIdleConns: 10
// ConnMaxLifetime: 1 hour
// ConnMaxIdleTime: 10 minutes

poolSettings.Apply(db)
```

### Cursor-Based Pagination

Use cursor-based pagination instead of OFFSET for better performance on large datasets:

```go
// BAD: OFFSET-based pagination (slow on large datasets)
db.Offset(page * limit).Limit(limit).Find(&users)

// GOOD: Cursor-based pagination
optimizer := database.NewQueryOptimizer(db)
result, err := optimizer.CursorPaginate(ctx, query, "id", cursor, 20, &users)
```

**Performance Impact**:
- OFFSET(10000): ~500ms
- Cursor-based: ~10ms

### Query Optimization Tips

1. **Select Only Needed Columns**

```go
// BAD: SELECT *
db.Find(&users)

// GOOD: Select specific fields
db.Select("id", "email", "name").Find(&users)
```

2. **Use Indexes**

```sql
-- Create composite indexes for frequent queries
CREATE INDEX idx_users_tenant_created ON users(tenant_id, created_at DESC);
CREATE INDEX idx_payments_status_date ON payments(status, created_at);

-- Partial indexes for filtered queries
CREATE INDEX idx_active_users ON users(tenant_id) WHERE status = 'active';
```

3. **Avoid N+1 Queries**

```go
// BAD: N+1 query
for _, user := range users {
    db.Where("user_id = ?", user.ID).Find(&user.Posts)
}

// GOOD: Preload associations
db.Preload("Posts").Find(&users)
```

4. **Batch Operations**

```go
optimizer := database.NewQueryOptimizer(db)

// Batch insert (1000 records at a time)
optimizer.BatchInsert(ctx, users, 1000)

// Bulk update
updates := map[string]interface{}{"status": "active"}
conditions := map[string]interface{}{"tenant_id": tenantID}
optimizer.BulkUpdate(ctx, &User{}, updates, conditions)
```

### Query Analysis

```go
// Analyze slow queries
optimizer := database.NewQueryOptimizer(db)

// Get query execution plan
plan, err := optimizer.ExplainQuery(ctx, query)
fmt.Println(plan)

// Get index recommendations
recommendations, err := optimizer.AnalyzeAndRecommend(ctx)
```

### Slow Query Logging

```go
threshold := 100 * time.Millisecond
slowLogger := database.NewSlowQueryLogger(threshold, logger)

db.Logger = slowLogger
// All queries > 100ms will be logged
```

## Caching Strategy

### Redis Cache Setup

```go
import "marimo/shared/cache"

// Initialize Redis cache
cache, err := cache.NewRedisCache(
    "localhost:6379",
    "",  // password
    0,   // database
    "marimo", // key prefix
)

// Create cache manager with cache-aside strategy
manager := cache.NewCacheManager(cache, cache.CacheAside)
```

### Cache-Aside Pattern

```go
var user User

// Get from cache or load from database
err := manager.GetOrSet(ctx, "user:123", &user, 15*time.Minute, func() (interface{}, error) {
    var u User
    err := db.First(&u, "id = ?", "123").Error
    return u, err
})
```

### TTL Strategy

Different TTL for different data types:

```go
ttl := cache.DefaultTTLStrategy()

// Frequently changing data (1-5 minutes)
cache.Set(ctx, "user_status:123", status, ttl.ShortTTL)

// Moderately stable data (5-60 minutes)
cache.Set(ctx, "user:123", user, ttl.MediumTTL)

// Stable data (1-24 hours)
cache.Set(ctx, "tenant_settings:456", settings, ttl.LongTTL)
```

### Cache Key Naming

```go
kb := cache.NewCacheKeyBuilder("marimo")

// Structured cache keys
userKey := kb.UserKey(userID)           // marimo:user:123
tenantKey := kb.TenantKey(tenantID)     // marimo:tenant:456
sessionKey := kb.SessionKey(sessionID)  // marimo:session:abc
queryKey := kb.QueryKey("users", params) // marimo:query:users:...
```

### Cache Invalidation by Tags

```go
tags := cache.NewCacheTags(cache)

// Store with tags
tags.Set(ctx, "user:123", user, 1*time.Hour, "users", "tenant:456")

// Invalidate all users in tenant
tags.InvalidateByTag(ctx, "tenant:456")
```

### Rate Limiting with Cache

```go
limiter := cache.NewRateLimiter(cache)

// Allow 100 requests per minute
allowed, err := limiter.Allow(ctx, "api:user:123", 100, 1*time.Minute)
if !allowed {
    return errors.New("rate limit exceeded")
}
```

### Cache Statistics

```go
stats, err := cache.GetStats(ctx)
fmt.Printf("Hit rate: %.2f%%\n", stats.HitRate * 100)
fmt.Printf("Keys: %d\n", stats.KeyCount)
```

### Cache Warming

Preload frequently accessed data:

```go
loaders := map[string]func() (interface{}, error){
    "tenant:main": func() (interface{}, error) {
        return loadMainTenant()
    },
    "config:app": func() (interface{}, error) {
        return loadAppConfig()
    },
}

manager.WarmUp(ctx, loaders)
```

## CDN Configuration

### CDN Setup

```go
import "marimo/shared/cdn"

config := &cdn.CDNConfig{
    Provider:    cdn.CloudFlare,
    BaseURL:     "https://cdn.marimo-erp.com",
    Enabled:     true,
    StaticPaths: []string{"static", "assets", "uploads"},
}

cdnClient := cdn.NewCDN(config)
```

### Generate CDN URLs

```go
// Static asset URL
cssURL := cdnClient.URL("/static/css/app.css")
// https://cdn.marimo-erp.com/static/css/app.css

// Image with transformations
imageURL := cdnClient.ImageURL("/uploads/photo.jpg", &cdn.ImageOptions{
    Width:   800,
    Height:  600,
    Quality: 85,
    Format:  "webp",
})
// https://cdn.marimo-erp.com/uploads/photo.jpg?w=800&h=600&q=85&f=webp
```

### Responsive Images

```go
// Generate responsive image set
responsiveSet := cdnClient.GenerateResponsiveSet(
    "/uploads/hero.jpg",
    []int{320, 640, 1024, 1920},
)

// Use in HTML
// <img src="..." srcset="... 320w, ... 640w, ... 1024w, ... 1920w">
```

### Cache-Control Headers

```go
// For versioned assets (app.abc123.js)
headers := cdn.CacheControlHeaders("immutable")
// Cache-Control: public, max-age=31536000, immutable

// For static assets
headers := cdn.CacheControlHeaders("static")
// Cache-Control: public, max-age=604800

// For dynamic content
headers := cdn.CacheControlHeaders("dynamic")
// Cache-Control: public, max-age=300, must-revalidate
```

### Asset Versioning

```go
versioning := cdn.NewAssetVersioning("manifest.json")

// Get versioned asset
versionedURL := versioning.Get("/static/app.js")
// /static/app.abc123.js
```

### HTTP/2 Server Push

```go
assets := []string{
    "/static/css/app.css",
    "/static/js/app.js",
    "/static/fonts/main.woff2",
}

headers := cdn.PreloadHeaders(assets)
// Link: </static/css/app.css>; rel=preload; as=style, ...
```

### Purge CDN Cache

```go
urls := []string{
    "https://cdn.marimo-erp.com/static/app.css",
    "https://cdn.marimo-erp.com/static/app.js",
}

cdnClient.PurgeCache(urls)
```

## Image Optimization

### Optimize Images

```go
import "marimo/shared/images"

optimizer := images.NewImageOptimizer()

opts := &images.OptimizeOptions{
    MaxWidth:  1920,
    MaxHeight: 1080,
    Quality:   85,
    Format:    images.FormatWebP,
    StripMeta: true,
}

err := optimizer.OptimizeFile("input.jpg", "output.webp", opts)
```

### Generate Thumbnails

```go
sizes := images.DefaultThumbnailSizes()
// small: 150x150, medium: 300x300, large: 600x600, xlarge: 1200x1200

thumbnails, err := optimizer.GenerateThumbnails(
    "original.jpg",
    "thumbnails/",
    sizes,
)

// thumbnails["small"] = "thumbnails/original_small.webp"
// thumbnails["medium"] = "thumbnails/original_medium.webp"
```

### Responsive Image Sets

```go
generator := images.NewResponsiveImageGenerator()

responsiveSet, err := generator.Generate(
    "original.jpg",
    "responsive/",
    []int{320, 640, 1024, 1920},
)

// Use in HTML
// srcset="responsive/original_320w.webp 320w, ..."
```

### Lazy Load Placeholders

```go
// Generate tiny blurred placeholder for lazy loading
placeholder, err := optimizer.LazyLoadPlaceholder("image.jpg")
// Returns base64-encoded 20px width JPEG
```

### Batch Optimization

```go
batchOptimizer := images.NewBatchOptimizer(4) // 4 workers

err := batchOptimizer.OptimizeDirectory(
    "uploads/",
    "optimized/",
    opts,
)
```

### Image Metadata

```go
metadata, err := images.GetMetadata("image.jpg")
fmt.Printf("Size: %dx%d\n", metadata.Width, metadata.Height)
fmt.Printf("Format: %s\n", metadata.Format)
fmt.Printf("File size: %d bytes\n", metadata.Size)
```

### Format Detection

```go
// Detect WebP support from Accept header
format := images.GetOptimalFormat(r.Header.Get("Accept"))
// Returns FormatWebP if supported, FormatJPEG otherwise
```

## Lazy Loading

### Lazy Load Images (React)

```tsx
import { LazyImage } from '@/components/LazyImage';

<LazyImage
  src="/uploads/photo.jpg"
  alt="Product photo"
  placeholder="/uploads/photo_thumb.jpg"
  srcSet="photo_320w.webp 320w, photo_640w.webp 640w"
  sizes="(max-width: 600px) 320px, 640px"
/>
```

### Progressive Image Loading

```tsx
import { ProgressiveImage } from '@/components/LazyImage';

<ProgressiveImage
  src="/uploads/high-res.jpg"
  placeholderSrc="/uploads/low-res.jpg"
  alt="Hero image"
/>
```

### Lazy Load Components

```tsx
import { lazyLoadComponent } from '@/utils/lazyLoad';

// Lazy load page component
const DashboardPage = lazyLoadComponent(
  () => import('./pages/Dashboard'),
  {
    fallback: <LoadingSpinner />,
    maxRetries: 3,
  }
);

// Use in routes
<Route path="/dashboard" component={DashboardPage} />
```

### Prefetch on Hover

```tsx
import { PrefetchLink } from '@/utils/lazyLoad';

<PrefetchLink
  to="/analytics"
  prefetchComponent={AnalyticsPage}
>
  Go to Analytics
</PrefetchLink>
```

### Intersection Observer Lazy Render

```tsx
import { LazyRender } from '@/utils/lazyLoad';

<LazyRender threshold={0.1} rootMargin="100px">
  <ExpensiveComponent />
</LazyRender>
```

### Code Splitting by Route

```tsx
// Routes are automatically code-split
const routes = [
  { path: '/dashboard', component: DashboardPage },    // dashboard.chunk.js
  { path: '/users', component: UsersPage },            // users.chunk.js
  { path: '/analytics', component: AnalyticsPage },    // analytics.chunk.js
];
```

## Monitoring and Metrics

### Performance Metrics

Key performance indicators:

```
Response Time Targets:
- API: < 200ms (p95)
- Database queries: < 50ms (p95)
- Cache hit rate: > 85%
- CDN offload: > 70% of static traffic
```

### Prometheus Metrics

```go
import "marimo/shared/monitoring"

metrics := monitoring.NewMetrics()

// HTTP request metrics
metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)

// Database metrics
metrics.DBQueriesTotal.WithLabelValues("select", "users").Inc()
metrics.DBQueryDuration.WithLabelValues("select").Observe(duration)

// Cache metrics
metrics.CacheHitsTotal.WithLabelValues("redis", "user").Inc()
metrics.CacheMissesTotal.WithLabelValues("redis", "user").Inc()
```

### Query Performance Monitoring

```sql
-- Enable pg_stat_statements
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- Find slow queries
SELECT
    query,
    calls,
    total_time,
    mean_time,
    max_time
FROM pg_stat_statements
WHERE mean_time > 100  -- queries > 100ms
ORDER BY mean_time DESC
LIMIT 20;
```

### Cache Hit Rate

```bash
# Redis cache stats
redis-cli INFO stats | grep hits
# keyspace_hits:1000000
# keyspace_misses:150000
# Hit rate: 87%
```

### CDN Analytics

Monitor CDN performance:
- Bandwidth savings
- Cache hit ratio
- Geographic distribution
- Response times

## Best Practices

### Database

1. ✅ Use indexes on frequently queried columns
2. ✅ Implement cursor-based pagination for large datasets
3. ✅ Batch insert/update operations
4. ✅ Use connection pooling
5. ✅ Monitor and optimize slow queries
6. ✅ Use EXPLAIN ANALYZE to understand query plans
7. ✅ Avoid SELECT *, select only needed columns
8. ✅ Use partial indexes for filtered queries
9. ✅ Regular VACUUM and ANALYZE
10. ✅ Archive old data

### Caching

1. ✅ Implement cache-aside pattern
2. ✅ Use appropriate TTL for different data types
3. ✅ Implement cache warming for critical data
4. ✅ Use cache tags for batch invalidation
5. ✅ Monitor cache hit rates
6. ✅ Use Redis for distributed caching
7. ✅ Implement graceful cache failures
8. ✅ Don't cache everything - profile first
9. ✅ Use cache stampede protection
10. ✅ Regular cache cleanup

### CDN

1. ✅ Serve all static assets through CDN
2. ✅ Use long cache TTL for versioned assets
3. ✅ Implement asset versioning/fingerprinting
4. ✅ Use HTTP/2 Server Push for critical resources
5. ✅ Enable gzip/brotli compression
6. ✅ Use WebP format for images
7. ✅ Implement responsive images
8. ✅ Set proper Cache-Control headers
9. ✅ Monitor CDN cache hit ratio
10. ✅ Purge cache on deployments

### Images

1. ✅ Always optimize images before upload
2. ✅ Use WebP format (fallback to JPEG)
3. ✅ Generate multiple sizes for responsive images
4. ✅ Implement lazy loading
5. ✅ Use low-quality placeholders
6. ✅ Strip metadata from images
7. ✅ Set appropriate quality (85 is good balance)
8. ✅ Use CDN for image delivery
9. ✅ Implement progressive image loading
10. ✅ Monitor image sizes and optimize outliers

### Frontend

1. ✅ Code splitting by route
2. ✅ Lazy load heavy components
3. ✅ Implement lazy loading for images
4. ✅ Use React.memo for expensive components
5. ✅ Prefetch routes on hover
6. ✅ Minimize bundle size
7. ✅ Tree-shake unused code
8. ✅ Use Web Workers for heavy computations
9. ✅ Implement virtual scrolling for long lists
10. ✅ Monitor bundle sizes

### General

1. ✅ Monitor performance metrics continuously
2. ✅ Set performance budgets
3. ✅ Regular performance audits
4. ✅ Use production profiling tools
5. ✅ Implement graceful degradation
6. ✅ Test with realistic data volumes
7. ✅ Optimize critical user journeys first
8. ✅ Document performance characteristics
9. ✅ Load test before production
10. ✅ Continuous performance monitoring

## Performance Checklist

Before deploying to production:

- [ ] Database indexes created for all query patterns
- [ ] Connection pool optimized
- [ ] Slow query logging enabled
- [ ] Redis cache configured and tested
- [ ] Cache warming implemented for critical data
- [ ] CDN configured for static assets
- [ ] All images optimized
- [ ] Responsive images generated
- [ ] Lazy loading implemented for images
- [ ] Code splitting implemented for routes
- [ ] Bundle size within budget (< 250KB gzipped)
- [ ] Cache-Control headers set correctly
- [ ] Compression enabled (gzip/brotli)
- [ ] HTTP/2 enabled
- [ ] Performance monitoring set up
- [ ] Load testing completed
- [ ] Performance budgets defined
- [ ] Alerts configured for performance degradation

## Resources

- [Database Query Optimizer](../shared/database/query_optimizer.go)
- [Cache Implementation](../shared/cache/cache.go)
- [CDN Configuration](../shared/cdn/cdn.go)
- [Image Optimization](../shared/images/optimizer.go)
- [Lazy Loading Utils](../frontend/src/utils/lazyLoad.tsx)
- [Prometheus Metrics](../shared/monitoring/metrics.go)

## Support

For performance issues:
1. Check Prometheus metrics dashboard
2. Review slow query logs
3. Analyze cache hit rates
4. Check CDN analytics
5. Profile frontend bundle
6. Contact DevOps team if needed
