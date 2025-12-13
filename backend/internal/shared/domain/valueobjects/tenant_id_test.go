package valueobjects

import (
	"easi/backend/internal/shared/domain"
	"testing"
)

func TestNewTenantID(t *testing.T) {
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
			name:    "valid tenant ID with numbers",
			value:   "tenant-123",
			wantErr: false,
		},
		{
			name:    "valid short tenant ID",
			value:   "abc",
			wantErr: false,
		},
		{
			name:    "valid max length tenant ID",
			value:   "a1234567890123456789012345678901234567890123456789", // 50 chars
			wantErr: false,
		},
		{
			name:    "default tenant ID",
			value:   "default",
			wantErr: false,
		},
		{
			name:    "special tenant ID - synthetic monitoring",
			value:   "synthetic-monitoring",
			wantErr: false,
		},
		{
			name:    "special tenant ID - synthetic load test",
			value:   "synthetic-load-test",
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
			name:    "too long",
			value:   "a12345678901234567890123456789012345678901234567890", // 51 chars
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
			name:    "contains spaces",
			value:   "acme corp",
			wantErr: true,
			errType: ErrInvalidTenantIDFormat,
		},
		{
			name:    "contains special characters",
			value:   "acme@corp",
			wantErr: true,
			errType: ErrInvalidTenantIDFormat,
		},
		{
			name:    "contains underscore",
			value:   "acme_corp",
			wantErr: true,
			errType: ErrInvalidTenantIDFormat,
		},
		{
			name:    "reserved ID - system",
			value:   "system",
			wantErr: true,
			errType: ErrReservedTenantID,
		},
		{
			name:    "reserved ID - admin",
			value:   "admin",
			wantErr: true,
			errType: ErrReservedTenantID,
		},
		{
			name:    "reserved ID - root",
			value:   "root",
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
				if tt.errType != nil && err != tt.errType && !isWrappedError(err, tt.errType) {
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

func TestTenantID_String(t *testing.T) {
	value := "acme-corp"
	tenantID, _ := NewTenantID(value)

	if tenantID.String() != value {
		t.Errorf("TenantID.String() = %v, want %v", tenantID.String(), value)
	}
}

func TestTenantID_IsDefault(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "default tenant",
			value: "default",
			want:  true,
		},
		{
			name:  "non-default tenant",
			value: "acme-corp",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantID, _ := NewTenantID(tt.value)
			if got := tenantID.IsDefault(); got != tt.want {
				t.Errorf("TenantID.IsDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTenantID_IsSpecial(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "default tenant",
			value: "default",
			want:  true,
		},
		{
			name:  "synthetic monitoring",
			value: "synthetic-monitoring",
			want:  true,
		},
		{
			name:  "synthetic load test",
			value: "synthetic-load-test",
			want:  true,
		},
		{
			name:  "regular tenant",
			value: "acme-corp",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantID, _ := NewTenantID(tt.value)
			if got := tenantID.IsSpecial(); got != tt.want {
				t.Errorf("TenantID.IsSpecial() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMustNewTenantID(t *testing.T) {
	t.Run("valid tenant ID", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("MustNewTenantID() panicked unexpectedly: %v", r)
			}
		}()
		tenantID := MustNewTenantID("acme-corp")
		if tenantID.Value() != "acme-corp" {
			t.Errorf("MustNewTenantID().Value() = %v, want %v", tenantID.Value(), "acme-corp")
		}
	})

	t.Run("invalid tenant ID panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustNewTenantID() should have panicked for invalid ID")
			}
		}()
		MustNewTenantID("INVALID")
	})
}

// isWrappedError checks if err wraps target error
func isWrappedError(err, target error) bool {
	if err == target {
		return true
	}
	// Check if error message contains target error message
	return err != nil && target != nil &&
		len(target.Error()) > 0 &&
		len(err.Error()) >= len(target.Error()) &&
		err.Error()[:len(target.Error())] == target.Error()
}
