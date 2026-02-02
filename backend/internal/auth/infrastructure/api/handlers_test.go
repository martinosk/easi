package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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
	t.Helper()
	idp := createMockIdPWithTokenEndpoint(t)
	return idp.server, idp.privateKey
}

func postSessionsWithEmail(t *testing.T, handlers *AuthHandlers, scsManager *scs.SessionManager, email string) (*httptest.ResponseRecorder, map[string]interface{}) {
	t.Helper()

	body := map[string]string{"email": email}
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
	return rec, response
}

func TestPostSessions_ValidEmail(t *testing.T) {
	idpServer, _ := createMockIdP(t)
	defer idpServer.Close()

	handlers, scsManager := setupTestHandler(t, idpServer)
	rec, response := postSessionsWithEmail(t, handlers, scsManager, "user@acme.com")

	assert.Equal(t, http.StatusOK, rec.Code)

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

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestPostSessions_AuthURLContainsPKCE(t *testing.T) {
	idpServer, _ := createMockIdP(t)
	defer idpServer.Close()

	handlers, scsManager := setupTestHandler(t, idpServer)
	_, response := postSessionsWithEmail(t, handlers, scsManager, "user@acme.com")

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

type idpWithTokenEndpoint struct {
	server     *httptest.Server
	privateKey *rsa.PrivateKey
	nonce      *string
}

func createMockIdPWithTokenEndpoint(t *testing.T) *idpWithTokenEndpoint {
	t.Helper()

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

	var storedNonce string
	idp := &idpWithTokenEndpoint{privateKey: privateKey, nonce: &storedNonce}

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	idp.server = server

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
		idp.handleTokenRequest(w)
	})

	return idp
}

func (idp *idpWithTokenEndpoint) handleTokenRequest(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")

	signer, _ := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: idp.privateKey}, &jose.SignerOptions{
		ExtraHeaders: map[jose.HeaderKey]interface{}{
			"kid": "test-key-1",
		},
	})

	claims := jwt.Claims{
		Issuer:   idp.server.URL,
		Subject:  "user-123",
		Audience: jwt.Audience{"test-client-id"},
		Expiry:   jwt.NewNumericDate(time.Now().Add(time.Hour)),
		IssuedAt: jwt.NewNumericDate(time.Now()),
	}
	customClaims := map[string]interface{}{
		"nonce": *idp.nonce,
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
}

func TestGetCallback_SuccessfulExchange(t *testing.T) {
	idp := createMockIdPWithTokenEndpoint(t)
	defer idp.server.Close()

	handlers, scsManager := setupTestHandler(t, idp.server)

	router := chi.NewRouter()
	router.Use(scsManager.LoadAndSave)
	router.Post("/auth/sessions", handlers.PostSessions)
	router.Get("/auth/callback", handlers.GetCallback)

	rec1, initResponse := postSessionsWithEmail(t, handlers, scsManager, "user@acme.com")
	require.Equal(t, http.StatusOK, rec1.Code)

	links := initResponse["_links"].(map[string]interface{})
	authURL := links["authorize"].(string)
	*idp.nonce = extractQueryParam(authURL, "nonce")
	storedState := extractQueryParam(authURL, "state")

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

func extractQueryParam(rawURL, param string) string {
	queryStart := strings.IndexByte(rawURL, '?')
	if queryStart < 0 {
		return ""
	}

	query := rawURL[queryStart+1:]
	for _, pair := range strings.Split(query, "&") {
		key, value, ok := strings.Cut(pair, "=")
		if ok && key == param {
			return value
		}
	}
	return ""
}
