package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"easi/backend/internal/architectureviews/application/commands"
	"easi/backend/internal/architectureviews/application/readmodels"
	"easi/backend/internal/architectureviews/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"github.com/go-chi/chi/v5"
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
	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		h.errorHandler.HandleError(w, err, "Failed to execute command")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ViewHandlers) dispatchCommandWithStatus(w http.ResponseWriter, r *http.Request, cmd cqrs.Command, errorStatus int) {
	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, errorStatus, err, "")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type elementParams struct {
	viewID      string
	elementID   string
	elementType string
}

func (h *ViewHandlers) updateElementColor(w http.ResponseWriter, r *http.Request, params elementParams) {
	var req UpdateElementColorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	cmd := &commands.UpdateElementColor{
		ViewID:      params.viewID,
		ElementID:   params.elementID,
		ElementType: params.elementType,
		Color:       req.Color,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ViewHandlers) clearElementColor(w http.ResponseWriter, r *http.Request, params elementParams) {
	cmd := &commands.ClearElementColor{
		ViewID:      params.viewID,
		ElementID:   params.elementID,
		ElementType: params.elementType,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, fmt.Sprintf("Failed to clear %s color", params.elementType))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ViewHandlers) decodeValidateAndDispatch(w http.ResponseWriter, r *http.Request, req interface{}, validate func() error, createCmd func() cqrs.Command) {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
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

type CreateViewRequest struct{
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

// CreateView godoc
// @Summary Create a new architecture view
// @Description Creates a new architecture view for organizing components
// @Tags views
// @Accept json
// @Produce json
// @Param view body CreateViewRequest true "View data"
// @Success 201 {object} readmodels.ArchitectureViewDTO
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /views [post]
func (h *ViewHandlers) CreateView(w http.ResponseWriter, r *http.Request) {
	var req CreateViewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorHandler.HandleValidationError(w, err)
		return
	}

	_, err := valueobjects.NewViewName(req.Name)
	if err != nil {
		h.errorHandler.HandleValidationError(w, err)
		return
	}

	cmd := &commands.CreateView{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create view")
		return
	}

	view, err := h.readModel.GetByID(r.Context(), cmd.ID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created view")
		return
	}

	if view == nil {
		location := fmt.Sprintf("/api/v1/views/%s", cmd.ID)
		w.Header().Set("Location", location)
		sharedAPI.RespondJSON(w, http.StatusCreated, map[string]string{
			"id":      cmd.ID,
			"message": "View created, processing",
		})
		return
	}

	view.Links = h.hateoas.ViewLinks(view.ID)

	location := fmt.Sprintf("/api/v1/views/%s", view.ID)
	w.Header().Set("Location", location)
	sharedAPI.RespondJSON(w, http.StatusCreated, view)
}

// GetAllViews godoc
// @Summary Get all architecture views
// @Description Retrieves all architecture views
// @Tags views
// @Produce json
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]easi_backend_internal_architectureviews_application_readmodels.ArchitectureViewDTO}
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /views [get]
func (h *ViewHandlers) GetAllViews(w http.ResponseWriter, r *http.Request) {
	views, err := h.readModel.GetAll(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve views")
		return
	}

	for i := range views {
		views[i].Links = h.hateoas.ViewLinks(views[i].ID)
		h.addElementLinks(&views[i])
	}

	links := map[string]string{
		"self": "/api/v1/views",
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
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /views/{id} [get]
func (h *ViewHandlers) GetViewByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	view, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		h.errorHandler.HandleError(w, err, "Failed to retrieve view")
		return
	}

	if view == nil {
		h.errorHandler.HandleNotFound(w, "View")
		return
	}

	view.Links = h.hateoas.ViewLinks(view.ID)
	h.addElementLinks(view)

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
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 409 {object} easi_backend_internal_shared_api.ErrorResponse "Component already in view"
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /views/{id}/components [post]
func (h *ViewHandlers) AddComponentToView(w http.ResponseWriter, r *http.Request) {
	viewID := chi.URLParam(r, "id")

	var req AddComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	cmd := commands.AddComponentToView{
		ViewID:      viewID,
		ComponentID: req.ComponentID,
		X:           req.X,
		Y:           req.Y,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	location := fmt.Sprintf("/api/v1/views/%s/components", viewID)
	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusCreated)
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
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse "View or component not found"
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /views/{id}/components/{componentId}/position [patch]
func (h *ViewHandlers) UpdateComponentPosition(w http.ResponseWriter, r *http.Request) {
	viewID := chi.URLParam(r, "id")
	componentID := chi.URLParam(r, "componentId")
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

func (h *ViewHandlers) UpdateMultiplePositions(w http.ResponseWriter, r *http.Request) {
	viewID := chi.URLParam(r, "id")

	var req UpdateMultiplePositionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
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

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /views/{id}/name [patch]
func (h *ViewHandlers) RenameView(w http.ResponseWriter, r *http.Request) {
	viewID := chi.URLParam(r, "id")
	var req RenameViewRequest

	h.decodeValidateAndDispatch(w, r, &req,
		func() error {
			_, err := valueobjects.NewViewName(req.Name)
			return err
		},
		func() cqrs.Command {
			return &commands.RenameView{
				ViewID:  viewID,
				NewName: req.Name,
			}
		},
	)
}

// DeleteView godoc
// @Summary Delete an architecture view
// @Description Deletes an architecture view (cannot delete default view)
// @Tags views
// @Produce json
// @Param id path string true "View ID"
// @Success 204
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 409 {object} easi_backend_internal_shared_api.ErrorResponse "Cannot delete default view"
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /views/{id} [delete]
func (h *ViewHandlers) DeleteView(w http.ResponseWriter, r *http.Request) {
	viewID := chi.URLParam(r, "id")

	h.dispatchCommandWithStatus(w, r, &commands.DeleteView{
		ViewID: viewID,
	}, http.StatusConflict)
}

// RemoveComponentFromView godoc
// @Summary Remove a component from a view
// @Description Removes a component from an architecture view without deleting the component
// @Tags views
// @Produce json
// @Param id path string true "View ID"
// @Param componentId path string true "Component ID"
// @Success 204
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse "View or component not found"
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /views/{id}/components/{componentId} [delete]
func (h *ViewHandlers) RemoveComponentFromView(w http.ResponseWriter, r *http.Request) {
	viewID := chi.URLParam(r, "id")
	componentID := chi.URLParam(r, "componentId")

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
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /views/{id}/default [put]
func (h *ViewHandlers) SetDefaultView(w http.ResponseWriter, r *http.Request) {
	viewID := chi.URLParam(r, "id")

	h.dispatchCommand(w, r, &commands.SetDefaultView{
		ViewID: viewID,
	})
}

func (h *ViewHandlers) UpdateEdgeType(w http.ResponseWriter, r *http.Request) {
	viewID := chi.URLParam(r, "id")
	var req UpdateEdgeTypeRequest

	h.decodeValidateAndDispatch(w, r, &req,
		func() error {
			_, err := valueobjects.NewEdgeType(req.EdgeType)
			return err
		},
		func() cqrs.Command {
			return &commands.UpdateViewEdgeType{
				ViewID:   viewID,
				EdgeType: req.EdgeType,
			}
		},
	)
}

func (h *ViewHandlers) UpdateLayoutDirection(w http.ResponseWriter, r *http.Request) {
	viewID := chi.URLParam(r, "id")
	var req UpdateLayoutDirectionRequest

	h.decodeValidateAndDispatch(w, r, &req,
		func() error {
			_, err := valueobjects.NewLayoutDirection(req.LayoutDirection)
			return err
		},
		func() cqrs.Command {
			return &commands.UpdateViewLayoutDirection{
				ViewID:          viewID,
				LayoutDirection: req.LayoutDirection,
			}
		},
	)
}

type AddCapabilityRequest struct {
	CapabilityID string  `json:"capabilityId"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
}

type UpdateCapabilityPositionRequest struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (h *ViewHandlers) AddCapabilityToView(w http.ResponseWriter, r *http.Request) {
	viewID := chi.URLParam(r, "id")

	var req AddCapabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	if req.CapabilityID == "" {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "capabilityId is required")
		return
	}

	if err := h.layoutRepo.AddCapabilityToView(r.Context(), viewID, req.CapabilityID, req.X, req.Y); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to add capability to view")
		return
	}

	location := fmt.Sprintf("/api/v1/views/%s/capabilities", viewID)
	w.Header().Set("Location", location)
	w.WriteHeader(http.StatusCreated)
}

func (h *ViewHandlers) UpdateCapabilityPosition(w http.ResponseWriter, r *http.Request) {
	viewID := chi.URLParam(r, "id")
	capabilityID := chi.URLParam(r, "capabilityId")

	var req UpdateCapabilityPositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	if err := h.layoutRepo.UpdateCapabilityPosition(r.Context(), viewID, capabilityID, req.X, req.Y); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to update capability position")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ViewHandlers) RemoveCapabilityFromView(w http.ResponseWriter, r *http.Request) {
	viewID := chi.URLParam(r, "id")
	capabilityID := chi.URLParam(r, "capabilityId")

	if err := h.layoutRepo.RemoveCapabilityFromView(r.Context(), viewID, capabilityID); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to remove capability from view")
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /views/{id}/color-scheme [patch]
func (h *ViewHandlers) UpdateColorScheme(w http.ResponseWriter, r *http.Request) {
	viewID := chi.URLParam(r, "id")

	var req UpdateColorSchemeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	_, err := valueobjects.NewColorScheme(req.ColorScheme)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	if err := h.commandBus.Dispatch(r.Context(), &commands.UpdateViewColorScheme{
		ViewID:      viewID,
		ColorScheme: req.ColorScheme,
	}); err != nil {
		h.errorHandler.HandleError(w, err, "Failed to update color scheme")
		return
	}

	response := struct {
		ColorScheme string            `json:"colorScheme"`
		Links       map[string]string `json:"_links"`
	}{
		ColorScheme: req.ColorScheme,
		Links: map[string]string{
			"self": fmt.Sprintf("/api/v1/views/%s/color-scheme", viewID),
			"view": fmt.Sprintf("/api/v1/views/%s", viewID),
		},
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
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse "Invalid hex color format"
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /views/{id}/components/{componentId}/color [patch]
func (h *ViewHandlers) UpdateComponentColor(w http.ResponseWriter, r *http.Request) {
	h.updateElementColor(w, r, elementParams{
		viewID:      chi.URLParam(r, "id"),
		elementID:   chi.URLParam(r, "componentId"),
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
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse "Invalid hex color format"
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /views/{id}/capabilities/{capabilityId}/color [patch]
func (h *ViewHandlers) UpdateCapabilityColor(w http.ResponseWriter, r *http.Request) {
	h.updateElementColor(w, r, elementParams{
		viewID:      chi.URLParam(r, "id"),
		elementID:   chi.URLParam(r, "capabilityId"),
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
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /views/{id}/components/{componentId}/color [delete]
func (h *ViewHandlers) ClearComponentColor(w http.ResponseWriter, r *http.Request) {
	h.clearElementColor(w, r, elementParams{
		viewID:      chi.URLParam(r, "id"),
		elementID:   chi.URLParam(r, "componentId"),
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
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /views/{id}/capabilities/{capabilityId}/color [delete]
func (h *ViewHandlers) ClearCapabilityColor(w http.ResponseWriter, r *http.Request) {
	h.clearElementColor(w, r, elementParams{
		viewID:      chi.URLParam(r, "id"),
		elementID:   chi.URLParam(r, "capabilityId"),
		elementType: "capability",
	})
}

func (h *ViewHandlers) addElementLinks(view *readmodels.ArchitectureViewDTO) {
	for i := range view.Components {
		componentID := view.Components[i].ComponentID
		view.Components[i].Links = map[string]string{
			"updateColor": fmt.Sprintf("/api/v1/views/%s/components/%s/color", view.ID, componentID),
			"clearColor":  fmt.Sprintf("/api/v1/views/%s/components/%s/color", view.ID, componentID),
		}
	}

	for i := range view.Capabilities {
		capabilityID := view.Capabilities[i].CapabilityID
		view.Capabilities[i].Links = map[string]string{
			"updateColor": fmt.Sprintf("/api/v1/views/%s/capabilities/%s/color", view.ID, capabilityID),
			"clearColor":  fmt.Sprintf("/api/v1/views/%s/capabilities/%s/color", view.ID, capabilityID),
		}
	}
}
