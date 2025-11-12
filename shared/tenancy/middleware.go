package tenancy

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// Context keys for tenant information
type contextKey string

const (
	TenantIDKey   contextKey = "tenant_id"
	TenantKey     contextKey = "tenant"
)

// TenantResolver resolves tenant from various sources
type TenantResolver struct {
	repo *TenantRepository
}

// NewTenantResolver creates a new tenant resolver
func NewTenantResolver(repo *TenantRepository) *TenantResolver {
	return &TenantResolver{repo: repo}
}

// ResolveFromRequest resolves tenant from HTTP request
// Priority: Header > Subdomain > Custom Domain
func (r *TenantResolver) ResolveFromRequest(req *http.Request) (*Tenant, error) {
	// 1. Try X-Tenant-ID header (for API requests)
	if tenantIDStr := req.Header.Get("X-Tenant-ID"); tenantIDStr != "" {
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return nil, ErrInvalidTenantID
		}
		return r.repo.GetByID(req.Context(), tenantID)
	}

	// 2. Try X-Tenant-Slug header
	if slug := req.Header.Get("X-Tenant-Slug"); slug != "" {
		return r.repo.GetBySlug(req.Context(), slug)
	}

	// 3. Try subdomain extraction from Host
	host := req.Host
	if strings.Contains(host, ":") {
		// Remove port
		host = strings.Split(host, ":")[0]
	}

	// Check for custom domain first
	if !strings.HasSuffix(host, ".marimo-erp.com") && host != "marimo-erp.com" {
		tenant, err := r.repo.GetByDomain(req.Context(), host)
		if err == nil {
			return tenant, nil
		}
		// If custom domain not found, fall through to subdomain check
	}

	// Extract subdomain
	if strings.HasSuffix(host, ".marimo-erp.com") {
		slug := strings.TrimSuffix(host, ".marimo-erp.com")
		// Avoid matching api.marimo-erp.com, www.marimo-erp.com, etc.
		if slug != "" && slug != "api" && slug != "www" {
			return r.repo.GetBySlug(req.Context(), slug)
		}
	}

	return nil, ErrTenantNotFound
}

// ResolveFromContext extracts tenant from context
func ResolveFromContext(ctx context.Context) (*Tenant, error) {
	tenant, ok := ctx.Value(TenantKey).(*Tenant)
	if !ok || tenant == nil {
		return nil, ErrNoTenantInContext
	}
	return tenant, nil
}

// GetTenantID extracts tenant ID from context
func GetTenantID(ctx context.Context) (uuid.UUID, error) {
	tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, ErrNoTenantInContext
	}
	return tenantID, nil
}

// WithTenant adds tenant to context
func WithTenant(ctx context.Context, tenant *Tenant) context.Context {
	ctx = context.WithValue(ctx, TenantKey, tenant)
	ctx = context.WithValue(ctx, TenantIDKey, tenant.ID)
	return ctx
}

// TenantMiddleware is HTTP middleware that resolves and validates tenant
type TenantMiddleware struct {
	resolver *TenantResolver
}

// NewTenantMiddleware creates a new tenant middleware
func NewTenantMiddleware(resolver *TenantResolver) *TenantMiddleware {
	return &TenantMiddleware{resolver: resolver}
}

// Middleware returns the HTTP middleware handler
func (m *TenantMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenant, err := m.resolver.ResolveFromRequest(r)
		if err != nil {
			if errors.Is(err, ErrTenantNotFound) {
				http.Error(w, "Tenant not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Invalid tenant", http.StatusBadRequest)
			return
		}

		// Validate tenant status
		if !tenant.IsActive() {
			if tenant.Status == TenantStatusSuspended {
				http.Error(w, "Tenant is suspended", http.StatusForbidden)
				return
			}
			http.Error(w, "Tenant is not active", http.StatusForbidden)
			return
		}

		// Check trial expiration
		if tenant.IsTrialExpired() {
			http.Error(w, "Trial period has expired", http.StatusPaymentRequired)
			return
		}

		// Add tenant to context
		ctx := WithTenant(r.Context(), tenant)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalMiddleware is like Middleware but doesn't fail if tenant is not found
func (m *TenantMiddleware) OptionalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenant, err := m.resolver.ResolveFromRequest(r)
		if err == nil && tenant != nil && tenant.IsActive() {
			ctx := WithTenant(r.Context(), tenant)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}
