package api

import (
	"net/http"

	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/types"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type ApplicationFitScoreHandlers struct {
	commandBus       cqrs.CommandBus
	fitScoreRM       *readmodels.ApplicationFitScoreReadModel
	fitComparisonRM  *readmodels.ComponentFitComparisonReadModel
	hateoas          *sharedAPI.HATEOASLinks
	sessionManager   *session.SessionManager
}

func NewApplicationFitScoreHandlers(
	commandBus cqrs.CommandBus,
	fitScoreRM *readmodels.ApplicationFitScoreReadModel,
	fitComparisonRM *readmodels.ComponentFitComparisonReadModel,
	hateoas *sharedAPI.HATEOASLinks,
	sessionManager *session.SessionManager,
) *ApplicationFitScoreHandlers {
	return &ApplicationFitScoreHandlers{
		commandBus:       commandBus,
		fitScoreRM:       fitScoreRM,
		fitComparisonRM:  fitComparisonRM,
		hateoas:          hateoas,
		sessionManager:   sessionManager,
	}
}

type SetApplicationFitScoreRequest struct {
	Score     int    `json:"score"`
	Rationale string `json:"rationale,omitempty"`
}

type ApplicationFitScoreResponse struct {
	ID            string      `json:"id"`
	ComponentID   string      `json:"componentId"`
	ComponentName string      `json:"componentName"`
	PillarID      string      `json:"pillarId"`
	PillarName    string      `json:"pillarName"`
	Score         int         `json:"score"`
	ScoreLabel    string      `json:"scoreLabel"`
	Rationale     string      `json:"rationale,omitempty"`
	ScoredAt      string      `json:"scoredAt"`
	ScoredBy      string      `json:"scoredBy"`
	Links         types.Links `json:"_links,omitempty"`
}

// GetFitScoresByComponent godoc
// @Summary Get strategic fit scores for a component
// @Description Retrieves all strategic fit scores for a specific application component
// @Tags application-fit-scores
// @Accept json
// @Produce json
// @Param id path string true "Component ID"
// @Success 200 {object} sharedAPI.CollectionResponse{data=[]ApplicationFitScoreResponse}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{id}/fit-scores [get]
func (h *ApplicationFitScoreHandlers) GetFitScoresByComponent(w http.ResponseWriter, r *http.Request) {
	componentID := chi.URLParam(r, "id")
	actor, _ := sharedctx.GetActor(r.Context())

	scores, err := h.fitScoreRM.GetByComponentID(r.Context(), componentID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve fit scores")
		return
	}

	data := h.buildFitScoreResponsesForActor(scores, actor)
	links := h.hateoas.FitScoresCollectionLinksForActor(componentID, actor)

	sharedAPI.RespondCollection(w, http.StatusOK, data, links)
}

// SetFitScore godoc
// @Summary Set or update strategic fit score for a component-pillar
// @Description Creates or updates a strategic fit score for a component-pillar combination. Requires components:write permission (Admin or Architect role).
// @Tags application-fit-scores
// @Accept json
// @Produce json
// @Param id path string true "Component ID"
// @Param pillarId path string true "Strategy Pillar ID"
// @Param fitScore body SetApplicationFitScoreRequest true "Fit score"
// @Success 200 {object} ApplicationFitScoreResponse
// @Success 201 {object} ApplicationFitScoreResponse
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid component/pillar ID format or validation error"
// @Failure 401 {object} sharedAPI.ErrorResponse "Authentication required"
// @Failure 403 {object} sharedAPI.ErrorResponse "Insufficient permissions - requires components:write"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security ApiKeyAuth
// @Router /components/{id}/fit-scores/{pillarId} [put]
func (h *ApplicationFitScoreHandlers) SetFitScore(w http.ResponseWriter, r *http.Request) {
	authSession, err := h.sessionManager.LoadAuthenticatedSession(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Authentication required")
		return
	}

	componentID := chi.URLParam(r, "id")
	pillarID := chi.URLParam(r, "pillarId")

	if _, err := uuid.Parse(componentID); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid component ID format")
		return
	}
	if _, err := uuid.Parse(pillarID); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid pillar ID format")
		return
	}

	req, ok := sharedAPI.DecodeRequestOrFail[SetApplicationFitScoreRequest](w, r)
	if !ok {
		return
	}

	existing, err := h.fitScoreRM.GetByComponentAndPillar(r.Context(), componentID, pillarID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to check existing fit score")
		return
	}

	var result cqrs.CommandResult
	var statusCode int
	userEmail := authSession.UserEmail()

	if existing != nil {
		cmd := &commands.UpdateApplicationFitScore{
			FitScoreID: existing.ID,
			Score:      req.Score,
			Rationale:  req.Rationale,
			UpdatedBy:  userEmail,
		}
		result, err = h.commandBus.Dispatch(r.Context(), cmd)
		statusCode = http.StatusOK
	} else {
		cmd := &commands.SetApplicationFitScore{
			ComponentID: componentID,
			PillarID:    pillarID,
			Score:       req.Score,
			Rationale:   req.Rationale,
			ScoredBy:    userEmail,
		}
		result, err = h.commandBus.Dispatch(r.Context(), cmd)
		statusCode = http.StatusCreated
	}

	if err != nil {
		httpStatus := sharedAPI.MapErrorToStatusCode(err, http.StatusBadRequest)
		sharedAPI.RespondError(w, httpStatus, err, "Failed to set fit score")
		return
	}

	var fitScoreID string
	if existing != nil {
		fitScoreID = existing.ID
	} else {
		fitScoreID = result.CreatedID
	}

	updated, err := h.fitScoreRM.GetByID(r.Context(), fitScoreID)
	if err != nil || updated == nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve fit score")
		return
	}

	actor, _ := sharedctx.GetActor(r.Context())
	response := h.buildFitScoreResponseForActor(*updated, actor)

	if statusCode == http.StatusCreated {
		location := "/api/v1/components/" + componentID + "/fit-scores/" + pillarID
		w.Header().Set("Location", location)
	}

	sharedAPI.RespondJSON(w, statusCode, response)
}

