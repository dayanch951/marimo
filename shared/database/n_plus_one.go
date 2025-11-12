package database

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// PreloadConfig defines preload configuration
type PreloadConfig struct {
	Associations []string
	Conditions   map[string]interface{}
}

// EagerLoader provides utilities to prevent N+1 queries
type EagerLoader struct {
	db *gorm.DB
}

// NewEagerLoader creates a new eager loader
func NewEagerLoader(db *gorm.DB) *EagerLoader {
	return &EagerLoader{db: db}
}

// WithPreload loads records with specified associations
func (el *EagerLoader) WithPreload(associations ...string) *gorm.DB {
	query := el.db

	for _, assoc := range associations {
		query = query.Preload(assoc)
	}

	return query
}

// WithPreloadConditions loads associations with conditions
func (el *EagerLoader) WithPreloadConditions(association string, conditions interface{}) *gorm.DB {
	return el.db.Preload(association, conditions)
}

// LoadUsersWithRelations loads users with all common relations
// Example of a pre-configured preload for common use case
func (el *EagerLoader) LoadUsersWithRelations(ctx context.Context, tenantID string) ([]interface{}, error) {
	var users []interface{}

	err := el.db.WithContext(ctx).
		Preload("Tenant").              // Load tenant relation
		Preload("CreatedBy").           // Load creator
		Preload("Roles").               // Load roles
		Preload("Permissions").         // Load permissions
		Where("tenant_id = ?", tenantID).
		Find(&users).Error

	return users, err
}

// DataLoader provides data loader pattern implementation
type DataLoader struct {
	cache map[string]interface{}
}

// NewDataLoader creates a new data loader
func NewDataLoader() *DataLoader {
	return &DataLoader{
		cache: make(map[string]interface{}),
	}
}

// Load loads data using cache
func (dl *DataLoader) Load(key string, loader func() (interface{}, error)) (interface{}, error) {
	// Check cache
	if data, ok := dl.cache[key]; ok {
		return data, nil
	}

	// Load data
	data, err := loader()
	if err != nil {
		return nil, err
	}

	// Cache data
	dl.cache[key] = data

	return data, nil
}

// BatchLoader loads multiple records in a single query
type BatchLoader struct {
	db *gorm.DB
}

// NewBatchLoader creates a new batch loader
func NewBatchLoader(db *gorm.DB) *BatchLoader {
	return &BatchLoader{db: db}
}

// LoadByIDs loads multiple records by IDs in a single query
// This prevents N queries when loading related records
func (bl *BatchLoader) LoadByIDs(ctx context.Context, model interface{}, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	return bl.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(model).Error
}

// LoadMap loads records and returns them as a map indexed by ID
// Useful for quick lookups without iteration
func (bl *BatchLoader) LoadMap(ctx context.Context, model interface{}, ids []string) (map[string]interface{}, error) {
	if len(ids) == 0 {
		return make(map[string]interface{}), nil
	}

	// This is a simplified version - in production, use reflection
	// to properly handle the generic type
	result := make(map[string]interface{})

	// Load records
	if err := bl.LoadByIDs(ctx, model, ids); err != nil {
		return nil, err
	}

	// Convert to map (simplified - needs proper implementation)
	// In production, use reflection to extract ID field and create map

	return result, nil
}

// QueryOptimizer provides query optimization helpers
type QueryOptimizer struct {
	db *gorm.DB
}

// SelectFields selects only necessary fields to reduce data transfer
func (qo *QueryOptimizer) SelectFields(fields ...string) *gorm.DB {
	return qo.db.Select(fields)
}

// JoinOptimized performs optimized join
func (qo *QueryOptimizer) JoinOptimized(joinType, table, condition string) *gorm.DB {
	return qo.db.Joins(fmt.Sprintf("%s JOIN %s ON %s", joinType, table, condition))
}

// Examples of N+1 query prevention

// BAD: N+1 query example
// func GetUsersWithPosts(db *gorm.DB) ([]User, error) {
//     var users []User
//     db.Find(&users)
//
//     // This creates N queries (one per user)
//     for i := range users {
//         db.Where("user_id = ?", users[i].ID).Find(&users[i].Posts)
//     }
//
//     return users, nil
// }

// GOOD: Single query with preload
// func GetUsersWithPosts(db *gorm.DB) ([]User, error) {
//     var users []User
//     err := db.Preload("Posts").Find(&users).Error
//     return users, err
// }

