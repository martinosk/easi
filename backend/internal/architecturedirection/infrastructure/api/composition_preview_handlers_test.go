package api

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	sharedAPI "easi/backend/internal/shared/api"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePreviewProvider struct {
	data       *CompositionPreviewData
	requestEC  string
	requestIDs []string
}

func (f *fakePreviewProvider) PreviewComposition(_ context.Context, ecID string, sourceIDs []string) (*CompositionPreviewData, error) {
	f.requestEC = ecID
	f.requestIDs = sourceIDs
	return f.data, nil
}

func previewRouter(provider CompositionPreviewProvider) *chi.Mux {
	h := NewCompositionPreviewHandlers(provider, sharedAPI.NewHATEOASLinks(""))
	r := chi.NewRouter()
	r.Post("/enterprise-capabilities/{id}/direction/composition-preview", h.PreviewComposition)
	return r
}

func TestPreviewComposition_ReturnsResolvedItemsEligibilityAndMeta(t *testing.T) {
	ecID := uuid.New().String()
	domainID := "dom-001"
	domainName := "Customer"
	reason := "Already an explicit source of an active direction on 'Take Payment'"
	provider := &fakePreviewProvider{data: &CompositionPreviewData{
		IncludedCapabilities: []PreviewIncludedCapabilityDTO{
			{CapabilityID: "cap-001", Name: "Customer Account Creation", Level: "L2", BusinessDomainID: &domainID, BusinessDomainName: &domainName, Role: "source"},
			{CapabilityID: "cap-002", Name: "Customer Fraud Prevention", Level: "L3", Role: "carved-out", CarvedOutBy: &CarvedOutByDTO{EnterpriseCapabilityID: "ec-pay", EnterpriseCapabilityName: "Take Payment"}},
		},
		SourceEligibility: []SourceEligibilityDTO{
			{CapabilityID: "cap-001", Eligible: true},
			{CapabilityID: "cap-002", Eligible: false, IneligibilityReason: &reason, ConflictingEnterpriseCapability: &ConflictingECDTO{ID: "ec-pay", Name: "Take Payment"}},
		},
		Meta: CompositionPreviewMetaDTO{SourceCount: 2, IncludedCount: 1, CarvedOutCount: 1},
	}}

	rec := postJSON(t, previewRouter(provider),
		"/enterprise-capabilities/"+ecID+"/direction/composition-preview",
		map[string]any{"sourceCapabilityIds": []string{"cap-001", "cap-002"}})

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, ecID, provider.requestEC)
	assert.Equal(t, []string{"cap-001", "cap-002"}, provider.requestIDs)

	var body map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	var included []map[string]any
	require.NoError(t, json.Unmarshal(body["includedCapabilities"], &included))
	require.Len(t, included, 2)
	assert.Equal(t, "source", included[0]["role"])
	assert.Nil(t, included[0]["carvedOutBy"], "carvedOutBy must be an explicit null for non-carved items")
	carvedOutBy, ok := included[1]["carvedOutBy"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "ec-pay", carvedOutBy["enterpriseCapabilityId"])

	var eligibility []map[string]any
	require.NoError(t, json.Unmarshal(body["sourceEligibility"], &eligibility))
	require.Len(t, eligibility, 2)
	assert.Equal(t, true, eligibility[0]["eligible"])
	assert.Nil(t, eligibility[0]["ineligibilityReason"])

	var meta map[string]int
	require.NoError(t, json.Unmarshal(body["meta"], &meta))
	assert.Equal(t, map[string]int{"sourceCount": 2, "includedCount": 1, "carvedOutCount": 1}, meta)

	var links sharedAPI.Links
	require.NoError(t, json.Unmarshal(body["_links"], &links))
	require.Contains(t, links, "self")
	assert.Equal(t, "POST", links["self"].Method)
}

func TestPreviewComposition_UnknownEC_404(t *testing.T) {
	provider := &fakePreviewProvider{data: nil}

	rec := postJSON(t, previewRouter(provider),
		"/enterprise-capabilities/"+uuid.New().String()+"/direction/composition-preview",
		map[string]any{"sourceCapabilityIds": []string{}})

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
