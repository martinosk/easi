package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/go-chi/chi/v5"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/auth/infrastructure/session"
)

type mockTenantOIDCRepository struct {
	config *repositories.TenantOIDCConfig
	err    error
}

func (m *mockTenantOIDCRepository) GetByEmailDomain(ctx context.Context, domain string) (*repositories.TenantOIDCConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.config, nil
}

func (m *mockTenantOIDCRepository) GetByTenantID(ctx context.Context, tenantID string) (*repositories.TenantOIDCConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.config, nil
}

func setupTestHandler(t *testing.T, idpServer *httptest.Server) (*AuthHandlers, *scs.SessionManager) {
	scsManager := scs.New()
	scsManager.Store = memstore.New()
	scsManager.Lifetime = time.Hour
	sessionManager := session.NewSessionManager(scsManager)

	mockRepo := &mockTenantOIDCRepository{
		config: &repositories.TenantOIDCConfig{
			TenantID:     "acme",
			DiscoveryURL: idpServer.URL,
			ClientID:     "test-client-id",
			AuthMethod:   "client_secret",
			Scopes:       "openid email profile offline_access",
		},
	}

	handlers := NewAuthHandlers(sessionManager, mockRepo, AuthHandlersConfig{
		ClientSecret:   "test-secret",
		RedirectURL:    idpServer.URL + "/callback",
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:5173"},
	})

	return handlers, scsManager
}

func createMockIdP(t *testing.T) (*httptest.Server, *rsa.PrivateKey) {
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

	return server, privateKey
}

