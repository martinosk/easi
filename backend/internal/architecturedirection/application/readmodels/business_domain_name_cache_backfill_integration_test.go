//go:build integration

package readmodels

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/infrastructure/database"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/eventsourcing/valueobjects"
)

const businessDomainNameCacheTestTenant = "ea-bdname-cache-test-tenant"

func openBusinessDomainNameCacheTestDB(t *testing.T) (*sql.DB, context.Context) {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	db, err := sql.Open("postgres", "host=localhost port=5432 user=easi password=easi dbname=easi sslmode=disable")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM architecturedirection.business_domain_name_cache WHERE tenant_id = $1", businessDomainNameCacheTestTenant)
		_, _ = db.Exec("DELETE FROM capabilitymapping.business_domains WHERE tenant_id = $1", businessDomainNameCacheTestTenant)
		_ = db.Close()
	})
	return db, sharedctx.WithTenant(context.Background(), valueobjects.MustNewTenantID(businessDomainNameCacheTestTenant))
}

func TestBusinessDomainNameCache_UpsertRenameDelete_RoundTrip(t *testing.T) {
	db, ctx := openBusinessDomainNameCacheTestDB(t)
	cache := NewBusinessDomainNameCacheReadModel(database.NewTenantAwareDB(db))

	require.NoError(t, cache.Upsert(ctx, "bd-1", "Finance"))
	name, err := cache.Name(ctx, "bd-1")
	require.NoError(t, err)
	assert.Equal(t, "Finance", name)

	require.NoError(t, cache.Upsert(ctx, "bd-1", "Finance & Risk"))
	name, err = cache.Name(ctx, "bd-1")
	require.NoError(t, err)
	assert.Equal(t, "Finance & Risk", name)

	require.NoError(t, cache.Delete(ctx, "bd-1"))
	name, err = cache.Name(ctx, "bd-1")
	require.NoError(t, err)
	assert.Empty(t, name)
}
