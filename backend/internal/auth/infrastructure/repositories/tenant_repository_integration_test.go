//go:build integration

package repositories

import (
	"context"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantRepository_ReadsNameAndDomainsFromAuthCache(t *testing.T) {
	db := openOIDCCacheTestDB(t)
	seedCachedTenant(t, db, "active")
	repo := NewTenantRepository(db)

	tenant, err := repo.GetByID(context.Background(), oidcCacheTestTenant)
	require.NoError(t, err)
	assert.Equal(t, oidcCacheTestTenant, tenant.ID)
	assert.Equal(t, "OIDC Test", tenant.Name)

	domains, err := repo.GetDomains(context.Background(), oidcCacheTestTenant)
	require.NoError(t, err)
	assert.Equal(t, []string{"auth-oidc-test.com"}, domains)
}

func TestTenantRepository_ReportsUnknownTenant(t *testing.T) {
	db := openOIDCCacheTestDB(t)

	_, err := NewTenantRepository(db).GetByID(context.Background(), "no-such-tenant")

	assert.ErrorIs(t, err, ErrTenantNotFound)
}
