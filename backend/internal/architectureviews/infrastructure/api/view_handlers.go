package api

import (
	"context"
	"fmt"
	"net/http"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/application/readmodels"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/types"
)

type ViewHandlers struct {
	commandBus   cqrs.CommandBus
	readModel    *readmodels.ArchitectureViewReadModel
	layoutRepo   LayoutRepository
	hateoas      *sharedAPI.HATEOASLinks
	errorHandler *sharedAPI.ErrorHandler
}

type LayoutRepository interface {
	AddCapabilityToView(ctx context.Context, viewID, capabilityID string, x, y float64) error
	UpdateCapabilityPosition(ctx context.Context, viewID, capabilityID string, x, y float64) error
	RemoveCapabilityFromView(ctx context.Context, viewID, capabilityID string) error
	AddOriginEntityToView(ctx context.Context, viewID, originEntityID string, x, y float64) error
	UpdateOriginEntityPosition(ctx context.Context, viewID, originEntityID string, x, y float64) error
	RemoveOriginEntityFromView(ctx context.Context, viewID, originEntityID string) error
}

func NewViewHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.ArchitectureViewReadModel,
	layoutRepo LayoutRepository,
	hateoas *sharedAPI.HATEOASLinks,
) *ViewHandlers {
	return &ViewHandlers{
		commandBus:   commandBus,
		readModel:    readModel,
		layoutRepo:   layoutRepo,
		hateoas:      hateoas,
		errorHandler: sharedAPI.NewErrorHandler(),
	}
}

func (h *ViewHandlers) dispatchCommand(w http.ResponseWriter, r *http.Request, cmd cqrs.Command) {
	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		w.WriteHeader(http.StatusNoContent)
	})
}

type elementParams struct {
	viewID      string
	elementID   string
	elementType string
}

func (h *ViewHandlers) updateElementColor(w http.ResponseWriter, r *http.Request, params elementParams) {
	if !h.checkViewEditPermission(w, r, params.viewID) {
		return
	}

	req, ok := sharedAPI.DecodeRequestOrFail[UpdateElementColorRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.UpdateElementColor{
		ViewID:      params.viewID,
		ElementID:   params.elementID,
		ElementType: params.elementType,
		Color:       req.Color,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	sharedAPI.RespondNoContent(w)
}

func (h *ViewHandlers) clearElementColor(w http.ResponseWriter, r *http.Request, params elementParams) {
	if !h.checkViewEditPermission(w, r, params.viewID) {
		return
	}

	cmd := &commands.ClearElementColor{
		ViewID:      params.viewID,
		ElementID:   params.elementID,
		ElementType: params.elementType,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, fmt.Sprintf("Failed to clear %s color", params.elementType))
		return
	}

	sharedAPI.RespondNoContent(w)
}

func (h *ViewHandlers) decodeValidateAndDispatch(w http.ResponseWriter, r *http.Request, req interface{}, validate func() error, createCmd func() cqrs.Command) {
	if err := sharedAPI.DecodeRequestInto(r, req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	if validate != nil {
		if err := validate(); err != nil {
			sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
			return
		}
	}

	h.dispatchCommand(w, r, createCmd())
}

func (h *ViewHandlers) handleViewUpdate(w http.ResponseWriter, r *http.Request, decode func() error, validate func() error, createCmd func() cqrs.Command) {
	viewID := sharedAPI.GetPathParam(r, "id")

	if !h.checkViewEditPermission(w, r, viewID) {
		return
	}

	if err := decode(); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	if err := validate(); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	h.dispatchCommand(w, r, createCmd())
}

type layoutOpParams struct {
	elementIDParam string
	operation      func(ctx context.Context, viewID, elementID string) error
	errorMsg       string
}

func (h *ViewHandlers) handleLayoutOp(w http.ResponseWriter, r *http.Request, p layoutOpParams) {
	viewID := sharedAPI.GetPathParam(r, "id")
	elementID := sharedAPI.GetPathParam(r, p.elementIDParam)

	if !h.checkViewEditPermission(w, r, viewID) {
		return
	}

	if err := p.operation(r.Context(), viewID, elementID); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, p.errorMsg)
		return
	}

	sharedAPI.RespondNoContent(w)
}

