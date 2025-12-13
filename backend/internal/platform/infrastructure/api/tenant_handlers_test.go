package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTenantRequest_Validation(t *testing.T) {
	tests := []struct {
		name        string
		request     CreateTenantRequest
		expectError bool
	}{
		{
			name: "valid request",
			request: CreateTenantRequest{
				ID:      "acme",
				Name:    "Acme Corporation",
				Domains: []string{"acme.com"},
				OIDCConfig: OIDCConfigRequest{
					DiscoveryURL: "https://login.microsoftonline.com/xxx/v2.0/.well-known/openid-configuration",
					ClientID:     "client-id",
					AuthMethod:   "client_secret",
					Scopes:       "openid email profile",
				},
				FirstAdminEmail: "admin@acme.com",
			},
			expectError: false,
		},
		{
			name: "empty id",
			request: CreateTenantRequest{
				ID:      "",
				Name:    "Acme Corporation",
				Domains: []string{"acme.com"},
				OIDCConfig: OIDCConfigRequest{
					DiscoveryURL: "https://login.microsoftonline.com/xxx/v2.0/.well-known/openid-configuration",
					ClientID:     "client-id",
					AuthMethod:   "client_secret",
				},
				FirstAdminEmail: "admin@acme.com",
			},
			expectError: true,
		},
		{
			name: "empty domains",
			request: CreateTenantRequest{
				ID:      "acme",
				Name:    "Acme Corporation",
				Domains: []string{},
				OIDCConfig: OIDCConfigRequest{
					DiscoveryURL: "https://login.microsoftonline.com/xxx/v2.0/.well-known/openid-configuration",
					ClientID:     "client-id",
					AuthMethod:   "client_secret",
				},
				FirstAdminEmail: "admin@acme.com",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTenantResponse_NoSecretsInResponse(t *testing.T) {
	response := TenantResponse{
		ID:     "acme",
		Name:   "Acme Corporation",
		Status: "active",
		OIDCConfig: &OIDCConfigResponse{
			DiscoveryURL: "https://example.com/.well-known/openid-configuration",
			ClientID:     "client-id",
			AuthMethod:   "private_key_jwt",
			Scopes:       "openid email profile",
		},
	}

	body, _ := json.Marshal(response)
	var decoded map[string]interface{}
	json.Unmarshal(body, &decoded)

	oidc := decoded["oidcConfig"].(map[string]interface{})
	_, hasSecret := oidc["clientSecret"]
	assert.False(t, hasSecret, "response should not contain clientSecret")
	_, hasPrivateKey := oidc["privateKey"]
	assert.False(t, hasPrivateKey, "response should not contain privateKey")
	assert.Equal(t, "private_key_jwt", oidc["authMethod"])
}

func TestParseCreateTenantRequest(t *testing.T) {
	body := `{
		"id": "acme",
		"name": "Acme Corporation",
		"domains": ["acme.com", "acme.co.uk"],
		"oidcConfig": {
			"discoveryUrl": "https://login.microsoftonline.com/xxx/v2.0/.well-known/openid-configuration",
			"clientId": "client-id",
			"authMethod": "private_key_jwt",
			"scopes": "openid email profile"
		},
		"firstAdminEmail": "john.doe@acme.com"
	}`

	req := httptest.NewRequest("POST", "/api/platform/v1/tenants", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	var parsed CreateTenantRequest
	err := json.NewDecoder(req.Body).Decode(&parsed)

	assert.NoError(t, err)
	assert.Equal(t, "acme", parsed.ID)
	assert.Equal(t, "Acme Corporation", parsed.Name)
	assert.Equal(t, []string{"acme.com", "acme.co.uk"}, parsed.Domains)
	assert.Equal(t, "private_key_jwt", parsed.OIDCConfig.AuthMethod)
	assert.Equal(t, "john.doe@acme.com", parsed.FirstAdminEmail)
}

func TestTenantHandlers_CreateTenant_InvalidJSON(t *testing.T) {
	handlers := &TenantHandlers{}

	req := httptest.NewRequest("POST", "/api/platform/v1/tenants", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handlers.CreateTenant(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestOIDCConfigResponse_IncludesSecretProvisioned(t *testing.T) {
	response := OIDCConfigResponse{
		DiscoveryURL:      "https://example.com/.well-known/openid-configuration",
		ClientID:          "client-id",
		AuthMethod:        "private_key_jwt",
		Scopes:            "openid email profile",
		SecretProvisioned: true,
	}

	body, _ := json.Marshal(response)
	var decoded map[string]interface{}
	json.Unmarshal(body, &decoded)

	assert.Equal(t, true, decoded["secretProvisioned"])
}

func TestTenantResponse_IncludesWarnings(t *testing.T) {
	response := TenantResponse{
		ID:     "acme",
		Name:   "Acme Corporation",
		Status: "active",
		OIDCConfig: &OIDCConfigResponse{
			DiscoveryURL:      "https://example.com/.well-known/openid-configuration",
			ClientID:          "client-id",
			AuthMethod:        "private_key_jwt",
			Scopes:            "openid email profile",
			SecretProvisioned: false,
		},
		Warnings: []string{"OIDC secret not provisioned"},
	}

	body, _ := json.Marshal(response)
	var decoded map[string]interface{}
	json.Unmarshal(body, &decoded)

	warnings := decoded["_warnings"].([]interface{})
	assert.Len(t, warnings, 1)
	assert.Equal(t, "OIDC secret not provisioned", warnings[0])
}
