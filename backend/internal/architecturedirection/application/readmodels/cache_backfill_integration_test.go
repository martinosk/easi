//go:build integration

package readmodels

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/infrastructure/database"
)

const backfillMigration = "138_backfill_architecturedirection_caches.sql"

func TestBackfillMigration_PopulatesDirectionCachesFromUpstreamTables(t *testing.T) {
	db, ctx := openCacheTestDB(t)
	t.Cleanup(func() {
		for _, table := range []string{
			"capabilitymapping.capabilities", "capabilitymapping.domain_capability_assignments",
			"capabilitymapping.business_domains", "enterprisearchitecture.enterprise_capabilities",
		} {
			_, _ = db.Exec("DELETE FROM "+table+" WHERE tenant_id = $1", cacheTestTenant)
		}
	})

	exec := func(query string, args ...any) {
		t.Helper()
		_, err := db.Exec(query, args...)
		require.NoError(t, err)
	}
	exec(`INSERT INTO capabilitymapping.business_domains (id, tenant_id, name, created_at) VALUES ('bd-1', $1, 'Finance', NOW())`, cacheTestTenant)
	exec(`INSERT INTO capabilitymapping.capabilities (id, tenant_id, name, level, parent_id, maturity_value, created_at) VALUES
		('l1', $1, 'Finance Ops', 'L1', NULL, 40, NOW()),
		('l2', $1, 'Billing', 'L2', 'l1', 55, NOW()),
		('l3', $1, 'Invoicing', 'L3', 'l2', 70, NOW())`, cacheTestTenant)
	exec(`INSERT INTO capabilitymapping.domain_capability_assignments (assignment_id, tenant_id, business_domain_id, business_domain_name, capability_id, capability_name, capability_level, assigned_at)
		VALUES ('a-1', $1, 'bd-1', 'Finance', 'l1', 'Finance Ops', 'L1', NOW())`, cacheTestTenant)
	exec(`INSERT INTO enterprisearchitecture.enterprise_capabilities (id, tenant_id, name, category, active, target_maturity, created_at)
		VALUES ('ec-1', $1, 'Payments', 'Finance', true, 80, NOW())`, cacheTestTenant)

	sqlBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "deploy-scripts", "migrations", backfillMigration))
	require.NoError(t, err)
	exec(string(sqlBytes))

	nodes := NewCapabilityNodeCacheReadModel(database.NewTenantAwareDB(db))
	leaf, err := nodes.GetByID(ctx, "l3")
	require.NoError(t, err)
	require.NotNil(t, leaf)
	assert.Equal(t, "l1", leaf.L1CapabilityID)
	assert.Equal(t, "l2", leaf.ParentID)
	assert.Equal(t, "bd-1", leaf.BusinessDomainID)
	assert.Equal(t, "Finance", leaf.BusinessDomainName)

	var maturity int
	require.NoError(t, db.QueryRow(`SELECT maturity_value FROM architecturedirection.capability_node_cache WHERE tenant_id = $1 AND capability_id = 'l3'`, cacheTestTenant).Scan(&maturity))
	assert.Equal(t, 70, maturity)

	ecs := NewEnterpriseCapabilityCacheReadModel(database.NewTenantAwareDB(db))
	ec, err := ecs.GetByID(ctx, "ec-1")
	require.NoError(t, err)
	require.NotNil(t, ec)
	assert.Equal(t, "Payments", ec.Name)
	assert.Equal(t, "Finance", ec.Category)
	require.NotNil(t, ec.TargetMaturity)
	assert.Equal(t, 80, *ec.TargetMaturity)

	exec(string(sqlBytes))
	var rows int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM architecturedirection.capability_node_cache WHERE tenant_id = $1`, cacheTestTenant).Scan(&rows))
	assert.Equal(t, 3, rows, "backfill is idempotent")
}

const referenceRealizationBackfillMigration = "146_backfill_architecturedirection_reference_and_realization_caches.sql"

func TestBackfillMigration_SeedsRealizationCacheAndReconcilesReferenceCache(t *testing.T) {
	db, ctx := openCacheTestDB(t)
	t.Cleanup(func() {
		for _, table := range []string{
			"capabilitymapping.capability_realizations", "capabilitymapping.business_domains",
			"architecturemodeling.application_components",
		} {
			_, _ = db.Exec("DELETE FROM "+table+" WHERE tenant_id = $1", cacheTestTenant)
		}
	})

	exec := func(query string, args ...any) {
		t.Helper()
		_, err := db.Exec(query, args...)
		require.NoError(t, err)
	}
	exec(`INSERT INTO capabilitymapping.business_domains (id, tenant_id, name, created_at) VALUES ('bd-live', $1, 'Finance', NOW())`, cacheTestTenant)
	exec(`INSERT INTO architecturemodeling.application_components (id, tenant_id, name, is_deleted, created_at) VALUES
		('comp-live', $1, 'Phoenix', FALSE, NOW()),
		('comp-gone', $1, 'Retired', TRUE, NOW())`, cacheTestTenant)
	exec(`INSERT INTO capabilitymapping.capability_realizations (id, tenant_id, capability_id, component_id, realization_level, origin, linked_at) VALUES
		('r-direct', $1, 'cap-1', 'comp-live', 'Full', 'Direct', NOW()),
		('r-inherited', $1, 'cap-2', 'comp-live', 'Full', 'Inherited', NOW())`, cacheTestTenant)
	exec(`INSERT INTO architecturedirection.reference_name_cache (tenant_id, entity_type, entity_id, name) VALUES
		($1, 'application', 'comp-gone', 'Retired'),
		($1, 'business_domain', 'bd-gone', 'Dissolved')`, cacheTestTenant)

	sqlBytes, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "deploy-scripts", "migrations", referenceRealizationBackfillMigration))
	require.NoError(t, err)
	exec(string(sqlBytes))

	realizations := NewRealizationCacheReadModel(database.NewTenantAwareDB(db))
	realizationID, found, err := realizations.DirectRealizationID(ctx, "cap-1", "comp-live")
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, RealizationID("r-direct"), realizationID)

	_, inheritedCached, err := realizations.DirectRealizationID(ctx, "cap-2", "comp-live")
	require.NoError(t, err)
	assert.False(t, inheritedCached, "only direct realizations are cached")

	references := NewReferenceCacheReadModel(database.NewTenantAwareDB(db))
	assertCached := func(entity ReferenceEntity, entityID string, expected bool) {
		t.Helper()
		exists, err := references.Exists(ctx, entity, entityID)
		require.NoError(t, err)
		assert.Equal(t, expected, exists, "%s %s", entity, entityID)
	}
	assertCached(ReferenceEntityApplication, "comp-live", true)
	assertCached(ReferenceEntityBusinessDomain, "bd-live", true)
	assertCached(ReferenceEntityApplication, "comp-gone", false)
	assertCached(ReferenceEntityBusinessDomain, "bd-gone", false)

	exec(string(sqlBytes))
	var rows int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM architecturedirection.realization_cache WHERE tenant_id = $1`, cacheTestTenant).Scan(&rows))
	assert.Equal(t, 1, rows, "backfill is idempotent")
}
