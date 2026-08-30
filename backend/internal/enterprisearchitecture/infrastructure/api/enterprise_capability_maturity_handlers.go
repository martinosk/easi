package api

import (
	"net/http"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	sharedAPI "easi/backend/internal/shared/api"
)

type SetTargetMaturityRequest struct {
	TargetMaturity int `json:"targetMaturity"`
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
