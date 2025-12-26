package api

import (
	"context"
	"log"
	"net/http"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/metamodel"
	sharedAPI "easi/backend/internal/shared/api"
)

type MaturityLevelDTO struct {
	Value    string `json:"value" example:"Genesis"`
	MinValue int    `json:"minValue" example:"0"`
	MaxValue int    `json:"maxValue" example:"24"`
	Order    int    `json:"order" example:"1"`
}

type StatusDTO struct {
	Value       string `json:"value" example:"Active"`
	DisplayName string `json:"displayName" example:"Active"`
	SortOrder   int    `json:"sortOrder" example:"1"`
}

type OwnershipModelDTO struct {
	Value       string `json:"value" example:"TribeOwned"`
	DisplayName string `json:"displayName" example:"Tribe Owned"`
}

type StrategyPillarDTO struct {
	Value       string `json:"value" example:"AlwaysOn"`
	DisplayName string `json:"displayName" example:"Always On"`
}

type MetadataIndexDTO struct {
	Links map[string]string `json:"_links"`
}

type MaturityScaleConfigProvider interface {
	GetMaturityScaleConfig(ctx context.Context) (*metamodel.MaturityScaleConfigDTO, error)
}

type MaturityLevelHandlers struct {
	gateway MaturityScaleConfigProvider
}

func NewMaturityLevelHandlers(gateway MaturityScaleConfigProvider) *MaturityLevelHandlers {
	return &MaturityLevelHandlers{
		gateway: gateway,
	}
}

func setCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "public, max-age=86400, stale-while-revalidate=3600")
	w.Header().Set("ETag", `"v1.0.0"`)
}

// GetMetadataIndex godoc
// @Summary List available metadata types
// @Description Returns HATEOAS links to all capability metadata endpoints
// @Tags capabilities
// @Produce json
// @Success 200 {object} MetadataIndexDTO
// @Router /capabilities/metadata [get]
func (h *MaturityLevelHandlers) GetMetadataIndex(w http.ResponseWriter, r *http.Request) {
	setCacheHeaders(w)

	response := MetadataIndexDTO{
		Links: map[string]string{
			"self":            "/api/v1/capabilities/metadata",
			"maturityLevels":  "/api/v1/capabilities/metadata/maturity-levels",
			"statuses":        "/api/v1/capabilities/metadata/statuses",
			"ownershipModels": "/api/v1/capabilities/metadata/ownership-models",
			"strategyPillars": "/api/v1/capabilities/metadata/strategy-pillars",
		},
	}

	sharedAPI.RespondJSON(w, http.StatusOK, response)
}

// GetMaturityLevels godoc
// @Summary Get valid maturity levels
// @Description Returns the list of valid maturity level options for capabilities based on Wardley mapping evolution stages (Genesis, Custom Build, Product, Commodity)
// @Tags capabilities
// @Produce json
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]MaturityLevelDTO}
// @Router /capabilities/metadata/maturity-levels [get]
func (h *MaturityLevelHandlers) GetMaturityLevels(w http.ResponseWriter, r *http.Request) {
	setCacheHeaders(w)

	config := h.getMaturityScaleConfigWithFallback(r.Context())

	levels := make([]MaturityLevelDTO, 0, len(config.Sections))
	for _, section := range config.Sections {
		levels = append(levels, MaturityLevelDTO{
			Value:    section.Name,
			MinValue: section.MinValue,
			MaxValue: section.MaxValue,
			Order:    section.Order,
		})
	}

	links := map[string]string{
		"self":        "/api/v1/capabilities/metadata/maturity-levels",
		"configureAt": "/api/v1/meta-model/maturity-scale",
	}

	sharedAPI.RespondCollection(w, http.StatusOK, levels, links)
}

func (h *MaturityLevelHandlers) getMaturityScaleConfigWithFallback(ctx context.Context) *metamodel.MaturityScaleConfigDTO {
	if h.gateway == nil {
		return metamodel.DefaultMaturityScaleConfig()
	}

	config, err := h.gateway.GetMaturityScaleConfig(ctx)
	if err != nil {
		log.Printf("Failed to fetch maturity scale config, using defaults: %v", err)
		return metamodel.DefaultMaturityScaleConfig()
	}

	if config == nil {
		return metamodel.DefaultMaturityScaleConfig()
	}

	return config
}

// GetStatuses godoc
// @Summary Get valid capability statuses
// @Description Returns lifecycle statuses for capabilities (Active, Planned, Deprecated)
// @Tags capabilities
// @Produce json
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]StatusDTO}
// @Router /capabilities/metadata/statuses [get]
func (h *MaturityLevelHandlers) GetStatuses(w http.ResponseWriter, r *http.Request) {
	setCacheHeaders(w)

	statuses := []StatusDTO{
		{Value: string(valueobjects.StatusActive), DisplayName: "Active", SortOrder: 1},
		{Value: string(valueobjects.StatusPlanned), DisplayName: "Planned", SortOrder: 2},
		{Value: string(valueobjects.StatusDeprecated), DisplayName: "Deprecated", SortOrder: 3},
	}

	links := map[string]string{
		"self": "/api/v1/capabilities/metadata/statuses",
	}

	sharedAPI.RespondCollection(w, http.StatusOK, statuses, links)
}

// GetOwnershipModels godoc
// @Summary Get valid ownership models
// @Description Returns ownership classification options (TribeOwned, TeamOwned, Shared, EnterpriseService)
// @Tags capabilities
// @Produce json
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]OwnershipModelDTO}
// @Router /capabilities/metadata/ownership-models [get]
func (h *MaturityLevelHandlers) GetOwnershipModels(w http.ResponseWriter, r *http.Request) {
	setCacheHeaders(w)

	models := []OwnershipModelDTO{
		{Value: string(valueobjects.OwnershipTribeOwned), DisplayName: "Tribe Owned"},
		{Value: string(valueobjects.OwnershipTeamOwned), DisplayName: "Team Owned"},
		{Value: string(valueobjects.OwnershipShared), DisplayName: "Shared"},
		{Value: string(valueobjects.OwnershipEnterpriseService), DisplayName: "Enterprise Service"},
	}

	links := map[string]string{
		"self": "/api/v1/capabilities/metadata/ownership-models",
	}

	sharedAPI.RespondCollection(w, http.StatusOK, models, links)
}

// GetStrategyPillars godoc
// @Summary Get valid strategy pillars
// @Description Returns strategic alignment categories (AlwaysOn, Grow, Transform)
// @Tags capabilities
// @Produce json
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]StrategyPillarDTO}
// @Router /capabilities/metadata/strategy-pillars [get]
func (h *MaturityLevelHandlers) GetStrategyPillars(w http.ResponseWriter, r *http.Request) {
	setCacheHeaders(w)

	pillars := []StrategyPillarDTO{
		{Value: string(valueobjects.PillarAlwaysOn), DisplayName: "Always On"},
		{Value: string(valueobjects.PillarGrow), DisplayName: "Grow"},
		{Value: string(valueobjects.PillarTransform), DisplayName: "Transform"},
	}

	links := map[string]string{
		"self": "/api/v1/capabilities/metadata/strategy-pillars",
	}

	sharedAPI.RespondCollection(w, http.StatusOK, pillars, links)
}
