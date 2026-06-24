package secrets

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTenantID = "acme-corp"

func writeTenantSecret(t *testing.T, secretName string, content []byte) *FileSecretProvider {
	t.Helper()
	tempDir := t.TempDir()
	tenantDir := filepath.Join(tempDir, testTenantID)
	require.NoError(t, os.MkdirAll(tenantDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tenantDir, secretName), content, 0600))
	return NewFileSecretProvider(tempDir)
}

func TestFileSecretProvider_GetClientSecret(t *testing.T) {
	tests := []struct {
		name     string
		stored   string
		expected string
	}{
		{"returns secret", "super-secret-client-secret", "super-secret-client-secret"},
		{"trims whitespace", "  secret-with-whitespace  \n", "secret-with-whitespace"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := writeTenantSecret(t, "client-secret", []byte(tt.stored))
			secret, err := provider.GetClientSecret(context.Background(), testTenantID)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, secret)
		})
	}
}

func TestFileSecretProvider_GetClientSecret_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	provider := NewFileSecretProvider(tempDir)

	_, err := provider.GetClientSecret(context.Background(), "nonexistent-tenant")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSecretNotFound)
}

func TestFileSecretProvider_GetPrivateKey(t *testing.T) {
	privateKey := []byte("-----BEGIN RSA PRIVATE KEY-----\ntest-key-content\n-----END RSA PRIVATE KEY-----")

	provider := writeTenantSecret(t, "private-key", privateKey)
	key, err := provider.GetPrivateKey(context.Background(), testTenantID)

	assert.NoError(t, err)
	assert.Equal(t, privateKey, key)
}

func TestFileSecretProvider_GetPrivateKey_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	provider := NewFileSecretProvider(tempDir)

	_, err := provider.GetPrivateKey(context.Background(), "nonexistent-tenant")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSecretNotFound)
}

func TestFileSecretProvider_GetCertificate(t *testing.T) {
	certificate := []byte("-----BEGIN CERTIFICATE-----\ntest-cert-content\n-----END CERTIFICATE-----")

	provider := writeTenantSecret(t, "certificate", certificate)
	cert, err := provider.GetCertificate(context.Background(), testTenantID)

	assert.NoError(t, err)
	assert.Equal(t, certificate, cert)
}

func TestFileSecretProvider_IsProvisioned(t *testing.T) {
	tests := []struct {
		name       string
		secretName string
	}{
		{"client secret", "client-secret"},
		{"private key jwt", "private-key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := writeTenantSecret(t, tt.secretName, []byte("secret"))

			assert.True(t, provider.IsProvisioned(context.Background(), testTenantID))
			assert.False(t, provider.IsProvisioned(context.Background(), "nonexistent-tenant"))
		})
	}
}
