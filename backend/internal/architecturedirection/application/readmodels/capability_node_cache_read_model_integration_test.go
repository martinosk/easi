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

const cacheTestTenant = "ad-cache-test-tenant"

func openCacheTestDB(t *testing.T) (*sql.DB, context.Context) {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test")
	}
	db, err := sql.Open("postgres", "host=localhost port=5432 user=easi password=easi dbname=easi sslmode=disable")
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM architecturedirection.capability_node_cache WHERE tenant_id = $1", cacheTestTenant)
		_, _ = db.Exec("DELETE FROM architecturedirection.reference_name_cache WHERE tenant_id = $1", cacheTestTenant)
		_, _ = db.Exec("DELETE FROM architecturedirection.realization_cache WHERE tenant_id = $1", cacheTestTenant)
		_ = db.Close()
	})
	return db, sharedctx.WithTenant(context.Background(), valueobjects.MustNewTenantID(cacheTestTenant))
}

func TestCapabilityNodeCache_DomainAssignmentPropagatesToSubtreeAndReparentingRecalculates(t *testing.T) {
	db, ctx := openCacheTestDB(t)
	cache := NewCapabilityNodeCacheReadModel(database.NewTenantAwareDB(db))

	require.NoError(t, cache.Insert(ctx, CapabilityNodeDTO{CapabilityID: "l1-a", CapabilityName: "A", CapabilityLevel: "L1", L1CapabilityID: "l1-a"}))
	require.NoError(t, cache.Insert(ctx, CapabilityNodeDTO{CapabilityID: "l1-b", CapabilityName: "B", CapabilityLevel: "L1", L1CapabilityID: "l1-b", BusinessDomainID: "bd-b", BusinessDomainName: "Domain B"}))
	require.NoError(t, cache.Insert(ctx, CapabilityNodeDTO{CapabilityID: "l2", CapabilityName: "A.1", CapabilityLevel: "L2", ParentID: "l1-a", L1CapabilityID: "l1-a"}))
	require.NoError(t, cache.Insert(ctx, CapabilityNodeDTO{CapabilityID: "l3", CapabilityName: "A.1.1", CapabilityLevel: "L3", ParentID: "l2", L1CapabilityID: "l1-a"}))

	require.NoError(t, cache.UpdateBusinessDomainForL1Subtree(ctx, "l1-a", BusinessDomainRef{ID: "bd-a", Name: "Domain A"}))
	leaf, err := cache.GetByID(ctx, "l3")
	require.NoError(t, err)
	assert.Equal(t, "bd-a", leaf.BusinessDomainID)
	assert.Equal(t, "Domain A", leaf.BusinessDomainName)

	require.NoError(t, cache.UpdateParentAndL1(ctx, ParentL1Update{CapabilityID: "l2", NewParentID: "l1-b", NewLevel: "L2", NewL1CapabilityID: "l2"}))
	require.NoError(t, cache.RecalculateL1ForSubtree(ctx, "l2"))

	moved, err := cache.GetByID(ctx, "l3")
	require.NoError(t, err)
	assert.Equal(t, "l1-b", moved.L1CapabilityID)
	assert.Equal(t, "bd-b", moved.BusinessDomainID)
	assert.Equal(t, "Domain B", moved.BusinessDomainName)
}

func TestCapabilityNodeCache_BusinessDomainNameComesFromReferenceNameCache(t *testing.T) {
	db, ctx := openCacheTestDB(t)
	cache := NewCapabilityNodeCacheReadModel(database.NewTenantAwareDB(db))
	_, err := db.Exec(`INSERT INTO architecturedirection.reference_name_cache (tenant_id, entity_type, entity_id, name) VALUES ($1, 'business_domain', 'bd-x', 'Domain X')`, cacheTestTenant)
	require.NoError(t, err)

	name, err := cache.BusinessDomainName(ctx, "bd-x")
	require.NoError(t, err)
	assert.Equal(t, "Domain X", name)

	missing, err := cache.BusinessDomainName(ctx, "bd-unknown")
	require.NoError(t, err)
	assert.Empty(t, missing)
}

func TestCapabilityNodeCache_RenamePreservesMaturityValue(t *testing.T) {
	db, ctx := openCacheTestDB(t)
	cache := NewCapabilityNodeCacheReadModel(database.NewTenantAwareDB(db))

	require.NoError(t, cache.Insert(ctx, CapabilityNodeDTO{
		CapabilityID: "cap", CapabilityName: "Before", CapabilityLevel: "L1", L1CapabilityID: "cap",
	}))
	require.NoError(t, cache.UpdateMaturityValue(ctx, "cap", 55))

	node, err := cache.GetByID(ctx, "cap")
	require.NoError(t, err)
	require.NotNil(t, node)
	require.Equal(t, 55, node.MaturityValue)

	node.CapabilityName = "After"
	require.NoError(t, cache.Insert(ctx, *node))

	renamed, err := cache.GetByID(ctx, "cap")
	require.NoError(t, err)
	require.NotNil(t, renamed)
	assert.Equal(t, "After", renamed.CapabilityName)
	assert.Equal(t, 55, renamed.MaturityValue)
}
