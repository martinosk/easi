package api

import (
	"net/http"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type FitComparisonHandlers struct {
	fitComparisonRM *readmodels.ComponentFitComparisonReadModel
}

func NewFitComparisonHandlers(fitComparisonRM *readmodels.ComponentFitComparisonReadModel) *FitComparisonHandlers {
	return &FitComparisonHandlers{
		fitComparisonRM: fitComparisonRM,
	}
}

type FitComparisonResponse struct {
	PillarID        string `json:"pillarId"`
	PillarName      string `json:"pillarName"`
	FitScore        int    `json:"fitScore"`
	FitScoreLabel   string `json:"fitScoreLabel"`
	Importance      int    `json:"importance"`
	ImportanceLabel string `json:"importanceLabel"`
	Gap             int    `json:"gap"`
	Category        string `json:"category"`
	FitRationale    string `json:"fitRationale,omitempty"`
}

// GetFitComparisons godoc
// @Summary Get fit comparisons for a component in a capability context
// @Description Returns fit scores compared with importance ratings for a component realizing a capability in a business domain
// @Tags application-fit-scores
// @Accept json
// @Produce json
// @Param id path string true "Component ID"
// @Param capabilityId query string true "Capability ID"
// @Param businessDomainId query string true "Business Domain ID"
// @Success 200 {object} sharedAPI.CollectionResponse{data=[]FitComparisonResponse}
// @Failure 400 {object} sharedAPI.ErrorResponse "Missing required query parameters"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{id}/fit-comparisons [get]
func (h *FitComparisonHandlers) GetFitComparisons(w http.ResponseWriter, r *http.Request) {
	componentID := chi.URLParam(r, "id")
	capabilityID := r.URL.Query().Get("capabilityId")
	businessDomainID := r.URL.Query().Get("businessDomainId")

	if _, err := uuid.Parse(componentID); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid component ID format")
		return
	}

	if capabilityID == "" {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "capabilityId query parameter is required")
		return
	}

	if businessDomainID == "" {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "businessDomainId query parameter is required")
		return
	}

	comparisons, err := h.fitComparisonRM.GetByComponentAndCapability(r.Context(), componentID, capabilityID, businessDomainID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve fit comparisons")
		return
	}

	data := make([]FitComparisonResponse, len(comparisons))
	for i, dto := range comparisons {
		data[i] = FitComparisonResponse{
			PillarID:        dto.PillarID,
			PillarName:      dto.PillarName,
			FitScore:        dto.FitScore,
			FitScoreLabel:   dto.FitScoreLabel,
			Importance:      dto.Importance,
			ImportanceLabel: dto.ImportanceLabel,
			Gap:             dto.Gap,
			Category:        dto.Category,
			FitRationale:    dto.FitRationale,
		}
	}

	links := sharedAPI.Links{
		"self": sharedAPI.NewLink("/api/v1/components/"+componentID+"/fit-comparisons?capabilityId="+capabilityID+"&businessDomainId="+businessDomainID, "GET"),
		"up":   sharedAPI.NewLink("/api/v1/components/"+componentID, "GET"),
	}

	sharedAPI.RespondCollection(w, http.StatusOK, data, links)
}
