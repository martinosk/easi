package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmailDomain_Valid(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple domain", "acme.com", "acme.com"},
		{"subdomain", "mail.acme.com", "mail.acme.com"},
		{"with numbers", "acme123.com", "acme123.com"},
		{"with hyphen", "acme-corp.com", "acme-corp.com"},
		{"multiple subdomains", "mail.server.acme.com", "mail.server.acme.com"},
		{"co.uk domain", "acme.co.uk", "acme.co.uk"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain, err := NewEmailDomain(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, domain.Value())
		})
	}
}

func TestNewEmailDomain_NormalizesToLowercase(t *testing.T) {
	domain, err := NewEmailDomain("ACME.COM")
	assert.NoError(t, err)
	assert.Equal(t, "acme.com", domain.Value())
}

func TestNewEmailDomain_TrimsWhitespace(t *testing.T) {
	domain, err := NewEmailDomain("  acme.com  ")
	assert.NoError(t, err)
	assert.Equal(t, "acme.com", domain.Value())
}

func TestNewEmailDomain_Empty(t *testing.T) {
	_, err := NewEmailDomain("")
	assert.Error(t, err)
	assert.Equal(t, ErrEmailDomainEmpty, err)
}

func TestNewEmailDomain_OnlyWhitespace(t *testing.T) {
	_, err := NewEmailDomain("   ")
	assert.Error(t, err)
	assert.Equal(t, ErrEmailDomainEmpty, err)
}

func TestNewEmailDomain_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"no tld", "acme"},
		{"starts with hyphen", "-acme.com"},
		{"ends with hyphen", "acme-.com"},
		{"starts with dot", ".acme.com"},
		{"ends with dot", "acme.com."},
		{"double dot", "acme..com"},
		{"contains space", "acme com.com"},
		{"contains underscore", "acme_corp.com"},
		{"too short", "a.b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewEmailDomain(tt.input)
			assert.Error(t, err)
			assert.Equal(t, ErrInvalidEmailDomain, err)
		})
	}
}

func TestEmailDomain_String(t *testing.T) {
	domain, _ := NewEmailDomain("acme.com")
	assert.Equal(t, "acme.com", domain.String())
}

func TestEmailDomain_Equals(t *testing.T) {
	domain1, _ := NewEmailDomain("acme.com")
	domain2, _ := NewEmailDomain("ACME.COM")
	domain3, _ := NewEmailDomain("other.com")

	assert.True(t, domain1.Equals(domain2))
	assert.False(t, domain1.Equals(domain3))
}

func TestNewEmailDomainList_Valid(t *testing.T) {
	domains, err := NewEmailDomainList([]string{"acme.com", "acme.co.uk"})
	assert.NoError(t, err)
	assert.Len(t, domains, 2)
	assert.Equal(t, "acme.com", domains[0].Value())
	assert.Equal(t, "acme.co.uk", domains[1].Value())
}

func TestNewEmailDomainList_Empty(t *testing.T) {
	_, err := NewEmailDomainList([]string{})
	assert.Error(t, err)
	assert.Equal(t, ErrEmailDomainListEmpty, err)
}

func TestNewEmailDomainList_WithInvalid(t *testing.T) {
	_, err := NewEmailDomainList([]string{"acme.com", "invalid"})
	assert.Error(t, err)
}

func TestNewEmailDomainList_Duplicates(t *testing.T) {
	_, err := NewEmailDomainList([]string{"acme.com", "ACME.COM"})
	assert.Error(t, err)
	assert.Equal(t, ErrDuplicateEmailDomain, err)
}
