//go:build integration
// +build integration

package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"easi/backend/internal/platform/application/handlers"
	"easi/backend/internal/platform/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type platformTestContext struct {
	db             *sql.DB
	testID         string
	createdTenants []string
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func setupPlatformTestDB(t *testing.T) (*platformTestContext, func()) {
	dbHost := getEnv("INTEGRATION_TEST_DB_HOST", "localhost")
	dbPort := getEnv("INTEGRATION_TEST_DB_PORT", "5432")
	dbUser := getEnv("INTEGRATION_TEST_DB_USER", "easi_app")
	dbPassword := getEnv("INTEGRATION_TEST_DB_PASSWORD", "localdev")
	dbName := getEnv("INTEGRATION_TEST_DB_NAME", "easi")
	dbSSLMode := getEnv("INTEGRATION_TEST_DB_SSLMODE", "disable")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	testID := fmt.Sprintf("test-%d", time.Now().UnixNano())

	ctx := &platformTestContext{
		db:             db,
		testID:         testID,
		createdTenants: make([]string, 0),
	}

	cleanup := func() {
		for _, id := range ctx.createdTenants {
			db.Exec("DELETE FROM invitations WHERE tenant_id = $1", id)
			db.Exec("DELETE FROM tenant_oidc_configs WHERE tenant_id = $1", id)
			db.Exec("DELETE FROM tenant_domains WHERE tenant_id = $1", id)
			db.Exec("DELETE FROM tenants WHERE id = $1", id)
		}
		db.Close()
	}

	return ctx, cleanup
}

func (ctx *platformTestContext) trackTenant(id string) {
	ctx.createdTenants = append(ctx.createdTenants, id)
}

func setupPlatformHandlers(db *sql.DB) (*TenantHandlers, chi.Router) {
	commandBus := cqrs.NewInMemoryCommandBus()
	tenantRepo := repositories.NewTenantRepository(db)

	createTenantHandler := handlers.NewCreateTenantHandler(tenantRepo)
	commandBus.Register("CreateTenant", createTenantHandler)

	tenantHandlers := NewTenantHandlers(commandBus, tenantRepo)

	r := chi.NewRouter()
	r.Use(PlatformAdminMiddleware("test-api-key"))
	r.Post("/tenants", tenantHandlers.CreateTenant)
	r.Get("/tenants", tenantHandlers.ListTenants)
	r.Get("/tenants/{id}", tenantHandlers.GetTenantByID)

	return tenantHandlers, r
}

func (ctx *platformTestContext) makeRequest(method, url string, body []byte, router chi.Router) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req := httptest.NewRequest(method, url, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Platform-Admin-Key", "test-api-key")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestCreateTenant_Integration(t *testing.T) {
	ctx, cleanup := setupPlatformTestDB(t)
	defer cleanup()

	_, router := setupPlatformHandlers(ctx.db)

	tenantID := fmt.Sprintf("acme-%d", time.Now().UnixNano()%10000)
	ctx.trackTenant(tenantID)

	reqBody := CreateTenantRequest{
		ID:   tenantID,
		Name: "Acme Corporation",
		Domains: []string{tenantID + ".com"},
		OIDCConfig: OIDCConfigRequest{
			DiscoveryURL: "https://login.microsoftonline.com/xxx/v2.0/.well-known/openid-configuration",
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			Scopes:       "openid email profile",
		},
		FirstAdminEmail: "admin@" + tenantID + ".com",
	}
	body, _ := json.Marshal(reqBody)

	w := ctx.makeRequest("POST", "/tenants", body, router)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/api/platform/v1/tenants/"+tenantID)

	var response TenantResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, tenantID, response.ID)
	assert.Equal(t, "Acme Corporation", response.Name)
	assert.Equal(t, "active", response.Status)
	assert.Contains(t, response.Domains, tenantID+".com")
	assert.NotNil(t, response.OIDCConfig)
	assert.Equal(t, "client-id", response.OIDCConfig.ClientID)
	assert.NotNil(t, response.Links)
	assert.Contains(t, response.Links["self"], tenantID)

	var exists bool
	err = ctx.db.QueryRow("SELECT EXISTS(SELECT 1 FROM tenants WHERE id = $1)", tenantID).Scan(&exists)
	require.NoError(t, err)
	assert.True(t, exists)

	var domainCount int
	err = ctx.db.QueryRow("SELECT COUNT(*) FROM tenant_domains WHERE tenant_id = $1", tenantID).Scan(&domainCount)
	require.NoError(t, err)
	assert.Equal(t, 1, domainCount)

	var invitationCount int
	err = ctx.db.QueryRow("SELECT COUNT(*) FROM invitations WHERE tenant_id = $1 AND email = $2 AND role = 'admin'",
		tenantID, "admin@"+tenantID+".com").Scan(&invitationCount)
	require.NoError(t, err)
	assert.Equal(t, 1, invitationCount)
}

