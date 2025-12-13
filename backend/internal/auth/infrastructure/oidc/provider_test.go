package oidc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestNewOIDCProvider_Discovery(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	jwks := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{
			{
				Key:       &privateKey.PublicKey,
				KeyID:     "test-key-1",
				Algorithm: string(jose.RS256),
				Use:       "sig",
			},
		},
	}

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		config := map[string]interface{}{
			"issuer":                 server.URL,
			"authorization_endpoint": server.URL + "/authorize",
			"token_endpoint":         server.URL + "/token",
			"jwks_uri":               server.URL + "/jwks",
			"userinfo_endpoint":      server.URL + "/userinfo",
		}
		json.NewEncoder(w).Encode(config)
	})

	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(jwks)
	})

	provider, err := NewOIDCProviderFromConfig(context.Background(), ProviderConfig{
		DiscoveryURL: server.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  server.URL + "/callback",
	})
	require.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestOIDCProvider_AuthCodeURL(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	jwks := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{
			{
				Key:       &privateKey.PublicKey,
				KeyID:     "test-key-1",
				Algorithm: string(jose.RS256),
				Use:       "sig",
			},
		},
	}

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		config := map[string]interface{}{
			"issuer":                 server.URL,
			"authorization_endpoint": server.URL + "/authorize",
			"token_endpoint":         server.URL + "/token",
			"jwks_uri":               server.URL + "/jwks",
		}
		json.NewEncoder(w).Encode(config)
	})

	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(jwks)
	})

	provider, err := NewOIDCProviderFromConfig(context.Background(), ProviderConfig{
		DiscoveryURL: server.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  server.URL + "/callback",
	})
	require.NoError(t, err)

	state := "test-state"
	nonce := "test-nonce"
	codeVerifier := oauth2.GenerateVerifier()

	url := provider.AuthCodeURL(state, nonce, codeVerifier)

	assert.Contains(t, url, server.URL+"/authorize")
	assert.Contains(t, url, "state="+state)
	assert.Contains(t, url, "nonce="+nonce)
	assert.Contains(t, url, "code_challenge=")
	assert.Contains(t, url, "code_challenge_method=S256")
}

func TestOIDCProvider_ExchangeCode(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	jwks := jose.JSONWebKeySet{
		Keys: []jose.JSONWebKey{
			{
				Key:       &privateKey.PublicKey,
				KeyID:     "test-key-1",
				Algorithm: string(jose.RS256),
				Use:       "sig",
			},
		},
	}

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	expectedNonce := "test-nonce"

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		config := map[string]interface{}{
			"issuer":                 server.URL,
			"authorization_endpoint": server.URL + "/authorize",
			"token_endpoint":         server.URL + "/token",
			"jwks_uri":               server.URL + "/jwks",
		}
		json.NewEncoder(w).Encode(config)
	})

	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(jwks)
	})

	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		assert.Equal(t, "authorization_code", r.Form.Get("grant_type"))
		assert.NotEmpty(t, r.Form.Get("code_verifier"))

		w.Header().Set("Content-Type", "application/json")

		signer, _ := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: privateKey}, &jose.SignerOptions{
			ExtraHeaders: map[jose.HeaderKey]interface{}{
				"kid": "test-key-1",
			},
		})

		claims := jwt.Claims{
			Issuer:   server.URL,
			Subject:  "user-123",
			Audience: jwt.Audience{"test-client-id"},
			Expiry:   jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		}
		customClaims := map[string]interface{}{
			"nonce": expectedNonce,
			"email": "user@example.com",
			"name":  "Test User",
		}

		idToken, _ := jwt.Signed(signer).Claims(claims).Claims(customClaims).Serialize()

		response := map[string]interface{}{
			"access_token":  "test-access-token",
			"token_type":    "Bearer",
			"refresh_token": "test-refresh-token",
			"expires_in":    3600,
			"id_token":      idToken,
		}
		json.NewEncoder(w).Encode(response)
	})

	provider, err := NewOIDCProviderFromConfig(context.Background(), ProviderConfig{
		DiscoveryURL: server.URL,
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  server.URL + "/callback",
	})
	require.NoError(t, err)

	codeVerifier := oauth2.GenerateVerifier()
	result, err := provider.ExchangeCode(context.Background(), "test-code", codeVerifier, expectedNonce)
	require.NoError(t, err)

	assert.Equal(t, "test-access-token", result.AccessToken)
	assert.Equal(t, "test-refresh-token", result.RefreshToken)
	assert.Equal(t, "user-123", result.Subject)
	assert.Equal(t, "user@example.com", result.Email)
	assert.Equal(t, "Test User", result.Name)
}
