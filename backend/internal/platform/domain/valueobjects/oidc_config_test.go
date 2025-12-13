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

func TestNewOIDCConfig_EmptyDiscoveryURL(t *testing.T) {
	_, err := NewOIDCConfig("", "client-id", OIDCAuthMethodClientSecret, "openid")
	assert.Error(t, err)
	assert.Equal(t, ErrOIDCDiscoveryURLEmpty, err)
}

func TestNewOIDCConfig_InvalidDiscoveryURL(t *testing.T) {
	_, err := NewOIDCConfig("not-a-url", "client-id", OIDCAuthMethodClientSecret, "openid")
	assert.Error(t, err)
	assert.Equal(t, ErrOIDCDiscoveryURLInvalid, err)
}

func TestNewOIDCConfig_HTTPDiscoveryURL(t *testing.T) {
	_, err := NewOIDCConfig("http://example.com/.well-known/openid-configuration", "client-id", OIDCAuthMethodClientSecret, "openid")
	assert.Error(t, err)
	assert.Equal(t, ErrOIDCDiscoveryURLNotHTTPS, err)
}

func TestNewOIDCConfig_EmptyClientID(t *testing.T) {
	_, err := NewOIDCConfig("https://example.com/.well-known/openid-configuration", "", OIDCAuthMethodClientSecret, "openid")
	assert.Error(t, err)
	assert.Equal(t, ErrOIDCClientIDEmpty, err)
}

func TestNewOIDCConfig_InvalidAuthMethod(t *testing.T) {
	_, err := NewOIDCConfig(
		"https://example.com/.well-known/openid-configuration",
		"client-id",
		OIDCAuthMethod("invalid"),
		"openid",
	)
	assert.Error(t, err)
	assert.Equal(t, ErrOIDCAuthMethodInvalid, err)
}

func TestNewOIDCConfig_EmptyAuthMethod(t *testing.T) {
	_, err := NewOIDCConfig(
		"https://example.com/.well-known/openid-configuration",
		"client-id",
		OIDCAuthMethod(""),
		"openid",
	)
	assert.Error(t, err)
	assert.Equal(t, ErrOIDCAuthMethodInvalid, err)
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
