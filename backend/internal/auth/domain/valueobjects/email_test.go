package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEmail_ValidFormats(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
		domain   string
	}{
		{
			name:     "simple email",
			input:    "user@example.com",
			expected: "user@example.com",
			domain:   "example.com",
		},
		{
			name:     "email with uppercase",
			input:    "User@Example.COM",
			expected: "user@example.com",
			domain:   "example.com",
		},
		{
			name:     "email with whitespace",
			input:    "  user@example.com  ",
			expected: "user@example.com",
			domain:   "example.com",
		},
		{
			name:     "email with subdomain",
			input:    "admin@mail.acme.com",
			expected: "admin@mail.acme.com",
			domain:   "mail.acme.com",
		},
		{
			name:     "email with plus addressing",
			input:    "user+tag@example.com",
			expected: "user+tag@example.com",
			domain:   "example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			email, err := NewEmail(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, email.Value())
			assert.Equal(t, tc.domain, email.Domain())
		})
	}
}

func TestNewEmail_InvalidFormats(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expectedErr error
	}{
		{
			name:        "empty string",
			input:       "",
			expectedErr: ErrEmailEmpty,
		},
		{
			name:        "whitespace only",
			input:       "   ",
			expectedErr: ErrEmailEmpty,
		},
		{
			name:        "missing @",
			input:       "userexample.com",
			expectedErr: ErrInvalidEmailFormat,
		},
		{
			name:        "missing local part",
			input:       "@example.com",
			expectedErr: ErrInvalidEmailFormat,
		},
		{
			name:        "missing domain",
			input:       "user@",
			expectedErr: ErrInvalidEmailFormat,
		},
		{
			name:        "multiple @",
			input:       "user@@example.com",
			expectedErr: ErrInvalidEmailFormat,
		},
		{
			name:        "invalid characters",
			input:       "user name@example.com",
			expectedErr: ErrInvalidEmailFormat,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewEmail(tc.input)
			assert.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestEmail_DomainExtraction(t *testing.T) {
	email, err := NewEmail("jane.doe@engineering.acme.com")
	require.NoError(t, err)
	assert.Equal(t, "engineering.acme.com", email.Domain())
}

func TestEmail_Equals(t *testing.T) {
	email1, _ := NewEmail("user@example.com")
	email2, _ := NewEmail("USER@example.com")
	email3, _ := NewEmail("other@example.com")

	assert.True(t, email1.Equals(email2))
	assert.False(t, email1.Equals(email3))
}
