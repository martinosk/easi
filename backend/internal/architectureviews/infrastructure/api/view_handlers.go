package api

import (
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

// ViewHandlers handles HTTP requests for architecture views
type ViewHandlers struct {
	commandBus cqrs.CommandBus
	readModel  *readmodels.ArchitectureViewReadModel
	hateoas    *sharedAPI.HATEOASLinks
}

// NewViewHandlers creates a new view handlers instance
func NewViewHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.ArchitectureViewReadModel,
	hateoas *sharedAPI.HATEOASLinks,
) *ViewHandlers {
	return &ViewHandlers{
		commandBus: commandBus,
		readModel:  readModel,
		hateoas:    hateoas,
	}
}

// CreateViewRequest represents the request body for creating a view
type CreateViewRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// AddComponentRequest represents the request body for adding a component to a view
type AddComponentRequest struct {
	ComponentID string  `json:"componentId"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
}

// UpdatePositionRequest represents the request body for updating component position
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

// RenameViewRequest represents the request body for renaming a view
type RenameViewRequest struct {
	Name string `json:"name"`
}

type UpdateEdgeTypeRequest struct {
	EdgeType string `json:"edgeType"`
}

type UpdateLayoutDirectionRequest struct {
	LayoutDirection string `json:"layoutDirection"`
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
	var req CreateViewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// Validate using domain value objects
	_, err := valueobjects.NewViewName(req.Name)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	// Create command (pass by reference so handler can set ID)
	cmd := &commands.CreateView{
		Name:        req.Name,
		Description: req.Description,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create view")
		return
	}

	// Retrieve the created view from read model
	view, err := h.readModel.GetByID(r.Context(), cmd.ID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created view")
		return
	}

	if view == nil {
		// If not yet in read model, return minimal response with Location header
		location := fmt.Sprintf("/api/v1/views/%s", cmd.ID)
		w.Header().Set("Location", location)
		sharedAPI.RespondJSON(w, http.StatusCreated, map[string]string{
			"id":      cmd.ID,
			"message": "View created, processing",
		})
		return
	}

	// Add HATEOAS links
	view.Links = h.hateoas.ViewLinks(view.ID)

	// Return created resource with Location header
	location := fmt.Sprintf("/api/v1/views/%s", view.ID)
	sharedAPI.RespondCreated(w, location, view, nil)
}

// GetAllViews godoc
// @Summary Get all architecture views
// @Description Retrieves all architecture views
// @Tags views
// @Produce json
// @Success 200 {array} readmodels.ArchitectureViewDTO
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views [get]
func (h *ViewHandlers) GetAllViews(w http.ResponseWriter, r *http.Request) {
	views, err := h.readModel.GetAll(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve views")
		return
	}

	// Add HATEOAS links to each view
	for i := range views {
		views[i].Links = h.hateoas.ViewLinks(views[i].ID)
	}

	sharedAPI.RespondSuccess(w, http.StatusOK, views, nil)
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
	id := chi.URLParam(r, "id")

	view, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve view")
		return
	}

	if view == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "View not found")
		return
	}

	// Add HATEOAS links
	view.Links = h.hateoas.ViewLinks(view.ID)

	sharedAPI.RespondSuccess(w, http.StatusOK, view, nil)
}

// AddComponentToView godoc
// @Summary Add a component to a view
// @Description Adds a component to an architecture view at a specific position
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param component body AddComponentRequest true "Component data"
// @Success 201 {object} sharedAPI.SuccessResponse
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse "Component already in view"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/components [post]
func (h *ViewHandlers) AddComponentToView(w http.ResponseWriter, r *http.Request) {
	viewID := chi.URLParam(r, "id")

	var req AddComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// Create command
	cmd := commands.AddComponentToView{
		ViewID:      viewID,
		ComponentID: req.ComponentID,
		X:           req.X,
		Y:           req.Y,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	// Add HATEOAS links for the created resource
	links := map[string]string{
		"self": fmt.Sprintf("/api/v1/views/%s/components", viewID),
		"view": fmt.Sprintf("/api/v1/views/%s", viewID),
	}

	sharedAPI.RespondCreated(w, fmt.Sprintf("/api/v1/views/%s/components", viewID),
		map[string]string{"message": "Component added to view successfully"}, links)
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
// @Success 200 {object} sharedAPI.SuccessResponse
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse "View or component not found"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/components/{componentId}/position [patch]
func (h *ViewHandlers) UpdateComponentPosition(w http.ResponseWriter, r *http.Request) {
	viewID := chi.URLParam(r, "id")
	componentID := chi.URLParam(r, "componentId")

	var req UpdatePositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// Create command
	cmd := commands.UpdateComponentPosition{
		ViewID:      viewID,
		ComponentID: componentID,
		X:           req.X,
		Y:           req.Y,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	// Add HATEOAS links
	links := map[string]string{
		"self": fmt.Sprintf("/api/v1/views/%s/components/%s/position", viewID, componentID),
		"view": fmt.Sprintf("/api/v1/views/%s", viewID),
	}

	sharedAPI.RespondSuccess(w, http.StatusOK,
		map[string]string{"message": "Component position updated successfully"}, links)
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

	links := map[string]string{
		"self": fmt.Sprintf("/api/v1/views/%s/layout", viewID),
		"view": fmt.Sprintf("/api/v1/views/%s", viewID),
	}

	sharedAPI.RespondSuccess(w, http.StatusOK,
		map[string]string{"message": "Component positions updated successfully"}, links)
}

// RenameView godoc
// @Summary Rename an architecture view
// @Description Renames an architecture view
// @Tags views
// @Accept json
// @Produce json
// @Param id path string true "View ID"
// @Param view body RenameViewRequest true "New view name"
// @Success 200 {object} sharedAPI.SuccessResponse
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/name [patch]
func (h *ViewHandlers) RenameView(w http.ResponseWriter, r *http.Request) {
	viewID := chi.URLParam(r, "id")

	var req RenameViewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// Validate using domain value objects
	_, err := valueobjects.NewViewName(req.Name)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	// Create command
	cmd := &commands.RenameView{
		ViewID:  viewID,
		NewName: req.Name,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	// Add HATEOAS links
	links := map[string]string{
		"self": fmt.Sprintf("/api/v1/views/%s", viewID),
	}

	sharedAPI.RespondSuccess(w, http.StatusOK,
		map[string]string{"message": "View renamed successfully"}, links)
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
	viewID := chi.URLParam(r, "id")

	// Create command
	cmd := &commands.DeleteView{
		ViewID: viewID,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusConflict, err, "")
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
	viewID := chi.URLParam(r, "id")
	componentID := chi.URLParam(r, "componentId")

	// Create command
	cmd := &commands.RemoveComponentFromView{
		ViewID:      viewID,
		ComponentID: componentID,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetDefaultView godoc
// @Summary Set a view as the default view
// @Description Sets an architecture view as the default view
// @Tags views
// @Produce json
// @Param id path string true "View ID"
// @Success 200 {object} sharedAPI.SuccessResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /views/{id}/default [put]
func (h *ViewHandlers) SetDefaultView(w http.ResponseWriter, r *http.Request) {
	viewID := chi.URLParam(r, "id")

	// Create command
	cmd := &commands.SetDefaultView{
		ViewID: viewID,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	// Add HATEOAS links
	links := map[string]string{
		"self": fmt.Sprintf("/api/v1/views/%s", viewID),
	}

	sharedAPI.RespondSuccess(w, http.StatusOK,
		map[string]string{"message": "Default view set successfully"}, links)
}

func (h *ViewHandlers) UpdateEdgeType(w http.ResponseWriter, r *http.Request) {
	viewID := chi.URLParam(r, "id")

	var req UpdateEdgeTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	_, err := valueobjects.NewEdgeType(req.EdgeType)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	cmd := &commands.UpdateViewEdgeType{
		ViewID:   viewID,
		EdgeType: req.EdgeType,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	links := map[string]string{
		"self": fmt.Sprintf("/api/v1/views/%s", viewID),
	}

	sharedAPI.RespondSuccess(w, http.StatusOK,
		map[string]string{"message": "Edge type updated successfully"}, links)
}

func (h *ViewHandlers) UpdateLayoutDirection(w http.ResponseWriter, r *http.Request) {
	viewID := chi.URLParam(r, "id")

	var req UpdateLayoutDirectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	_, err := valueobjects.NewLayoutDirection(req.LayoutDirection)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	cmd := &commands.UpdateViewLayoutDirection{
		ViewID:          viewID,
		LayoutDirection: req.LayoutDirection,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	links := map[string]string{
		"self": fmt.Sprintf("/api/v1/views/%s", viewID),
	}

	sharedAPI.RespondSuccess(w, http.StatusOK,
		map[string]string{"message": "Layout direction updated successfully"}, links)
}
