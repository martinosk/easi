package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/viewlayouts/domain"
	"easi/backend/internal/viewlayouts/domain/aggregates"
	"easi/backend/internal/viewlayouts/domain/valueobjects"
	"easi/backend/internal/viewlayouts/infrastructure/repositories"

	"github.com/go-chi/chi/v5"
)

type LayoutHandlers struct {
	repo    domain.LayoutContainerRepository
	hateoas *sharedAPI.HATEOASLinks
}

func NewLayoutHandlers(repo domain.LayoutContainerRepository, hateoas *sharedAPI.HATEOASLinks) *LayoutHandlers {
	return &LayoutHandlers{
		repo:    repo,
		hateoas: hateoas,
	}
}

type UpsertLayoutRequest struct {
	Preferences map[string]interface{} `json:"preferences,omitempty"`
}

type UpdatePreferencesRequest struct {
	Preferences map[string]interface{} `json:"preferences"`
}

type ElementPositionInput struct {
	X           float64  `json:"x"`
	Y           float64  `json:"y"`
	Width       *float64 `json:"width,omitempty"`
	Height      *float64 `json:"height,omitempty"`
	CustomColor *string  `json:"customColor,omitempty"`
	SortOrder   *int     `json:"sortOrder,omitempty"`
}

type BatchUpdateItem struct {
	ElementID   string   `json:"elementId"`
	X           float64  `json:"x"`
	Y           float64  `json:"y"`
	Width       *float64 `json:"width,omitempty"`
	Height      *float64 `json:"height,omitempty"`
	CustomColor *string  `json:"customColor,omitempty"`
	SortOrder   *int     `json:"sortOrder,omitempty"`
}

type BatchUpdateRequest struct {
	Updates []BatchUpdateItem `json:"updates"`
}

type ElementPositionDTO struct {
	ElementID   string             `json:"elementId"`
	X           float64            `json:"x"`
	Y           float64            `json:"y"`
	Width       *float64           `json:"width,omitempty"`
	Height      *float64           `json:"height,omitempty"`
	CustomColor *string            `json:"customColor,omitempty"`
	SortOrder   *int               `json:"sortOrder,omitempty"`
	Links       map[string]LinkDTO `json:"_links"`
}

type LinkDTO struct {
	Href   string `json:"href"`
	Method string `json:"method,omitempty"`
}

type LayoutContainerDTO struct {
	ID          string                 `json:"id"`
	ContextType string                 `json:"contextType"`
	ContextRef  string                 `json:"contextRef"`
	Preferences map[string]interface{} `json:"preferences"`
	Elements    []ElementPositionDTO   `json:"elements"`
	Version     int                    `json:"version"`
	CreatedAt   string                 `json:"createdAt"`
	UpdatedAt   string                 `json:"updatedAt"`
	Links       map[string]LinkDTO     `json:"_links"`
}

type LayoutContainerSummaryDTO struct {
	ID          string                 `json:"id"`
	ContextType string                 `json:"contextType"`
	ContextRef  string                 `json:"contextRef"`
	Preferences map[string]interface{} `json:"preferences"`
	Version     int                    `json:"version"`
	Links       map[string]LinkDTO     `json:"_links"`
}

type BatchUpdateResponse struct {
	Updated  int                  `json:"updated"`
	Elements []ElementPositionDTO `json:"elements"`
	Links    map[string]LinkDTO   `json:"_links"`
}

func (h *LayoutHandlers) getPathParams(r *http.Request) (valueobjects.LayoutContextType, valueobjects.ContextRef, error) {
	contextTypeStr := chi.URLParam(r, "contextType")
	contextRefStr := chi.URLParam(r, "contextRef")

	contextType, err := valueobjects.NewLayoutContextType(contextTypeStr)
	if err != nil {
		return "", valueobjects.ContextRef{}, err
	}

	contextRef, err := valueobjects.NewContextRef(contextRefStr)
	if err != nil {
		return "", valueobjects.ContextRef{}, err
	}

	return contextType, contextRef, nil
}

