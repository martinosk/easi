//go:build integration

package readmodels

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/infrastructure/database"
)

func TestReferenceCache_ExistenceFollowsSaveAndRemove(t *testing.T) {
	db, ctx := openCacheTestDB(t)
	cache := NewReferenceCacheReadModel(database.NewTenantAwareDB(db))

	exists, err := cache.Exists(ctx, ReferenceEntityApplication, "comp-1")
	require.NoError(t, err)
	assert.False(t, exists)

	require.NoError(t, cache.SaveReference(ctx, ReferenceEntityApplication, "comp-1", "Phoenix"))
	require.NoError(t, cache.SaveReference(ctx, ReferenceEntityBusinessDomain, "bd-1", "Finance"))

	exists, err = cache.Exists(ctx, ReferenceEntityApplication, "comp-1")
	require.NoError(t, err)
	assert.True(t, exists)

	crossKind, err := cache.Exists(ctx, ReferenceEntityBusinessDomain, "comp-1")
	require.NoError(t, err)
	assert.False(t, crossKind, "existence is scoped to the referenced kind")

	require.NoError(t, cache.SaveReference(ctx, ReferenceEntityApplication, "comp-1", "Phoenix v2"))
	var name string
	require.NoError(t, db.QueryRow(
		`SELECT name FROM architecturedirection.reference_name_cache
		 WHERE tenant_id = $1 AND entity_type = 'application' AND entity_id = 'comp-1'`, cacheTestTenant).Scan(&name))
	assert.Equal(t, "Phoenix v2", name)

	require.NoError(t, cache.RemoveReference(ctx, ReferenceEntityApplication, "comp-1"))
	exists, err = cache.Exists(ctx, ReferenceEntityApplication, "comp-1")
	require.NoError(t, err)
	assert.False(t, exists)

	stillThere, err := cache.Exists(ctx, ReferenceEntityBusinessDomain, "bd-1")
	require.NoError(t, err)
	assert.True(t, stillThere)
}

func TestRealizationCache_DirectRealizationLookupAndRemovals(t *testing.T) {
	db, ctx := openCacheTestDB(t)
	cache := NewRealizationCacheReadModel(database.NewTenantAwareDB(db))

	_, found, err := cache.DirectRealizationID(ctx, "cap-1", "comp-1")
	require.NoError(t, err)
	assert.False(t, found)

	require.NoError(t, cache.SaveDirectRealization(ctx, DirectRealizationDTO{RealizationID: "r-1", CapabilityID: "cap-1", ComponentID: "comp-1"}))
	require.NoError(t, cache.SaveDirectRealization(ctx, DirectRealizationDTO{RealizationID: "r-2", CapabilityID: "cap-1", ComponentID: "comp-2"}))
	require.NoError(t, cache.SaveDirectRealization(ctx, DirectRealizationDTO{RealizationID: "r-3", CapabilityID: "cap-2", ComponentID: "comp-1"}))

	realizationID, found, err := cache.DirectRealizationID(ctx, "cap-1", "comp-1")
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, RealizationID("r-1"), realizationID)

	require.NoError(t, cache.RemoveRealization(ctx, "r-1"))
	_, found, err = cache.DirectRealizationID(ctx, "cap-1", "comp-1")
	require.NoError(t, err)
	assert.False(t, found)

	require.NoError(t, cache.RemoveRealizationsOfCapability(ctx, "cap-1"))
	_, found, err = cache.DirectRealizationID(ctx, "cap-1", "comp-2")
	require.NoError(t, err)
	assert.False(t, found)

	require.NoError(t, cache.RemoveRealizationsOfComponent(ctx, "comp-1"))
	_, found, err = cache.DirectRealizationID(ctx, "cap-2", "comp-1")
	require.NoError(t, err)
	assert.False(t, found)
}
