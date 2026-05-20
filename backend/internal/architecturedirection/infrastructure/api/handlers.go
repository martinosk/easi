package api

import (
	"context"
	"net/http"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/application/readmodels"
	authPL "easi/backend/internal/auth/publishedlanguage"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type DirectionQueries interface {
	GetByID(ctx context.Context, id string) (*readmodels.DirectionDTO, error)
	GetActiveByEnterpriseCapabilityID(ctx context.Context, enterpriseCapabilityID string) (*readmodels.DirectionDTO, error)
}

type DirectionHandlers struct {
	commandBus      cqrs.CommandBus
	queries         DirectionQueries
	sessionProvider authPL.SessionProvider
	hateoas         *DirectionLinks
}

func NewDirectionHandlers(commandBus cqrs.CommandBus, queries DirectionQueries, sessionProvider authPL.SessionProvider, hateoas *DirectionLinks) *DirectionHandlers {
	return &DirectionHandlers{
		commandBus:      commandBus,
		queries:         queries,
		sessionProvider: sessionProvider,
		hateoas:         hateoas,
	}
}

type PlacementRequest struct {
	TargetBusinessDomainID string `json:"targetBusinessDomainId"`
	ResultingName          string `json:"resultingName,omitempty"`
}

type CaptureDirectionRequest struct {
	Type                string             `json:"type"`
	SourceCapabilityIDs []string           `json:"sourceCapabilityIds"`
	Placements          []PlacementRequest `json:"placements"`
	Horizon             string             `json:"horizon"`
	Narrative           string             `json:"narrative,omitempty"`
}

type UpdateDirectionRequest struct {
	SourceCapabilityIDs *[]string           `json:"sourceCapabilityIds,omitempty"`
	Placements          *[]PlacementRequest `json:"placements,omitempty"`
	Horizon             *string             `json:"horizon,omitempty"`
	Narrative           *string             `json:"narrative,omitempty"`
}

type ECDirectionResponse struct {
	Direction *readmodels.DirectionDTO `json:"direction"`
	Links     sharedAPI.Links          `json:"_links,omitempty"`
}

// GetDirectionForEnterpriseCapability godoc
// @Summary Get the direction for an enterprise capability
// @Description Returns the current active direction on an enterprise capability, or null if none.
// @Tags directions
// @Produce json
// @Security CookieAuth
// @Param id path string true "Enterprise capability ID"
// @Success 200 {object} ECDirectionResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/direction [get]
func (h *DirectionHandlers) GetDirectionForEnterpriseCapability(w http.ResponseWriter, r *http.Request) {
	ecID := sharedAPI.GetPathParam(r, "id")
	direction, err := h.queries.GetActiveByEnterpriseCapabilityID(r.Context(), ecID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	envelope := ECDirectionResponse{
		Direction: direction,
		Links:     h.hateoas.EnterpriseCapabilityDirectionLinks(ecID, direction, actor),
	}
	if direction != nil {
		direction.Links = h.hateoas.DirectionForActor(ecID, direction.Status, actor)
	}
	sharedAPI.RespondJSON(w, http.StatusOK, envelope)
}

// CaptureDirection godoc
// @Summary Capture a draft direction on an enterprise capability
// @Description Creates a new direction in draft status; rejected if an active direction already exists.
// @Tags directions
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param id path string true "Enterprise capability ID"
// @Param body body CaptureDirectionRequest true "Direction data"
// @Success 201 {object} easi_backend_internal_architecturedirection_application_readmodels.DirectionDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/direction [post]
func (h *DirectionHandlers) CaptureDirection(w http.ResponseWriter, r *http.Request) {
	ecID := sharedAPI.GetPathParam(r, "id")
	req, ok := sharedAPI.DecodeRequestOrFail[CaptureDirectionRequest](w, r)
	if !ok {
		return
	}
	cmd := &commands.CaptureDirection{
		EnterpriseCapabilityID: ecID,
		Type:                   req.Type,
		SourceCapabilityIDs:    req.SourceCapabilityIDs,
		Placements:             placementsFromRequest(req.Placements),
		Horizon:                req.Horizon,
		Narrative:              req.Narrative,
	}
	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	h.respondWithActiveDirection(w, r, ecID, http.StatusCreated)
}

// UpdateDirection godoc
// @Summary Update the active direction on an enterprise capability
// @Tags directions
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param id path string true "Enterprise capability ID"
// @Param body body UpdateDirectionRequest true "Direction updates"
// @Success 200 {object} easi_backend_internal_architecturedirection_application_readmodels.DirectionDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/direction [put]
func (h *DirectionHandlers) UpdateDirection(w http.ResponseWriter, r *http.Request) {
	ecID := sharedAPI.GetPathParam(r, "id")
	req, ok := sharedAPI.DecodeRequestOrFail[UpdateDirectionRequest](w, r)
	if !ok {
		return
	}
	direction, ok := h.resolveActiveDirection(w, r, ecID)
	if !ok {
		return
	}
	for _, cmd := range buildUpdateCommands(direction.ID, req) {
		if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
			sharedAPI.HandleError(w, err)
			return
		}
	}
	h.respondWithActiveDirection(w, r, ecID, http.StatusOK)
}

// ProposeDirection godoc
// @Summary Advance the active direction to proposed
// @Tags directions
// @Produce json
// @Security CookieAuth
// @Param id path string true "Enterprise capability ID"
// @Success 200 {object} easi_backend_internal_architecturedirection_application_readmodels.DirectionDTO
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/direction/propose [post]
func (h *DirectionHandlers) ProposeDirection(w http.ResponseWriter, r *http.Request) {
	h.advance(w, r, "proposed")
}

// AgreeDirection godoc
// @Summary Advance the active direction to agreed
// @Tags directions
// @Produce json
// @Security CookieAuth
// @Param id path string true "Enterprise capability ID"
// @Success 200 {object} easi_backend_internal_architecturedirection_application_readmodels.DirectionDTO
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/direction/agree [post]
func (h *DirectionHandlers) AgreeDirection(w http.ResponseWriter, r *http.Request) {
	h.advance(w, r, "agreed")
}

// RejectDirection godoc
// @Summary Reject the active direction
// @Description Rejects the active direction on the enterprise capability. Returns the rejected direction with status set to rejected; the enterprise capability then has no active direction until a new one is captured.
// @Tags directions
// @Produce json
// @Security CookieAuth
// @Param id path string true "Enterprise capability ID"
// @Success 200 {object} easi_backend_internal_architecturedirection_application_readmodels.DirectionDTO
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/direction/reject [post]
func (h *DirectionHandlers) RejectDirection(w http.ResponseWriter, r *http.Request) {
	ecID := sharedAPI.GetPathParam(r, "id")
	direction, ok := h.resolveActiveDirection(w, r, ecID)
	if !ok {
		return
	}
	if _, err := h.commandBus.Dispatch(r.Context(), &commands.RejectDirection{DirectionID: direction.ID}); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	rejected, err := h.queries.GetByID(r.Context(), direction.ID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	h.respondWithDirection(w, r, directionResponse{ecID: ecID, direction: rejected, statusCode: http.StatusOK})
}

func (h *DirectionHandlers) advance(w http.ResponseWriter, r *http.Request, target string) {
	ecID := sharedAPI.GetPathParam(r, "id")
	direction, ok := h.resolveActiveDirection(w, r, ecID)
	if !ok {
		return
	}
	cmd := &commands.AdvanceDirection{DirectionID: direction.ID, TargetStatus: target}
	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	h.respondWithActiveDirection(w, r, ecID, http.StatusOK)
}

func (h *DirectionHandlers) resolveActiveDirection(w http.ResponseWriter, r *http.Request, ecID string) (*readmodels.DirectionDTO, bool) {
	direction, err := h.queries.GetActiveByEnterpriseCapabilityID(r.Context(), ecID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return nil, false
	}
	if direction == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, ErrNoActiveDirection, "No active direction on this enterprise capability")
		return nil, false
	}
	return direction, true
}

