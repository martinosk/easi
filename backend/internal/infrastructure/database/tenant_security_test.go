package database

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLInjectionProtection_TenantID(t *testing.T) {
	injectionAttempts := []struct {
		name       string
		tenantID   string
		shouldFail bool
	}{
		{"SQL comment injection attempt with hyphens", "tenant--drop-table", false},
		{"Single quote SQL injection", "tenant'; DROP TABLE users; --", true},
		{"Semicolon command separator", "tenant; DROP TABLE users", true},
		{"Union-based injection", "tenant' UNION SELECT * FROM users--", true},
		{"Null byte injection", "tenant\x00malicious", true},
		{"Unicode normalization attack", "tenant\u2019", true},
		{"Case manipulation attack", "Tenant-Corp", true},
		{"Space injection", "tenant id", true},
		{"Backslash escape attempt", "tenant\\", true},
		{"Valid tenant with hyphens", "acme-corp-123", false},
	}

	for _, tt := range injectionAttempts {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sharedvo.NewTenantID(tt.tenantID)
			if tt.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEscapeTenantID_DefenseInDepth(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedEscaped string
	}{
		{"normal tenant", "acme-corp", "acme-corp"},
		{"single quote escape", "tenant'test", "tenant''test"},
		{"multiple quotes", "a'b'c", "a''b''c"},
		{"quote at start", "'tenant", "''tenant"},
		{"quote at end", "tenant'", "tenant''"},
		{"consecutive quotes", "tenant''test", "tenant''''test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeTenantID(tt.input)
			assert.Equal(t, tt.expectedEscaped, result)
		})
	}
}

func TestBuildSetTenantSQL_SafeConstruction(t *testing.T) {
	tests := []struct {
		name        string
		tenantID    string
		expectedSQL string
	}{
		{"normal tenant", "acme-corp", "SET app.current_tenant = 'acme-corp'"},
		{"escaped quotes", "tenant'malicious", "SET app.current_tenant = 'tenant''malicious'"},
		{"attempted SQL injection", "'; DROP TABLE users; --", "SET app.current_tenant = '''; DROP TABLE users; --'"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql := buildSetTenantSQL(tt.tenantID)
			assert.Equal(t, tt.expectedSQL, sql)
		})
	}
}

func TestBuildSetLocalTenantSQL_TransactionScoped(t *testing.T) {
	sql := buildSetLocalTenantSQL("acme-corp")
	assert.Equal(t, "SET LOCAL app.current_tenant = 'acme-corp'", sql)
	assert.Contains(t, sql, "SET LOCAL")
}

func TestTenantContextPropagation_MissingContext(t *testing.T) {
	mockDB := &sql.DB{}
	tenantDB := NewTenantAwareDB(mockDB)

	ctx := context.Background()
	conn := &sql.Conn{}

	err := tenantDB.setTenantContext(ctx, conn)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get tenant from context")
}

func TestTenantContextPropagation_ValidContext(t *testing.T) {
	tenantID := sharedvo.MustNewTenantID("acme-corp")
	ctx := sharedctx.WithTenant(context.Background(), tenantID)

	retrieved, err := sharedctx.GetTenant(ctx)
	require.NoError(t, err)
	assert.Equal(t, tenantID.Value(), retrieved.Value())
}

func TestNullTenantID_Handling(t *testing.T) {
	_, err := sharedvo.NewTenantID("")
	assert.Error(t, err)
}

func TestReservedTenantIDs_Prevention(t *testing.T) {
	reservedIDs := []string{"system", "admin", "root"}

	for _, id := range reservedIDs {
		t.Run(id, func(t *testing.T) {
			_, err := sharedvo.NewTenantID(id)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "reserved")
		})
	}
}