type addEntityParams struct {
	entityIDField   string
	subResourcePath string
	addFn           func(ctx context.Context, viewID, entityID string, x, y float64) error
	errorMsg        string
}

func (h *ViewHandlers) handleAddEntity(w http.ResponseWriter, r *http.Request, decodeAndExtract func() (entityID string, x, y float64, ok bool), p addEntityParams) {
	viewID := sharedAPI.GetPathParam(r, "id")
	if !h.checkViewEditPermission(w, r, viewID) {
		return
	}

	entityID, x, y, ok := decodeAndExtract()
	if !ok {
		return
	}

	if entityID == "" {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, p.entityIDField+" is required")
		return
	}

	if err := p.addFn(r.Context(), viewID, entityID, x, y); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, p.errorMsg)
		return
	}

	location := sharedAPI.BuildSubResourceLink(sharedAPI.ResourcePath("/views"), sharedAPI.ResourceID(viewID), sharedAPI.ResourcePath(p.subResourcePath))
	sharedAPI.RespondCreatedNoBody(w, location)
}

type updatePositionParams struct {
	elementIDParam string
	updateFn       func(ctx context.Context, viewID, elementID string, x, y float64) error
	errorMsg       string
}

func (h *ViewHandlers) handleUpdatePosition(w http.ResponseWriter, r *http.Request, p updatePositionParams) {
	viewID := sharedAPI.GetPathParam(r, "id")
	elementID := sharedAPI.GetPathParam(r, p.elementIDParam)

	if !h.checkViewEditPermission(w, r, viewID) {
		return
	}

	req, ok := sharedAPI.DecodeRequestOrFail[UpdatePositionRequest](w, r)
	if !ok {
		return
	}

	if err := p.updateFn(r.Context(), viewID, elementID, req.X, req.Y); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, p.errorMsg)
		return
	}

	sharedAPI.RespondNoContent(w)
}

type CreateViewRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
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

type RenameViewRequest struct {
	Name string `json:"name"`
}

type UpdateEdgeTypeRequest struct {
	EdgeType string `json:"edgeType"`
}

type UpdateLayoutDirectionRequest struct {
	LayoutDirection string `json:"layoutDirection"`
}

type UpdateColorSchemeRequest struct {
	ColorScheme string `json:"colorScheme"`
}

type ChangeVisibilityRequest struct {
	IsPrivate bool `json:"isPrivate"`
}

// CreateView godoc
// @Summary Create a new architecture view
// @Description Creates a new architecture view for organizing components
// @Tags views
// @Accept json
// @Produce json
// @Param view body CreateViewRequest true "View data"
// @Success 201 {object} readmodels.ArchitectureViewDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views [post]
func (h *ViewHandlers) CreateView(w http.ResponseWriter, r *http.Request) {
	req, ok := sharedAPI.DecodeRequestOrFail[CreateViewRequest](w, r)
	if !ok {
		return
	}

	if _, err := valueobjects.NewViewName(req.Name); err != nil {
		h.errorHandler.HandleValidationError(w, err)
		return
	}

	cmd := &commands.CreateView{
		Name:        req.Name,
		Description: req.Description,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create view")
		return
	}

	location := sharedAPI.BuildResourceLink(sharedAPI.ResourcePath("/views"), sharedAPI.ResourceID(result.CreatedID))
	view, err := h.readModel.GetByID(r.Context(), result.CreatedID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created view")
		return
	}

	if view == nil {
		sharedAPI.RespondCreated(w, location, map[string]string{
			"id":      result.CreatedID,
			"message": "View created, processing",
		})
		return
	}

	view.Links = h.buildViewLinks(r, view)
	sharedAPI.RespondCreated(w, location, view)
}

