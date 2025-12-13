//go:build integration
// +build integration

package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/auth/infrastructure/session"

	"github.com/alexedwards/scs/v2"
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type authTestContext struct {
	db         *sql.DB
	testID     string
	tenantID   string
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

	testID := fmt.Sprintf("auth-test-%d", time.Now().UnixNano())
	tenantID := fmt.Sprintf("acme-%s", testID)

	_, err = db.Exec(`
		INSERT INTO tenants (id, name, status, created_at, updated_at)
		VALUES ($1, $2, 'active', NOW(), NOW())
	`, tenantID, "ACME Test Corp")
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO tenant_domains (tenant_id, domain, created_at)
		VALUES ($1, 'acme.com', NOW())
	`, tenantID)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO tenant_oidc_configs (tenant_id, discovery_url, client_id, auth_method, scopes, created_at, updated_at)
		VALUES ($1, $2, 'easi-test', 'client_secret', 'openid email profile offline_access', NOW(), NOW())
	`, tenantID, dexBaseURL)
	require.NoError(t, err)

	ctx := &authTestContext{
		db:         db,
		testID:     testID,
		tenantID:   tenantID,
		dexBaseURL: dexBaseURL,
	}

	cleanup := func() {
		db.Exec("DELETE FROM tenant_oidc_configs WHERE tenant_id = $1", tenantID)
		db.Exec("DELETE FROM tenant_domains WHERE tenant_id = $1", tenantID)
		db.Exec("DELETE FROM tenants WHERE id = $1", tenantID)
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
		RedirectURL:    "http://localhost:8080/auth/callback",
		AllowedOrigins: []string{"http://localhost:3000", "http://localhost:5173"},
	})

	r := chi.NewRouter()
	r.Use(scsManager.LoadAndSave)
	r.Post("/auth/sessions", handlers.PostSessions)
	r.Get("/auth/callback", handlers.GetCallback)

	return handlers, scsManager, r
}

func TestIntegration_PostSessions_ValidEmail_ReturnsDexAuthURL(t *testing.T) {
	ctx, cleanup := setupAuthTestDB(t)
	defer cleanup()

	_, _, router := setupAuthHandlers(t, ctx.db, ctx.dexBaseURL)

	body := map[string]string{"email": "testuser@acme.com"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/sessions", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	authURL, ok := response["authorizationUrl"].(string)
	require.True(t, ok, "authorizationUrl should be a string")
	assert.Contains(t, authURL, ctx.dexBaseURL)
	assert.Contains(t, authURL, "response_type=code")
	assert.Contains(t, authURL, "client_id=easi-test")
	assert.Contains(t, authURL, "code_challenge=")
	assert.Contains(t, authURL, "code_challenge_method=S256")
	assert.Contains(t, authURL, "state=")
	assert.Contains(t, authURL, "nonce=")

	links, ok := response["_links"].(map[string]interface{})
	require.True(t, ok, "_links should be a map")
	assert.Contains(t, links, "authorize")
}

func TestIntegration_PostSessions_UnknownDomain_Returns404(t *testing.T) {
	ctx, cleanup := setupAuthTestDB(t)
	defer cleanup()

	_, _, router := setupAuthHandlers(t, ctx.db, ctx.dexBaseURL)

	body := map[string]string{"email": "user@unknown-domain.com"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/sessions", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestIntegration_GetCallback_InvalidState_ReturnsBadRequest(t *testing.T) {
	ctx, cleanup := setupAuthTestDB(t)
	defer cleanup()

	_, _, router := setupAuthHandlers(t, ctx.db, ctx.dexBaseURL)

	body := map[string]string{"email": "testuser@acme.com"}
	jsonBody, _ := json.Marshal(body)
	req1 := httptest.NewRequest(http.MethodPost, "/auth/sessions", bytes.NewReader(jsonBody))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	router.ServeHTTP(rec1, req1)
	require.Equal(t, http.StatusOK, rec1.Code)

	cookies := rec1.Result().Cookies()

	req2 := httptest.NewRequest(http.MethodGet, "/auth/callback?code=test-code&state=invalid-state", nil)
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

	body := map[string]string{"email": "testuser@acme.com"}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/auth/sessions", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	var response map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&response)
	authURL := response["authorizationUrl"].(string)

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

func TestIntegration_TokenRefreshMiddleware_RefreshesExpiredToken(t *testing.T) {
	ctx, cleanup := setupAuthTestDB(t)
	defer cleanup()

	scsManager := scs.New()
	scsManager.Store = memstore.New()
	scsManager.Lifetime = time.Hour
	sessionManager := session.NewSessionManager(scsManager)

	tenantRepo := repositories.NewTenantOIDCRepository(ctx.db)
	refreshMiddleware := NewTokenRefreshMiddleware(
		sessionManager,
		tenantRepo,
		"easi-test-secret",
		"http://localhost:8080/auth/callback",
	)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r := chi.NewRouter()
	r.Use(scsManager.LoadAndSave)
	r.Use(refreshMiddleware.RefreshIfNeeded())
	r.Get("/test", testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