func TestTenantIDLength_Constraints(t *testing.T) {
	tests := []struct {
		name     string
		tenantID string
		valid    bool
	}{
		{"too short - 2 chars", "ab", false},
		{"minimum valid - 3 chars", "abc", true},
		{"maximum valid - 50 chars", "a123456789-123456789-123456789-123456789-123456789", true},
		{"too long - 51 chars", "a123456789-123456789-123456789-123456789-1234567890", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sharedvo.NewTenantID(tt.tenantID)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestSpecialCharacters_Prevention(t *testing.T) {
	specialChars := []struct {
		name     string
		tenantID string
	}{
		{"underscore", "tenant_id"},
		{"period", "tenant.id"},
		{"slash", "tenant/id"},
		{"backslash", "tenant\\id"},
		{"asterisk", "tenant*id"},
		{"percent", "tenant%id"},
		{"dollar", "tenant$id"},
		{"at sign", "tenant@id"},
		{"exclamation", "tenant!id"},
		{"parentheses", "tenant(id)"},
		{"brackets", "tenant[id]"},
		{"braces", "tenant{id}"},
	}

	for _, tt := range specialChars {
		t.Run(tt.name, func(t *testing.T) {
			_, err := sharedvo.NewTenantID(tt.tenantID)
			assert.Error(t, err)
		})
	}
}

func TestHeaderInjection_Prevention(t *testing.T) {
	injectionAttempts := []string{
		"tenant\r\nX-Admin: true",
		"tenant\nX-Admin: true",
		"tenant\r\n\r\nGET /admin",
		"tenant%0d%0aX-Admin:%20true",
	}

	for _, attempt := range injectionAttempts {
		t.Run(fmt.Sprintf("Attempt: %s", attempt), func(t *testing.T) {
			_, err := sharedvo.NewTenantID(attempt)
			assert.Error(t, err)
		})
	}
}

func TestSetLocalVsSet_TransactionScope(t *testing.T) {
	t.Run("SET LOCAL is used for transactions", func(t *testing.T) {
		sql := buildSetLocalTenantSQL("acme-corp")
		assert.Contains(t, sql, "SET LOCAL")
		assert.NotContains(t, sql, "SET app.current_tenant")
	})

	t.Run("SET is used for connection scope", func(t *testing.T) {
		sql := buildSetTenantSQL("acme-corp")
		assert.Contains(t, sql, "SET app.current_tenant")
		assert.NotContains(t, sql, "LOCAL")
	})
}

func TestDefaultTenant_Handling(t *testing.T) {
	defaultTenant := sharedvo.DefaultTenantID()
	assert.Equal(t, "default", defaultTenant.Value())
	assert.True(t, defaultTenant.IsDefault())
	assert.True(t, defaultTenant.IsSpecial())
}

func TestTenantIDPattern_Validation(t *testing.T) {
	t.Run("Must start with alphanumeric", func(t *testing.T) {
		_, err := sharedvo.NewTenantID("-tenant")
		assert.Error(t, err)

		_, err = sharedvo.NewTenantID("_tenant")
		assert.Error(t, err)
	})

	t.Run("Must end with alphanumeric", func(t *testing.T) {
		_, err := sharedvo.NewTenantID("tenant-")
		assert.Error(t, err)

		_, err = sharedvo.NewTenantID("tenant_")
		assert.Error(t, err)
	})

	t.Run("Can contain hyphens in middle", func(t *testing.T) {
		_, err := sharedvo.NewTenantID("tenant-id-123")
		assert.NoError(t, err)

		_, err = sharedvo.NewTenantID("a-b-c-d-e")
		assert.NoError(t, err)
	})

	t.Run("Minimum length enforced", func(t *testing.T) {
		_, err := sharedvo.NewTenantID("--")
		assert.Error(t, err)

		_, err = sharedvo.NewTenantID("ab")
		assert.Error(t, err)
	})

	t.Run("Three character tenant IDs", func(t *testing.T) {
		_, err := sharedvo.NewTenantID("abc")
		assert.NoError(t, err)

		_, err = sharedvo.NewTenantID("123")
		assert.NoError(t, err)

		_, err = sharedvo.NewTenantID("a-b")
		assert.Error(t, err)
	})
}

func TestRLSPolicyNaming_Consistency(t *testing.T) {
	settingName := "app.current_tenant"
	setSQL := buildSetTenantSQL("test")
	setLocalSQL := buildSetLocalTenantSQL("test")

	assert.Contains(t, setSQL, settingName)
	assert.Contains(t, setLocalSQL, settingName)
}

func TestEmptyTenantValue_AfterValidation(t *testing.T) {
	tenantID, err := sharedvo.NewTenantID("valid-tenant")
	require.NoError(t, err)

	value := tenantID.Value()
	assert.NotEmpty(t, value)
	assert.Len(t, value, len("valid-tenant"))
}

func TestTenantEquality_ValueObject(t *testing.T) {
	tenant1 := sharedvo.MustNewTenantID("acme-corp")
	tenant2 := sharedvo.MustNewTenantID("acme-corp")
	tenant3 := sharedvo.MustNewTenantID("other-corp")

	assert.True(t, tenant1.Equals(tenant2))
	assert.False(t, tenant1.Equals(tenant3))
}