// GetAllViews godoc
// @Summary Get all architecture views
// @Description Retrieves all architecture views
// @Tags views
// @Produce json
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_architectureviews_application_readmodels.ArchitectureViewDTO}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views [get]
func (h *ViewHandlers) GetAllViews(w http.ResponseWriter, r *http.Request) {
	views, err := h.readModel.GetAll(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve views")
		return
	}

	for i := range views {
		views[i].Links = h.buildViewLinks(r, &views[i])
		h.addElementLinks(&views[i], h.canEditView(r, &views[i]))
	}

	links := sharedAPI.Links{
		"self": sharedAPI.NewLink("/api/v1/views", "GET"),
	}

	sharedAPI.RespondCollection(w, http.StatusOK, views, links)
}

// GetViewByID godoc
// @Summary Get an architecture view by ID
// @Description Retrieves a specific architecture view by its ID with all component positions
// @Tags views
// @Produce json
// @Param id path string true "View ID"
// @Success 200 {object} readmodels.ArchitectureViewDTO
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id} [get]
func (h *ViewHandlers) GetViewByID(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	view, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		h.errorHandler.HandleError(w, err, "Failed to retrieve view")
		return
	}

	if view == nil {
		h.errorHandler.HandleNotFound(w, "View")
		return
	}

	view.Links = h.buildViewLinks(r, view)
	h.addElementLinks(view, h.canEditView(r, view))

	sharedAPI.RespondJSON(w, http.StatusOK, view)
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
func (h *ViewHandlers) AddComponentToView(w http.ResponseWriter, r *http.Request) {
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
func (h *ViewHandlers) UpdateComponentPosition(w http.ResponseWriter, r *http.Request) {
	viewID := sharedAPI.GetPathParam(r, "id")
	componentID := sharedAPI.GetPathParam(r, "componentId")

	if !h.checkViewEditPermission(w, r, viewID) {
		return
	}

	var req UpdatePositionRequest

	h.decodeValidateAndDispatch(w, r, &req, nil, func() cqrs.Command {
		return commands.UpdateComponentPosition{
			ViewID:      viewID,
			ComponentID: componentID,
			X:           req.X,
			Y:           req.Y,
		}
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
func (h *ViewHandlers) UpdateMultiplePositions(w http.ResponseWriter, r *http.Request) {
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

// RenameView godoc
// @Summary Rename an architecture view
// @Description Renames an architecture view
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param view body RenameViewRequest true "New view name"
// @Success 204
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/name [patch]
func (h *ViewHandlers) RenameView(w http.ResponseWriter, r *http.Request) {
	var req RenameViewRequest
	viewID := sharedAPI.GetPathParam(r, "id")

	h.handleViewUpdate(w, r,
		func() error { return sharedAPI.DecodeRequestInto(r, &req) },
		func() error { _, err := valueobjects.NewViewName(req.Name); return err },
		func() cqrs.Command { return &commands.RenameView{ViewID: viewID, NewName: req.Name} },
	)
}

// DeleteView godoc
// @Summary Delete an architecture view
// @Description Deletes an architecture view (cannot delete default view)
// @Tags views
// @Produce json
// @Param id path string true "View ID"
// @Success 204
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse "Cannot delete default view"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id} [delete]
func (h *ViewHandlers) DeleteView(w http.ResponseWriter, r *http.Request) {
	viewID := sharedAPI.GetPathParam(r, "id")

	if !h.checkViewEditPermission(w, r, viewID) {
		return
	}

	h.dispatchCommand(w, r, &commands.DeleteView{
		ViewID: viewID,
	})
}

// ChangeVisibility godoc
// @Summary Change view visibility
// @Description Toggle view between private and public
// @Tags views
// @Accept json
// @Param id path string true "View ID"
// @Param request body ChangeVisibilityRequest true "Visibility request"
// @Success 204
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/visibility [patch]
func (h *ViewHandlers) ChangeVisibility(w http.ResponseWriter, r *http.Request) {
	viewID := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[ChangeVisibilityRequest](w, r)
	if !ok {
		return
	}

	h.dispatchCommand(w, r, &commands.ChangeViewVisibility{
		ViewID:    viewID,
		IsPrivate: req.IsPrivate,
	})
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
func (h *ViewHandlers) RemoveComponentFromView(w http.ResponseWriter, r *http.Request) {
	viewID := sharedAPI.GetPathParam(r, "id")
	componentID := sharedAPI.GetPathParam(r, "componentId")

	if !h.checkViewEditPermission(w, r, viewID) {
		return
	}

	h.dispatchCommand(w, r, &commands.RemoveComponentFromView{
		ViewID:      viewID,
		ComponentID: componentID,
	})
}

// SetDefaultView godoc
// @Summary Set a view as the default view
// @Description Sets an architecture view as the default view
// @Tags views
// @Produce json
// @Param id path string true "View ID"
// @Success 204
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/default [put]
func (h *ViewHandlers) SetDefaultView(w http.ResponseWriter, r *http.Request) {
	viewID := sharedAPI.GetPathParam(r, "id")

	if !h.checkViewEditPermission(w, r, viewID) {
		return
	}

	h.dispatchCommand(w, r, &commands.SetDefaultView{
		ViewID: viewID,
	})
}

// UpdateEdgeType godoc
// @Summary Update edge type for a view
// @Description Updates the edge rendering style for an architecture view
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param edgeType body UpdateEdgeTypeRequest true "Edge type update"
// @Success 204
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/edge-type [patch]
func (h *ViewHandlers) UpdateEdgeType(w http.ResponseWriter, r *http.Request) {
	var req UpdateEdgeTypeRequest
	var edgeType valueobjects.EdgeType
	viewID := sharedAPI.GetPathParam(r, "id")

	h.handleViewUpdate(w, r,
		func() error { return sharedAPI.DecodeRequestInto(r, &req) },
		func() error { var err error; edgeType, err = valueobjects.NewEdgeType(req.EdgeType); return err },
		func() cqrs.Command { return &commands.UpdateViewEdgeType{ViewID: viewID, EdgeType: edgeType.String()} },
	)
}

// UpdateLayoutDirection godoc
// @Summary Update layout direction for a view
// @Description Updates the auto-layout direction for an architecture view
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param layoutDirection body UpdateLayoutDirectionRequest true "Layout direction update"
// @Success 204
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/layout-direction [patch]
func (h *ViewHandlers) UpdateLayoutDirection(w http.ResponseWriter, r *http.Request) {
	var req UpdateLayoutDirectionRequest
	var layoutDir valueobjects.LayoutDirection
	viewID := sharedAPI.GetPathParam(r, "id")

	h.handleViewUpdate(w, r,
		func() error { return sharedAPI.DecodeRequestInto(r, &req) },
		func() error { var err error; layoutDir, err = valueobjects.NewLayoutDirection(req.LayoutDirection); return err },
		func() cqrs.Command { return &commands.UpdateViewLayoutDirection{ViewID: viewID, LayoutDirection: layoutDir.String()} },
	)
}

type AddCapabilityRequest struct {
	CapabilityID string  `json:"capabilityId"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
}


// AddCapabilityToView godoc
// @Summary Add a capability to a view
// @Description Adds a capability node to an architecture view at the specified position
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param capability body AddCapabilityRequest true "Capability to add with position"
// @Success 201
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/capabilities [post]
func (h *ViewHandlers) AddCapabilityToView(w http.ResponseWriter, r *http.Request) {
	h.handleAddEntity(w, r,
		func() (string, float64, float64, bool) {
			req, ok := sharedAPI.DecodeRequestOrFail[AddCapabilityRequest](w, r)
			if !ok {
				return "", 0, 0, false
			}
			return req.CapabilityID, req.X, req.Y, true
		},
		addEntityParams{
			entityIDField:   "capabilityId",
			subResourcePath: "/capabilities",
			addFn:           h.layoutRepo.AddCapabilityToView,
			errorMsg:        "Failed to add capability to view",
		})
}

// UpdateCapabilityPosition godoc
// @Summary Update capability position in a view
// @Description Updates the position of a capability node in an architecture view
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param capabilityId path string true "Capability ID"
// @Param position body UpdatePositionRequest true "New position"
// @Success 204
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/capabilities/{capabilityId}/position [patch]
func (h *ViewHandlers) UpdateCapabilityPosition(w http.ResponseWriter, r *http.Request) {
	h.handleUpdatePosition(w, r, updatePositionParams{
		elementIDParam: "capabilityId",
		updateFn:       h.layoutRepo.UpdateCapabilityPosition,
		errorMsg:       "Failed to update capability position",
	})
}

// RemoveCapabilityFromView godoc
// @Summary Remove a capability from a view
// @Description Removes a capability node from an architecture view
// @Tags views
// @Produce json
// @Param id path string true "View ID"
// @Param capabilityId path string true "Capability ID"
// @Success 204
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/capabilities/{capabilityId} [delete]
func (h *ViewHandlers) RemoveCapabilityFromView(w http.ResponseWriter, r *http.Request) {
	h.handleLayoutOp(w, r, layoutOpParams{
		elementIDParam: "capabilityId",
		operation:      h.layoutRepo.RemoveCapabilityFromView,
		errorMsg:       "Failed to remove capability from view",
	})
}

type AddOriginEntityRequest struct {
	OriginEntityID string  `json:"originEntityId"`
	X              float64 `json:"x"`
	Y              float64 `json:"y"`
}

// AddOriginEntityToView godoc
// @Summary Add an origin entity to a view
// @Description Adds an origin entity node to an architecture view at the specified position
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param originEntity body AddOriginEntityRequest true "Origin entity to add with position"
// @Success 201
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/origin-entities [post]
func (h *ViewHandlers) AddOriginEntityToView(w http.ResponseWriter, r *http.Request) {
	h.handleAddEntity(w, r,
		func() (string, float64, float64, bool) {
			req, ok := sharedAPI.DecodeRequestOrFail[AddOriginEntityRequest](w, r)
			if !ok {
				return "", 0, 0, false
			}
			return req.OriginEntityID, req.X, req.Y, true
		},
		addEntityParams{
			entityIDField:   "originEntityId",
			subResourcePath: "/origin-entities",
			addFn:           h.layoutRepo.AddOriginEntityToView,
			errorMsg:        "Failed to add origin entity to view",
		})
}

// UpdateOriginEntityPosition godoc
// @Summary Update origin entity position in a view
// @Description Updates the position of an origin entity node in an architecture view
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param originEntityId path string true "Origin Entity ID"
// @Param position body UpdatePositionRequest true "New position"
// @Success 204
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/origin-entities/{originEntityId}/position [patch]
func (h *ViewHandlers) UpdateOriginEntityPosition(w http.ResponseWriter, r *http.Request) {
	h.handleUpdatePosition(w, r, updatePositionParams{
		elementIDParam: "originEntityId",
		updateFn:       h.layoutRepo.UpdateOriginEntityPosition,
		errorMsg:       "Failed to update origin entity position",
	})
}

// RemoveOriginEntityFromView godoc
// @Summary Remove an origin entity from a view
// @Description Removes an origin entity node from an architecture view
// @Tags views
// @Produce json
// @Param id path string true "View ID"
// @Param originEntityId path string true "Origin Entity ID"
// @Success 204
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/origin-entities/{originEntityId} [delete]
func (h *ViewHandlers) RemoveOriginEntityFromView(w http.ResponseWriter, r *http.Request) {
	h.handleLayoutOp(w, r, layoutOpParams{
		elementIDParam: "originEntityId",
		operation:      h.layoutRepo.RemoveOriginEntityFromView,
		errorMsg:       "Failed to remove origin entity from view",
	})
}

// UpdateColorScheme godoc
// @Summary Update color scheme for a view
// @Description Updates the color scheme for an architecture view. Valid schemes: maturity, classic, custom
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param colorScheme body UpdateColorSchemeRequest true "Color scheme update request"
// @Success 200 {object} object{colorScheme=string,_links=map[string]string}
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/color-scheme [patch]
func (h *ViewHandlers) UpdateColorScheme(w http.ResponseWriter, r *http.Request) {
	viewID := sharedAPI.GetPathParam(r, "id")

	if !h.checkViewEditPermission(w, r, viewID) {
		return
	}

	req, ok := sharedAPI.DecodeRequestOrFail[UpdateColorSchemeRequest](w, r)
	if !ok {
		return
	}

	if _, err := valueobjects.NewColorScheme(req.ColorScheme); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	if _, err := h.commandBus.Dispatch(r.Context(), &commands.UpdateViewColorScheme{
		ViewID:      viewID,
		ColorScheme: req.ColorScheme,
	}); err != nil {
		h.errorHandler.HandleError(w, err, "Failed to update color scheme")
		return
	}

	links := sharedAPI.NewResourceLinks().
		SelfSubResource(sharedAPI.ResourcePath("/views"), sharedAPI.ResourceID(viewID), sharedAPI.ResourcePath("/color-scheme")).
		Related(sharedAPI.LinkRelation("view"), sharedAPI.ResourcePath("/views"), sharedAPI.ResourceID(viewID)).
		Build()

	response := struct {
		ColorScheme string      `json:"colorScheme"`
		Links       types.Links `json:"_links"`
	}{
		ColorScheme: req.ColorScheme,
		Links:       links,
	}

	sharedAPI.RespondJSON(w, http.StatusOK, response)
}

type UpdateElementColorRequest struct {
	Color string `json:"color" example:"#FF5733"`
}

// UpdateComponentColor godoc
// @Summary Update custom color for a component in a view
// @Description Sets a custom hex color for a component when using the custom color scheme
// @Tags views
// @Accept json
// @Param id path string true "View ID"
// @Param componentId path string true "Component ID"
// @Param color body UpdateElementColorRequest true "Color update request with hex color (e.g., #FF5733)"
// @Success 204
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid hex color format"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/components/{componentId}/color [patch]
func (h *ViewHandlers) UpdateComponentColor(w http.ResponseWriter, r *http.Request) {
	h.updateElementColor(w, r, elementParams{
		viewID:      sharedAPI.GetPathParam(r, "id"),
		elementID:   sharedAPI.GetPathParam(r, "componentId"),
		elementType: "component",
	})
}

// UpdateCapabilityColor godoc
// @Summary Update custom color for a capability in a view
// @Description Sets a custom hex color for a capability when using the custom color scheme
// @Tags views
// @Accept json
// @Param id path string true "View ID"
// @Param capabilityId path string true "Capability ID"
// @Param color body UpdateElementColorRequest true "Color update request with hex color (e.g., #FF5733)"
// @Success 204
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid hex color format"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/capabilities/{capabilityId}/color [patch]
func (h *ViewHandlers) UpdateCapabilityColor(w http.ResponseWriter, r *http.Request) {
	h.updateElementColor(w, r, elementParams{
		viewID:      sharedAPI.GetPathParam(r, "id"),
		elementID:   sharedAPI.GetPathParam(r, "capabilityId"),
		elementType: "capability",
	})
}

// ClearComponentColor godoc
// @Summary Clear custom color for a component in a view
// @Description Removes the custom color from a component, returning it to the default color scheme
// @Tags views
// @Param id path string true "View ID"
// @Param componentId path string true "Component ID"
// @Success 204
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/components/{componentId}/color [delete]
func (h *ViewHandlers) ClearComponentColor(w http.ResponseWriter, r *http.Request) {
	h.clearElementColor(w, r, elementParams{
		viewID:      sharedAPI.GetPathParam(r, "id"),
		elementID:   sharedAPI.GetPathParam(r, "componentId"),
		elementType: "component",
	})
}

// ClearCapabilityColor godoc
// @Summary Clear custom color for a capability in a view
// @Description Removes the custom color from a capability, returning it to the default color scheme
// @Tags views
// @Param id path string true "View ID"
// @Param capabilityId path string true "Capability ID"
// @Success 204
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/capabilities/{capabilityId}/color [delete]
func (h *ViewHandlers) ClearCapabilityColor(w http.ResponseWriter, r *http.Request) {
	h.clearElementColor(w, r, elementParams{
		viewID:      sharedAPI.GetPathParam(r, "id"),
		elementID:   sharedAPI.GetPathParam(r, "capabilityId"),
		elementType: "capability",
	})
}

func (h *ViewHandlers) buildViewLinks(r *http.Request, view *readmodels.ArchitectureViewDTO) sharedAPI.Links {
	actor, _ := sharedctx.GetActor(r.Context())
	viewInfo := sharedAPI.ViewInfo{
		ID:          view.ID,
		IsPrivate:   view.IsPrivate,
		IsDefault:   view.IsDefault,
		OwnerUserID: view.OwnerUserID,
	}
	return h.hateoas.ViewLinksForActor(viewInfo, actor)
}

func isOwnerOfView(ownerUserID *string, actorID string) bool {
	return ownerUserID != nil && *ownerUserID == actorID
}

func (h *ViewHandlers) canEditView(r *http.Request, view *readmodels.ArchitectureViewDTO) bool {
	actor, _ := sharedctx.GetActor(r.Context())
	isOwner := view.OwnerUserID != nil && *view.OwnerUserID == actor.ID
	canEditThisView := !view.IsPrivate || isOwner
	return canEditThisView && actor.CanWrite("views")
}

func (h *ViewHandlers) checkViewEditPermission(w http.ResponseWriter, r *http.Request, viewID string) bool {
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

func (h *ViewHandlers) buildElementLinks(viewID, elementType, elementID string, canEdit bool) sharedAPI.Links {
	links := sharedAPI.Links{}
	if !canEdit {
		return links
	}
	basePath := fmt.Sprintf("/api/v1/views/%s/%s/%s", viewID, elementType, elementID)
	links["x-update-color"] = sharedAPI.NewLink(basePath+"/color", "PATCH")
	links["x-clear-color"] = sharedAPI.NewLink(basePath+"/color", "DELETE")
	links["x-update-position"] = sharedAPI.NewLink(basePath+"/position", "PATCH")
	links["x-remove"] = sharedAPI.NewLink(basePath, "DELETE")
	return links
}

func (h *ViewHandlers) addElementLinks(view *readmodels.ArchitectureViewDTO, canEdit bool) {
	for i := range view.Components {
		view.Components[i].Links = h.buildElementLinks(view.ID, "components", view.Components[i].ComponentID, canEdit)
	}

	for i := range view.Capabilities {
		view.Capabilities[i].Links = h.buildElementLinks(view.ID, "capabilities", view.Capabilities[i].CapabilityID, canEdit)
	}

	for i := range view.OriginEntities {
		view.OriginEntities[i].Links = h.buildOriginEntityLinks(view.ID, view.OriginEntities[i].OriginEntityID, canEdit)
	}
}

func (h *ViewHandlers) buildOriginEntityLinks(viewID, originEntityID string, canEdit bool) sharedAPI.Links {
	links := sharedAPI.Links{}
	if !canEdit {
		return links
	}
	basePath := fmt.Sprintf("/api/v1/views/%s/origin-entities/%s", viewID, originEntityID)
	links["x-update-position"] = sharedAPI.NewLink(basePath+"/position", "PATCH")
	links["x-remove"] = sharedAPI.NewLink(basePath, "DELETE")
	return links
}
