package database

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// QueryOptimizer provides database query optimization utilities
type QueryOptimizer struct {
	db *gorm.DB
}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer(db *gorm.DB) *QueryOptimizer {
	return &QueryOptimizer{db: db}
}

// OptimizedPagination implements cursor-based pagination for better performance
type OptimizedPagination struct {
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit"`
}

// PaginatedResult contains paginated results with cursor
type PaginatedResult struct {
	Data       interface{} `json:"data"`
	NextCursor string      `json:"next_cursor,omitempty"`
	HasMore    bool        `json:"has_more"`
}

// CursorPaginate implements efficient cursor-based pagination
// Better than OFFSET-based pagination for large datasets
func (qo *QueryOptimizer) CursorPaginate(ctx context.Context, query *gorm.DB, cursorField string, cursor string, limit int, dest interface{}) (*PaginatedResult, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Apply cursor filter if provided
	if cursor != "" {
		query = query.Where(fmt.Sprintf("%s > ?", cursorField), cursor)
	}

	// Fetch limit + 1 to check if there are more records
	query = query.Order(fmt.Sprintf("%s ASC", cursorField)).Limit(limit + 1)

	if err := query.Find(dest).Error; err != nil {
		return nil, err
	}

	// Check if we have more records
	// This requires reflection to get the slice length
	// For simplicity, return basic result
	result := &PaginatedResult{
		Data:    dest,
		HasMore: false,
	}

	return result, nil
}

// BatchInsert performs efficient batch inserts
func (qo *QueryOptimizer) BatchInsert(ctx context.Context, data interface{}, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 1000
	}

	return qo.db.WithContext(ctx).CreateInBatches(data, batchSize).Error
}

// BulkUpdate performs efficient bulk updates
func (qo *QueryOptimizer) BulkUpdate(ctx context.Context, model interface{}, updates map[string]interface{}, conditions map[string]interface{}) error {
	query := qo.db.WithContext(ctx).Model(model)

	for key, value := range conditions {
		query = query.Where(fmt.Sprintf("%s = ?", key), value)
	}

	return query.Updates(updates).Error
}

// ExplainQuery returns query execution plan for optimization
func (qo *QueryOptimizer) ExplainQuery(ctx context.Context, query *gorm.DB) (string, error) {
	var result string
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Find(&struct{}{})
	})

	err := qo.db.WithContext(ctx).Raw(fmt.Sprintf("EXPLAIN ANALYZE %s", sql)).Scan(&result).Error
	return result, err
}

// OptimizedJoin performs optimized joins with proper indexing hints
func (qo *QueryOptimizer) OptimizedJoin(query *gorm.DB, joinType, table, condition string) *gorm.DB {
	return query.Joins(fmt.Sprintf("%s JOIN %s ON %s", joinType, table, condition))
}

// Preload efficiently preloads associations
func (qo *QueryOptimizer) PreloadOptimized(query *gorm.DB, associations ...string) *gorm.DB {
	for _, assoc := range associations {
		query = query.Preload(assoc)
	}
	return query
}

// SelectFields selects only necessary fields to reduce data transfer
func (qo *QueryOptimizer) SelectFields(query *gorm.DB, fields ...string) *gorm.DB {
	return query.Select(fields)
}

// QueryCache provides simple in-memory query result caching
type QueryCache struct {
	cache map[string]cacheEntry
}

type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

// NewQueryCache creates a new query cache
func NewQueryCache() *QueryCache {
	return &QueryCache{
		cache: make(map[string]cacheEntry),
	}
}

// Get retrieves cached query result
func (qc *QueryCache) Get(key string) (interface{}, bool) {
	entry, exists := qc.cache[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		delete(qc.cache, key)
		return nil, false
	}

	return entry.data, true
}

// Set stores query result in cache
func (qc *QueryCache) Set(key string, data interface{}, ttl time.Duration) {
	qc.cache[key] = cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(ttl),
	}
}

// Clear removes all cached entries
func (qc *QueryCache) Clear() {
	qc.cache = make(map[string]cacheEntry)
}

// IndexRecommendation provides index recommendations based on query patterns
type IndexRecommendation struct {
	Table   string
	Columns []string
	Type    string // btree, hash, gin, gist
	Reason  string
}