func TestPostSessions_ValidEmail(t *testing.T) {
	idpServer, _ := createMockIdP(t)
	defer idpServer.Close()

	handlers, scsManager := setupTestHandler(t, idpServer)

	body := map[string]string{"email": "user@acme.com"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/sessions", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Post("/auth/sessions", handlers.PostSessions)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&response)
	assert.Contains(t, response, "_links")
	links := response["_links"].(map[string]interface{})
	assert.Contains(t, links, "self")
	assert.Contains(t, links, "authorize")
	assert.Contains(t, links["authorize"].(string), idpServer.URL+"/authorize")
}

func TestPostSessions_InvalidEmail(t *testing.T) {
	idpServer, _ := createMockIdP(t)
	defer idpServer.Close()

	handlers, scsManager := setupTestHandler(t, idpServer)

	body := map[string]string{"email": "invalid-email"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/sessions", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Post("/auth/sessions", handlers.PostSessions)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPostSessions_DomainNotFound(t *testing.T) {
	idpServer, _ := createMockIdP(t)
	defer idpServer.Close()

	scsManager := scs.New()
	scsManager.Store = memstore.New()
	scsManager.Lifetime = time.Hour
	sessionManager := session.NewSessionManager(scsManager)

	mockRepo := &mockTenantOIDCRepository{
		err: repositories.ErrDomainNotFound,
	}

	handlers := NewAuthHandlers(sessionManager, mockRepo, AuthHandlersConfig{
		ClientSecret:   "test-secret",
		RedirectURL:    idpServer.URL + "/callback",
		AllowedOrigins: []string{"http://localhost:3000"},
	})

	body := map[string]string{"email": "user@unknown.com"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/sessions", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Post("/auth/sessions", handlers.PostSessions)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestPostSessions_AuthURLContainsPKCE(t *testing.T) {
	idpServer, _ := createMockIdP(t)
	defer idpServer.Close()

	handlers, scsManager := setupTestHandler(t, idpServer)

	body := map[string]string{"email": "user@acme.com"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/sessions", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Post("/auth/sessions", handlers.PostSessions)
	router.ServeHTTP(rec, req)

	var response map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&response)
	links := response["_links"].(map[string]interface{})
	authURL := links["authorize"].(string)

	assert.Contains(t, authURL, "code_challenge=")
	assert.Contains(t, authURL, "code_challenge_method=S256")
	assert.Contains(t, authURL, "state=")
	assert.Contains(t, authURL, "nonce=")
}

func TestGetCallback_InvalidState(t *testing.T) {
	idpServer, _ := createMockIdP(t)
	defer idpServer.Close()

	handlers, scsManager := setupTestHandler(t, idpServer)

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Post("/auth/sessions", handlers.PostSessions)
	router.Get("/auth/callback", handlers.GetCallback)

	body := map[string]string{"email": "user@acme.com"}
	jsonBody, _ := json.Marshal(body)
	req1 := httptest.NewRequest(http.MethodPost, "/auth/sessions", bytes.NewReader(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)
	require.Equal(t, http.StatusOK, rec1.Code)

	cookies := rec1.Result().Cookies()

	req2 := httptest.NewRequest(http.MethodGet, "/auth/callback?code=test-code&state=wrong-state", nil)
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusBadRequest, rec2.Code)
}

func TestGetCallback_MissingCode(t *testing.T) {
	idpServer, _ := createMockIdP(t)
	defer idpServer.Close()

	handlers, scsManager := setupTestHandler(t, idpServer)

	req := httptest.NewRequest(http.MethodGet, "/auth/callback?state=some-state", nil)
	rec := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Get("/auth/callback", handlers.GetCallback)
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestGetCallback_SuccessfulExchange(t *testing.T) {
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

	var storedState, storedNonce string

	mux := http.NewServeMux()
	idpServer := httptest.NewServer(mux)
	defer idpServer.Close()

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		config := map[string]interface{}{
			"issuer":                 idpServer.URL,
			"authorization_endpoint": idpServer.URL + "/authorize",
			"token_endpoint":         idpServer.URL + "/token",
			"jwks_uri":               idpServer.URL + "/jwks",
		}
		json.NewEncoder(w).Encode(config)
	})

	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(jwks)
	})

	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		signer, _ := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: privateKey}, &jose.SignerOptions{
			ExtraHeaders: map[jose.HeaderKey]interface{}{
				"kid": "test-key-1",
			},
		})

		claims := jwt.Claims{
			Issuer:   idpServer.URL,
			Subject:  "user-123",
			Audience: jwt.Audience{"test-client-id"},
			Expiry:   jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		}
		customClaims := map[string]interface{}{
			"nonce": storedNonce,
			"email": "user@acme.com",
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

	scsManager := scs.New()
	scsManager.Store = memstore.New()
	scsManager.Lifetime = time.Hour
	sessionManager := session.NewSessionManager(scsManager)

	mockRepo := &mockTenantOIDCRepository{
		config: &repositories.TenantOIDCConfig{
			TenantID:     "acme",
			DiscoveryURL: idpServer.URL,
			ClientID:     "test-client-id",
			AuthMethod:   "client_secret",
			Scopes:       "openid email profile offline_access",
		},
	}

	handlers := NewAuthHandlers(sessionManager, mockRepo, AuthHandlersConfig{
		ClientSecret:   "test-secret",
		RedirectURL:    idpServer.URL + "/callback",
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:5173"},
	})

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Post("/auth/sessions", handlers.PostSessions)
	router.Get("/auth/callback", handlers.GetCallback)

	body := map[string]string{"email": "user@acme.com"}
	jsonBody, _ := json.Marshal(body)
	req1 := httptest.NewRequest(http.MethodPost, "/auth/sessions", bytes.NewReader(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)
	require.Equal(t, http.StatusOK, rec1.Code)

	var initResponse map[string]interface{}
	json.NewDecoder(rec1.Body).Decode(&initResponse)
	links := initResponse["_links"].(map[string]interface{})
	authURL := links["authorize"].(string)

	storedState = extractQueryParam(authURL, "state")
	storedNonce = extractQueryParam(authURL, "nonce")

	cookies := rec1.Result().Cookies()

	req2 := httptest.NewRequest(http.MethodGet, "/auth/callback?code=test-code&state="+storedState, nil)
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusFound, rec2.Code)
	assert.Equal(t, "/easi/", rec2.Header().Get("Location"))
}

func extractQueryParam(url, param string) string {
	parts := make(map[string]string)
	if idx := len(url) - 1; idx > 0 {
		queryStart := 0
		for i, c := range url {
			if c == '?' {
				queryStart = i + 1
				break
			}
		}
		if queryStart > 0 {
			query := url[queryStart:]
			for _, pair := range splitString(query, '&') {
				kv := splitString(pair, '=')
				if len(kv) == 2 {
					parts[kv[0]] = kv[1]
				}
			}
		}
	}
	return parts[param]
}

func splitString(s string, sep rune) []string {
	var result []string
	current := ""
	for _, c := range s {
		if c == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
