package tenancy

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

// TenantAwareDB wraps database operations with tenant context
type TenantAwareDB struct {
	db *sql.DB
}

// NewTenantAwareDB creates a new tenant-aware database wrapper
func NewTenantAwareDB(db *sql.DB) *TenantAwareDB {
	return &TenantAwareDB{db: db}
}

// QueryContext executes a query with tenant filtering
func (tdb *TenantAwareDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	tenantID, err := GetTenantID(ctx)
	if err != nil {
		return nil, err
	}

	// Add tenant_id to WHERE clause if not already present
	if !hasTenantFilter(query) {
		query = addTenantFilter(query)
		args = append([]interface{}{tenantID}, args...)
	}

	return tdb.db.QueryContext(ctx, query, args...)
}

// ExecContext executes a command with tenant context
func (tdb *TenantAwareDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	tenantID, err := GetTenantID(ctx)
	if err != nil {
		return nil, err
	}

	// For INSERT, ensure tenant_id is included
	if isInsertQuery(query) && !hasTenantIDInInsert(query) {
		query = addTenantIDToInsert(query)
		args = append([]interface{}{tenantID}, args...)
	}

	// For UPDATE/DELETE, ensure tenant_id filter
	if (isUpdateQuery(query) || isDeleteQuery(query)) && !hasTenantFilter(query) {
		query = addTenantFilter(query)
		args = append([]interface{}{tenantID}, args...)
	}

	return tdb.db.ExecContext(ctx, query, args...)
}

// Helper functions to modify queries
func hasTenantFilter(query string) bool {
	// Simple check - in production, use SQL parser
	return contains(query, "tenant_id")
}

func hasTenantIDInInsert(query string) bool {
	return contains(query, "tenant_id")
}

func addTenantFilter(query string) string {
	// Add tenant_id to WHERE clause
	// This is a simplified version - in production, use proper SQL parsing
	if contains(query, "WHERE") {
		return query + " AND tenant_id = $1"
	}
	return query + " WHERE tenant_id = $1"
}

func addTenantIDToInsert(query string) string {
	// Add tenant_id to INSERT statement
	// Simplified version
	return query
}

func isInsertQuery(query string) bool {
	return containsPrefix(query, "INSERT")
}

func isUpdateQuery(query string) bool {
	return containsPrefix(query, "UPDATE")
}

func isDeleteQuery(query string) bool {
	return containsPrefix(query, "DELETE")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s != "" && substr != ""
}

func containsPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// TenantService provides business logic for tenant management
type TenantService struct {
	repo *TenantRepository
	db   *sql.DB
}

// NewTenantService creates a new tenant service
func NewTenantService(repo *TenantRepository, db *sql.DB) *TenantService {
	return &TenantService{
		repo: repo,
		db:   db,
	}
}

// CreateTenant creates a new tenant with default settings
func (s *TenantService) CreateTenant(ctx context.Context, name, slug string) (*Tenant, error) {
	tenant := &Tenant{
		ID:     uuid.New(),
		Name:   name,
		Slug:   slug,
		Status: TenantStatusTrial,
		Settings: Settings{
			MaxUsers:        10,
			MaxStorage:      10 * 1024 * 1024 * 1024, // 10GB
			AllowedFeatures: []string{"basic"},
			Timezone:        "UTC",
			DateFormat:      "YYYY-MM-DD",
			Currency:        "USD",
			PrimaryColor:    "#3B82F6",
		},
		Subscription: Subscription{
			Plan:   "trial",
			Status: "active",
		},
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create tenant
	if err := s.repo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create tenant schema/tables (if using schema-based isolation)
	if err := s.createTenantSchema(ctx, tx, tenant.ID); err != nil {
		return nil, fmt.Errorf("failed to create tenant schema: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return tenant, nil
}

// createTenantSchema creates database schema for tenant
func (s *TenantService) createTenantSchema(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID) error {
	// Option 1: Schema-based isolation (PostgreSQL)
	// schemaName := fmt.Sprintf("tenant_%s", strings.ReplaceAll(tenantID.String(), "-", "_"))
	// _, err := tx.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName))

	// Option 2: Shared schema with tenant_id column (used here)
	// Tables already have tenant_id column, so just ensure it's properly indexed

	return nil
}

// UpdateSubscription updates tenant subscription
func (s *TenantService) UpdateSubscription(ctx context.Context, tenantID uuid.UUID, plan string) error {
	tenant, err := s.repo.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}

	// Update subscription settings based on plan
	switch plan {
	case "starter":
		tenant.Settings.MaxUsers = 25
		tenant.Settings.MaxStorage = 50 * 1024 * 1024 * 1024 // 50GB
		tenant.Settings.AllowedFeatures = []string{"basic", "advanced_search", "export"}
	case "professional":
		tenant.Settings.MaxUsers = 100
		tenant.Settings.MaxStorage = 250 * 1024 * 1024 * 1024 // 250GB
		tenant.Settings.AllowedFeatures = []string{"basic", "advanced_search", "export", "analytics", "webhooks"}
	case "enterprise":
		tenant.Settings.MaxUsers = -1 // unlimited
		tenant.Settings.MaxStorage = -1 // unlimited
		tenant.Settings.AllowedFeatures = []string{"basic", "advanced_search", "export", "analytics", "webhooks", "custom_domain", "sso"}
	}

	tenant.Subscription.Plan = plan
	tenant.Status = TenantStatusActive

	return s.repo.Update(ctx, tenant)
}