func (h *LayoutHandlers) buildLayoutDTO(container *aggregates.LayoutContainer) LayoutContainerDTO {
	basePath := fmt.Sprintf("/api/v1/layouts/%s/%s",
		container.ContextType().Value(),
		container.ContextRef().Value())

	elements := make([]ElementPositionDTO, 0, len(container.Elements()))
	for _, elem := range container.Elements() {
		elements = append(elements, h.buildElementDTO(elem, basePath))
	}

	return LayoutContainerDTO{
		ID:          container.ID().Value(),
		ContextType: container.ContextType().Value(),
		ContextRef:  container.ContextRef().Value(),
		Preferences: container.Preferences().ToMap(),
		Elements:    elements,
		Version:     container.Version(),
		CreatedAt:   container.CreatedAt().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   container.UpdatedAt().Format("2006-01-02T15:04:05Z"),
		Links: map[string]LinkDTO{
			"self":              {Href: basePath},
			"updatePreferences": {Href: basePath + "/preferences", Method: "PATCH"},
			"batchUpdate":       {Href: basePath + "/elements", Method: "PATCH"},
			"delete":            {Href: basePath, Method: "DELETE"},
		},
	}
}

func (h *LayoutHandlers) buildElementDTO(elem valueobjects.ElementPosition, basePath string) ElementPositionDTO {
	dto := ElementPositionDTO{
		ElementID: elem.ElementID().Value(),
		X:         elem.X(),
		Y:         elem.Y(),
		Width:     elem.Width(),
		Height:    elem.Height(),
		SortOrder: elem.SortOrder(),
		Links: map[string]LinkDTO{
			"self":   {Href: fmt.Sprintf("%s/elements/%s", basePath, elem.ElementID().Value())},
			"layout": {Href: basePath},
			"update": {Href: fmt.Sprintf("%s/elements/%s", basePath, elem.ElementID().Value()), Method: "PUT"},
			"delete": {Href: fmt.Sprintf("%s/elements/%s", basePath, elem.ElementID().Value()), Method: "DELETE"},
		},
	}

	if elem.CustomColor() != nil {
		color := elem.CustomColor().Value()
		dto.CustomColor = &color
	}

	return dto
}

// GetLayout godoc
// @Summary Get a layout container
// @Description Retrieves a layout container with all element positions for a specific context
// @Tags layouts
// @Produce json
// @Param contextType path string true "Context type (e.g., 'view', 'dashboard')"
// @Param contextRef path string true "Context reference ID"
// @Success 200 {object} LayoutContainerDTO "Layout container with elements"
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid path parameters"
// @Failure 404 {object} map[string]interface{} "Layout not found (includes create link)"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /layouts/{contextType}/{contextRef} [get]
func (h *LayoutHandlers) GetLayout(w http.ResponseWriter, r *http.Request) {
	contextType, contextRef, err := h.getPathParams(r)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid path parameters")
		return
	}

	container, err := h.repo.GetByContext(r.Context(), contextType, contextRef)
	if err != nil {
		if errors.Is(err, repositories.ErrContainerNotFound) {
			basePath := fmt.Sprintf("/api/v1/layouts/%s/%s", contextType.Value(), contextRef.Value())
			response := map[string]interface{}{
				"error":   "Not Found",
				"message": "Layout not found",
				"_links": map[string]LinkDTO{
					"create": {Href: basePath, Method: "PUT"},
				},
			}
			sharedAPI.RespondJSON(w, http.StatusNotFound, response)
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve layout")
		return
	}

	dto := h.buildLayoutDTO(container)
	w.Header().Set("ETag", fmt.Sprintf(`"%d"`, container.Version()))
	sharedAPI.RespondJSON(w, http.StatusOK, dto)
}

func (h *LayoutHandlers) createLayout(
	contextType valueobjects.LayoutContextType,
	contextRef valueobjects.ContextRef,
	prefs map[string]interface{},
) (*aggregates.LayoutContainer, error) {
	preferences := valueobjects.NewLayoutPreferences(prefs)
	return aggregates.NewLayoutContainer(contextType, contextRef, preferences)
}

func (h *LayoutHandlers) updateLayout(
	existing *aggregates.LayoutContainer,
	prefs map[string]interface{},
) *aggregates.LayoutContainer {
	newPrefs := valueobjects.NewLayoutPreferences(prefs)
	_ = existing.UpdatePreferences(newPrefs)
	existing.IncrementVersion()
	return existing
}