func TestCreateTenant_DuplicateID_Integration(t *testing.T) {
	ctx, cleanup := setupPlatformTestDB(t)
	defer cleanup()

	_, router := setupPlatformHandlers(ctx.db)

	tenantID := fmt.Sprintf("dup-%d", time.Now().UnixNano()%10000)
	ctx.trackTenant(tenantID)

	reqBody := CreateTenantRequest{
		ID:   tenantID,
		Name: "First Tenant",
		Domains: []string{tenantID + ".com"},
		OIDCConfig: OIDCConfigRequest{
			DiscoveryURL: "https://login.microsoftonline.com/xxx/v2.0/.well-known/openid-configuration",
			ClientID:     "client-id",
			ClientSecret: "client-secret",
		},
		FirstAdminEmail: "admin@" + tenantID + ".com",
	}
	body, _ := json.Marshal(reqBody)

	w1 := ctx.makeRequest("POST", "/tenants", body, router)
	assert.Equal(t, http.StatusCreated, w1.Code)

	w2 := ctx.makeRequest("POST", "/tenants", body, router)
	assert.Equal(t, http.StatusConflict, w2.Code)
}

func TestGetTenantByID_Integration(t *testing.T) {
	ctx, cleanup := setupPlatformTestDB(t)
	defer cleanup()

	_, router := setupPlatformHandlers(ctx.db)

	tenantID := fmt.Sprintf("get-%d", time.Now().UnixNano()%10000)
	ctx.trackTenant(tenantID)

	reqBody := CreateTenantRequest{
		ID:   tenantID,
		Name: "Get Test Corp",
		Domains: []string{tenantID + ".com"},
		OIDCConfig: OIDCConfigRequest{
			DiscoveryURL: "https://login.microsoftonline.com/xxx/v2.0/.well-known/openid-configuration",
			ClientID:     "client-id",
			ClientSecret: "client-secret",
		},
		FirstAdminEmail: "admin@" + tenantID + ".com",
	}
	body, _ := json.Marshal(reqBody)
	ctx.makeRequest("POST", "/tenants", body, router)

	w := ctx.makeRequest("GET", "/tenants/"+tenantID, nil, router)

	assert.Equal(t, http.StatusOK, w.Code)

	var response TenantResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, tenantID, response.ID)
	assert.Equal(t, "Get Test Corp", response.Name)
}

func TestGetTenantByID_NotFound_Integration(t *testing.T) {
	ctx, cleanup := setupPlatformTestDB(t)
	defer cleanup()

	_, router := setupPlatformHandlers(ctx.db)

	w := ctx.makeRequest("GET", "/tenants/nonexistent-tenant", nil, router)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestListTenants_Integration(t *testing.T) {
	ctx, cleanup := setupPlatformTestDB(t)
	defer cleanup()

	_, router := setupPlatformHandlers(ctx.db)

	tenantID := fmt.Sprintf("list-%d", time.Now().UnixNano()%10000)
	ctx.trackTenant(tenantID)

	reqBody := CreateTenantRequest{
		ID:   tenantID,
		Name: "List Test Corp",
		Domains: []string{tenantID + ".com"},
		OIDCConfig: OIDCConfigRequest{
			DiscoveryURL: "https://login.microsoftonline.com/xxx/v2.0/.well-known/openid-configuration",
			ClientID:     "client-id",
			ClientSecret: "client-secret",
		},
		FirstAdminEmail: "admin@" + tenantID + ".com",
	}
	body, _ := json.Marshal(reqBody)
	ctx.makeRequest("POST", "/tenants", body, router)

	w := ctx.makeRequest("GET", "/tenants", nil, router)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response["data"])
	assert.NotNil(t, response["pagination"])
}

func TestPlatformAdminMiddleware_MissingKey_Integration(t *testing.T) {
	ctx, cleanup := setupPlatformTestDB(t)
	defer cleanup()

	_, router := setupPlatformHandlers(ctx.db)

	req := httptest.NewRequest("GET", "/tenants", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestPlatformAdminMiddleware_InvalidKey_Integration(t *testing.T) {
	ctx, cleanup := setupPlatformTestDB(t)
	defer cleanup()

	_, router := setupPlatformHandlers(ctx.db)

	req := httptest.NewRequest("GET", "/tenants", nil)
	req.Header.Set("X-Platform-Admin-Key", "wrong-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
