package api

import (
	"net/http"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/types"
)

type SetTargetMaturityRequest struct {
	TargetMaturity int `json:"targetMaturity"`
}

type MaturityAnalysisResponse struct {
	Summary readmodels.MaturityAnalysisSummaryDTO     `json:"summary"`
	Data    []readmodels.MaturityAnalysisCandidateDTO `json:"data"`
	Links   types.Links                               `json:"_links,omitempty"`
}

// SetTargetMaturity godoc
// @Summary Set target maturity for enterprise capability
// @Description Sets the target maturity level (0-99) for an enterprise capability used in gap analysis
// @Tags enterprise-capabilities
// @Accept json
// @Produce json
// @Param id path string true "Enterprise capability ID"
// @Param maturity body SetTargetMaturityRequest true "Target maturity data"
// @Success 200 {object} easi_backend_internal_enterprisearchitecture_application_readmodels.EnterpriseCapabilityDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/target-maturity [put]
func (h *EnterpriseCapabilityHandlers) SetTargetMaturity(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	if h.getCapabilityOrNotFound(w, r, id) == nil {
		return
	}

	req, ok := sharedAPI.DecodeRequestOrFail[SetTargetMaturityRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.SetTargetMaturity{
		ID:             id,
		TargetMaturity: req.TargetMaturity,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		h.respondWithCapability(w, r, id, http.StatusOK)
	})
}

// GetMaturityAnalysisCandidates godoc
// @Summary Get enterprise capabilities with maturity gaps
// @Description Retrieves enterprise capabilities that have 2+ implementations with varying maturity levels
// @Tags enterprise-capabilities
// @Produce json
// @Param sortBy query string false "Sort order: 'gap' or 'implementations' (default: gap)"
// @Success 200 {object} object{summary=easi_backend_internal_enterprisearchitecture_application_readmodels.MaturityAnalysisSummaryDTO,data=[]easi_backend_internal_enterprisearchitecture_application_readmodels.MaturityAnalysisCandidateDTO}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/maturity-analysis [get]
func (h *EnterpriseCapabilityHandlers) GetMaturityAnalysisCandidates(w http.ResponseWriter, r *http.Request) {
	sortBy := r.URL.Query().Get("sortBy")

	candidates, summary, err := h.readModels.MaturityAnalysis.GetMaturityAnalysisCandidates(r.Context(), sortBy)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	for i := range candidates {
		candidates[i].Links = h.hateoas.MaturityAnalysisCandidateLinks(candidates[i].EnterpriseCapabilityID)
	}

	sharedAPI.RespondJSON(w, http.StatusOK, MaturityAnalysisResponse{
		Summary: summary,
		Data:    candidates,
		Links:   h.hateoas.MaturityAnalysisCollectionLinks(),
	})
}

// GetMaturityGapDetail godoc
// @Summary Get detailed maturity gap analysis
// @Description Retrieves detailed maturity gap analysis for a specific enterprise capability
// @Tags enterprise-capabilities
// @Produce json
// @Param id path string true "Enterprise capability ID"
// @Success 200 {object} easi_backend_internal_enterprisearchitecture_application_readmodels.MaturityGapDetailDTO
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/maturity-gap [get]
func (h *EnterpriseCapabilityHandlers) GetMaturityGapDetail(w http.ResponseWriter, r *http.Request) {
	enterpriseCapabilityID := sharedAPI.GetPathParam(r, "id")

	detail, err := h.readModels.MaturityAnalysis.GetMaturityGapDetail(r.Context(), enterpriseCapabilityID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	if detail == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Enterprise capability not found")
		return
	}

	detail.Links = h.hateoas.MaturityGapDetailLinks(enterpriseCapabilityID)

	sharedAPI.RespondJSON(w, http.StatusOK, detail)
}