// UpsertLayout godoc
// @Summary Create or update a layout container
// @Description Creates a new layout container or updates an existing one with preferences
// @Tags layouts
// @Accept json
// @Produce json
// @Param contextType path string true "Context type (e.g., 'view', 'dashboard')"
// @Param contextRef path string true "Context reference ID"
// @Param request body UpsertLayoutRequest true "Layout preferences"
// @Success 200 {object} LayoutContainerDTO "Layout updated"
// @Success 201 {object} LayoutContainerDTO "Layout created"
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid request"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /layouts/{contextType}/{contextRef} [put]
func (h *LayoutHandlers) UpsertLayout(w http.ResponseWriter, r *http.Request) {
	contextType, contextRef, err := h.getPathParams(r)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid path parameters")
		return
	}

	var req UpsertLayoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	existing, err := h.repo.GetByContext(r.Context(), contextType, contextRef)
	isNew := errors.Is(err, repositories.ErrContainerNotFound)

	if err != nil && !isNew {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to check existing layout")
		return
	}

	var container *aggregates.LayoutContainer
	if isNew {
		container, err = h.createLayout(contextType, contextRef, req.Preferences)
		if err != nil {
			sharedAPI.RespondError(w, http.StatusBadRequest, err, "Failed to create layout")
			return
		}
	} else {
		container = h.updateLayout(existing, req.Preferences)
	}

	if err := h.repo.Save(r.Context(), container); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to save layout")
		return
	}

	savedContainer, err := h.repo.GetByContext(r.Context(), contextType, contextRef)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve saved layout")
		return
	}

	dto := h.buildLayoutDTO(savedContainer)
	w.Header().Set("ETag", fmt.Sprintf(`"%d"`, savedContainer.Version()))

	if isNew {
		basePath := fmt.Sprintf("/api/v1/layouts/%s/%s", contextType.Value(), contextRef.Value())
		w.Header().Set("Location", basePath)
		sharedAPI.RespondJSON(w, http.StatusCreated, dto)
	} else {
		sharedAPI.RespondJSON(w, http.StatusOK, dto)
	}
}

// DeleteLayout godoc
// @Summary Delete a layout container
// @Description Deletes a layout container and all its element positions
// @Tags layouts
// @Param contextType path string true "Context type (e.g., 'view', 'dashboard')"
// @Param contextRef path string true "Context reference ID"
// @Success 204 "Layout deleted"
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid path parameters"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /layouts/{contextType}/{contextRef} [delete]
func (h *LayoutHandlers) DeleteLayout(w http.ResponseWriter, r *http.Request) {
	contextType, contextRef, err := h.getPathParams(r)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid path parameters")
		return
	}

	if err := h.repo.DeleteByContextRef(r.Context(), contextType, contextRef); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to delete layout")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdatePreferences godoc
// @Summary Update layout preferences
// @Description Updates the preferences of an existing layout container with optimistic locking
// @Tags layouts
// @Accept json
// @Produce json
// @Param contextType path string true "Context type (e.g., 'view', 'dashboard')"
// @Param contextRef path string true "Context reference ID"
// @Param If-Match header string true "ETag for optimistic locking"
// @Param request body UpdatePreferencesRequest true "Preferences to update"
// @Success 200 {object} LayoutContainerSummaryDTO "Preferences updated"
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid request or ETag format"
// @Failure 404 {object} sharedAPI.ErrorResponse "Layout not found"
// @Failure 412 {object} map[string]interface{} "Version conflict"
// @Failure 428 {object} sharedAPI.ErrorResponse "If-Match header required"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /layouts/{contextType}/{contextRef}/preferences [patch]
func (h *LayoutHandlers) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	contextType, contextRef, err := h.getPathParams(r)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid path parameters")
		return
	}

	ifMatch := r.Header.Get("If-Match")
	if ifMatch == "" {
		sharedAPI.RespondError(w, http.StatusPreconditionRequired, nil, "If-Match header required for optimistic locking")
		return
	}

	expectedVersion, err := parseETag(ifMatch)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid ETag format")
		return
	}

	var req UpdatePreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	container, err := h.repo.GetByContext(r.Context(), contextType, contextRef)
	if err != nil {
		if errors.Is(err, repositories.ErrContainerNotFound) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Layout not found")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve layout")
		return
	}

	if container.Version() != expectedVersion {
		basePath := fmt.Sprintf("/api/v1/layouts/%s/%s", contextType.Value(), contextRef.Value())
		response := map[string]interface{}{
			"error":          "Precondition Failed",
			"message":        "Version conflict: layout was modified",
			"currentVersion": container.Version(),
			"_links": map[string]LinkDTO{
				"current": {Href: basePath},
			},
		}
		sharedAPI.RespondJSON(w, http.StatusPreconditionFailed, response)
		return
	}

	newPrefs := container.Preferences().WithUpdated(req.Preferences)
	_ = container.UpdatePreferences(newPrefs)
	container.IncrementVersion()

	if err := h.repo.Save(r.Context(), container); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to save preferences")
		return
	}

	basePath := fmt.Sprintf("/api/v1/layouts/%s/%s", contextType.Value(), contextRef.Value())
	summary := LayoutContainerSummaryDTO{
		ID:          container.ID().Value(),
		ContextType: container.ContextType().Value(),
		ContextRef:  container.ContextRef().Value(),
		Preferences: container.Preferences().ToMap(),
		Version:     container.Version(),
		Links: map[string]LinkDTO{
			"self": {Href: basePath},
		},
	}

	w.Header().Set("ETag", fmt.Sprintf(`"%d"`, container.Version()))
	sharedAPI.RespondJSON(w, http.StatusOK, summary)
}

