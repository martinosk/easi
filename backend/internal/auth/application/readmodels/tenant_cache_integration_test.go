//go:build integration

package readmodels

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

const (
	tenantCacheTestTenant = "auth-cache-test"
	backfillMigration     = "140_backfill_auth_tenant_caches.sql"
)

func openTenantCacheTestDB(t *testing.T) (*sql.DB, context.Context) {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	db, err := sql.Open("postgres", "host=localhost port=5432 user=easi_app password=localdev dbname=easi sslmode=disable")
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM auth.tenant_domain_cache WHERE tenant_id = $1", tenantCacheTestTenant)
		_, _ = db.Exec("DELETE FROM auth.tenant_oidc_cache WHERE tenant_id = $1", tenantCacheTestTenant)
		_, _ = db.Exec("DELETE FROM auth.tenant_cache WHERE tenant_id = $1", tenantCacheTestTenant)
		_ = db.Close()
	})
	return db, sharedctx.WithTenant(context.Background(), sharedvo.MustNewTenantID(tenantCacheTestTenant))
}

func cachedTestTenant() TenantCacheEntry {
	return TenantCacheEntry{
		TenantID:     tenantCacheTestTenant,
		Name:         "Auth Cache Test",
		Status:       "active",
		Domains:      []string{"auth-cache-test.com"},
		DiscoveryURL: "https://login.example.com/v2.0/.well-known/openid-configuration",
		IssuerURL:    "https://login.example.com/v2.0",
		ClientID:     "client-id",
		AuthMethod:   "client_secret",
		Scopes:       "openid email profile",
	}
}

func TestTenantCacheReadModel_SaveIsIdempotentAndTenantBecomesKnown(t *testing.T) {
	db, ctx := openTenantCacheTestDB(t)
	cache := NewTenantCacheReadModel(db)

	require.NoError(t, cache.Save(ctx, cachedTestTenant()))
	require.NoError(t, cache.Save(ctx, cachedTestTenant()))

	known, err := cache.Exists(ctx, tenantCacheTestTenant)
	require.NoError(t, err)
	assert.True(t, known)

	unknown, err := cache.Exists(ctx, "no-such-tenant")
	require.NoError(t, err)
	assert.False(t, unknown)

	var domains int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM auth.tenant_domain_cache WHERE tenant_id = $1", tenantCacheTestTenant).Scan(&domains))
	assert.Equal(t, 1, domains)
}

func TestTenantDomainChecker_ReadsTheAuthDomainCache(t *testing.T) {
	db, ctx := openTenantCacheTestDB(t)
	require.NoError(t, NewTenantCacheReadModel(db).Save(ctx, cachedTestTenant()))

	checker := NewTenantDomainChecker(database.NewTenantAwareDB(db))

	allowed, err := checker.IsDomainAllowed(ctx, "someone@auth-cache-test.com")
	require.NoError(t, err)
	assert.True(t, allowed)

	foreign, err := checker.IsDomainAllowed(ctx, "someone@elsewhere.com")
	require.NoError(t, err)
	assert.False(t, foreign)
}

func TestBackfillMigration_PopulatesAuthTenantCachesFromPlatformTables(t *testing.T) {
	db, ctx := openTenantCacheTestDB(t)
	seedPlatformTenant(t, db)

	sqlBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "deploy-scripts", "migrations", backfillMigration))
	require.NoError(t, err)
	_, err = db.Exec(string(sqlBytes))
	require.NoError(t, err)

	known, err := NewTenantCacheReadModel(db).Exists(ctx, tenantCacheTestTenant)
	require.NoError(t, err)
	assert.True(t, known)

	var discoveryURL, issuerURL, clientID string
	require.NoError(t, db.QueryRow(
		"SELECT discovery_url, issuer_url, client_id FROM auth.tenant_oidc_cache WHERE tenant_id = $1",
		tenantCacheTestTenant).Scan(&discoveryURL, &issuerURL, &clientID))
	assert.Equal(t, "https://login.example.com/v2.0/.well-known/openid-configuration", discoveryURL)
	assert.Equal(t, "https://login.example.com/v2.0", issuerURL)
	assert.Equal(t, "client-id", clientID)

	allowed, err := NewTenantDomainChecker(database.NewTenantAwareDB(db)).IsDomainAllowed(ctx, "someone@auth-cache-test.com")
	require.NoError(t, err)
	assert.True(t, allowed)

	_, err = db.Exec(string(sqlBytes))
	require.NoError(t, err, "backfill is idempotent")
}

func seedPlatformTenant(t *testing.T, db *sql.DB) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM platform.tenant_oidc_configs WHERE tenant_id = $1", tenantCacheTestTenant)
		_, _ = db.Exec("DELETE FROM platform.tenant_domains WHERE tenant_id = $1", tenantCacheTestTenant)
		_, _ = db.Exec("DELETE FROM platform.tenants WHERE id = $1", tenantCacheTestTenant)
	})

	exec := func(query string, args ...any) {
		t.Helper()
		_, err := db.Exec(query, args...)
		require.NoError(t, err)
	}
	exec(`INSERT INTO platform.tenants (id, name, status) VALUES ($1, 'Auth Cache Test', 'active')`, tenantCacheTestTenant)
	exec(`INSERT INTO platform.tenant_domains (domain, tenant_id) VALUES ('auth-cache-test.com', $1)`, tenantCacheTestTenant)
	exec(`INSERT INTO platform.tenant_oidc_configs (tenant_id, discovery_url, issuer_url, client_id, auth_method, scopes)
	      VALUES ($1, 'https://login.example.com/v2.0/.well-known/openid-configuration', 'https://login.example.com/v2.0', 'client-id', 'client_secret', 'openid email profile')`, tenantCacheTestTenant)
}
