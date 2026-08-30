package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTenantStatus_Valid(t *testing.T) {
	tests := []struct {
		input    string
		expected TenantStatus
	}{
		{"active", TenantStatusActive},
		{"suspended", TenantStatusSuspended},
		{"archived", TenantStatusArchived},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			status, err := NewTenantStatus(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, status)
		})
	}
}

func TestNewTenantStatus_Invalid(t *testing.T) {
	_, err := NewTenantStatus("invalid")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidTenantStatus, err)
}

func TestNewTenantStatus_Empty(t *testing.T) {
	_, err := NewTenantStatus("")
	assert.Error(t, err)
}

func TestTenantStatus_String(t *testing.T) {
	assert.Equal(t, "active", TenantStatusActive.String())
	assert.Equal(t, "suspended", TenantStatusSuspended.String())
	assert.Equal(t, "archived", TenantStatusArchived.String())
}

func TestTenantStatus_IsActive(t *testing.T) {
	assert.True(t, TenantStatusActive.IsActive())
	assert.False(t, TenantStatusSuspended.IsActive())
	assert.False(t, TenantStatusArchived.IsActive())
}
