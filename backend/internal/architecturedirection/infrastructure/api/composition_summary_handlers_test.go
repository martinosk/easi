package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appservices "easi/backend/internal/architecturedirection/application/services"
	domainservices "easi/backend/internal/architecturedirection/domain/services"
	sharedAPI "easi/backend/internal/shared/api"
)

type fakeCompositionSummaries struct {
	summaries appservices.CompositionSummaries
	loadCalls int
}

func (f *fakeCompositionSummaries) SummariesForAll(_ context.Context) (appservices.CompositionSummaries, error) {
	f.loadCalls++
	return f.summaries, nil
}

func TestGetCompositionSummaries_OneRowPerEnterpriseCapability(t *testing.T) {
	source := &fakeCompositionSummaries{summaries: appservices.CompositionSummaries{
		EnterpriseCapabilityIDs: []string{"ec-1", "ec-2"},
		Counts:                  map[string]domainservices.CompositionCounts{"ec-1": {SourceCount: 2, IncludedCount: 5, CarvedOutCount: 1, DomainCount: 2}},
		Statuses:                map[string]string{"ec-1": "agreed"},
	}}
	handlers := NewCompositionSummaryHandlers(source, sharedAPI.NewHATEOASLinks("/api/v1"))

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
	assert.Equal(t, http.MethodGet, byID["ec-1"].Links["x-composition"].Method)
	assert.Equal(t, "/api/v1/enterprise-capabilities/ec-1/direction", byID["ec-1"].Links["x-direction"].Href)
	assert.Equal(t, http.MethodGet, byID["ec-1"].Links["x-direction"].Method)
	assert.Equal(t, "/api/v1/enterprise-capabilities/ec-2/direction", byID["ec-2"].Links["x-direction"].Href)
	assert.False(t, byID["ec-2"].HasActiveDirection)
	assert.Equal(t, 0, byID["ec-2"].IncludedCount)
	assert.Equal(t, "/api/v1/enterprise-capability-compositions", body.Links["self"].Href)
	assert.Equal(t, 1, source.loadCalls, "the handler must load composition inputs exactly once per request")
}
