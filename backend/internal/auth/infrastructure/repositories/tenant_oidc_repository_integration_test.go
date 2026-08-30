//go:build integration

package repositories

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const oidcCacheTestTenant = "auth-oidc-test"

func openOIDCCacheTestDB(t *testing.T) *sql.DB {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	db, err := sql.Open("postgres", "host=localhost port=5432 user=easi_app password=localdev dbname=easi sslmode=disable")
	require.NoError(t, err)
	require.NoError(t, db.Ping())
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM auth.tenant_oidc_configs WHERE tenant_id = $1", oidcCacheTestTenant)
		_, _ = db.Exec("DELETE FROM auth.tenant_domains WHERE tenant_id = $1", oidcCacheTestTenant)
		_, _ = db.Exec("DELETE FROM auth.tenants WHERE id = $1", oidcCacheTestTenant)
		_ = db.Close()
	})
	return db
}

func seedCachedTenant(t *testing.T, db *sql.DB, status string) {
	t.Helper()
	exec := func(query string, args ...any) {
		t.Helper()
		_, err := db.Exec(query, args...)
		require.NoError(t, err)
	}
	exec(`INSERT INTO auth.tenants (id, name, status) VALUES ($1, 'OIDC Test', $2)`, oidcCacheTestTenant, status)
	exec(`INSERT INTO auth.tenant_domains (domain, tenant_id) VALUES ('auth-oidc-test.com', $1)`, oidcCacheTestTenant)
	exec(`INSERT INTO auth.tenant_oidc_configs (tenant_id, discovery_url, issuer_url, client_id, auth_method, scopes)
	      VALUES ($1, 'https://login.example.com/v2.0/.well-known/openid-configuration', 'https://login.example.com/v2.0', 'client-id', 'client_secret', 'openid email profile')`, oidcCacheTestTenant)
}

func TestTenantOIDCRepository_ResolvesConfigurationFromAuthTablesByEmailDomain(t *testing.T) {
	db := openOIDCCacheTestDB(t)
	seedCachedTenant(t, db, "active")

	config, err := NewTenantOIDCRepository(db).GetByEmailDomain(context.Background(), "auth-oidc-test.com")

	require.NoError(t, err)
	assert.Equal(t, oidcCacheTestTenant, config.TenantID)
	assert.Equal(t, "https://login.example.com/v2.0/.well-known/openid-configuration", config.DiscoveryURL)
	assert.Equal(t, "https://login.example.com/v2.0", config.IssuerURL)
	assert.Equal(t, "client-id", config.ClientID)
	assert.Equal(t, "client_secret", config.AuthMethod)
	assert.Equal(t, "openid email profile", config.Scopes)
}

func TestTenantOIDCRepository_ResolvesConfigurationFromAuthTablesByTenantID(t *testing.T) {
	db := openOIDCCacheTestDB(t)
	seedCachedTenant(t, db, "active")

	config, err := NewTenantOIDCRepository(db).GetByTenantID(context.Background(), oidcCacheTestTenant)

	require.NoError(t, err)
	assert.Equal(t, "client-id", config.ClientID)
}

func TestTenantOIDCRepository_RejectsUnknownDomainAndInactiveTenant(t *testing.T) {
	db := openOIDCCacheTestDB(t)

	_, err := NewTenantOIDCRepository(db).GetByEmailDomain(context.Background(), "auth-oidc-test.com")
	assert.ErrorIs(t, err, ErrDomainNotFound)

	seedCachedTenant(t, db, "suspended")

	_, err = NewTenantOIDCRepository(db).GetByEmailDomain(context.Background(), "auth-oidc-test.com")
	assert.ErrorIs(t, err, ErrTenantInactive)

	_, err = NewTenantOIDCRepository(db).GetByTenantID(context.Background(), "no-such-tenant")
	assert.ErrorIs(t, err, ErrTenantNotFound)
}
