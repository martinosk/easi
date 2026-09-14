package api

import (
	"testing"

	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func containmentAffordancesFor(t *testing.T, role sharedctx.Role, component *readmodels.ApplicationComponentDTO) sharedAPI.Links {
	t.Helper()
	links := NewArchitectureModelingLinks(sharedAPI.NewHATEOASLinks("/api/v1"))
	actor := sharedctx.NewActor("u1", "u@example.com", role)
	result := sharedAPI.Links{}
	links.AddContainmentAffordances(result, component, actor)
	return result
}

func standaloneComponent() *readmodels.ApplicationComponentDTO {
	return &readmodels.ApplicationComponentDTO{ID: "c1", Name: "Quoting"}
}

func partComponent() *readmodels.ApplicationComponentDTO {
	return &readmodels.ApplicationComponentDTO{
		ID:     "c1",
		Name:   "Quoting",
		PartOf: &readmodels.ContainmentParentDTO{ID: "p1", Name: "CRM Suite", Kind: valueobjects.ContainmentComposition},
	}
}

func parentComponent() *readmodels.ApplicationComponentDTO {
	return &readmodels.ApplicationComponentDTO{
		ID:    "p1",
		Name:  "CRM Suite",
		Parts: []readmodels.ContainmentPartDTO{{ID: "c1", Name: "Quoting", Kind: valueobjects.ContainmentComposition}},
	}
}

func TestContainmentAffordances_StandaloneGetsAttachOnly(t *testing.T) {
	result := containmentAffordancesFor(t, sharedctx.RoleArchitect, standaloneComponent())

	attach, ok := result["x-attach-to"]
	require.True(t, ok)
	assert.Equal(t, "PUT", attach.Method)
	assert.Equal(t, "/api/v1/components/c1/containment", attach.Href)
	assert.NotContains(t, result, "x-detach")
}

func TestContainmentAffordances_PartGetsDetachOnly(t *testing.T) {
	result := containmentAffordancesFor(t, sharedctx.RoleArchitect, partComponent())

	detach, ok := result["x-detach"]
	require.True(t, ok)
	assert.Equal(t, "DELETE", detach.Method)
	assert.Equal(t, "/api/v1/components/c1/containment", detach.Href)
	assert.NotContains(t, result, "x-attach-to")
}

func TestContainmentAffordances_PopulatedParentGetsNeither(t *testing.T) {
	result := containmentAffordancesFor(t, sharedctx.RoleArchitect, parentComponent())

	assert.Empty(t, result)
}

func TestContainmentAffordances_StakeholderGetsNone(t *testing.T) {
	for name, component := range map[string]*readmodels.ApplicationComponentDTO{
		"standalone": standaloneComponent(),
		"part":       partComponent(),
	} {
		t.Run(name, func(t *testing.T) {
			assert.Empty(t, containmentAffordancesFor(t, sharedctx.RoleStakeholder, component))
		})
	}
}

func TestEnrichWithLinks_AddsContainmentAffordance(t *testing.T) {
	h := &ComponentHandlers{
		hateoas: NewArchitectureModelingLinks(sharedAPI.NewHATEOASLinks("/api/v1")),
	}
	dto := standaloneComponent()
	dto.OwnershipState = valueobjects.OwnershipStateUnknown
	dto.Hosting = valueobjects.HostingUnknown

	h.enrichWithLinks(architectRequest(), dto)

	assert.Contains(t, dto.Links, "x-attach-to")
}
