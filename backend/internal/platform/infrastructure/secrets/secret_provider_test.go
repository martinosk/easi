package secrets

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileSecretProvider_GetClientSecret(t *testing.T) {
	tempDir := t.TempDir()
	tenantID := "acme-corp"
	secretValue := "super-secret-client-secret"

	tenantDir := filepath.Join(tempDir, tenantID)
	require.NoError(t, os.MkdirAll(tenantDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tenantDir, "client-secret"), []byte(secretValue), 0600))

	provider := NewFileSecretProvider(tempDir)
	secret, err := provider.GetClientSecret(context.Background(), tenantID)

	assert.NoError(t, err)
	assert.Equal(t, secretValue, secret)
}

func TestFileSecretProvider_GetClientSecret_TrimsWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	tenantID := "acme-corp"
	secretValue := "  secret-with-whitespace  \n"

	tenantDir := filepath.Join(tempDir, tenantID)
	require.NoError(t, os.MkdirAll(tenantDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tenantDir, "client-secret"), []byte(secretValue), 0600))

	provider := NewFileSecretProvider(tempDir)
	secret, err := provider.GetClientSecret(context.Background(), tenantID)

	assert.NoError(t, err)
	assert.Equal(t, "secret-with-whitespace", secret)
}

func TestFileSecretProvider_GetClientSecret_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	provider := NewFileSecretProvider(tempDir)

	_, err := provider.GetClientSecret(context.Background(), "nonexistent-tenant")

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrSecretNotFound)
}

func TestFileSecretProvider_GetPrivateKey(t *testing.T) {
	tempDir := t.TempDir()
	tenantID := "acme-corp"
	privateKey := []byte("-----BEGIN RSA PRIVATE KEY-----\ntest-key-content\n-----END RSA PRIVATE KEY-----")

	tenantDir := filepath.Join(tempDir, tenantID)
	require.NoError(t, os.MkdirAll(tenantDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tenantDir, "private-key"), privateKey, 0600))

	provider := NewFileSecretProvider(tempDir)
	key, err := provider.GetPrivateKey(context.Background(), tenantID)

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
	tempDir := t.TempDir()
	tenantID := "acme-corp"
	certificate := []byte("-----BEGIN CERTIFICATE-----\ntest-cert-content\n-----END CERTIFICATE-----")

	tenantDir := filepath.Join(tempDir, tenantID)
	require.NoError(t, os.MkdirAll(tenantDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tenantDir, "certificate"), certificate, 0600))

	provider := NewFileSecretProvider(tempDir)
	cert, err := provider.GetCertificate(context.Background(), tenantID)

	assert.NoError(t, err)
	assert.Equal(t, certificate, cert)
}

func TestFileSecretProvider_IsProvisioned(t *testing.T) {
	tempDir := t.TempDir()
	tenantID := "acme-corp"

	tenantDir := filepath.Join(tempDir, tenantID)
	require.NoError(t, os.MkdirAll(tenantDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tenantDir, "client-secret"), []byte("secret"), 0600))

	provider := NewFileSecretProvider(tempDir)

	assert.True(t, provider.IsProvisioned(context.Background(), tenantID))
	assert.False(t, provider.IsProvisioned(context.Background(), "nonexistent-tenant"))
}

func TestFileSecretProvider_IsProvisioned_PrivateKeyJWT(t *testing.T) {
	tempDir := t.TempDir()
	tenantID := "acme-corp"

	tenantDir := filepath.Join(tempDir, tenantID)
	require.NoError(t, os.MkdirAll(tenantDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tenantDir, "private-key"), []byte("key"), 0600))

	provider := NewFileSecretProvider(tempDir)

	assert.True(t, provider.IsProvisioned(context.Background(), tenantID))
}
