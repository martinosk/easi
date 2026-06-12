package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	appservices "easi/backend/internal/enterprisearchitecture/application/services"
	domainservices "easi/backend/internal/enterprisearchitecture/domain/services"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeCompositionQueries struct {
	composition appservices.CompositionResult
	candidates  appservices.SourceCandidatesResult
	lastQuery   appservices.SourceCandidatesQuery
}

func (f *fakeCompositionQueries) CompositionForEC(_ context.Context, _ string) (appservices.CompositionResult, error) {
	return f.composition, nil
}

func (f *fakeCompositionQueries) SourceCandidates(_ context.Context, q appservices.SourceCandidatesQuery) (appservices.SourceCandidatesResult, error) {
	f.lastQuery = q
	return f.candidates, nil
}

type fakeECQueries struct {
	byID map[string]*readmodels.EnterpriseCapabilityDTO
}

func (f *fakeECQueries) GetAll(_ context.Context) ([]readmodels.EnterpriseCapabilityDTO, error) {
	return nil, nil
}

func (f *fakeECQueries) GetByID(_ context.Context, id string) (*readmodels.EnterpriseCapabilityDTO, error) {
	return f.byID[id], nil
}

func architectActorEA() sharedctx.Actor {
	return sharedctx.NewActor("u1", "user@example.com", sharedctx.RoleArchitect)
}

func compositionRouter(h *CompositionHandlers) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/enterprise-capabilities/{id}/composition", h.GetComposition)
	r.Get("/capabilities/source-candidates", h.GetSourceCandidates)
	return r
}

func performGet(r http.Handler, target string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, target, nil)
	req = req.WithContext(sharedctx.WithActor(req.Context(), architectActorEA()))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func newCompositionHandlers(queries *fakeCompositionQueries, ecs *fakeECQueries) *CompositionHandlers {
	return NewCompositionHandlers(queries, ecs, NewEnterpriseArchLinks(sharedAPI.NewHATEOASLinks("")))
}

func domainNode(id, name, domainID, domainName string) domainservices.CapabilityNode {
	return domainservices.CapabilityNode{
		ID: id, Name: name, Level: "L2",
		BusinessDomainID: domainID, BusinessDomainName: domainName,
	}
}

func resolvedItem(node domainservices.CapabilityNode, role string, carvedOutBy *domainservices.CarvedOutBy) domainservices.ResolvedCapability {
	return domainservices.ResolvedCapability{
		Node:        node,
		Role:        domainservices.CompositionRole(role),
		CarvedOutBy: carvedOutBy,
	}
}

