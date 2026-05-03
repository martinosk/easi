package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/architecturemodeling/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func marshaledXRelated(t *testing.T, dto any) []map[string]any {
	t.Helper()
	data, err := json.Marshal(dto)
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(data, &out))
	links, ok := out["_links"].(map[string]any)
	require.True(t, ok, "_links missing or not an object: %s", string(data))
	xr, ok := links["x-related"].([]any)
	require.True(t, ok, "_links.x-related missing or not an array: %s", string(data))
	entries := make([]map[string]any, len(xr))
	for i, e := range xr {
		entries[i] = e.(map[string]any)
	}
	return entries
}

func componentRelatedFor(t *testing.T, role sharedctx.Role) []types.RelatedLink {
	t.Helper()
	h := sharedAPI.NewHATEOASLinks("/api/v1")
	links := NewArchitectureModelingLinks(h)
	actor := sharedctx.NewActor("u1", "u@example.com", role)
	return links.ComponentXRelatedForActor(actor)
}

func findRelated(items []types.RelatedLink, relationType string) *types.RelatedLink {
	for i := range items {
		if items[i].RelationType == relationType {
			return &items[i]
		}
	}
	return nil
}

func TestComponentXRelatedForActor_ArchitectGetsTriggersAndServesEntries(t *testing.T) {
	related := componentRelatedFor(t, sharedctx.RoleArchitect)

	cases := []struct{ relationType, title string }{
		{"component-triggers", "Component (triggers)"},
		{"component-serves", "Component (serves)"},
	}
	for _, tc := range cases {
		t.Run(tc.relationType, func(t *testing.T) {
			entry := findRelated(related, tc.relationType)
			require.NotNil(t, entry, "expected x-related entry with relationType=%s", tc.relationType)
			assert.Equal(t, "/api/v1/components", entry.Href)
			assert.Contains(t, entry.Methods, "POST")
			assert.Equal(t, "component", entry.TargetType)
			assert.Equal(t, tc.title, entry.Title)
		})
	}
	assert.Nil(t, findRelated(related, "component-relation"),
		"the legacy component-relation entry must be replaced by component-triggers/component-serves")
}

func TestComponentXRelatedForActor_ArchitectGetsOriginFromComponentEntries(t *testing.T) {
	related := componentRelatedFor(t, sharedctx.RoleArchitect)

	cases := []struct{ relationType, targetType, href, title string }{
		{"origin-acquired-via", "acquiredEntity", "/api/v1/acquired-entities", "Acquired Entity (acquired-via)"},
		{"origin-purchased-from", "vendor", "/api/v1/vendors", "Vendor (purchased-from)"},
		{"origin-built-by", "internalTeam", "/api/v1/internal-teams", "Internal Team (built-by)"},
	}
	for _, tc := range cases {
		t.Run(tc.relationType, func(t *testing.T) {
			entry := findRelated(related, tc.relationType)
			require.NotNil(t, entry, "expected origin-from-component entry with relationType=%s", tc.relationType)
			assert.Equal(t, tc.href, entry.Href)
			assert.Contains(t, entry.Methods, "POST")
			assert.Equal(t, tc.targetType, entry.TargetType)
			assert.Equal(t, tc.title, entry.Title)
		})
	}
}

func TestComponentXRelatedForActor_StakeholderGetsNoEntries(t *testing.T) {
	related := componentRelatedFor(t, sharedctx.RoleStakeholder)

	assert.Empty(t, related,
		"stakeholder must see no x-related entries (none advertise POST and no GET-only entries are emitted today)")
}

func TestEnrichWithLinks_PopulatesXRelatedFromActor(t *testing.T) {
	h := &ComponentHandlers{
		hateoas: NewArchitectureModelingLinks(sharedAPI.NewHATEOASLinks("/api/v1")),
	}
	architect := sharedctx.NewActor("u1", "u@example.com", sharedctx.RoleArchitect)
	req := httptest.NewRequest("GET", "/api/v1/components/c1", nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), architect))
	dto := &readmodels.ApplicationComponentDTO{ID: "c1", Name: "Comp"}

	h.enrichWithLinks(req, dto)

	require.NotEmpty(t, dto.XRelated, "expected XRelated to be populated for architect")
	require.NotNil(t, findRelated(dto.XRelated, "component-triggers"))
	require.NotNil(t, findRelated(dto.XRelated, "component-serves"))
	require.NotNil(t, findRelated(dto.XRelated, "origin-acquired-via"))
	require.NotNil(t, findRelated(dto.XRelated, "origin-purchased-from"))
	require.NotNil(t, findRelated(dto.XRelated, "origin-built-by"))
}

