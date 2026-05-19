package api

import (
	"context"
	"errors"
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
// @Summary Get direction for enterprise capability
// @Description Returns the current active direction on an enterprise capability, or null if none
// @Tags directions
// @Produce json
// @Param id path string true "Enterprise capability ID"
// @Success 200 {object} ECDirectionResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
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
		Links:     h.hateoas.EnterpriseCapabilityDirectionLinks(ecID, direction != nil, actor),
	}
	if direction != nil {
		direction.Links = h.hateoas.DirectionForActor(direction.ID, direction.EnterpriseCapabilityID, direction.Status, actor)
	}
	sharedAPI.RespondJSON(w, http.StatusOK, envelope)
}

// CaptureDirection godoc
// @Summary Capture a draft direction on an enterprise capability
// @Description Creates a new direction in draft status; rejected if an active direction already exists
// @Tags directions
// @Accept json
// @Produce json
// @Param id path string true "Enterprise capability ID"
// @Param body body CaptureDirectionRequest true "Direction data"
// @Success 201 {object} easi_backend_internal_architecturedirection_application_readmodels.DirectionDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
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
	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	h.respondWithDirection(w, r, result.CreatedID, http.StatusCreated)
}

// AdvanceDirection godoc
// @Summary Advance a direction to proposed or agreed
// @Tags directions
// @Param id path string true "Direction ID"
// @Param target path string true "Target status: proposed or agreed"
// @Success 200 {object} easi_backend_internal_architecturedirection_application_readmodels.DirectionDTO
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Router /directions/{id}/advance/{target} [post]
func (h *DirectionHandlers) AdvanceDirection(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")
	target := sharedAPI.GetPathParam(r, "target")
	h.dispatchAndRespond(w, r, id, &commands.AdvanceDirection{DirectionID: id, TargetStatus: target})
}

// RejectDirection godoc
// @Summary Reject a direction
// @Tags directions
// @Param id path string true "Direction ID"
// @Success 200 {object} easi_backend_internal_architecturedirection_application_readmodels.DirectionDTO
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Router /directions/{id}/reject [post]
func (h *DirectionHandlers) RejectDirection(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")
	h.dispatchAndRespond(w, r, id, &commands.RejectDirection{DirectionID: id})
}

// UpdateDirection godoc
// @Summary Update a draft or proposed direction
// @Tags directions
// @Accept json
// @Produce json
// @Param id path string true "Direction ID"
// @Param body body UpdateDirectionRequest true "Updates"
// @Success 200 {object} easi_backend_internal_architecturedirection_application_readmodels.DirectionDTO
// @Router /directions/{id} [put]
func (h *DirectionHandlers) UpdateDirection(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")
	req, ok := sharedAPI.DecodeRequestOrFail[UpdateDirectionRequest](w, r)
	if !ok {
		return
	}
	for _, cmd := range buildUpdateCommands(id, req) {
		if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
			sharedAPI.HandleError(w, err)
			return
		}
	}
	h.respondWithDirection(w, r, id, http.StatusOK)
}

// GetDirection godoc
// @Summary Get a direction by ID
// @Tags directions
// @Param id path string true "Direction ID"
// @Success 200 {object} easi_backend_internal_architecturedirection_application_readmodels.DirectionDTO
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Router /directions/{id} [get]
func (h *DirectionHandlers) GetDirection(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")
	h.respondWithDirection(w, r, id, http.StatusOK)
}

func (h *DirectionHandlers) dispatchAndRespond(w http.ResponseWriter, r *http.Request, id string, cmd cqrs.Command) {
	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	h.respondWithDirection(w, r, id, http.StatusOK)
}

func (h *DirectionHandlers) respondWithDirection(w http.ResponseWriter, r *http.Request, id string, statusCode int) {
	direction, err := h.queries.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	if direction == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, errors.New("not found"), "Direction not found")
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	direction.Links = h.hateoas.DirectionForActor(direction.ID, direction.EnterpriseCapabilityID, direction.Status, actor)
	if statusCode == http.StatusCreated {
		location := sharedAPI.BuildResourceLink(sharedAPI.ResourcePath("/directions"), sharedAPI.ResourceID(direction.ID))
		sharedAPI.RespondCreated(w, location, direction)
		return
	}
	sharedAPI.RespondJSON(w, statusCode, direction)
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
