package api

import (
	"net/http"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type ViewComponentHandlers struct {
	commandBus   cqrs.CommandBus
	readModel    *readmodels.ArchitectureViewReadModel
	errorHandler *sharedAPI.ErrorHandler
}

func NewViewComponentHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.ArchitectureViewReadModel,
) *ViewComponentHandlers {
	return &ViewComponentHandlers{
		commandBus:   commandBus,
		readModel:    readModel,
		errorHandler: sharedAPI.NewErrorHandler(),
	}
}

func (h *ViewComponentHandlers) checkViewEditPermission(w http.ResponseWriter, r *http.Request, viewID string) bool {
	actor, ok := sharedctx.GetActor(r.Context())
	if !ok {
		sharedAPI.RespondError(w, http.StatusUnauthorized, nil, "Authentication required")
		return false
	}

	authInfo, err := h.readModel.GetAuthInfo(r.Context(), viewID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to check permissions")
		return false
	}
	if authInfo == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "View not found")
		return false
	}

	if !authInfo.IsPrivate {
		return true
	}

	if !isOwnerOfView(authInfo.OwnerUserID, actor.ID) {
		sharedAPI.RespondError(w, http.StatusForbidden, nil, "Access denied")
		return false
	}

	return true
}

type AddComponentRequest struct {
	ComponentID string  `json:"componentId"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
}

type UpdatePositionRequest struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type PositionUpdateItem struct {
	ComponentID string  `json:"componentId"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
}

type UpdateMultiplePositionsRequest struct {
	Positions []PositionUpdateItem `json:"positions"`
}

// AddComponentToView godoc
// @Summary Add a component to a view
// @Description Adds a component to an architecture view at a specific position
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param component body AddComponentRequest true "Component data"
// @Success 201
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse "Component already in view"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/components [post]
func (h *ViewComponentHandlers) AddComponentToView(w http.ResponseWriter, r *http.Request) {
	viewID := sharedAPI.GetPathParam(r, "id")

	if !h.checkViewEditPermission(w, r, viewID) {
		return
	}

	req, ok := sharedAPI.DecodeRequestOrFail[AddComponentRequest](w, r)
	if !ok {
		return
	}

	cmd := commands.AddComponentToView{
		ViewID:      viewID,
		ComponentID: req.ComponentID,
		X:           req.X,
		Y:           req.Y,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	location := sharedAPI.BuildSubResourceLink(sharedAPI.ResourcePath("/views"), sharedAPI.ResourceID(viewID), sharedAPI.ResourcePath("/components"))
	sharedAPI.RespondCreatedNoBody(w, location)
}

// UpdateComponentPosition godoc
// @Summary Update component position in a view
// @Description Updates the position of a component in an architecture view
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param componentId path string true "Component ID"
// @Param position body UpdatePositionRequest true "Position data"
// @Success 204
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse "View or component not found"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/components/{componentId}/position [patch]
func (h *ViewComponentHandlers) UpdateComponentPosition(w http.ResponseWriter, r *http.Request) {
	viewID := sharedAPI.GetPathParam(r, "id")
	componentID := sharedAPI.GetPathParam(r, "componentId")

	if !h.checkViewEditPermission(w, r, viewID) {
		return
	}

	req, ok := sharedAPI.DecodeRequestOrFail[UpdatePositionRequest](w, r)
	if !ok {
		return
	}

	cmd := commands.UpdateComponentPosition{
		ViewID:      viewID,
		ComponentID: componentID,
		X:           req.X,
		Y:           req.Y,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		w.WriteHeader(http.StatusNoContent)
	})
}

// UpdateMultiplePositions godoc
// @Summary Update multiple component positions
// @Description Updates positions for multiple components in a view in a single operation
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param positions body UpdateMultiplePositionsRequest true "Position updates"
// @Success 204
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/positions [patch]
func (h *ViewComponentHandlers) UpdateMultiplePositions(w http.ResponseWriter, r *http.Request) {
	viewID := sharedAPI.GetPathParam(r, "id")

	if !h.checkViewEditPermission(w, r, viewID) {
		return
	}

	req, ok := sharedAPI.DecodeRequestOrFail[UpdateMultiplePositionsRequest](w, r)
	if !ok {
		return
	}

	positions := make([]commands.PositionUpdate, len(req.Positions))
	for i, pos := range req.Positions {
		positions[i] = commands.PositionUpdate{
			ComponentID: pos.ComponentID,
			X:           pos.X,
			Y:           pos.Y,
		}
	}

	cmd := commands.UpdateMultiplePositions{
		ViewID:    viewID,
		Positions: positions,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	sharedAPI.RespondNoContent(w)
}

// RemoveComponentFromView godoc
// @Summary Remove a component from a view
// @Description Removes a component from an architecture view without deleting the component
// @Tags views
// @Produce json
// @Param id path string true "View ID"
// @Param componentId path string true "Component ID"
// @Success 204
// @Failure 404 {object} sharedAPI.ErrorResponse "View or component not found"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/components/{componentId} [delete]
func (h *ViewComponentHandlers) RemoveComponentFromView(w http.ResponseWriter, r *http.Request) {
	viewID := sharedAPI.GetPathParam(r, "id")
	componentID := sharedAPI.GetPathParam(r, "componentId")

	if !h.checkViewEditPermission(w, r, viewID) {
		return
	}

	cmd := &commands.RemoveComponentFromView{
		ViewID:      viewID,
		ComponentID: componentID,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		w.WriteHeader(http.StatusNoContent)
	})
}
