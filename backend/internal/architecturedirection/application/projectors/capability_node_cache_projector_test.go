package projectors

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/architecturedirection/application/readmodels"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
)

type recordingCapabilityNodeCache struct {
	nodes            map[string]*readmodels.CapabilityNodeDTO
	inserted         []readmodels.CapabilityNodeDTO
	deleted          []string
	parentUpdates    []readmodels.ParentL1Update
	levelUpdates     map[string]string
	subtreeDomains   []subtreeDomainUpdate
	recalculatedFrom []string
	maturities       map[string]int
	renamedDomains   map[string]string
}

type subtreeDomainUpdate struct {
	L1CapabilityID string
	Domain         readmodels.BusinessDomainRef
}

func newRecordingNodeCache(nodes ...readmodels.CapabilityNodeDTO) *recordingCapabilityNodeCache {
	cache := &recordingCapabilityNodeCache{
		nodes: map[string]*readmodels.CapabilityNodeDTO{}, levelUpdates: map[string]string{},
		maturities: map[string]int{}, renamedDomains: map[string]string{},
	}
	for i := range nodes {
		node := nodes[i]
		cache.nodes[node.CapabilityID] = &node
	}
	return cache
}

func (c *recordingCapabilityNodeCache) GetByID(_ context.Context, capabilityID string) (*readmodels.CapabilityNodeDTO, error) {
	return c.nodes[capabilityID], nil
}

func (c *recordingCapabilityNodeCache) Insert(_ context.Context, dto readmodels.CapabilityNodeDTO) error {
	c.inserted = append(c.inserted, dto)
	return nil
}

func (c *recordingCapabilityNodeCache) Delete(_ context.Context, capabilityID string) error {
	c.deleted = append(c.deleted, capabilityID)
	return nil
}

func (c *recordingCapabilityNodeCache) UpdateParentAndL1(_ context.Context, update readmodels.ParentL1Update) error {
	c.parentUpdates = append(c.parentUpdates, update)
	return nil
}

func (c *recordingCapabilityNodeCache) UpdateLevel(_ context.Context, capabilityID, newLevel string) error {
	c.levelUpdates[capabilityID] = newLevel
	return nil
}

func (c *recordingCapabilityNodeCache) UpdateBusinessDomainForL1Subtree(_ context.Context, l1CapabilityID string, domain readmodels.BusinessDomainRef) error {
	c.subtreeDomains = append(c.subtreeDomains, subtreeDomainUpdate{L1CapabilityID: l1CapabilityID, Domain: domain})
	return nil
}

func (c *recordingCapabilityNodeCache) UpdateBusinessDomainName(_ context.Context, domain readmodels.BusinessDomainRef) error {
	c.renamedDomains[domain.ID] = domain.Name
	return nil
}

func (c *recordingCapabilityNodeCache) RecalculateL1ForSubtree(_ context.Context, capabilityID string) error {
	c.recalculatedFrom = append(c.recalculatedFrom, capabilityID)
	return nil
}

func (c *recordingCapabilityNodeCache) UpdateMaturityValue(_ context.Context, capabilityID string, maturityValue int) error {
	c.maturities[capabilityID] = maturityValue
	return nil
}

func domainNameLookup(names map[string]string) BusinessDomainNameLookup {
	return func(_ context.Context, id string) (string, error) { return names[id], nil }
}

func projectNodeEvent(t *testing.T, p *CapabilityNodeCacheProjector, eventType string, payload map[string]any) {
	t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	require.NoError(t, p.ProjectEvent(context.Background(), eventType, data))
}

func TestCapabilityNodeCacheProjector_CreatedChild_InheritsL1AndDomainFromParent(t *testing.T) {
	cache := newRecordingNodeCache(readmodels.CapabilityNodeDTO{
		CapabilityID: "l1", CapabilityLevel: "L1", L1CapabilityID: "l1", BusinessDomainID: "bd-1", BusinessDomainName: "Finance",
	})
	projector := NewCapabilityNodeCacheProjector(cache, domainNameLookup(nil))

	projectNodeEvent(t, projector, cmPL.CapabilityCreated, map[string]any{"id": "l2", "name": "Billing", "parentId": "l1", "level": "L2", "maturityValue": 12})

	require.Len(t, cache.inserted, 1)
	assert.Equal(t, readmodels.CapabilityNodeDTO{
		CapabilityID: "l2", CapabilityName: "Billing", CapabilityLevel: "L2", ParentID: "l1",
		L1CapabilityID: "l1", BusinessDomainID: "bd-1", BusinessDomainName: "Finance", MaturityValue: 12,
	}, cache.inserted[0])
}

