package api

import (
	"net/http"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/types"

	"github.com/go-chi/chi/v5"
)

type StrategyImportanceHandlers struct {
	commandBus   cqrs.CommandBus
	importanceRM *readmodels.StrategyImportanceReadModel
	hateoas      *sharedAPI.HATEOASLinks
}

func NewStrategyImportanceHandlers(
	commandBus cqrs.CommandBus,
	importanceRM *readmodels.StrategyImportanceReadModel,
	hateoas *sharedAPI.HATEOASLinks,
) *StrategyImportanceHandlers {
	return &StrategyImportanceHandlers{
		commandBus:   commandBus,
		importanceRM: importanceRM,
		hateoas:      hateoas,
	}
}

type SetStrategyImportanceRequest struct {
	PillarID   string `json:"pillarId"`
	Importance int    `json:"importance"`
	Rationale  string `json:"rationale,omitempty"`
}

type UpdateStrategyImportanceRequest struct {
	Importance int    `json:"importance"`
	Rationale  string `json:"rationale,omitempty"`
}

type StrategyImportanceResponse struct {
	ID                 string      `json:"id"`
	BusinessDomainID   string      `json:"businessDomainId"`
	BusinessDomainName string      `json:"businessDomainName"`
	CapabilityID       string      `json:"capabilityId"`
	CapabilityName     string      `json:"capabilityName"`
	PillarID           string      `json:"pillarId"`
	PillarName         string      `json:"pillarName"`
	Importance         int         `json:"importance"`
	ImportanceLabel    string      `json:"importanceLabel"`
	Rationale          string      `json:"rationale,omitempty"`
	Links              types.Links `json:"_links,omitempty"`
}

// GetImportanceByDomainAndCapability godoc
// @Summary Get strategic importance ratings for a capability in a domain
// @Description Retrieves all strategic importance ratings for a specific capability within a business domain
// @Tags strategy-importance
// @Accept json
// @Produce json
// @Param domainId path string true "Business Domain ID"
// @Param capabilityId path string true "Capability ID"
// @Success 200 {object} sharedAPI.CollectionResponse{data=[]StrategyImportanceResponse}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /business-domains/{domainId}/capabilities/{capabilityId}/importance [get]
func (h *StrategyImportanceHandlers) GetImportanceByDomainAndCapability(w http.ResponseWriter, r *http.Request) {
	domainID := chi.URLParam(r, "id")
	capabilityID := chi.URLParam(r, "capabilityId")

	actor, _ := sharedctx.GetActor(r.Context())
	collectionLinks := h.hateoas.StrategyImportanceCollectionLinksForActor(domainID, capabilityID, actor)

	h.respondWithImportanceCollection(w, r, importanceCollectionQuery{
		fetcher:  func() ([]readmodels.StrategyImportanceDTO, error) { return h.importanceRM.GetByDomainAndCapability(r.Context(), domainID, capabilityID) },
		domainID: domainID,
		links:    collectionLinks,
	})
}

// SetImportance godoc
// @Summary Set strategic importance for a capability in a domain
// @Description Creates a new strategic importance rating for a capability-pillar combination
// @Tags strategy-importance
// @Accept json
// @Produce json
// @Param domainId path string true "Business Domain ID"
// @Param capabilityId path string true "Capability ID"
// @Param importance body SetStrategyImportanceRequest true "Importance rating"
// @Success 201 {object} StrategyImportanceResponse
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /business-domains/{domainId}/capabilities/{capabilityId}/importance [post]
func (h *StrategyImportanceHandlers) SetImportance(w http.ResponseWriter, r *http.Request) {
	domainID := chi.URLParam(r, "id")
	capabilityID := chi.URLParam(r, "capabilityId")

	req, ok := sharedAPI.DecodeRequestOrFail[SetStrategyImportanceRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.SetStrategyImportance{
		BusinessDomainID: domainID,
		CapabilityID:     capabilityID,
		PillarID:         req.PillarID,
		Importance:       req.Importance,
		Rationale:        req.Rationale,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		statusCode := sharedAPI.MapErrorToStatusCode(err, http.StatusBadRequest)
		sharedAPI.RespondError(w, statusCode, err, "Failed to set importance")
		return
	}

	created, err := h.importanceRM.GetByID(r.Context(), result.CreatedID)
	if err != nil || created == nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created importance")
		return
	}

	actor, _ := sharedctx.GetActor(r.Context())
	response := h.buildImportanceResponse(*created, domainID, actor)
	location := "/api/v1/business-domains/" + domainID + "/capabilities/" + capabilityID + "/importance/" + result.CreatedID
	w.Header().Set("Location", location)
	sharedAPI.RespondJSON(w, http.StatusCreated, response)
}

// UpdateImportance godoc
// @Summary Update strategic importance rating
// @Description Updates an existing strategic importance rating
// @Tags strategy-importance
// @Accept json
// @Produce json
// @Param domainId path string true "Business Domain ID"
// @Param capabilityId path string true "Capability ID"
// @Param importanceId path string true "Importance ID"
// @Param importance body UpdateStrategyImportanceRequest true "Updated importance"
// @Success 200 {object} StrategyImportanceResponse
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /business-domains/{domainId}/capabilities/{capabilityId}/importance/{importanceId} [put]
func (h *StrategyImportanceHandlers) UpdateImportance(w http.ResponseWriter, r *http.Request) {
	domainID := chi.URLParam(r, "id")
	importanceID := chi.URLParam(r, "importanceId")

	req, ok := sharedAPI.DecodeRequestOrFail[UpdateStrategyImportanceRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.UpdateStrategyImportance{
		ImportanceID: importanceID,
		Importance:   req.Importance,
		Rationale:    req.Rationale,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		statusCode := sharedAPI.MapErrorToStatusCode(err, http.StatusBadRequest)
		sharedAPI.RespondError(w, statusCode, err, "Failed to update importance")
		return
	}

	updated, err := h.importanceRM.GetByID(r.Context(), importanceID)
	if err != nil || updated == nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve updated importance")
		return
	}

	actor, _ := sharedctx.GetActor(r.Context())
	response := h.buildImportanceResponse(*updated, domainID, actor)
	sharedAPI.RespondJSON(w, http.StatusOK, response)
}

