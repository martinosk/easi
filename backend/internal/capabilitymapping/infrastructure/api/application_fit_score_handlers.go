package api

import (
	"context"
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
	commandBus     cqrs.CommandBus
	fitScoreRM     *readmodels.ApplicationFitScoreReadModel
	hateoas        *sharedAPI.HATEOASLinks
	sessionManager *session.SessionManager
}

func NewApplicationFitScoreHandlers(
	commandBus cqrs.CommandBus,
	fitScoreRM *readmodels.ApplicationFitScoreReadModel,
	hateoas *sharedAPI.HATEOASLinks,
	sessionManager *session.SessionManager,
) *ApplicationFitScoreHandlers {
	return &ApplicationFitScoreHandlers{
		commandBus:     commandBus,
		fitScoreRM:     fitScoreRM,
		hateoas:        hateoas,
		sessionManager: sessionManager,
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
	links := h.hateoas.FitScoresCollectionLinksForActor(componentID, actor)
	h.fetchAndRespondFitScores(w, r, actor, links, func() ([]readmodels.ApplicationFitScoreDTO, error) {
		return h.fitScoreRM.GetByComponentID(r.Context(), componentID)
	})
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
	if !validateUUIDs(w, componentID, pillarID) {
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

	dispatchResult, err := h.dispatchFitScoreCommand(r.Context(), fitScoreDispatchParams{
		existing:    existing,
		req:         req,
		componentID: componentID,
		pillarID:    pillarID,
		userEmail:   authSession.UserEmail(),
	})
	if err != nil {
		httpStatus := sharedAPI.MapErrorToStatusCode(err, http.StatusBadRequest)
		sharedAPI.RespondError(w, httpStatus, err, "Failed to set fit score")
		return
	}

	updated, err := h.fitScoreRM.GetByID(r.Context(), dispatchResult.fitScoreID)
	if err != nil || updated == nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve fit score")
		return
	}

	h.respondWithFitScore(w, r, *updated, dispatchResult)
}

type fitScoreDispatchParams struct {
	existing    *readmodels.ApplicationFitScoreDTO
	req         SetApplicationFitScoreRequest
	componentID string
	pillarID    string
	userEmail   string
}

type fitScoreResult struct {
	fitScoreID string
	statusCode int
	location   string
}

func (h *ApplicationFitScoreHandlers) dispatchFitScoreCommand(ctx context.Context, params fitScoreDispatchParams) (fitScoreResult, error) {
	if params.existing != nil {
		cmd := &commands.UpdateApplicationFitScore{
			FitScoreID: params.existing.ID,
			Score:      params.req.Score,
			Rationale:  params.req.Rationale,
			UpdatedBy:  params.userEmail,
		}
		_, err := h.commandBus.Dispatch(ctx, cmd)
		return fitScoreResult{fitScoreID: params.existing.ID, statusCode: http.StatusOK}, err
	}

	cmd := &commands.SetApplicationFitScore{
		ComponentID: params.componentID,
		PillarID:    params.pillarID,
		Score:       params.req.Score,
		Rationale:   params.req.Rationale,
		ScoredBy:    params.userEmail,
	}
	result, err := h.commandBus.Dispatch(ctx, cmd)
	return fitScoreResult{
		fitScoreID: result.CreatedID,
		statusCode: http.StatusCreated,
		location:   "/api/v1/components/" + params.componentID + "/fit-scores/" + params.pillarID,
	}, err
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
	if !validateUUIDs(w, componentID, pillarID) {
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
	links := sharedAPI.Links{
		"self": sharedAPI.NewLink("/api/v1/strategy-pillars/"+pillarID+"/fit-scores", "GET"),
	}
	h.fetchAndRespondFitScores(w, r, actor, links, func() ([]readmodels.ApplicationFitScoreDTO, error) {
		return h.fitScoreRM.GetByPillarID(r.Context(), pillarID)
	})
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

func (h *ApplicationFitScoreHandlers) respondWithFitScore(w http.ResponseWriter, r *http.Request, dto readmodels.ApplicationFitScoreDTO, result fitScoreResult) {
	actor, _ := sharedctx.GetActor(r.Context())
	response := h.buildFitScoreResponseForActor(dto, actor)

	if result.location != "" {
		w.Header().Set("Location", result.location)
	}

	sharedAPI.RespondJSON(w, result.statusCode, response)
}

func (h *ApplicationFitScoreHandlers) buildFitScoreResponsesForActor(dtos []readmodels.ApplicationFitScoreDTO, actor sharedctx.Actor) []ApplicationFitScoreResponse {
	responses := make([]ApplicationFitScoreResponse, len(dtos))
	for i, dto := range dtos {
		responses[i] = h.buildFitScoreResponseForActor(dto, actor)
	}
	return responses
}

func (h *ApplicationFitScoreHandlers) fetchAndRespondFitScores(
	w http.ResponseWriter, r *http.Request,
	actor sharedctx.Actor, links sharedAPI.Links,
	query func() ([]readmodels.ApplicationFitScoreDTO, error),
) {
	scores, err := query()
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve fit scores")
		return
	}
	data := h.buildFitScoreResponsesForActor(scores, actor)
	sharedAPI.RespondCollection(w, http.StatusOK, data, links)
}

func validateUUIDs(w http.ResponseWriter, componentID, pillarID string) bool {
	if _, err := uuid.Parse(componentID); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid component ID format")
		return false
	}
	if _, err := uuid.Parse(pillarID); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid pillar ID format")
		return false
	}
	return true
}