func TestCapabilityNodeCacheProjector_CreatedL1_IsItsOwnL1(t *testing.T) {
	cache := newRecordingNodeCache()
	projector := NewCapabilityNodeCacheProjector(cache, domainNameLookup(nil))

	projectNodeEvent(t, projector, cmPL.CapabilityCreated, map[string]any{"id": "l1", "name": "Finance Ops", "level": "L1", "maturityValue": 12})

	require.Len(t, cache.inserted, 1)
	assert.Equal(t, "l1", cache.inserted[0].L1CapabilityID)
	assert.Empty(t, cache.inserted[0].BusinessDomainID)
}

func TestCapabilityNodeCacheProjector_Created_WithMaturityValue_UsesProvidedValue(t *testing.T) {
	cache := newRecordingNodeCache()
	projector := NewCapabilityNodeCacheProjector(cache, domainNameLookup(nil))

	projectNodeEvent(t, projector, cmPL.CapabilityCreated, map[string]any{"id": "l1", "name": "Finance Ops", "level": "L1", "maturityValue": 62})

	require.Len(t, cache.inserted, 1)
	assert.Equal(t, 62, cache.inserted[0].MaturityValue)
}

func TestCapabilityNodeCacheProjector_Created_MissingMaturityValue_DefaultsToGenesis(t *testing.T) {
	cache := newRecordingNodeCache()
	projector := NewCapabilityNodeCacheProjector(cache, domainNameLookup(nil))

	projectNodeEvent(t, projector, cmPL.CapabilityCreated, map[string]any{"id": "l1", "name": "Finance Ops", "level": "L1"})

	require.Len(t, cache.inserted, 1)
	assert.Equal(t, 12, cache.inserted[0].MaturityValue, "a CapabilityCreated event published before this field existed must default to Genesis (12), not zero")
}

func TestCapabilityNodeCacheProjector_Updated_RenamesExistingNode(t *testing.T) {
	cache := newRecordingNodeCache(readmodels.CapabilityNodeDTO{CapabilityID: "l1", CapabilityName: "Old", CapabilityLevel: "L1", L1CapabilityID: "l1"})
	projector := NewCapabilityNodeCacheProjector(cache, domainNameLookup(nil))

	projectNodeEvent(t, projector, cmPL.CapabilityUpdated, map[string]any{"id": "l1", "name": "New"})

	require.Len(t, cache.inserted, 1)
	assert.Equal(t, "New", cache.inserted[0].CapabilityName)
	assert.Equal(t, "l1", cache.inserted[0].L1CapabilityID)
}

func TestCapabilityNodeCacheProjector_Updated_UnknownNodeIsIgnored(t *testing.T) {
	cache := newRecordingNodeCache()
	projector := NewCapabilityNodeCacheProjector(cache, domainNameLookup(nil))

	projectNodeEvent(t, projector, cmPL.CapabilityUpdated, map[string]any{"id": "ghost", "name": "New"})

	assert.Empty(t, cache.inserted)
}

func TestCapabilityNodeCacheProjector_Deleted_RemovesNode(t *testing.T) {
	cache := newRecordingNodeCache()
	projector := NewCapabilityNodeCacheProjector(cache, domainNameLookup(nil))

	projectNodeEvent(t, projector, cmPL.CapabilityDeleted, map[string]any{"id": "l2"})

	assert.Equal(t, []string{"l2"}, cache.deleted)
}

func TestCapabilityNodeCacheProjector_ParentChanged_UpdatesParentThenRecalculatesSubtree(t *testing.T) {
	cache := newRecordingNodeCache()
	projector := NewCapabilityNodeCacheProjector(cache, domainNameLookup(nil))

	projectNodeEvent(t, projector, cmPL.CapabilityParentChanged, map[string]any{
		"capabilityId": "l3", "oldParentId": "l2a", "newParentId": "l2b", "oldLevel": "L3", "newLevel": "L3",
	})

	require.Len(t, cache.parentUpdates, 1)
	assert.Equal(t, readmodels.ParentL1Update{CapabilityID: "l3", NewParentID: "l2b", NewLevel: "L3", NewL1CapabilityID: "l3"}, cache.parentUpdates[0])
	assert.Equal(t, []string{"l3"}, cache.recalculatedFrom)
}

