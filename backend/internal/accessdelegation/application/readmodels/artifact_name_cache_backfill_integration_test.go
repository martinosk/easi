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

const (
	nameCacheTestTenant = "ad-name-cache-test-tenant"
	backfillMigration   = "142_backfill_accessdelegation_artifact_name_cache.sql"
)

var nameCacheSourceTables = []string{
	"capabilitymapping.capabilities",
	"capabilitymapping.business_domains",
	"architecturemodeling.application_components",
	"architecturemodeling.vendors",
	"architecturemodeling.acquired_entities",
	"architecturemodeling.internal_teams",
	"architectureviews.architecture_views",
	"accessdelegation.artifact_name_cache",
}

func openNameCacheTestDB(t *testing.T) (*sql.DB, context.Context) {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	db, err := sql.Open("postgres", "host=localhost port=5432 user=easi password=easi dbname=easi sslmode=disable")
	require.NoError(t, err)
	t.Cleanup(func() {
		for _, table := range nameCacheSourceTables {
			_, _ = db.Exec("DELETE FROM "+table+" WHERE tenant_id = $1", nameCacheTestTenant)
		}
		_ = db.Close()
	})
	return db, sharedctx.WithTenant(context.Background(), valueobjects.MustNewTenantID(nameCacheTestTenant))
}

func runNameCacheBackfill(t *testing.T, db *sql.DB) {
	t.Helper()
	sqlBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "deploy-scripts", "migrations", backfillMigration))
	require.NoError(t, err)
	_, err = db.Exec(string(sqlBytes))
	require.NoError(t, err)
}

func seedNameCacheSources(t *testing.T, db *sql.DB) {
	t.Helper()
	exec := func(query string, args ...any) {
		t.Helper()
		_, err := db.Exec(query, args...)
		require.NoError(t, err)
	}
	exec(`INSERT INTO capabilitymapping.capabilities (id, tenant_id, name, level, created_at) VALUES ('cap-1', $1, 'Customer Onboarding', 'L1', NOW())`, nameCacheTestTenant)
	exec(`INSERT INTO capabilitymapping.business_domains (id, tenant_id, name, created_at) VALUES ('bd-1', $1, 'Sales', NOW())`, nameCacheTestTenant)
	exec(`INSERT INTO architecturemodeling.application_components (id, tenant_id, name, created_at, is_deleted) VALUES
		('comp-1', $1, 'Payment Service', NOW(), FALSE),
		('comp-gone', $1, 'Retired Service', NOW(), TRUE)`, nameCacheTestTenant)
	exec(`INSERT INTO architecturemodeling.vendors (id, tenant_id, name, created_at, is_deleted) VALUES ('ven-1', $1, 'Acme', NOW(), FALSE)`, nameCacheTestTenant)
	exec(`INSERT INTO architecturemodeling.acquired_entities (id, tenant_id, name, created_at, is_deleted) VALUES ('ae-1', $1, 'Widget Co', NOW(), FALSE)`, nameCacheTestTenant)
	exec(`INSERT INTO architecturemodeling.internal_teams (id, tenant_id, name, created_at, is_deleted) VALUES ('team-1', $1, 'Platform', NOW(), FALSE)`, nameCacheTestTenant)
	exec(`INSERT INTO architectureviews.architecture_views (id, tenant_id, name, created_at, is_deleted) VALUES ('view-1', $1, 'Integration Map', NOW(), FALSE)`, nameCacheTestTenant)
}

func TestBackfillMigration_PopulatesArtifactNameCacheFromOwningContexts(t *testing.T) {
	db, ctx := openNameCacheTestDB(t)
	seedNameCacheSources(t, db)

	runNameCacheBackfill(t, db)

	cache := NewArtifactNameCacheReadModel(database.NewTenantAwareDB(db))
	expected := map[string]map[string]string{
		"capability":      {"cap-1": "Customer Onboarding"},
		"domain":          {"bd-1": "Sales"},
		"component":       {"comp-1": "Payment Service"},
		"vendor":          {"ven-1": "Acme"},
		"acquired_entity": {"ae-1": "Widget Co"},
		"internal_team":   {"team-1": "Platform"},
		"view":            {"view-1": "Integration Map"},
	}
	for artifactType, names := range expected {
		ids := make([]string, 0, len(names))
		for id := range names {
			ids = append(ids, id)
		}
		cached, err := cache.NamesByIDs(ctx, artifactType, ids)
		require.NoError(t, err)
		assert.Equal(t, names, cached, artifactType)
	}

	deleted, err := cache.NamesByIDs(ctx, "component", []string{"comp-gone"})
	require.NoError(t, err)
	assert.Empty(t, deleted, "soft-deleted artifacts are not cached")
}

func TestBackfillMigration_IsIdempotentAndRefreshesRenamedArtifacts(t *testing.T) {
	db, ctx := openNameCacheTestDB(t)
	seedNameCacheSources(t, db)
	runNameCacheBackfill(t, db)

	_, err := db.Exec(`UPDATE capabilitymapping.capabilities SET name = 'Onboarding' WHERE tenant_id = $1 AND id = 'cap-1'`, nameCacheTestTenant)
	require.NoError(t, err)
	runNameCacheBackfill(t, db)

	var rows int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM accessdelegation.artifact_name_cache WHERE tenant_id = $1`, nameCacheTestTenant).Scan(&rows))
	assert.Equal(t, 7, rows, "backfill is idempotent")

	cache := NewArtifactNameCacheReadModel(database.NewTenantAwareDB(db))
	names, err := cache.NamesByIDs(ctx, "capability", []string{"cap-1"})
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"cap-1": "Onboarding"}, names)
}
