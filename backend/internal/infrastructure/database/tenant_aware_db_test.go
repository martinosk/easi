package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEscapeTenantID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal tenant ID",
			input:    "acme-corp",
			expected: "acme-corp",
		},
		{
			name:     "tenant ID with numbers",
			input:    "tenant-123",
			expected: "tenant-123",
		},
		{
			name:     "hypothetical single quote injection attempt",
			input:    "tenant'; DROP TABLE users; --",
			expected: "tenant''; DROP TABLE users; --",
		},
		{
			name:     "multiple single quotes",
			input:    "ten''ant",
			expected: "ten''''ant",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeTenantID(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildSetTenantSQL(t *testing.T) {
	sql := buildSetTenantSQL("acme-corp")
	assert.Equal(t, "SET app.current_tenant = 'acme-corp'", sql)
}

func TestBuildSetTenantSQL_EscapesQuotes(t *testing.T) {
	sql := buildSetTenantSQL("tenant'; DROP TABLE users; --")
	assert.Equal(t, "SET app.current_tenant = 'tenant''; DROP TABLE users; --'", sql)
}

func TestBuildSetLocalTenantSQL(t *testing.T) {
	sql := buildSetLocalTenantSQL("acme-corp")
	assert.Equal(t, "SET LOCAL app.current_tenant = 'acme-corp'", sql)
}