// RemoveFitScore godoc
// @Summary Remove strategic fit score
// @Description Deletes an existing strategic fit score. Requires components:write permission (Admin or Architect role).
// @Tags application-fit-scores
// @Accept json
// @Produce json
// @Param id path string true "Component ID"
// @Param pillarId path string true "Strategy Pillar ID"
// @Success 204
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid component/pillar ID format"
// @Failure 401 {object} sharedAPI.ErrorResponse "Authentication required"
// @Failure 403 {object} sharedAPI.ErrorResponse "Insufficient permissions - requires components:write"
// @Failure 404 {object} sharedAPI.ErrorResponse "Fit score not found"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security ApiKeyAuth
// @Router /components/{id}/fit-scores/{pillarId} [delete]
func (h *ApplicationFitScoreHandlers) RemoveFitScore(w http.ResponseWriter, r *http.Request) {
	authSession, err := h.sessionManager.LoadAuthenticatedSession(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusUnauthorized, err, "Authentication required")
		return
	}

	componentID := chi.URLParam(r, "id")
	pillarID := chi.URLParam(r, "pillarId")

	if _, err := uuid.Parse(componentID); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid component ID format")
		return
	}
	if _, err := uuid.Parse(pillarID); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid pillar ID format")
		return
	}

	existing, err := h.fitScoreRM.GetByComponentAndPillar(r.Context(), componentID, pillarID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to check existing fit score")
		return
	}

	if existing == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Fit score not found")
		return
	}

	cmd := &commands.RemoveApplicationFitScore{
		FitScoreID: existing.ID,
		RemovedBy:  authSession.UserEmail(),
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		httpStatus := sharedAPI.MapErrorToStatusCode(err, http.StatusBadRequest)
		sharedAPI.RespondError(w, httpStatus, err, "Failed to remove fit score")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetFitScoresByPillar godoc
// @Summary Get all fit scores for a strategy pillar
// @Description Retrieves all strategic fit scores for a specific strategy pillar across all components
// @Tags application-fit-scores
// @Accept json
// @Produce json
// @Param pillarId path string true "Strategy Pillar ID"
// @Success 200 {object} sharedAPI.CollectionResponse{data=[]ApplicationFitScoreResponse}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /strategy-pillars/{pillarId}/fit-scores [get]
func (h *ApplicationFitScoreHandlers) GetFitScoresByPillar(w http.ResponseWriter, r *http.Request) {
	pillarID := chi.URLParam(r, "pillarId")
	actor, _ := sharedctx.GetActor(r.Context())

	scores, err := h.fitScoreRM.GetByPillarID(r.Context(), pillarID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve fit scores")
		return
	}

	data := h.buildFitScoreResponsesForActor(scores, actor)
	links := sharedAPI.Links{
		"self": sharedAPI.NewLink("/api/v1/strategy-pillars/"+pillarID+"/fit-scores", "GET"),
	}

	sharedAPI.RespondCollection(w, http.StatusOK, data, links)
}

func (h *ApplicationFitScoreHandlers) buildFitScoreResponseForActor(dto readmodels.ApplicationFitScoreDTO, actor sharedctx.Actor) ApplicationFitScoreResponse {
	return ApplicationFitScoreResponse{
		ID:            dto.ID,
		ComponentID:   dto.ComponentID,
		ComponentName: dto.ComponentName,
		PillarID:      dto.PillarID,
		PillarName:    dto.PillarName,
		Score:         dto.Score,
		ScoreLabel:    dto.ScoreLabel,
		Rationale:     dto.Rationale,
		ScoredAt:      dto.ScoredAt.Format("2006-01-02T15:04:05Z"),
		ScoredBy:      dto.ScoredBy,
		Links:         h.hateoas.FitScoreLinksForActor(dto.ComponentID, dto.PillarID, actor),
	}
}

func (h *ApplicationFitScoreHandlers) buildFitScoreResponsesForActor(dtos []readmodels.ApplicationFitScoreDTO, actor sharedctx.Actor) []ApplicationFitScoreResponse {
	responses := make([]ApplicationFitScoreResponse, len(dtos))
	for i, dto := range dtos {
		responses[i] = h.buildFitScoreResponseForActor(dto, actor)
	}
	return responses
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
func (h *ApplicationFitScoreHandlers) GetFitComparisons(w http.ResponseWriter, r *http.Request) {
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