// UpsertElementPosition godoc
// @Summary Create or update an element position
// @Description Creates a new element position or updates an existing one within a layout
// @Tags layouts
// @Accept json
// @Produce json
// @Param contextType path string true "Context type (e.g., 'view', 'dashboard')"
// @Param contextRef path string true "Context reference ID"
// @Param elementId path string true "Element ID"
// @Param request body ElementPositionInput true "Element position data"
// @Success 200 {object} ElementPositionDTO "Element position updated"
// @Success 201 {object} ElementPositionDTO "Element position created"
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid request"
// @Failure 404 {object} sharedAPI.ErrorResponse "Layout not found"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /layouts/{contextType}/{contextRef}/elements/{elementId} [put]
func (h *LayoutHandlers) UpsertElementPosition(w http.ResponseWriter, r *http.Request) {
	contextType, contextRef, err := h.getPathParams(r)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid path parameters")
		return
	}

	elementIDStr := chi.URLParam(r, "elementId")
	elementID, err := valueobjects.NewElementID(elementIDStr)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid element ID")
		return
	}

	var req ElementPositionInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	container, err := h.repo.GetByContext(r.Context(), contextType, contextRef)
	if err != nil {
		if errors.Is(err, repositories.ErrContainerNotFound) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Layout not found")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve layout")
		return
	}

	existingPos := container.GetElement(elementID)
	isNew := existingPos == nil

	var customColor *valueobjects.HexColor
	if req.CustomColor != nil {
		color, err := valueobjects.NewHexColor(*req.CustomColor)
		if err != nil {
			sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid custom color")
			return
		}
		customColor = &color
	}

	position, _ := valueobjects.NewElementPositionWithOptions(
		elementID, req.X, req.Y,
		req.Width, req.Height, customColor, req.SortOrder,
	)

	if err := h.repo.UpsertElementPosition(r.Context(), container.ID(), position); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to save element position")
		return
	}

	basePath := fmt.Sprintf("/api/v1/layouts/%s/%s", contextType.Value(), contextRef.Value())
	dto := h.buildElementDTO(position, basePath)

	if isNew {
		w.Header().Set("Location", dto.Links["self"].Href)
		sharedAPI.RespondJSON(w, http.StatusCreated, dto)
	} else {
		sharedAPI.RespondJSON(w, http.StatusOK, dto)
	}
}

