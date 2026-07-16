package api

import (
	"context"
	"net/http"
	"strings"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/application/handlers"
	"easi/backend/internal/architecturedirection/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type TimeAssessmentQueries interface {
	GetByPair(ctx context.Context, capabilityID, componentID string) (*readmodels.TimeAssessmentDTO, error)
	GetByCapabilityIDs(ctx context.Context, capabilityIDs []string) ([]readmodels.TimeAssessmentDTO, error)
	GetAll(ctx context.Context) ([]readmodels.TimeAssessmentDTO, error)
	GetRollupsByComponentIDs(ctx context.Context, componentIDs []string) ([]readmodels.TimeAssessmentRollupDTO, error)
}

type TimeAssessmentHandlers struct {
	commandBus cqrs.CommandBus
	queries    TimeAssessmentQueries
	hateoas    *TimeAssessmentLinks
}

func NewTimeAssessmentHandlers(commandBus cqrs.CommandBus, queries TimeAssessmentQueries, hateoas *TimeAssessmentLinks) *TimeAssessmentHandlers {
	return &TimeAssessmentHandlers{commandBus: commandBus, queries: queries, hateoas: hateoas}
}

type AssessRealizationRequest struct {
	Grade     string `json:"grade"`
	Rationale string `json:"rationale,omitempty"`
}

// GetTimeAssessments godoc
// @Summary Bulk-fetch current TIME assessments for a capability set
// @Description Returns the current assessment for every (capability, component) pair among the given capabilities.
// @Tags time-assessments
// @Produce json
// @Security CookieAuth
// @Param capabilityIds query string false "Comma-separated capability IDs; omit to return the whole collection"
// @Success 200 {object} sharedAPI.CollectionResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /time-assessments [get]
func (h *TimeAssessmentHandlers) GetTimeAssessments(w http.ResponseWriter, r *http.Request) {
	assessments, err := h.fetchAssessments(r)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	for i := range assessments {
		assessments[i].Links = h.hateoas.ItemLinks(assessments[i].CapabilityID, assessments[i].ComponentID, actor)
	}
	links := h.hateoas.CollectionLinks(string(timeAssessmentsPath), actor)
	sharedAPI.RespondCollection(w, http.StatusOK, assessments, links)
}

func (h *TimeAssessmentHandlers) fetchAssessments(r *http.Request) ([]readmodels.TimeAssessmentDTO, error) {
	if capabilityIDs, filtered := capabilityIDFilter(r); filtered {
		return h.queries.GetByCapabilityIDs(r.Context(), capabilityIDs)
	}
	return h.queries.GetAll(r.Context())
}

// GetTimeAssessmentRollups godoc
// @Summary Per-application TIME grade rollups across the landscape
// @Description Returns, for each given application component, the count of current assessments per grade.
// @Tags time-assessments
// @Produce json
// @Security CookieAuth
// @Param componentIds query string true "Comma-separated application component IDs"
// @Success 200 {object} sharedAPI.CollectionResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /time-assessments/rollups [get]
func (h *TimeAssessmentHandlers) GetTimeAssessmentRollups(w http.ResponseWriter, r *http.Request) {
	componentIDs := parseIDList(r.URL.Query().Get("componentIds"))
	rollups, err := h.queries.GetRollupsByComponentIDs(r.Context(), componentIDs)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	links := sharedAPI.Links{"self": h.hateoas.Get(string(timeAssessmentsPath) + "/rollups")}
	sharedAPI.RespondCollection(w, http.StatusOK, rollups, links)
}