func originLinks(t *testing.T) *ArchitectureModelingLinks {
	t.Helper()
	return NewArchitectureModelingLinks(sharedAPI.NewHATEOASLinks("/api/v1"))
}

func TestOriginXRelatedForActor_ArchitectGetsComponentPOST(t *testing.T) {
	cases := []struct {
		name         string
		invoke       func(*ArchitectureModelingLinks, sharedctx.Actor) []types.RelatedLink
		relationType string
		title        string
	}{
		{
			name: "AcquiredEntity",
			invoke: func(l *ArchitectureModelingLinks, a sharedctx.Actor) []types.RelatedLink {
				return l.AcquiredEntityXRelatedForActor(a)
			},
			relationType: "origin-acquired-via",
			title:        "Component (acquired-via)",
		},
		{
			name: "Vendor",
			invoke: func(l *ArchitectureModelingLinks, a sharedctx.Actor) []types.RelatedLink {
				return l.VendorXRelatedForActor(a)
			},
			relationType: "origin-purchased-from",
			title:        "Component (purchased-from)",
		},
		{
			name: "InternalTeam",
			invoke: func(l *ArchitectureModelingLinks, a sharedctx.Actor) []types.RelatedLink {
				return l.InternalTeamXRelatedForActor(a)
			},
			relationType: "origin-built-by",
			title:        "Component (built-by)",
		},
	}
	architect := sharedctx.NewActor("u1", "u@example.com", sharedctx.RoleArchitect)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entry := findRelated(tc.invoke(originLinks(t), architect), tc.relationType)
			require.NotNil(t, entry)
			assert.Equal(t, "/api/v1/components", entry.Href)
			assert.Contains(t, entry.Methods, "POST")
			assert.Equal(t, "component", entry.TargetType)
			assert.Equal(t, tc.title, entry.Title)
		})
	}
}

func TestOriginXRelatedForActor_StakeholderGetsNothing(t *testing.T) {
	stakeholder := sharedctx.NewActor("u1", "u@example.com", sharedctx.RoleStakeholder)
	links := originLinks(t)

	assert.Empty(t, links.AcquiredEntityXRelatedForActor(stakeholder),
		"stakeholder must see no x-related entries on acquired-entity")
	assert.Empty(t, links.VendorXRelatedForActor(stakeholder),
		"stakeholder must see no x-related entries on vendor")
	assert.Empty(t, links.InternalTeamXRelatedForActor(stakeholder),
		"stakeholder must see no x-related entries on internal-team")
}

func architectRequest() *http.Request {
	architect := sharedctx.NewActor("u1", "u@example.com", sharedctx.RoleArchitect)
	req := httptest.NewRequest("GET", "/api/v1/foo", nil)
	return req.WithContext(sharedctx.WithActor(req.Context(), architect))
}

func TestOriginHandlers_EnrichToMarshaledJSON_AdvertisesXRelated(t *testing.T) {
	cases := []struct {
		name         string
		enrich       func(*http.Request) any
		relationType string
	}{
		{
			name: "AcquiredEntity",
			enrich: func(r *http.Request) any {
				h := &AcquiredEntityHandlers{hateoas: originLinks(t)}
				dto := &readmodels.AcquiredEntityDTO{ID: "ae1", Name: "Acme"}
				h.enrichWithLinks(r, dto)
				return dto
			},
			relationType: "origin-acquired-via",
		},
		{
			name: "Vendor",
			enrich: func(r *http.Request) any {
				h := &VendorHandlers{hateoas: originLinks(t)}
				dto := &readmodels.VendorDTO{ID: "v1", Name: "Vendor"}
				h.enrichWithLinks(r, dto)
				return dto
			},
			relationType: "origin-purchased-from",
		},
		{
			name: "InternalTeam",
			enrich: func(r *http.Request) any {
				h := &InternalTeamHandlers{hateoas: originLinks(t)}
				dto := &readmodels.InternalTeamDTO{ID: "t1", Name: "Team"}
				h.enrichWithLinks(r, dto)
				return dto
			},
			relationType: "origin-built-by",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entries := marshaledXRelated(t, tc.enrich(architectRequest()))
			require.Len(t, entries, 1)
			assert.Equal(t, tc.relationType, entries[0]["relationType"])
			assert.Equal(t, "/api/v1/components", entries[0]["href"])
			assert.Equal(t, []any{"POST"}, entries[0]["methods"])
			assert.Equal(t, "component", entries[0]["targetType"])
		})
	}
}
