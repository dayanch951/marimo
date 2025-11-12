package tenancy

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrTenantNotFound     = errors.New("tenant not found")
	ErrTenantInactive     = errors.New("tenant is inactive")
	ErrTenantSuspended    = errors.New("tenant is suspended")
	ErrInvalidTenantID    = errors.New("invalid tenant ID")
	ErrNoTenantInContext  = errors.New("no tenant in context")
)

// TenantStatus represents the status of a tenant
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusInactive  TenantStatus = "inactive"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusTrial     TenantStatus = "trial"
)

// Tenant represents an organization/company in the multi-tenant system
type Tenant struct {
	ID             uuid.UUID    `json:"id"`
	Name           string       `json:"name"`
	Slug           string       `json:"slug"` // Used in subdomain: {slug}.marimo-erp.com
	Domain         *string      `json:"domain,omitempty"` // Custom domain
	Status         TenantStatus `json:"status"`
	Settings       Settings     `json:"settings"`
	Subscription   Subscription `json:"subscription"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	TrialEndsAt    *time.Time   `json:"trial_ends_at,omitempty"`
	SuspendedAt    *time.Time   `json:"suspended_at,omitempty"`
	SuspendReason  *string      `json:"suspend_reason,omitempty"`
}

// Settings contains tenant-specific settings
type Settings struct {
	MaxUsers         int    `json:"max_users"`
	MaxStorage       int64  `json:"max_storage"` // in bytes
	AllowedFeatures  []string `json:"allowed_features"`
	Timezone         string `json:"timezone"`
	DateFormat       string `json:"date_format"`
	Currency         string `json:"currency"`
	Logo             *string `json:"logo,omitempty"`
	PrimaryColor     string `json:"primary_color"`
	CustomCSS        *string `json:"custom_css,omitempty"`
}

// Subscription contains subscription information
type Subscription struct {
	Plan            string     `json:"plan"` // free, starter, professional, enterprise
	Status          string     `json:"status"` // active, past_due, canceled
	CurrentPeriodStart time.Time `json:"current_period_start"`
	CurrentPeriodEnd   time.Time `json:"current_period_end"`
	CancelAt        *time.Time `json:"cancel_at,omitempty"`
	StripeCustomerID *string   `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID *string `json:"stripe_subscription_id,omitempty"`
}

// TenantRepository handles tenant data access
type TenantRepository struct {
	db *sql.DB
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *sql.DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// GetByID retrieves a tenant by ID
func (r *TenantRepository) GetByID(ctx context.Context, tenantID uuid.UUID) (*Tenant, error) {
	query := `
		SELECT id, name, slug, domain, status, settings, subscription,
		       created_at, updated_at, trial_ends_at, suspended_at, suspend_reason
		FROM tenants
		WHERE id = $1 AND deleted_at IS NULL
	`

	var tenant Tenant
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(
		&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Domain,
		&tenant.Status, &tenant.Settings, &tenant.Subscription,
		&tenant.CreatedAt, &tenant.UpdatedAt,
		&tenant.TrialEndsAt, &tenant.SuspendedAt, &tenant.SuspendReason,
	)

	if err == sql.ErrNoRows {
		return nil, ErrTenantNotFound
	}
	if err != nil {
		return nil, err
	}

	return &tenant, nil
}

// GetBySlug retrieves a tenant by slug
func (r *TenantRepository) GetBySlug(ctx context.Context, slug string) (*Tenant, error) {
	query := `
		SELECT id, name, slug, domain, status, settings, subscription,
		       created_at, updated_at, trial_ends_at, suspended_at, suspend_reason
		FROM tenants
		WHERE slug = $1 AND deleted_at IS NULL
	`

	var tenant Tenant
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Domain,
		&tenant.Status, &tenant.Settings, &tenant.Subscription,
		&tenant.CreatedAt, &tenant.UpdatedAt,
		&tenant.TrialEndsAt, &tenant.SuspendedAt, &tenant.SuspendReason,
	)

	if err == sql.ErrNoRows {
		return nil, ErrTenantNotFound
	}
	if err != nil {
		return nil, err
	}

	return &tenant, nil
}

// GetByDomain retrieves a tenant by custom domain
func (r *TenantRepository) GetByDomain(ctx context.Context, domain string) (*Tenant, error) {
	query := `
		SELECT id, name, slug, domain, status, settings, subscription,
		       created_at, updated_at, trial_ends_at, suspended_at, suspend_reason
		FROM tenants
		WHERE domain = $1 AND deleted_at IS NULL
	`

	var tenant Tenant
	err := r.db.QueryRowContext(ctx, query, domain).Scan(
		&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Domain,
		&tenant.Status, &tenant.Settings, &tenant.Subscription,
		&tenant.CreatedAt, &tenant.UpdatedAt,
		&tenant.TrialEndsAt, &tenant.SuspendedAt, &tenant.SuspendReason,
	)

	if err == sql.ErrNoRows {
		return nil, ErrTenantNotFound
	}
	if err != nil {
		return nil, err
	}

	return &tenant, nil
}

// Create creates a new tenant
func (r *TenantRepository) Create(ctx context.Context, tenant *Tenant) error {
	query := `
		INSERT INTO tenants (id, name, slug, domain, status, settings, subscription,
		                     created_at, updated_at, trial_ends_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.db.ExecContext(ctx, query,
		tenant.ID, tenant.Name, tenant.Slug, tenant.Domain,
		tenant.Status, tenant.Settings, tenant.Subscription,
		tenant.CreatedAt, tenant.UpdatedAt, tenant.TrialEndsAt,
	)

	return err
}

// Update updates an existing tenant
func (r *TenantRepository) Update(ctx context.Context, tenant *Tenant) error {
	query := `
		UPDATE tenants
		SET name = $2, slug = $3, domain = $4, status = $5,
		    settings = $6, subscription = $7, updated_at = $8,
		    trial_ends_at = $9, suspended_at = $10, suspend_reason = $11
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		tenant.ID, tenant.Name, tenant.Slug, tenant.Domain,
		tenant.Status, tenant.Settings, tenant.Subscription,
		tenant.UpdatedAt, tenant.TrialEndsAt,
		tenant.SuspendedAt, tenant.SuspendReason,
	)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrTenantNotFound
	}

	return nil
}

// Delete soft deletes a tenant
func (r *TenantRepository) Delete(ctx context.Context, tenantID uuid.UUID) error {
	query := `
		UPDATE tenants
		SET deleted_at = $2, updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, tenantID, time.Now())
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrTenantNotFound
	}

	return nil
}

// IsActive checks if tenant is active and can be used
func (t *Tenant) IsActive() bool {
	return t.Status == TenantStatusActive || t.Status == TenantStatusTrial
}

// IsTrialExpired checks if trial period has expired
func (t *Tenant) IsTrialExpired() bool {
	if t.Status != TenantStatusTrial || t.TrialEndsAt == nil {
		return false
	}
	return time.Now().After(*t.TrialEndsAt)
}

// CanAccessFeature checks if tenant has access to a specific feature
func (t *Tenant) CanAccessFeature(feature string) bool {
	for _, f := range t.Settings.AllowedFeatures {
		if f == feature {
			return true
		}
	}
	return false
}