// GOOD: Preload with conditions
// func GetUsersWithPublishedPosts(db *gorm.DB) ([]User, error) {
//     var users []User
//     err := db.Preload("Posts", "status = ?", "published").Find(&users).Error
//     return users, err
// }

// GOOD: Multiple preloads
// func GetUsersWithAllRelations(db *gorm.DB) ([]User, error) {
//     var users []User
//     err := db.
//         Preload("Posts").
//         Preload("Posts.Comments").
//         Preload("Profile").
//         Preload("Roles").
//         Find(&users).Error
//     return users, err
// }

// GOOD: Nested preloads
// func GetUsersWithNestedRelations(db *gorm.DB) ([]User, error) {
//     var users []User
//     err := db.
//         Preload("Posts.Comments.Author").
//         Preload("Posts.Tags").
//         Find(&users).Error
//     return users, err
// }

// PreloadHelper provides common preload patterns
type PreloadHelper struct {
	db *gorm.DB
}

// NewPreloadHelper creates a new preload helper
func NewPreloadHelper(db *gorm.DB) *PreloadHelper {
	return &PreloadHelper{db: db}
}

// UserWithTenant preloads user with tenant
func (ph *PreloadHelper) UserWithTenant() *gorm.DB {
	return ph.db.Preload("Tenant")
}

// UserWithRoles preloads user with roles
func (ph *PreloadHelper) UserWithRoles() *gorm.DB {
	return ph.db.Preload("Roles")
}

// UserWithAll preloads user with all common relations
func (ph *PreloadHelper) UserWithAll() *gorm.DB {
	return ph.db.
		Preload("Tenant").
		Preload("Roles").
		Preload("Permissions").
		Preload("CreatedBy")
}

// TenantWithUsers preloads tenant with users
func (ph *PreloadHelper) TenantWithUsers() *gorm.DB {
	return ph.db.Preload("Users")
}

// TenantWithActiveUsers preloads tenant with active users only
func (ph *PreloadHelper) TenantWithActiveUsers() *gorm.DB {
	return ph.db.Preload("Users", "status = ?", "active")
}

// WebhookWithDeliveries preloads webhook with recent deliveries
func (ph *PreloadHelper) WebhookWithDeliveries(limit int) *gorm.DB {
	return ph.db.Preload("Deliveries", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at DESC").Limit(limit)
	})
}

// DashboardWithWidgets preloads dashboard with widgets
func (ph *PreloadHelper) DashboardWithWidgets() *gorm.DB {
	return ph.db.Preload("Widgets").Preload("Widgets.Query")
}

// N+1 Detection Helper
type N1Detector struct {
	enabled bool
	queries []string
}

// NewN1Detector creates a new N+1 detector
func NewN1Detector(enabled bool) *N1Detector {
	return &N1Detector{
		enabled: enabled,
		queries: make([]string, 0),
	}
}

// RecordQuery records a query for analysis
func (n *N1Detector) RecordQuery(query string) {
	if !n.enabled {
		return
	}

	n.queries = append(n.queries, query)
}

// Analyze analyzes recorded queries for N+1 patterns
func (n *N1Detector) Analyze() []string {
	if !n.enabled {
		return nil
	}

	warnings := make([]string, 0)

	// Simple analysis: look for repeated similar queries
	queryCount := make(map[string]int)

	for _, query := range n.queries {
		// Simplified - in production, normalize queries before counting
		queryCount[query]++
	}

	// Report queries that appear more than 5 times
	for query, count := range queryCount {
		if count > 5 {
			warnings = append(warnings, fmt.Sprintf("Potential N+1: Query executed %d times: %s", count, query))
		}
	}

	return warnings
}

// Clear clears recorded queries
func (n *N1Detector) Clear() {
	n.queries = make([]string, 0)
}

// Best practices for preventing N+1 queries:
//
// 1. Always use Preload for associations
// 2. Use Joins when you need to filter by association fields
// 3. Select only needed fields
// 4. Use batch loading for multiple records
// 5. Implement data loader pattern for GraphQL
// 6. Monitor query counts in tests
// 7. Use database query logging in development
// 8. Profile queries in production
// 9. Set up alerts for high query counts
// 10. Regular code reviews for query patterns
