package api

import (
	"context"
	"net/http"

	"easi/backend/internal/architecturedirection/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/types"
)

type MaturityAnalysisQueries interface {
	GetMaturityAnalysisCandidates(ctx context.Context, sortBy string) ([]readmodels.MaturityAnalysisCandidateDTO, readmodels.MaturityAnalysisSummaryDTO, error)
	GetMaturityGapDetail(ctx context.Context, enterpriseCapabilityID string) (*readmodels.MaturityGapDetailDTO, error)
}

type MaturityAnalysisResponse struct {
	Summary readmodels.MaturityAnalysisSummaryDTO     `json:"summary"`
	Data    []readmodels.MaturityAnalysisCandidateDTO `json:"data"`
	Links   types.Links                               `json:"_links,omitempty"`
}

type MaturityAnalysisHandlers struct {
	queries MaturityAnalysisQueries
	hateoas *sharedAPI.HATEOASLinks
}

func NewMaturityAnalysisHandlers(queries MaturityAnalysisQueries, hateoas *sharedAPI.HATEOASLinks) *MaturityAnalysisHandlers {
	return &MaturityAnalysisHandlers{queries: queries, hateoas: hateoas}
}

// GetMaturityAnalysisCandidates godoc
// @Summary Get enterprise capabilities with maturity gaps
// @Description Retrieves enterprise capabilities whose included domain capabilities fall below the target maturity, derived from the active direction's composition
// @Tags architecturedirection
// @Produce json
// @Param sortBy query string false "Sort order: 'gap' or 'implementations' (default: gap)"
// @Success 200 {object} MaturityAnalysisResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /enterprise-capabilities/maturity-analysis [get]
func (h *MaturityAnalysisHandlers) GetMaturityAnalysisCandidates(w http.ResponseWriter, r *http.Request) {
	candidates, summary, err := h.queries.GetMaturityAnalysisCandidates(r.Context(), r.URL.Query().Get("sortBy"))
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	for i := range candidates {
		candidates[i].Links = h.candidateLinks(candidates[i].EnterpriseCapabilityID)
	}
	sharedAPI.RespondJSON(w, http.StatusOK, MaturityAnalysisResponse{
		Summary: summary,
		Data:    candidates,
		Links:   types.Links{"self": h.hateoas.Get("/enterprise-capabilities/maturity-analysis")},
	})
}

// GetMaturityGapDetail godoc
// @Summary Get detailed maturity gap analysis
// @Description Retrieves each included domain capability's current maturity versus the enterprise capability's target maturity
// @Tags architecturedirection
// @Produce json
// @Param id path string true "Enterprise capability ID"
// @Success 200 {object} readmodels.MaturityGapDetailDTO
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /enterprise-capabilities/{id}/maturity-gap [get]
func (h *MaturityAnalysisHandlers) GetMaturityGapDetail(w http.ResponseWriter, r *http.Request) {
	enterpriseCapabilityID := sharedAPI.GetPathParam(r, "id")
	detail, err := h.queries.GetMaturityGapDetail(r.Context(), enterpriseCapabilityID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	if detail == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Enterprise capability not found")
		return
	}
	detail.Links = h.gapDetailLinks(enterpriseCapabilityID)
	sharedAPI.RespondJSON(w, http.StatusOK, detail)
}

func (h *MaturityAnalysisHandlers) candidateLinks(ecID string) types.Links {
	return types.Links{
		"self":           h.hateoas.Get("/enterprise-capabilities/" + ecID),
		"x-maturity-gap": h.hateoas.Get("/enterprise-capabilities/" + ecID + "/maturity-gap"),
	}
}

func (h *MaturityAnalysisHandlers) gapDetailLinks(ecID string) types.Links {
	base := "/enterprise-capabilities/" + ecID
	return types.Links{
		"self":                  h.hateoas.Get(base + "/maturity-gap"),
		"up":                    h.hateoas.Get(base),
		"x-set-target-maturity": h.hateoas.Put(base + "/target-maturity"),
	}
}
