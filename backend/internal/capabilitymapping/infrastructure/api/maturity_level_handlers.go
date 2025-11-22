package api

import (
	"net/http"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
)

type MaturityLevelDTO struct {
	Value        string `json:"value" example:"Genesis"`
	NumericValue int    `json:"numericValue" example:"1"`
}

type MaturityLevelHandlers struct{}

func NewMaturityLevelHandlers() *MaturityLevelHandlers {
	return &MaturityLevelHandlers{}
}

// GetMaturityLevels godoc
// @Summary Get valid maturity levels
// @Description Returns the list of valid maturity level options for capabilities based on Wardley mapping evolution stages (Genesis, Custom Build, Product, Commodity)
// @Tags capabilities
// @Produce json
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]MaturityLevelDTO}
// @Router /capabilities/metadata/maturity-levels [get]
func (h *MaturityLevelHandlers) GetMaturityLevels(w http.ResponseWriter, r *http.Request) {
	levels := []MaturityLevelDTO{
		{Value: string(valueobjects.MaturityGenesis), NumericValue: valueobjects.MaturityGenesis.NumericValue()},
		{Value: string(valueobjects.MaturityCustomBuild), NumericValue: valueobjects.MaturityCustomBuild.NumericValue()},
		{Value: string(valueobjects.MaturityProduct), NumericValue: valueobjects.MaturityProduct.NumericValue()},
		{Value: string(valueobjects.MaturityCommodity), NumericValue: valueobjects.MaturityCommodity.NumericValue()},
	}

	links := map[string]string{
		"self": "/api/v1/capabilities/metadata/maturity-levels",
	}

	sharedAPI.RespondCollection(w, http.StatusOK, levels, links)
}
