package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOIDCConfig_ValidWithClientSecretMethod(t *testing.T) {
	config, err := NewOIDCConfig(
		"https://login.microsoftonline.com/xxx/v2.0/.well-known/openid-configuration",
		"client-id",
		OIDCAuthMethodClientSecret,
		"openid email profile",
	)
	assert.NoError(t, err)
	assert.Equal(t, "https://login.microsoftonline.com/xxx/v2.0/.well-known/openid-configuration", config.DiscoveryURL())
	assert.Equal(t, "client-id", config.ClientID())
	assert.Equal(t, OIDCAuthMethodClientSecret, config.AuthMethod())
	assert.Equal(t, "openid email profile", config.Scopes())
}

func TestNewOIDCConfig_ValidWithPrivateKeyJWTMethod(t *testing.T) {
	config, err := NewOIDCConfig(
		"https://login.microsoftonline.com/xxx/v2.0/.well-known/openid-configuration",
		"client-id",
		OIDCAuthMethodPrivateKeyJWT,
		"openid email profile offline_access",
	)
	assert.NoError(t, err)
	assert.Equal(t, OIDCAuthMethodPrivateKeyJWT, config.AuthMethod())
}

func TestNewOIDCConfig_ValidationErrors(t *testing.T) {
	const validURL = "https://example.com/.well-known/openid-configuration"

	tests := []struct {
		name         string
		discoveryURL string
		clientID     string
		authMethod   OIDCAuthMethod
		expectedErr  error
	}{
		{"empty discovery url", "", "client-id", OIDCAuthMethodClientSecret, ErrOIDCDiscoveryURLEmpty},
		{"invalid discovery url", "not-a-url", "client-id", OIDCAuthMethodClientSecret, ErrOIDCDiscoveryURLInvalid},
		{"http discovery url", "http://example.com/.well-known/openid-configuration", "client-id", OIDCAuthMethodClientSecret, ErrOIDCDiscoveryURLNotHTTPS},
		{"empty client id", validURL, "", OIDCAuthMethodClientSecret, ErrOIDCClientIDEmpty},
		{"invalid auth method", validURL, "client-id", OIDCAuthMethod("invalid"), ErrOIDCAuthMethodInvalid},
		{"empty auth method", validURL, "client-id", OIDCAuthMethod(""), ErrOIDCAuthMethodInvalid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewOIDCConfig(tt.discoveryURL, tt.clientID, tt.authMethod, "openid")
			assert.Error(t, err)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestNewOIDCConfig_DefaultScopes(t *testing.T) {
	config, err := NewOIDCConfig(
		"https://example.com/.well-known/openid-configuration",
		"client-id",
		OIDCAuthMethodClientSecret,
		"",
	)
	assert.NoError(t, err)
	assert.Equal(t, "openid email profile", config.Scopes())
}

func TestOIDCAuthMethod_IsValid(t *testing.T) {
	tests := []struct {
		method   OIDCAuthMethod
		expected bool
	}{
		{OIDCAuthMethodClientSecret, true},
		{OIDCAuthMethodPrivateKeyJWT, true},
		{OIDCAuthMethod("invalid"), false},
		{OIDCAuthMethod(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.method), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.method.IsValid())
		})
	}
}

func TestOIDCConfig_Equals(t *testing.T) {
	config1, _ := NewOIDCConfig(
		"https://example.com/.well-known/openid-configuration",
		"client-id",
		OIDCAuthMethodClientSecret,
		"openid",
	)
	config2, _ := NewOIDCConfig(
		"https://example.com/.well-known/openid-configuration",
		"client-id",
		OIDCAuthMethodClientSecret,
		"openid",
	)
	config3, _ := NewOIDCConfig(
		"https://example.com/.well-known/openid-configuration",
		"client-id",
		OIDCAuthMethodPrivateKeyJWT,
		"openid",
	)

	assert.True(t, config1.Equals(config2))
	assert.False(t, config1.Equals(config3))
}
