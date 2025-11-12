package valueobjects

import (
	"easi/backend/internal/shared/domain"
	"fmt"
	"regexp"
)

var (
	// tenantIDPattern matches lowercase alphanumeric and hyphens, 3-50 characters
	tenantIDPattern = regexp.MustCompile(`^[a-z0-9-]{3,50}$`)

	// reservedTenantIDs are IDs that cannot be used for regular tenants
	reservedTenantIDs = map[string]bool{
		"system": true,
		"admin":  true,
		"root":   true,
	}

	// ErrInvalidTenantIDFormat is returned when tenant ID doesn't match required pattern
	ErrInvalidTenantIDFormat = fmt.Errorf("%w: tenant ID must be lowercase alphanumeric with hyphens, 3-50 characters", domain.ErrInvalidValue)

	// ErrReservedTenantID is returned when trying to use a reserved tenant ID
	ErrReservedTenantID = fmt.Errorf("%w: tenant ID is reserved for system use", domain.ErrInvalidValue)
)

// TenantID represents a unique identifier for a tenant
// Tenants are completely isolated from each other at all layers
type TenantID struct {
	value string
}

// DefaultTenantID returns the default tenant ID used for single-tenant deployments
func DefaultTenantID() TenantID {
	return TenantID{value: "default"}
}

// NewTenantID creates a new tenant ID from a string value
// Returns error if the value doesn't match the required pattern or is reserved
func NewTenantID(value string) (TenantID, error) {
	if value == "" {
		return TenantID{}, domain.ErrEmptyValue
	}

	// Validate pattern
	if !tenantIDPattern.MatchString(value) {
		return TenantID{}, ErrInvalidTenantIDFormat
	}

	// Check for reserved IDs
	if reservedTenantIDs[value] {
		return TenantID{}, ErrReservedTenantID
	}

	return TenantID{value: value}, nil
}

// MustNewTenantID creates a new tenant ID from a string value
// Panics if the value is invalid - use only for constants and tests
func MustNewTenantID(value string) TenantID {
	tid, err := NewTenantID(value)
	if err != nil {
		panic(fmt.Sprintf("invalid tenant ID: %v", err))
	}
	return tid
}

// Value returns the string value of the tenant ID
func (t TenantID) Value() string {
	return t.value
}

// Equals checks if two tenant IDs are equal
func (t TenantID) Equals(other domain.ValueObject) bool {
	if otherID, ok := other.(TenantID); ok {
		return t.value == otherID.value
	}
	return false
}

// String implements the Stringer interface
func (t TenantID) String() string {
	return t.value
}

// IsDefault returns true if this is the default tenant
func (t TenantID) IsDefault() bool {
	return t.value == "default"
}

// IsSpecial returns true if this is a special tenant (synthetic monitoring, load test, etc.)
func (t TenantID) IsSpecial() bool {
	return t.value == "synthetic-monitoring" ||
		t.value == "synthetic-load-test" ||
		t.value == "default"
}
