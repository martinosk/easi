//go:build integration
// +build integration

package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/auth/infrastructure/session"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type authTestContext struct {
	db         *sql.DB
	testID     string
	tenantID   string
	testDomain string
	dexBaseURL string
}

func getTestEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func setupAuthTestDB(t *testing.T) (*authTestContext, func()) {
	dbHost := getTestEnv("INTEGRATION_TEST_DB_HOST", "localhost")
	dbPort := getTestEnv("INTEGRATION_TEST_DB_PORT", "5432")
	dbUser := getTestEnv("INTEGRATION_TEST_DB_USER", "easi_app")
	dbPassword := getTestEnv("INTEGRATION_TEST_DB_PASSWORD", "localdev")
	dbName := getTestEnv("INTEGRATION_TEST_DB_NAME", "easi")
	dbSSLMode := getTestEnv("INTEGRATION_TEST_DB_SSLMODE", "disable")
	dexBaseURL := getTestEnv("DEX_BASE_URL", "http://localhost:5556/dex")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	testID := fmt.Sprintf("%d", time.Now().UnixNano())
	tenantID := fmt.Sprintf("test-%s", testID)

	_, err = db.Exec(`
		INSERT INTO platform.tenants (id, name, status, created_at, updated_at)
		VALUES ($1, $2, 'active', NOW(), NOW())
	`, tenantID, "ACME Test Corp")
	require.NoError(t, err)

	testDomain := fmt.Sprintf("test%s.example.com", testID)
	_, err = db.Exec(`
		INSERT INTO platform.tenant_domains (tenant_id, domain, created_at)
		VALUES ($1, $2, NOW())
	`, tenantID, testDomain)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO platform.tenant_oidc_configs (tenant_id, discovery_url, client_id, auth_method, scopes, created_at, updated_at)
		VALUES ($1, $2, 'easi-test', 'client_secret', 'openid email profile offline_access', NOW(), NOW())
	`, tenantID, dexBaseURL)
	require.NoError(t, err)

	ctx := &authTestContext{
		db:         db,
		testID:     testID,
		tenantID:   tenantID,
		testDomain: testDomain,
		dexBaseURL: dexBaseURL,
	}

	cleanup := func() {
		db.Exec("DELETE FROM platform.tenant_oidc_configs WHERE tenant_id = $1", tenantID)
		db.Exec("DELETE FROM platform.tenant_domains WHERE tenant_id = $1", tenantID)
		db.Exec("DELETE FROM platform.tenants WHERE id = $1", tenantID)
		db.Close()
	}

	return ctx, cleanup
}

func setupAuthHandlers(t *testing.T, db *sql.DB, dexBaseURL string) (*AuthHandlers, *scs.SessionManager, chi.Router) {
	scsManager := scs.New()
	scsManager.Store = memstore.New()
	scsManager.Lifetime = time.Hour
	sessionManager := session.NewSessionManager(scsManager)

	tenantRepo := repositories.NewTenantOIDCRepository(db)

	handlers := NewAuthHandlers(sessionManager, tenantRepo, AuthHandlersConfig{
		ClientSecret:   "easi-test-secret",
		RedirectURL:    "http://localhost:8080/api/v1/auth/callback",
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:5173"},
	})

	r := chi.NewRouter()
	r.Use(scsManager.LoadAndSave)
	r.Post("/api/v1/auth/sessions", handlers.PostSessions)
	r.Get("/api/v1/auth/callback", handlers.GetCallback)

	return handlers, scsManager, r
}

func TestIntegration_PostSessions_ValidEmail_ReturnsDexAuthURL(t *testing.T) {
	ctx, cleanup := setupAuthTestDB(t)
	defer cleanup()

	_, _, router := setupAuthHandlers(t, ctx.db, ctx.dexBaseURL)

	body := map[string]string{"email": fmt.Sprintf("testuser@%s", ctx.testDomain)}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sessions", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	links, ok := response["_links"].(map[string]interface{})
	require.True(t, ok, "_links should be a map")
	assert.Contains(t, links, "authorize")
	assert.Contains(t, links, "self")

	authURL, ok := links["authorize"].(string)
	require.True(t, ok, "authorize link should be a string")
	assert.Contains(t, authURL, ctx.dexBaseURL)
	assert.Contains(t, authURL, "response_type=code")
	assert.Contains(t, authURL, "client_id=easi-test")
	assert.Contains(t, authURL, "code_challenge=")
	assert.Contains(t, authURL, "code_challenge_method=S256")
	assert.Contains(t, authURL, "state=")
	assert.Contains(t, authURL, "nonce=")
}

func TestIntegration_PostSessions_UnknownDomain_Returns404(t *testing.T) {
	ctx, cleanup := setupAuthTestDB(t)
	defer cleanup()

	_, _, router := setupAuthHandlers(t, ctx.db, ctx.dexBaseURL)

	body := map[string]string{"email": "user@unknown-domain.com"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sessions", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestIntegration_GetCallback_InvalidState_ReturnsBadRequest(t *testing.T) {
	ctx, cleanup := setupAuthTestDB(t)
	defer cleanup()

	_, _, router := setupAuthHandlers(t, ctx.db, ctx.dexBaseURL)

	body := map[string]string{"email": fmt.Sprintf("testuser@%s", ctx.testDomain)}
	jsonBody, _ := json.Marshal(body)
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sessions", bytes.NewReader(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)
	require.Equal(t, http.StatusOK, rec1.Code)

	cookies := rec1.Result().Cookies()

	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/auth/callback?code=test-code&state=invalid-state", nil)
	for _, c := range cookies {
		req2.AddCookie(c)
	}
	rec2 := httptest.NewRecorder()
	router.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusBadRequest, rec2.Code)
}

func TestIntegration_PostSessions_AuthURLContainsPKCE(t *testing.T) {
	ctx, cleanup := setupAuthTestDB(t)
	defer cleanup()

	_, _, router := setupAuthHandlers(t, ctx.db, ctx.dexBaseURL)

	body := map[string]string{"email": fmt.Sprintf("testuser@%s", ctx.testDomain)}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/sessions", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	links, ok := response["_links"].(map[string]interface{})
	require.True(t, ok, "_links should be a map")
	authURL, ok := links["authorize"].(string)
	require.True(t, ok, "authorize link should be a string")

	assert.True(t, strings.Contains(authURL, "code_challenge="), "URL should contain code_challenge")
	assert.True(t, strings.Contains(authURL, "code_challenge_method=S256"), "URL should use S256 challenge method")

	challengeStart := strings.Index(authURL, "code_challenge=") + len("code_challenge=")
	challengeEnd := strings.Index(authURL[challengeStart:], "&")
	if challengeEnd == -1 {
		challengeEnd = len(authURL) - challengeStart
	}
	challenge := authURL[challengeStart : challengeStart+challengeEnd]
	assert.True(t, len(challenge) >= 43, "code_challenge should be at least 43 characters (base64url encoded SHA256)")
}

type mockTenantRepoForRefresh struct {
	config *repositories.TenantOIDCConfig
}

func (m *mockTenantRepoForRefresh) GetByEmailDomain(ctx context.Context, domain string) (*repositories.TenantOIDCConfig, error) {
	return m.config, nil
}

func (m *mockTenantRepoForRefresh) GetByTenantID(ctx context.Context, tenantID string) (*repositories.TenantOIDCConfig, error) {
	return m.config, nil
}

func createMockIdPWithTokenRefresh(t *testing.T, refreshShouldSucceed bool) *httptest.Server {
	var refreshCallCount atomic.Int32

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		config := map[string]interface{}{
			"issuer":                 server.URL,
			"authorization_endpoint": server.URL + "/authorize",
			"token_endpoint":         server.URL + "/token",
			"jwks_uri":               server.URL + "/jwks",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(config)
	})

	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		jwks := map[string]interface{}{"keys": []interface{}{}}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwks)
	})

	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		refreshCallCount.Add(1)
		w.Header().Set("Content-Type", "application/json")

		if !refreshShouldSucceed {
			w.WriteHeader(http.StatusBadRequest)
			errorResp := map[string]string{
				"error":             "invalid_grant",
				"error_description": "refresh token is expired or revoked",
			}
			json.NewEncoder(w).Encode(errorResp)
			return
		}

		response := map[string]interface{}{
			"access_token":  "new-access-token-" + fmt.Sprintf("%d", refreshCallCount.Load()),
			"token_type":    "Bearer",
			"refresh_token": "new-refresh-token",
			"expires_in":    3600,
		}
		json.NewEncoder(w).Encode(response)
	})

	return server
}

func createTestAuthSession(tenantID string, tokenExpired bool) session.AuthSession {
	var tokenExpiry time.Time
	var accessToken string
	if tokenExpired {
		tokenExpiry = time.Now().Add(-10 * time.Minute)
		accessToken = "expired-access-token"
	} else {
		tokenExpiry = time.Now().Add(1 * time.Hour)
		accessToken = "valid-access-token"
	}

	sessionData := fmt.Sprintf(`{
		"tenantId": "%s",
		"state": "",
		"nonce": "",
		"codeVerifier": "",
		"expectedEmailDomain": "",
		"returnUrl": "",
		"userId": "%s",
		"userEmail": "testuser@acme.com",
		"accessToken": "%s",
		"refreshToken": "valid-refresh-token",
		"tokenExpiry": "%s",
		"authenticated": true
	}`, tenantID, uuid.New().String(), accessToken, tokenExpiry.Format(time.RFC3339))

	authSession, _ := session.UnmarshalAuthSession([]byte(sessionData))
	return authSession
}

type tokenRefreshTestFixture struct {
	idpServer         *httptest.Server
	mockRepo          *mockTenantRepoForRefresh
	sessionManager    *session.SessionManager
	refreshMiddleware *TokenRefreshMiddleware
	scsManager        *scs.SessionManager
	tenantID          string
}

func setupTokenRefreshTest(t *testing.T, refreshShouldSucceed bool) *tokenRefreshTestFixture {
	idpServer := createMockIdPWithTokenRefresh(t, refreshShouldSucceed)
	tenantID := "test-tenant"

	mockRepo := &mockTenantRepoForRefresh{
		config: &repositories.TenantOIDCConfig{
			TenantID:     tenantID,
			DiscoveryURL: idpServer.URL,
			ClientID:     "test-client-id",
			AuthMethod:   "client_secret",
			Scopes:       "openid email profile offline_access",
		},
	}

	scsManager := scs.New()
	scsManager.Store = memstore.New()
	scsManager.Lifetime = time.Hour
	sessionManager := session.NewSessionManager(scsManager)

	refreshMiddleware := NewTokenRefreshMiddleware(
		sessionManager,
		mockRepo,
		"test-client-secret",
		idpServer.URL+"/callback",
	)

	return &tokenRefreshTestFixture{
		idpServer:         idpServer,
		mockRepo:          mockRepo,
		sessionManager:    sessionManager,
		refreshMiddleware: refreshMiddleware,
		scsManager:        scsManager,
		tenantID:          tenantID,
	}
}

func (f *tokenRefreshTestFixture) createRouterWithSession(authSession session.AuthSession, testHandler http.HandlerFunc) chi.Router {
	setupHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.sessionManager.StoreAuthenticatedSession(r.Context(), authSession)
		w.WriteHeader(http.StatusOK)
	})

	r := chi.NewRouter()
	r.Use(f.scsManager.LoadAndSave)
	r.Post("/setup", setupHandler)
	r.Group(func(r chi.Router) {
		r.Use(f.refreshMiddleware.RefreshIfNeeded())
		r.Get("/test", testHandler)
	})
	return r
}

func (f *tokenRefreshTestFixture) setupSessionAndGetCookies(t *testing.T, router chi.Router) []*http.Cookie {
	req := httptest.NewRequest(http.MethodPost, "/setup", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	cookies := rec.Result().Cookies()
	require.NotEmpty(t, cookies, "Should have session cookie")
	return cookies
}

func (f *tokenRefreshTestFixture) executeTestRequest(router chi.Router, cookies []*http.Cookie) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	router.ServeHTTP(rec, req)
	return rec
}

func TestIntegration_TokenRefreshMiddleware_TransparentRefresh_WhenAccessTokenExpired(t *testing.T) {
	fixture := setupTokenRefreshTest(t, true)
	defer fixture.idpServer.Close()

	var handlerCalled bool
	var sessionAfterRefresh session.AuthSession
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		sessionAfterRefresh, _ = fixture.sessionManager.LoadAuthenticatedSession(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	expiredSession := createTestAuthSession(fixture.tenantID, true)
	router := fixture.createRouterWithSession(expiredSession, testHandler)
	cookies := fixture.setupSessionAndGetCookies(t, router)
	rec := fixture.executeTestRequest(router, cookies)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, handlerCalled, "Handler should have been called after transparent refresh")
	assert.True(t, sessionAfterRefresh.IsAuthenticated(), "Session should be authenticated")
	assert.Contains(t, sessionAfterRefresh.AccessToken(), "new-access-token",
		"Session should have new access token after refresh")
}

func TestIntegration_TokenRefreshMiddleware_Returns401_WhenRefreshTokenExpired(t *testing.T) {
	fixture := setupTokenRefreshTest(t, false)
	defer fixture.idpServer.Close()

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	expiredSession := createTestAuthSession(fixture.tenantID, true)
	router := fixture.createRouterWithSession(expiredSession, testHandler)
	cookies := fixture.setupSessionAndGetCookies(t, router)
	rec := fixture.executeTestRequest(router, cookies)

	assert.Equal(t, http.StatusUnauthorized, rec.Code, "Should return 401 when refresh token is expired")
	assert.False(t, handlerCalled, "Handler should not be called when session expires")
}

func TestIntegration_TokenRefreshMiddleware_NoRefresh_WhenTokenValid(t *testing.T) {
	fixture := setupTokenRefreshTest(t, true)
	defer fixture.idpServer.Close()

	handlerCalled := false
	var accessTokenInHandler string
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		authSession, err := fixture.sessionManager.LoadAuthenticatedSession(r.Context())
		if err == nil {
			accessTokenInHandler = authSession.AccessToken()
		}
		w.WriteHeader(http.StatusOK)
	})

	validSession := createTestAuthSession(fixture.tenantID, false)
	router := fixture.createRouterWithSession(validSession, testHandler)
	cookies := fixture.setupSessionAndGetCookies(t, router)
	rec := fixture.executeTestRequest(router, cookies)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, handlerCalled, "Handler should be called")
	assert.Equal(t, "valid-access-token", accessTokenInHandler,
		"Access token should remain unchanged when not expired")
}

func TestIntegration_TokenRefreshMiddleware_NoSession_PassesThrough(t *testing.T) {
	fixture := setupTokenRefreshTest(t, true)
	defer fixture.idpServer.Close()

	handlerCalled := false
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	r := chi.NewRouter()
	r.Use(fixture.scsManager.LoadAndSave)
	r.Use(fixture.refreshMiddleware.RefreshIfNeeded())
	r.Get("/test", testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, handlerCalled, "Handler should be called even without session")
}
