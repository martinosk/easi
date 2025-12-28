package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"testing"
)

func TestNewTenantID_Validation(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
		errType error
	}{
		{
			name:    "valid tenant ID",
			value:   "acme-corp",
			wantErr: false,
		},
		{
			name:    "empty value",
			value:   "",
			wantErr: true,
			errType: domain.ErrEmptyValue,
		},
		{
			name:    "too short",
			value:   "ab",
			wantErr: true,
			errType: ErrInvalidTenantIDFormat,
		},
		{
			name:    "contains uppercase",
			value:   "Acme-Corp",
			wantErr: true,
			errType: ErrInvalidTenantIDFormat,
		},
		{
			name:    "reserved ID - system",
			value:   "system",
			wantErr: true,
			errType: ErrReservedTenantID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantID, err := NewTenantID(tt.value)
			if tt.wantErr {
				if err == nil {
					t.Errorf("NewTenantID() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if tt.errType != nil && err != tt.errType {
					t.Errorf("NewTenantID() error = %v, want error type %v", err, tt.errType)
				}
				return
			}
			if err != nil {
				t.Errorf("NewTenantID() unexpected error = %v", err)
				return
			}
			if tenantID.Value() != tt.value {
				t.Errorf("NewTenantID().Value() = %v, want %v", tenantID.Value(), tt.value)
			}
		})
	}
}

func TestDefaultTenantID(t *testing.T) {
	tenantID := DefaultTenantID()
	if tenantID.Value() != "default" {
		t.Errorf("DefaultTenantID().Value() = %v, want %v", tenantID.Value(), "default")
	}
	if !tenantID.IsDefault() {
		t.Error("DefaultTenantID().IsDefault() = false, want true")
	}
}

func TestTenantID_Equals(t *testing.T) {
	tenant1, _ := NewTenantID("acme-corp")
	tenant2, _ := NewTenantID("acme-corp")
	tenant3, _ := NewTenantID("other-corp")

	if !tenant1.Equals(tenant2) {
		t.Error("Expected equal tenant IDs to be equal")
	}
	if tenant1.Equals(tenant3) {
		t.Error("Expected different tenant IDs to not be equal")
	}
}

func TestTenantID_IsSpecial(t *testing.T) {
	defaultTenant := DefaultTenantID()
	if !defaultTenant.IsSpecial() {
		t.Error("Default tenant should be special")
	}

	syntheticMonitoring, _ := NewTenantID("synthetic-monitoring")
	if !syntheticMonitoring.IsSpecial() {
		t.Error("synthetic-monitoring should be special")
	}

	regularTenant, _ := NewTenantID("acme-corp")
	if regularTenant.IsSpecial() {
		t.Error("Regular tenant should not be special")
	}
}

func TestMustNewTenantID_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustNewTenantID() should have panicked for invalid ID")
		}
	}()
	MustNewTenantID("INVALID")
}