// GetTimeAssessment godoc
// @Summary Get the current TIME assessment for a realisation
// @Description Returns the current assessment for the given (capability, component) pair, or 404 when unassessed.
// @Tags time-assessments
// @Produce json
// @Security CookieAuth
// @Param id path string true "Capability ID"
// @Param componentId path string true "Application component ID"
// @Success 200 {object} readmodels.TimeAssessmentDTO
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id}/components/{componentId}/time-assessment [get]
func (h *TimeAssessmentHandlers) GetTimeAssessment(w http.ResponseWriter, r *http.Request) {
	capabilityID := sharedAPI.GetPathParam(r, "id")
	componentID := sharedAPI.GetPathParam(r, "componentId")
	dto, err := h.queries.GetByPair(r.Context(), capabilityID, componentID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	if dto == nil {
		sharedAPI.HandleError(w, handlers.ErrTimeAssessmentNotFoundForPair)
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	dto.Links = h.hateoas.ItemLinks(capabilityID, componentID, actor)
	sharedAPI.RespondJSON(w, http.StatusOK, dto)
}

// PutTimeAssessment godoc
// @Summary Assess or re-assess a realisation's TIME grade
// @Description Records the architect's grade for the (capability, component) pair. Requires a direct realisation. 201 on first assessment, 200 on re-assessment.
// @Tags time-assessments
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param id path string true "Capability ID"
// @Param componentId path string true "Application component ID"
// @Param body body AssessRealizationRequest true "Assessment data"
// @Success 200 {object} readmodels.TimeAssessmentDTO
// @Success 201 {object} readmodels.TimeAssessmentDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id}/components/{componentId}/time-assessment [put]
func (h *TimeAssessmentHandlers) PutTimeAssessment(w http.ResponseWriter, r *http.Request) {
	capabilityID := sharedAPI.GetPathParam(r, "id")
	componentID := sharedAPI.GetPathParam(r, "componentId")
	req, ok := sharedAPI.DecodeRequestOrFail[AssessRealizationRequest](w, r)
	if !ok {
		return
	}
	existedBefore, err := h.queries.GetByPair(r.Context(), capabilityID, componentID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	cmd := &commands.AssessRealization{
		CapabilityID: capabilityID,
		ComponentID:  componentID,
		Grade:        req.Grade,
		Rationale:    req.Rationale,
		AssessedBy:   actor.Email,
	}
	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	statusCode := http.StatusOK
	if existedBefore == nil {
		statusCode = http.StatusCreated
	}
	h.respondWithCurrentAssessment(w, r, timeAssessmentPairID{CapabilityID: capabilityID, ComponentID: componentID}, statusCode)
}

// DeleteTimeAssessment godoc
// @Summary Remove a TIME assessment
// @Description Removes the current assessment for the pair; the realisation presents as unassessed. Recorded as a discrete TimeAssessmentRemoved event.
// @Tags time-assessments
// @Security CookieAuth
// @Param id path string true "Capability ID"
// @Param componentId path string true "Application component ID"
// @Success 204 "No Content"
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id}/components/{componentId}/time-assessment [delete]
func (h *TimeAssessmentHandlers) DeleteTimeAssessment(w http.ResponseWriter, r *http.Request) {
	capabilityID := sharedAPI.GetPathParam(r, "id")
	componentID := sharedAPI.GetPathParam(r, "componentId")
	actor, _ := sharedctx.GetActor(r.Context())
	cmd := &commands.RemoveTimeAssessment{
		CapabilityID: capabilityID,
		ComponentID:  componentID,
		RemovedBy:    actor.Email,
	}
	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	sharedAPI.RespondNoContent(w)
}

type timeAssessmentPairID struct {
	CapabilityID string
	ComponentID  string
}

func (h *TimeAssessmentHandlers) respondWithCurrentAssessment(w http.ResponseWriter, r *http.Request, pair timeAssessmentPairID, statusCode int) {
	dto, err := h.queries.GetByPair(r.Context(), pair.CapabilityID, pair.ComponentID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	if dto == nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, handlers.ErrTimeAssessmentNotFoundForPair, "failed to load time assessment")
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	dto.Links = h.hateoas.ItemLinks(pair.CapabilityID, pair.ComponentID, actor)
	if statusCode == http.StatusCreated {
		location := sharedAPI.APIVersionPrefix + timeAssessmentItemResourcePath(pair.CapabilityID, pair.ComponentID)
		sharedAPI.RespondCreated(w, location, dto)
		return
	}
	sharedAPI.RespondJSON(w, statusCode, dto)
}

func capabilityIDFilter(r *http.Request) ([]string, bool) {
	if !r.URL.Query().Has("capabilityIds") {
		return nil, false
	}
	return parseIDList(r.URL.Query().Get("capabilityIds")), true
}

func parseIDList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	ids := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			ids = append(ids, trimmed)
		}
	}
	return ids
}