func TestCapabilityNodeCacheProjector_LevelChanged_UpdatesLevel(t *testing.T) {
	cache := newRecordingNodeCache()
	projector := NewCapabilityNodeCacheProjector(cache, domainNameLookup(nil))

	projectNodeEvent(t, projector, cmPL.CapabilityLevelChanged, map[string]any{"capabilityId": "l2", "newLevel": "L3"})

	assert.Equal(t, "L3", cache.levelUpdates["l2"])
}

func TestCapabilityNodeCacheProjector_AssignedToDomain_PropagatesNamedDomainToL1Subtree(t *testing.T) {
	cache := newRecordingNodeCache(readmodels.CapabilityNodeDTO{CapabilityID: "l1", CapabilityLevel: "L1", L1CapabilityID: "l1"})
	projector := NewCapabilityNodeCacheProjector(cache, domainNameLookup(map[string]string{"bd-1": "Finance"}))

	projectNodeEvent(t, projector, cmPL.CapabilityAssignedToDomain, map[string]any{"id": "a-1", "businessDomainId": "bd-1", "capabilityId": "l1"})

	require.Len(t, cache.subtreeDomains, 1)
	assert.Equal(t, subtreeDomainUpdate{L1CapabilityID: "l1", Domain: readmodels.BusinessDomainRef{ID: "bd-1", Name: "Finance"}}, cache.subtreeDomains[0])
	assert.Equal(t, []string{"l1"}, cache.recalculatedFrom)
}

func TestCapabilityNodeCacheProjector_AssignedToDomain_NonL1Node_IsSkipped(t *testing.T) {
	cache := newRecordingNodeCache(readmodels.CapabilityNodeDTO{
		CapabilityID: "l2", CapabilityLevel: "L2", ParentID: "l1", L1CapabilityID: "l1",
	})
	projector := NewCapabilityNodeCacheProjector(cache, domainNameLookup(map[string]string{"bd-1": "Finance"}))

	projectNodeEvent(t, projector, cmPL.CapabilityAssignedToDomain, map[string]any{"id": "a-1", "businessDomainId": "bd-1", "capabilityId": "l2"})

	assert.Empty(t, cache.subtreeDomains, "only L1 capabilities can be assigned to a business domain (R14); a non-L1 event is a corrupted invariant and must not propagate")
	assert.Empty(t, cache.recalculatedFrom)
}

func TestCapabilityNodeCacheProjector_UnassignedFromDomain_ClearsDomainOnL1Subtree(t *testing.T) {
	cache := newRecordingNodeCache(readmodels.CapabilityNodeDTO{CapabilityID: "l1", CapabilityLevel: "L1", L1CapabilityID: "l1", BusinessDomainID: "bd-1"})
	projector := NewCapabilityNodeCacheProjector(cache, domainNameLookup(nil))

	projectNodeEvent(t, projector, cmPL.CapabilityUnassignedFromDomain, map[string]any{"id": "a-1", "businessDomainId": "bd-1", "capabilityId": "l1"})

	require.Len(t, cache.subtreeDomains, 1)
	assert.Equal(t, readmodels.BusinessDomainRef{}, cache.subtreeDomains[0].Domain)
}

func TestCapabilityNodeCacheProjector_MetadataUpdated_StoresMaturity(t *testing.T) {
	cache := newRecordingNodeCache()
	projector := NewCapabilityNodeCacheProjector(cache, domainNameLookup(nil))

	projectNodeEvent(t, projector, cmPL.CapabilityMetadataUpdated, map[string]any{"id": "l2", "maturityValue": 63})

	assert.Equal(t, 63, cache.maturities["l2"])
}

func TestCapabilityNodeCacheProjector_BusinessDomainUpdated_RenamesDomainOnAllNodes(t *testing.T) {
	cache := newRecordingNodeCache()
	projector := NewCapabilityNodeCacheProjector(cache, domainNameLookup(nil))

	projectNodeEvent(t, projector, cmPL.BusinessDomainUpdated, map[string]any{"id": "bd-1", "name": "Finance & Risk"})

	assert.Equal(t, "Finance & Risk", cache.renamedDomains["bd-1"])
}
