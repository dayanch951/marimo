package tenancy

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenant_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   TenantStatus
		expected bool
	}{
		{
			name:     "active tenant",
			status:   TenantStatusActive,
			expected: true,
		},
		{
			name:     "trial tenant",
			status:   TenantStatusTrial,
			expected: true,
		},
		{
			name:     "inactive tenant",
			status:   TenantStatusInactive,
			expected: false,
		},
		{
			name:     "suspended tenant",
			status:   TenantStatusSuspended,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant := &Tenant{
				Status: tt.status,
			}

			assert.Equal(t, tt.expected, tenant.IsActive())
		})
	}
}

func TestTenant_IsTrialExpired(t *testing.T) {
	now := time.Now()
	pastDate := now.AddDate(0, 0, -1)
	futureDate := now.AddDate(0, 0, 1)

	tests := []struct {
		name        string
		status      TenantStatus
		trialEndsAt *time.Time
		expected    bool
	}{
		{
			name:        "trial expired",
			status:      TenantStatusTrial,
			trialEndsAt: &pastDate,
			expected:    true,
		},
		{
			name:        "trial active",
			status:      TenantStatusTrial,
			trialEndsAt: &futureDate,
			expected:    false,
		},
		{
			name:        "not on trial",
			status:      TenantStatusActive,
			trialEndsAt: nil,
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenant := &Tenant{
				Status:      tt.status,
				TrialEndsAt: tt.trialEndsAt,
			}

			assert.Equal(t, tt.expected, tenant.IsTrialExpired())
		})
	}
}

func TestTenant_CanAccessFeature(t *testing.T) {
	tenant := &Tenant{
		Settings: Settings{
			AllowedFeatures: []string{"basic", "analytics", "export"},
		},
	}

	tests := []struct {
		name     string
		feature  string
		expected bool
	}{
		{
			name:     "has feature",
			feature:  "analytics",
			expected: true,
		},
		{
			name:     "no feature",
			feature:  "webhooks",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tenant.CanAccessFeature(tt.feature))
		})
	}
}

func TestWithTenant(t *testing.T) {
	ctx := context.Background()
	tenant := &Tenant{
		ID:   uuid.New(),
		Name: "Test Tenant",
	}

	// Add tenant to context
	ctx = WithTenant(ctx, tenant)

	// Retrieve tenant from context
	retrieved, err := ResolveFromContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, tenant.ID, retrieved.ID)
	assert.Equal(t, tenant.Name, retrieved.Name)

	// Retrieve tenant ID
	tenantID, err := GetTenantID(ctx)
	require.NoError(t, err)
	assert.Equal(t, tenant.ID, tenantID)
}

func TestResolveFromContext_NoTenant(t *testing.T) {
	ctx := context.Background()

	_, err := ResolveFromContext(ctx)
	assert.Error(t, err)
	assert.Equal(t, ErrNoTenantInContext, err)

	_, err = GetTenantID(ctx)
	assert.Error(t, err)
	assert.Equal(t, ErrNoTenantInContext, err)
}
