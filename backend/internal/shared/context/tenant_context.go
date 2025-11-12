package context

import (
	"context"
	"errors"

	sharedvo "easi/backend/internal/shared/domain/valueobjects"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// TenantContextKey is the key used to store tenant ID in context
const TenantContextKey contextKey = "tenant_id"

var (
	// ErrTenantContextNotFound is returned when tenant context is missing
	ErrTenantContextNotFound = errors.New("tenant context not found")
)

// WithTenant adds a tenant ID to the context
func WithTenant(ctx context.Context, tenantID sharedvo.TenantID) context.Context {
	return context.WithValue(ctx, TenantContextKey, tenantID)
}

// GetTenant retrieves the tenant ID from context
// Returns error if tenant context is not set
func GetTenant(ctx context.Context) (sharedvo.TenantID, error) {
	tenantID, ok := ctx.Value(TenantContextKey).(sharedvo.TenantID)
	if !ok {
		return sharedvo.TenantID{}, ErrTenantContextNotFound
	}
	return tenantID, nil
}

// GetTenantOrDefault retrieves the tenant ID from context
// Returns default tenant if not set
func GetTenantOrDefault(ctx context.Context) sharedvo.TenantID {
	tenantID, err := GetTenant(ctx)
	if err != nil {
		return sharedvo.DefaultTenantID()
	}
	return tenantID
}

// MustGetTenant retrieves the tenant ID from context
// Panics if tenant context is not set - use only when you're certain context has tenant
func MustGetTenant(ctx context.Context) sharedvo.TenantID {
	tenantID, err := GetTenant(ctx)
	if err != nil {
		panic(err)
	}
	return tenantID
}
