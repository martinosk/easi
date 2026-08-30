package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/architecturedirection/application/readmodels"
	domainservices "easi/backend/internal/architecturedirection/domain/services"
	sharedAPI "easi/backend/internal/shared/api"
)

type fakeCompositionCounts struct {
	counts map[string]domainservices.CompositionCounts
	status map[string]string
}

func (f *fakeCompositionCounts) CountsForAll(_ context.Context) (map[string]domainservices.CompositionCounts, error) {
	return f.counts, nil
}

func (f *fakeCompositionCounts) DirectionStatusByEC(_ context.Context) (map[string]string, error) {
	return f.status, nil
}

type fakeECNames struct {
	names map[string]string
}

func (f *fakeECNames) ActiveEnterpriseCapabilityNames(_ context.Context) (map[string]string, error) {
	return f.names, nil
}

func (f *fakeECNames) GetByID(_ context.Context, id string) (*readmodels.EnterpriseCapabilityCacheDTO, error) {
	if _, ok := f.names[id]; !ok {
		return nil, nil
	}
	return &readmodels.EnterpriseCapabilityCacheDTO{ID: id, Name: f.names[id], Active: true}, nil
}

func TestGetCompositionSummaries_OneRowPerEnterpriseCapability(t *testing.T) {
	counts := &fakeCompositionCounts{
		counts: map[string]domainservices.CompositionCounts{"ec-1": {SourceCount: 2, IncludedCount: 5, CarvedOutCount: 1, DomainCount: 2}},
		status: map[string]string{"ec-1": "agreed"},
	}
	handlers := NewCompositionSummaryHandlers(counts, &fakeECNames{names: map[string]string{"ec-1": "Payments", "ec-2": "Identity"}}, sharedAPI.NewHATEOASLinks("/api/v1"))

	rec := httptest.NewRecorder()
	handlers.GetCompositionSummaries(rec, httptest.NewRequest(http.MethodGet, "/enterprise-capability-compositions", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	var body CompositionSummariesResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Data, 2)
	byID := map[string]CompositionSummaryDTO{}
	for _, row := range body.Data {
		byID[row.EnterpriseCapabilityID] = row
	}
	assert.Equal(t, CompositionSummaryDTO{
		EnterpriseCapabilityID: "ec-1", SourceCount: 2, IncludedCount: 5, CarvedOutCount: 1, DomainCount: 2,
		HasActiveDirection: true, DirectionStatus: "agreed", Links: byID["ec-1"].Links,
	}, byID["ec-1"])
	assert.Equal(t, "/api/v1/enterprise-capabilities/ec-1/composition", byID["ec-1"].Links["x-composition"].Href)
	assert.False(t, byID["ec-2"].HasActiveDirection)
	assert.Equal(t, 0, byID["ec-2"].IncludedCount)
	assert.Equal(t, "/api/v1/enterprise-capability-compositions", body.Links["self"].Href)
}
