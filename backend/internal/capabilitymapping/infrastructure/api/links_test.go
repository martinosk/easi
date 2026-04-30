package api

import (
	"encoding/json"
	"testing"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func capabilityRelatedFor(t *testing.T, role sharedctx.Role, level string) []types.RelatedLink {
	t.Helper()
	h := sharedAPI.NewHATEOASLinks("/api/v1")
	links := NewCapabilityMappingLinks(h)
	actor := sharedctx.NewActor("u1", "u@example.com", role)
	return links.CapabilityXRelatedForActor(level, actor)
}

func capabilityRelatedForPerms(t *testing.T, perms map[string]bool, level string) []types.RelatedLink {
	t.Helper()
	h := sharedAPI.NewHATEOASLinks("/api/v1")
	links := NewCapabilityMappingLinks(h)
	actor := sharedctx.Actor{ID: "u1", Email: "u@example.com", Permissions: perms}
	return links.CapabilityXRelatedForActor(level, actor)
}

func findRelatedLink(items []types.RelatedLink, relationType string) *types.RelatedLink {
	for i := range items {
		if items[i].RelationType == relationType {
			return &items[i]
		}
	}
	return nil
}

func TestCapabilityXRelatedForActor_ArchitectAtL2GetsParentPOST(t *testing.T) {
	related := capabilityRelatedFor(t, sharedctx.RoleArchitect, "L2")

	entry := findRelatedLink(related, "capability-parent")
	require.NotNil(t, entry, "expected capability-parent entry")
	assert.Equal(t, "/api/v1/capabilities", entry.Href)
	assert.Contains(t, entry.Methods, "POST")
	assert.Equal(t, "capability", entry.TargetType)
	assert.Equal(t, "Capability (child of)", entry.Title)
}

func TestCapabilityXRelatedForActor_AtL4HasNoParentEntry(t *testing.T) {
	related := capabilityRelatedFor(t, sharedctx.RoleArchitect, "L4")

	entry := findRelatedLink(related, "capability-parent")
	assert.Nil(t, entry, "L4 capability must omit the capability-parent entry entirely")
}

func TestCapabilityXRelatedForActor_ArchitectGetsRealizationPOST(t *testing.T) {
	related := capabilityRelatedFor(t, sharedctx.RoleArchitect, "L3")

	entry := findRelatedLink(related, "capability-realization")
	require.NotNil(t, entry, "expected capability-realization entry")
	assert.Equal(t, "/api/v1/components", entry.Href)
	assert.Contains(t, entry.Methods, "POST")
	assert.Equal(t, "component", entry.TargetType)
	assert.Equal(t, "Component (realization)", entry.Title)
}

func TestCapabilityXRelatedForActor_RealizationRequiresBothPermissions(t *testing.T) {
	componentsOnly := map[string]bool{"components:write": true}
	related := capabilityRelatedForPerms(t, componentsOnly, "L2")
	assert.Nil(t, findRelatedLink(related, "capability-realization"),
		"actor with components:write but no capabilities:write must not see capability-realization (source mutation perm missing)")
	assert.Nil(t, findRelatedLink(related, "capability-parent"),
		"actor with components:write but no capabilities:write must not see capability-parent")

	capabilitiesOnly := map[string]bool{"capabilities:write": true}
	related = capabilityRelatedForPerms(t, capabilitiesOnly, "L2")
	assert.Nil(t, findRelatedLink(related, "capability-realization"),
		"actor with capabilities:write but no components:write must not see capability-realization (target create perm missing)")
	assert.NotNil(t, findRelatedLink(related, "capability-parent"),
		"actor with capabilities:write must still see capability-parent (target create + source mutation both satisfied by capabilities:write)")
}

func TestCapabilityXRelatedForActor_StakeholderHasNoPOSTAffordances(t *testing.T) {
	related := capabilityRelatedFor(t, sharedctx.RoleStakeholder, "L2")

	assert.Empty(t, related,
		"stakeholder must see no x-related entries (none advertise POST and no GET-only entries are emitted today)")
}

func TestAddLinksToCapability_EnrichToMarshaledJSON_AdvertisesXRelated(t *testing.T) {
	h := &CapabilityHandlers{
		hateoas: NewCapabilityMappingLinks(sharedAPI.NewHATEOASLinks("/api/v1")),
	}
	architect := sharedctx.NewActor("u1", "u@example.com", sharedctx.RoleArchitect)
	dto := &readmodels.CapabilityDTO{ID: "cap1", Name: "Cap", Level: "L2"}

	h.addLinksToCapability(dto, architect)

	data, err := json.Marshal(dto)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(data, &out))
	links := out["_links"].(map[string]any)
	xr, ok := links["x-related"].([]any)
	require.True(t, ok, "x-related missing in marshaled JSON: %s", string(data))

	relTypes := make([]string, 0, len(xr))
	for _, e := range xr {
		relTypes = append(relTypes, e.(map[string]any)["relationType"].(string))
	}
	assert.Contains(t, relTypes, "capability-parent")
	assert.Contains(t, relTypes, "capability-realization")
}
