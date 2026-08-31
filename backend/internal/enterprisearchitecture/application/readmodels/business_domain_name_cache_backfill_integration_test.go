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
	"easi/backend/internal/shared/eventsourcing/valueobjects"
)

const businessDomainNameCacheTestTenant = "ea-bdname-cache-test-tenant"
const businessDomainNameCacheBackfillMigration = "144_backfill_enterprisearchitecture_business_domain_name_cache.sql"

func openBusinessDomainNameCacheTestDB(t *testing.T) (*sql.DB, context.Context) {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	db, err := sql.Open("postgres", "host=localhost port=5432 user=easi password=easi dbname=easi sslmode=disable")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM enterprisearchitecture.business_domain_name_cache WHERE tenant_id = $1", businessDomainNameCacheTestTenant)
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

func TestBackfillMigration_PopulatesBusinessDomainNameCacheFromCapabilityMapping(t *testing.T) {
	db, ctx := openBusinessDomainNameCacheTestDB(t)

	exec := func(query string, args ...any) {
		t.Helper()
		_, err := db.Exec(query, args...)
		require.NoError(t, err)
	}
	exec(`INSERT INTO capabilitymapping.business_domains (id, tenant_id, name, created_at) VALUES ('bd-1', $1, 'Finance', NOW())`, businessDomainNameCacheTestTenant)

	sqlBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "deploy-scripts", "migrations", businessDomainNameCacheBackfillMigration))
	require.NoError(t, err)
	exec(string(sqlBytes))

	cache := NewBusinessDomainNameCacheReadModel(database.NewTenantAwareDB(db))
	name, err := cache.Name(ctx, "bd-1")
	require.NoError(t, err)
	assert.Equal(t, "Finance", name)

	exec(`UPDATE capabilitymapping.business_domains SET name = 'Finance & Risk' WHERE tenant_id = $1 AND id = 'bd-1'`, businessDomainNameCacheTestTenant)
	exec(string(sqlBytes))

	name, err = cache.Name(ctx, "bd-1")
	require.NoError(t, err)
	assert.Equal(t, "Finance & Risk", name, "backfill is idempotent and re-syncs the name")

	var rows int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM enterprisearchitecture.business_domain_name_cache WHERE tenant_id = $1`, businessDomainNameCacheTestTenant).Scan(&rows))
	assert.Equal(t, 1, rows)
}