type directionResponse struct {
	ecID       string
	direction  *readmodels.DirectionDTO
	statusCode int
}

func (h *DirectionHandlers) respondWithActiveDirection(w http.ResponseWriter, r *http.Request, ecID string, statusCode int) {
	direction, err := h.queries.GetActiveByEnterpriseCapabilityID(r.Context(), ecID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	h.respondWithDirection(w, r, directionResponse{ecID: ecID, direction: direction, statusCode: statusCode})
}

func (h *DirectionHandlers) respondWithDirection(w http.ResponseWriter, r *http.Request, resp directionResponse) {
	if resp.direction == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, ErrNoActiveDirection, "Direction not found")
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	resp.direction.Links = h.hateoas.DirectionForActor(resp.ecID, resp.direction.Status, actor)
	if resp.statusCode == http.StatusCreated {
		location := sharedAPI.BuildSubResourceLink(enterpriseCapabilitiesPath, sharedAPI.ResourceID(resp.ecID), directionSubPath)
		sharedAPI.RespondCreated(w, location, resp.direction)
		return
	}
	sharedAPI.RespondJSON(w, resp.statusCode, resp.direction)
}

func buildUpdateCommands(directionID string, req UpdateDirectionRequest) []cqrs.Command {
	cmds := make([]cqrs.Command, 0, 4)
	if req.Narrative != nil {
		cmds = append(cmds, &commands.UpdateDirectionNarrative{DirectionID: directionID, Narrative: *req.Narrative})
	}
	if req.Horizon != nil {
		cmds = append(cmds, &commands.UpdateDirectionHorizon{DirectionID: directionID, Horizon: *req.Horizon})
	}
	if req.SourceCapabilityIDs != nil {
		cmds = append(cmds, &commands.UpdateDirectionSourceCapabilities{DirectionID: directionID, SourceCapabilityIDs: *req.SourceCapabilityIDs})
	}
	if req.Placements != nil {
		cmds = append(cmds, &commands.UpdateDirectionPlacements{DirectionID: directionID, Placements: placementsFromRequest(*req.Placements)})
	}
	return cmds
}

func placementsFromRequest(input []PlacementRequest) []commands.PlacementInput {
	out := make([]commands.PlacementInput, len(input))
	for i, p := range input {
		out[i] = commands.PlacementInput{
			TargetBusinessDomainID: p.TargetBusinessDomainID,
			ResultingName:          p.ResultingName,
		}
	}
	return out
}
