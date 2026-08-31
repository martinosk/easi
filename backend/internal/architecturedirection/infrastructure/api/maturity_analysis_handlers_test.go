package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/architecturedirection/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
)

type fakeMaturityAnalysis struct {
	candidates []readmodels.MaturityAnalysisCandidateDTO
	summary    readmodels.MaturityAnalysisSummaryDTO
	detail     *readmodels.MaturityGapDetailDTO
	sortBy     string
}

func (f *fakeMaturityAnalysis) GetMaturityAnalysisCandidates(_ context.Context, sortBy string) ([]readmodels.MaturityAnalysisCandidateDTO, readmodels.MaturityAnalysisSummaryDTO, error) {
	f.sortBy = sortBy
	return f.candidates, f.summary, nil
}

func (f *fakeMaturityAnalysis) GetMaturityGapDetail(_ context.Context, _ string) (*readmodels.MaturityGapDetailDTO, error) {
	return f.detail, nil
}

func maturityRouter(h *MaturityAnalysisHandlers) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/enterprise-capabilities/maturity-analysis", h.GetMaturityAnalysisCandidates)
	r.Get("/enterprise-capabilities/{id}/maturity-gap", h.GetMaturityGapDetail)
	return r
}

func TestGetMaturityAnalysisCandidates_LinksEachCandidateAndCollection(t *testing.T) {
	analysis := &fakeMaturityAnalysis{candidates: []readmodels.MaturityAnalysisCandidateDTO{{EnterpriseCapabilityID: "ec-1"}}}
	handlers := NewMaturityAnalysisHandlers(analysis, sharedAPI.NewHATEOASLinks("/api/v1"))

	rec := httptest.NewRecorder()
	maturityRouter(handlers).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/enterprise-capabilities/maturity-analysis?sortBy=implementations", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "implementations", analysis.sortBy)
	var body MaturityAnalysisResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Data, 1)
	assert.Equal(t, "/api/v1/enterprise-capabilities/ec-1/maturity-gap", body.Data[0].Links["x-maturity-gap"].Href)
	assert.Equal(t, "/api/v1/enterprise-capabilities/maturity-analysis", body.Links["self"].Href)
}

func TestGetMaturityGapDetail_NotFoundWhenAbsent(t *testing.T) {
	handlers := NewMaturityAnalysisHandlers(&fakeMaturityAnalysis{}, sharedAPI.NewHATEOASLinks("/api/v1"))

	rec := httptest.NewRecorder()
	maturityRouter(handlers).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/enterprise-capabilities/ec-9/maturity-gap", nil))

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetMaturityGapDetail_LinksDetailToCapabilityAndTargetMaturity(t *testing.T) {
	analysis := &fakeMaturityAnalysis{detail: &readmodels.MaturityGapDetailDTO{}}
	handlers := NewMaturityAnalysisHandlers(analysis, sharedAPI.NewHATEOASLinks("/api/v1"))

	rec := httptest.NewRecorder()
	maturityRouter(handlers).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/enterprise-capabilities/ec-1/maturity-gap", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	var body readmodels.MaturityGapDetailDTO
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "/api/v1/enterprise-capabilities/ec-1/maturity-gap", body.Links["self"].Href)
	assert.Equal(t, "/api/v1/enterprise-capabilities/ec-1", body.Links["up"].Href)
	assert.Equal(t, "PUT", body.Links["x-set-target-maturity"].Method)
}