// DeleteElementPosition godoc
// @Summary Delete an element position
// @Description Removes an element position from a layout
// @Tags layouts
// @Param contextType path string true "Context type (e.g., 'view', 'dashboard')"
// @Param contextRef path string true "Context reference ID"
// @Param elementId path string true "Element ID"
// @Success 204 "Element position deleted"
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid path parameters"
// @Failure 404 {object} sharedAPI.ErrorResponse "Layout not found"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /layouts/{contextType}/{contextRef}/elements/{elementId} [delete]
func (h *LayoutHandlers) DeleteElementPosition(w http.ResponseWriter, r *http.Request) {
	contextType, contextRef, err := h.getPathParams(r)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid path parameters")
		return
	}

	elementIDStr := chi.URLParam(r, "elementId")
	elementID, err := valueobjects.NewElementID(elementIDStr)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid element ID")
		return
	}

	container, err := h.repo.GetByContext(r.Context(), contextType, contextRef)
	if err != nil {
		if errors.Is(err, repositories.ErrContainerNotFound) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Layout not found")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve layout")
		return
	}

	if err := h.repo.DeleteElementPosition(r.Context(), container.ID(), elementID); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to delete element position")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func transformToPosition(update BatchUpdateItem) (valueobjects.ElementPosition, error) {
	elementID, err := valueobjects.NewElementID(update.ElementID)
	if err != nil {
		return valueobjects.ElementPosition{}, fmt.Errorf("invalid element ID %s: %w", update.ElementID, err)
	}

	var customColor *valueobjects.HexColor
	if update.CustomColor != nil {
		color, err := valueobjects.NewHexColor(*update.CustomColor)
		if err != nil {
			return valueobjects.ElementPosition{}, fmt.Errorf("invalid color for element %s: %w", update.ElementID, err)
		}
		customColor = &color
	}

	position, _ := valueobjects.NewElementPositionWithOptions(
		elementID, update.X, update.Y,
		update.Width, update.Height, customColor, update.SortOrder,
	)
	return position, nil
}

func (h *LayoutHandlers) buildBatchResponse(positions []valueobjects.ElementPosition, basePath string) BatchUpdateResponse {
	elements := make([]ElementPositionDTO, 0, len(positions))
	for _, pos := range positions {
		elements = append(elements, h.buildElementDTO(pos, basePath))
	}

	return BatchUpdateResponse{
		Updated:  len(positions),
		Elements: elements,
		Links: map[string]LinkDTO{
			"self":   {Href: basePath + "/elements"},
			"layout": {Href: basePath},
		},
	}
}

// BatchUpdateElements godoc
// @Summary Batch update element positions
// @Description Updates multiple element positions in a single request
// @Tags layouts
// @Accept json
// @Produce json
// @Param contextType path string true "Context type (e.g., 'view', 'dashboard')"
// @Param contextRef path string true "Context reference ID"
// @Param request body BatchUpdateRequest true "Batch update items"
// @Success 200 {object} BatchUpdateResponse "Elements updated"
// @Failure 400 {object} sharedAPI.ErrorResponse "Invalid request"
// @Failure 404 {object} sharedAPI.ErrorResponse "Layout not found"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /layouts/{contextType}/{contextRef}/elements [patch]
func (h *LayoutHandlers) BatchUpdateElements(w http.ResponseWriter, r *http.Request) {
	contextType, contextRef, err := h.getPathParams(r)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid path parameters")
		return
	}

	var req BatchUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	if len(req.Updates) == 0 {
		sharedAPI.RespondError(w, http.StatusBadRequest, nil, "At least one update is required")
		return
	}

	container, err := h.repo.GetByContext(r.Context(), contextType, contextRef)
	if err != nil {
		if errors.Is(err, repositories.ErrContainerNotFound) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Layout not found")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve layout")
		return
	}

	positions := make([]valueobjects.ElementPosition, 0, len(req.Updates))
	for _, update := range req.Updates {
		position, err := transformToPosition(update)
		if err != nil {
			sharedAPI.RespondError(w, http.StatusBadRequest, err, err.Error())
			return
		}
		positions = append(positions, position)
	}

	if err := h.repo.BatchUpdatePositions(r.Context(), container.ID(), positions); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to batch update positions")
		return
	}

	basePath := fmt.Sprintf("/api/v1/layouts/%s/%s", contextType.Value(), contextRef.Value())
	sharedAPI.RespondJSON(w, http.StatusOK, h.buildBatchResponse(positions, basePath))
}

func parseETag(etag string) (int, error) {
	if len(etag) < 3 || etag[0] != '"' || etag[len(etag)-1] != '"' {
		return 0, errors.New("invalid ETag format")
	}
	versionStr := etag[1 : len(etag)-1]
	return strconv.Atoi(versionStr)
}