func TestGetComposition_UnknownEC_404(t *testing.T) {
	h := newCompositionHandlers(&fakeCompositionQueries{}, &fakeECQueries{})

	rec := performGet(compositionRouter(h), "/enterprise-capabilities/ec-missing/composition")

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetComposition_GroupsByDomainWithMetaAndLinks(t *testing.T) {
	ecs := &fakeECQueries{byID: map[string]*readmodels.EnterpriseCapabilityDTO{
		"ec-1": {ID: "ec-1", Name: "Customer Identity", Active: true},
	}}
	queries := &fakeCompositionQueries{composition: appservices.CompositionResult{
		HasActiveDirection: true,
		DirectionStatus:    "draft",
		Resolved: []domainservices.ResolvedCapability{
			resolvedItem(domainNode("cap-001", "Customer Account Creation", "dom-001", "Customer"), "source", nil),
			resolvedItem(domainNode("cap-002", "Customer Fraud Prevention", "dom-001", "Customer"), "carved-out", &domainservices.CarvedOutBy{EnterpriseCapabilityID: "ec-pay", EnterpriseCapabilityName: "Take Payment"}),
			resolvedItem(domainNode("cap-003", "Order Handling", "", ""), "implicit", nil),
		},
		Counts: domainservices.CompositionCounts{SourceCount: 1, IncludedCount: 2, CarvedOutCount: 1, DomainCount: 1},
	}}
	h := newCompositionHandlers(queries, ecs)

	rec := performGet(compositionRouter(h), "/enterprise-capabilities/ec-1/composition")

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Data []struct {
			BusinessDomainID   *string `json:"businessDomainId"`
			BusinessDomainName *string `json:"businessDomainName"`
			Items              []struct {
				CapabilityID string          `json:"capabilityId"`
				Role         string          `json:"role"`
				CarvedOutBy  *CarvedOutByDTO `json:"carvedOutBy"`
				Links        sharedAPI.Links `json:"_links"`
			} `json:"items"`
		} `json:"data"`
		Meta  map[string]int  `json:"meta"`
		Links sharedAPI.Links `json:"_links"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	require.Len(t, body.Data, 2, "items are grouped by business domain")
	require.NotNil(t, body.Data[0].BusinessDomainID)
	assert.Equal(t, "dom-001", *body.Data[0].BusinessDomainID)
	require.Len(t, body.Data[0].Items, 2)
	assert.Nil(t, body.Data[1].BusinessDomainID, "unassigned capabilities group under a null domain")

	source := body.Data[0].Items[0]
	assert.Equal(t, "cap-001", source.CapabilityID)
	require.Contains(t, source.Links, "x-exclude", "explicit sources of a non-agreed direction expose the exclude affordance")
	assert.Equal(t, "DELETE", source.Links["x-exclude"].Method)
	assert.Equal(t, "/api/v1/enterprise-capabilities/ec-1/direction/sources/cap-001", source.Links["x-exclude"].Href)
	assert.Equal(t, "/api/v1/capabilities/cap-001", source.Links["self"].Href)

	carved := body.Data[0].Items[1]
	require.NotNil(t, carved.CarvedOutBy)
	assert.Equal(t, "ec-pay", carved.CarvedOutBy.EnterpriseCapabilityID)
	require.Contains(t, carved.Links, "x-owning-ec")
	assert.Equal(t, "/api/v1/enterprise-capabilities/ec-pay", carved.Links["x-owning-ec"].Href)
	assert.NotContains(t, carved.Links, "x-exclude")

	assert.Equal(t, map[string]int{"sourceCount": 1, "includedCount": 2, "carvedOutCount": 1, "domainCount": 1}, body.Meta)
	assert.Contains(t, body.Links, "self")
	assert.Contains(t, body.Links, "up")
	assert.Contains(t, body.Links, "x-direction")
	assert.NotContains(t, body.Links, "x-capture-direction", "an active direction exists")
}

func TestGetComposition_NonDraftDirectionHidesExclude(t *testing.T) {
	cases := []struct {
		status string
		reason string
	}{
		{"proposed", "source set frozen once proposed"},
		{"agreed", "agreed directions are immutable (R5)"},
	}
	for _, tc := range cases {
		t.Run(tc.status, func(t *testing.T) {
			ecs := &fakeECQueries{byID: map[string]*readmodels.EnterpriseCapabilityDTO{
				"ec-1": {ID: "ec-1", Name: "Customer Identity", Active: true},
			}}
			queries := &fakeCompositionQueries{composition: appservices.CompositionResult{
				HasActiveDirection: true,
				DirectionStatus:    tc.status,
				Resolved: []domainservices.ResolvedCapability{
					resolvedItem(domainNode("cap-001", "Customer Account Creation", "dom-001", "Customer"), "source", nil),
				},
				Counts: domainservices.CompositionCounts{SourceCount: 1, IncludedCount: 1},
			}}
			h := newCompositionHandlers(queries, ecs)

			rec := performGet(compositionRouter(h), "/enterprise-capabilities/ec-1/composition")

			require.Equal(t, http.StatusOK, rec.Code)
			assert.NotContains(t, rec.Body.String(), "x-exclude", tc.reason)
		})
	}
}

func TestGetComposition_NoActiveDirection_EmptyDataAndCaptureLink(t *testing.T) {
	ecs := &fakeECQueries{byID: map[string]*readmodels.EnterpriseCapabilityDTO{
		"ec-1": {ID: "ec-1", Name: "Customer Identity", Active: true},
	}}
	h := newCompositionHandlers(&fakeCompositionQueries{}, ecs)

	rec := performGet(compositionRouter(h), "/enterprise-capabilities/ec-1/composition")

	require.Equal(t, http.StatusOK, rec.Code)
	var body struct {
		Data  []any           `json:"data"`
		Meta  map[string]int  `json:"meta"`
		Links sharedAPI.Links `json:"_links"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.NotNil(t, body.Data)
	assert.Empty(t, body.Data)
	assert.Equal(t, 0, body.Meta["includedCount"])
	assert.Contains(t, body.Links, "x-capture-direction")
}

func TestGetSourceCandidates_MissingEcId_400(t *testing.T) {
	h := newCompositionHandlers(&fakeCompositionQueries{}, &fakeECQueries{})

	rec := performGet(compositionRouter(h), "/capabilities/source-candidates?q=customer")
	require.Equal(t, http.StatusBadRequest, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "BadRequest", body["error"])
	assert.Equal(t, "ecId is required", body["message"])
}

func TestGetSourceCandidates_EmptyQ_ReturnsAllCandidates(t *testing.T) {
	h := newCompositionHandlers(&fakeCompositionQueries{candidates: appservices.SourceCandidatesResult{
		Candidates: []appservices.SourceCandidate{{Node: domainservices.CapabilityNode{ID: "cap-1", Name: "Billing"}}},
		HasMore:    false,
	}}, &fakeECQueries{})

	rec := performGet(compositionRouter(h), "/capabilities/source-candidates?ecId=ec-1")
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestGetSourceCandidates_ReturnsCandidatesWithEligibilityLinks(t *testing.T) {
	reason := "Already an explicit source of an active direction on 'Take Payment'"
	queries := &fakeCompositionQueries{candidates: appservices.SourceCandidatesResult{
		Candidates: []appservices.SourceCandidate{
			{
				Node:     domainservices.CapabilityNode{ID: "cap-1", Name: "Customer Account Creation", Level: "L2", ParentID: "cap-parent", BusinessDomainID: "dom-001", BusinessDomainName: "Customer"},
				Eligible: true,
			},
			{
				Node:                domainservices.CapabilityNode{ID: "cap-ineligible", Name: "Customer Fraud Prevention", Level: "L3"},
				Eligible:            false,
				IneligibilityReason: &reason,
				ConflictingEnterpriseCapability: &domainservices.CarvedOutBy{
					EnterpriseCapabilityID: "ec-pay", EnterpriseCapabilityName: "Take Payment",
				},
			},
		},
		HasMore: true,
	}}
	h := newCompositionHandlers(queries, &fakeECQueries{})

	rec := performGet(compositionRouter(h), "/capabilities/source-candidates?q=customer&ecId=ec-1&domainId=dom-001&limit=5")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, appservices.SourceCandidatesQuery{
		EnterpriseCapabilityID: "ec-1", Search: "customer", BusinessDomainID: "dom-001", Limit: 5,
	}, queries.lastQuery)

	var body struct {
		Data []struct {
			CapabilityID                    string            `json:"capabilityId"`
			ParentID                        *string           `json:"parentId"`
			BusinessDomainID                *string           `json:"businessDomainId"`
			Eligible                        bool              `json:"eligible"`
			IneligibilityReason             *string           `json:"ineligibilityReason"`
			ConflictingEnterpriseCapability *ConflictingECDTO `json:"conflictingEnterpriseCapability"`
			Links                           sharedAPI.Links   `json:"_links"`
		} `json:"data"`
		Pagination struct {
			HasMore bool   `json:"hasMore"`
			Limit   int    `json:"limit"`
			Cursor  string `json:"cursor"`
		} `json:"pagination"`
		Links sharedAPI.Links `json:"_links"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	require.Len(t, body.Data, 2)
	eligible := body.Data[0]
	require.NotNil(t, eligible.ParentID)
	assert.Equal(t, "cap-parent", *eligible.ParentID)
	assert.True(t, eligible.Eligible)
	assert.Nil(t, eligible.IneligibilityReason)
	assert.Nil(t, eligible.ConflictingEnterpriseCapability)
	assert.NotContains(t, eligible.Links, "x-conflicting-ec")
	assert.Equal(t, "/api/v1/capabilities/cap-1", eligible.Links["self"].Href)

	conflicted := body.Data[1]
	assert.False(t, conflicted.Eligible)
	require.NotNil(t, conflicted.ConflictingEnterpriseCapability)
	assert.Equal(t, "ec-pay", conflicted.ConflictingEnterpriseCapability.ID)
	require.Contains(t, conflicted.Links, "x-conflicting-ec")
	assert.Equal(t, "/api/v1/enterprise-capabilities/ec-pay", conflicted.Links["x-conflicting-ec"].Href)

	assert.True(t, body.Pagination.HasMore)
	assert.Equal(t, 5, body.Pagination.Limit)
	assert.Equal(t, "", body.Pagination.Cursor)
	assert.Contains(t, body.Links, "self")
}

func TestGetSourceCandidates_DefaultLimitIs20(t *testing.T) {
	queries := &fakeCompositionQueries{}
	h := newCompositionHandlers(queries, &fakeECQueries{})

	rec := performGet(compositionRouter(h), "/capabilities/source-candidates?q=x&ecId=ec-1")

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 20, queries.lastQuery.Limit)
}