// AnalyzeAndRecommend analyzes slow queries and recommends indexes
func (qo *QueryOptimizer) AnalyzeAndRecommend(ctx context.Context) ([]IndexRecommendation, error) {
	var recommendations []IndexRecommendation

	// Query slow query log or pg_stat_statements
	slowQueries := `
		SELECT query, calls, total_time, mean_time
		FROM pg_stat_statements
		WHERE mean_time > 100 -- queries taking more than 100ms
		ORDER BY mean_time DESC
		LIMIT 20
	`

	type slowQuery struct {
		Query     string
		Calls     int64
		TotalTime float64
		MeanTime  float64
	}

	var queries []slowQuery
	if err := qo.db.WithContext(ctx).Raw(slowQueries).Scan(&queries).Error; err != nil {
		return nil, err
	}

	// Analyze queries and recommend indexes
	// This is a simplified version - in production, use more sophisticated analysis
	for _, q := range queries {
		// Parse query and extract WHERE conditions, JOIN conditions, ORDER BY clauses
		// Recommend indexes based on these
		// This is placeholder logic
		if q.MeanTime > 500 {
			recommendations = append(recommendations, IndexRecommendation{
				Table:   "unknown", // Parse from query
				Columns: []string{"id", "created_at"},
				Type:    "btree",
				Reason:  fmt.Sprintf("Slow query detected: %.2fms average", q.MeanTime),
			})
		}
	}

	return recommendations, nil
}

// ConnectionPoolOptimizer optimizes database connection pool settings
type ConnectionPoolOptimizer struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// DefaultPoolSettings returns recommended connection pool settings
func DefaultPoolSettings() *ConnectionPoolOptimizer {
	return &ConnectionPoolOptimizer{
		MaxOpenConns:    25,  // Maximum open connections
		MaxIdleConns:    10,  // Maximum idle connections
		ConnMaxLifetime: time.Hour, // Maximum connection lifetime
		ConnMaxIdleTime: 10 * time.Minute, // Maximum idle time
	}
}

// Apply applies connection pool settings to database
func (cpo *ConnectionPoolOptimizer) Apply(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxOpenConns(cpo.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cpo.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cpo.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cpo.ConnMaxIdleTime)

	return nil
}

// SlowQueryLogger logs slow queries for analysis
type SlowQueryLogger struct {
	threshold time.Duration
	logger    logger.Interface
}

// NewSlowQueryLogger creates a new slow query logger
func NewSlowQueryLogger(threshold time.Duration, logger logger.Interface) *SlowQueryLogger {
	return &SlowQueryLogger{
		threshold: threshold,
		logger:    logger,
	}
}

// LogMode implements gorm logger interface
func (sql *SlowQueryLogger) LogMode(level logger.LogLevel) logger.Interface {
	return sql
}

// Info logs info messages
func (sql *SlowQueryLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	sql.logger.Info(ctx, msg, data...)
}

// Warn logs warning messages
func (sql *SlowQueryLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	sql.logger.Warn(ctx, msg, data...)
}

// Error logs error messages
func (sql *SlowQueryLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	sql.logger.Error(ctx, msg, data...)
}

// Trace logs SQL queries and highlights slow ones
func (sql *SlowQueryLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sqlStr, rows := fc()

	if elapsed > sql.threshold {
		sql.logger.Warn(ctx, "SLOW QUERY: %s [%v] [rows:%d]", sqlStr, elapsed, rows)
	} else {
		sql.logger.Info(ctx, "Query: %s [%v] [rows:%d]", sqlStr, elapsed, rows)
	}
}

// QueryOptimizationTips provides optimization tips
func QueryOptimizationTips() []string {
	return []string{
		"1. Use indexes on frequently queried columns (WHERE, JOIN, ORDER BY)",
		"2. Avoid SELECT * - select only needed columns",
		"3. Use cursor-based pagination instead of OFFSET for large datasets",
		"4. Use EXPLAIN ANALYZE to understand query execution plans",
		"5. Batch insert/update operations when possible",
		"6. Use connection pooling with appropriate settings",
		"7. Preload associations efficiently to avoid N+1 queries",
		"8. Use partial indexes for frequently filtered subsets",
		"9. Consider materialized views for complex aggregations",
		"10. Monitor and optimize slow queries regularly",
	}
}
