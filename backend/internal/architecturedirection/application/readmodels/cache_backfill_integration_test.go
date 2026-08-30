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
