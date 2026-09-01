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

func hostingAffordancesFor(t *testing.T, role sharedctx.Role) sharedAPI.Links {
	t.Helper()
	links := NewArchitectureModelingLinks(sharedAPI.NewHATEOASLinks("/api/v1"))
	actor := sharedctx.NewActor("u1", "u@example.com", role)
	result := sharedAPI.Links{}
	links.AddHostingAffordances(result, "c1", actor)
	return result
}

func TestHostingAffordances_ArchitectGetsClassify(t *testing.T) {
	result := hostingAffordancesFor(t, sharedctx.RoleArchitect)

	classify, ok := result["x-classify-hosting"]
	require.True(t, ok)
	assert.Equal(t, "PUT", classify.Method)
	assert.Equal(t, "/api/v1/components/c1/hosting", classify.Href)
}

func TestHostingAffordances_StakeholderGetsNone(t *testing.T) {
	result := hostingAffordancesFor(t, sharedctx.RoleStakeholder)

	assert.Empty(t, result)
}

func TestEnrichWithLinks_AddsHostingAffordance(t *testing.T) {
	h := &ComponentHandlers{
		hateoas: NewArchitectureModelingLinks(sharedAPI.NewHATEOASLinks("/api/v1")),
	}
	dto := &readmodels.ApplicationComponentDTO{ID: "c1", Name: "Comp", OwnershipState: valueobjects.OwnershipStateUnknown, Hosting: valueobjects.HostingUnknown}

	h.enrichWithLinks(architectRequest(), dto)

	assert.Contains(t, dto.Links, "x-classify-hosting")
}