// RemoveImportance godoc
// @Summary Remove strategic importance rating
// @Description Deletes an existing strategic importance rating
// @Tags strategy-importance
// @Accept json
// @Produce json
// @Param domainId path string true "Business Domain ID"
// @Param capabilityId path string true "Capability ID"
// @Param importanceId path string true "Importance ID"
// @Success 204
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /business-domains/{domainId}/capabilities/{capabilityId}/importance/{importanceId} [delete]
func (h *StrategyImportanceHandlers) RemoveImportance(w http.ResponseWriter, r *http.Request) {
	importanceID := chi.URLParam(r, "importanceId")

	cmd := &commands.RemoveStrategyImportance{
		ImportanceID: importanceID,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		statusCode := sharedAPI.MapErrorToStatusCode(err, http.StatusBadRequest)
		sharedAPI.RespondError(w, statusCode, err, "Failed to remove importance")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetImportanceByDomain godoc
// @Summary Get all strategic importance ratings for a domain
// @Description Retrieves all strategic importance ratings for all capabilities in a business domain
// @Tags strategy-importance
// @Accept json
// @Produce json
// @Param domainId path string true "Business Domain ID"
// @Success 200 {object} sharedAPI.CollectionResponse{data=[]StrategyImportanceResponse}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /business-domains/{domainId}/importance [get]
func (h *StrategyImportanceHandlers) GetImportanceByDomain(w http.ResponseWriter, r *http.Request) {
	domainID := chi.URLParam(r, "id")

	h.respondWithImportanceCollection(w, r, importanceCollectionQuery{
		fetcher:  func() ([]readmodels.StrategyImportanceDTO, error) { return h.importanceRM.GetByDomain(r.Context(), domainID) },
		domainID: domainID,
		links:    selfOnlyLinks("/api/v1/business-domains/" + domainID + "/importance"),
	})
}

// GetImportanceByCapability godoc
// @Summary Get all strategic importance ratings for a capability
// @Description Retrieves all strategic importance ratings for a capability across all domains
// @Tags strategy-importance
// @Accept json
// @Produce json
// @Param id path string true "Capability ID"
// @Success 200 {object} sharedAPI.CollectionResponse{data=[]StrategyImportanceResponse}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id}/importance [get]
func (h *StrategyImportanceHandlers) GetImportanceByCapability(w http.ResponseWriter, r *http.Request) {
	capabilityID := chi.URLParam(r, "id")

	h.respondWithImportanceCollection(w, r, importanceCollectionQuery{
		fetcher: func() ([]readmodels.StrategyImportanceDTO, error) { return h.importanceRM.GetByCapability(r.Context(), capabilityID) },
		links:   selfOnlyLinks("/api/v1/capabilities/" + capabilityID + "/importance"),
	})
}

func selfOnlyLinks(href string) sharedAPI.Links {
	return sharedAPI.Links{"self": sharedAPI.NewLink(href, "GET")}
}

type importanceCollectionQuery struct {
	fetcher  func() ([]readmodels.StrategyImportanceDTO, error)
	domainID string
	links    sharedAPI.Links
}

func (h *StrategyImportanceHandlers) respondWithImportanceCollection(w http.ResponseWriter, r *http.Request, q importanceCollectionQuery) {
	ratings, err := q.fetcher()
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve importance ratings")
		return
	}

	actor, _ := sharedctx.GetActor(r.Context())
	data := h.buildImportanceResponses(ratings, q.domainID, actor)

	sharedAPI.RespondCollection(w, http.StatusOK, data, q.links)
}

func (h *StrategyImportanceHandlers) buildImportanceResponse(dto readmodels.StrategyImportanceDTO, domainID string, actor sharedctx.Actor) StrategyImportanceResponse {
	effectiveDomainID := domainID
	if effectiveDomainID == "" {
		effectiveDomainID = dto.BusinessDomainID
	}

	return StrategyImportanceResponse{
		ID:                 dto.ID,
		BusinessDomainID:   dto.BusinessDomainID,
		BusinessDomainName: dto.BusinessDomainName,
		CapabilityID:       dto.CapabilityID,
		CapabilityName:     dto.CapabilityName,
		PillarID:           dto.PillarID,
		PillarName:         dto.PillarName,
		Importance:         dto.Importance,
		ImportanceLabel:    dto.ImportanceLabel,
		Rationale:          dto.Rationale,
		Links:              h.hateoas.StrategyImportanceLinksForActor(effectiveDomainID, dto.CapabilityID, dto.ID, actor),
	}
}

func (h *StrategyImportanceHandlers) buildImportanceResponses(dtos []readmodels.StrategyImportanceDTO, domainID string, actor sharedctx.Actor) []StrategyImportanceResponse {
	responses := make([]StrategyImportanceResponse, len(dtos))
	for i, dto := range dtos {
		responses[i] = h.buildImportanceResponse(dto, domainID, actor)
	}
	return responses
}
