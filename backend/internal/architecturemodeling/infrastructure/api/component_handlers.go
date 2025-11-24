package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"github.com/go-chi/chi/v5"
)

type ComponentHandlers struct {
	commandBus cqrs.CommandBus
	readModel  *readmodels.ApplicationComponentReadModel
	hateoas    *sharedAPI.HATEOASLinks
}

func NewComponentHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.ApplicationComponentReadModel,
	hateoas *sharedAPI.HATEOASLinks,
) *ComponentHandlers {
	return &ComponentHandlers{
		commandBus: commandBus,
		readModel:  readModel,
		hateoas:    hateoas,
	}
}

type CreateApplicationComponentRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type UpdateApplicationComponentRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CreateApplicationComponent godoc
// @Summary Create a new application component
// @Description Creates a new application component in the system
// @Tags components
// @Accept json
// @Produce json
// @Param component body CreateApplicationComponentRequest true "Component data"
// @Success 201 {object} readmodels.ApplicationComponentDTO
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /components [post]
func (h *ComponentHandlers) CreateApplicationComponent(w http.ResponseWriter, r *http.Request) {
	var req CreateApplicationComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	_, err := valueobjects.NewComponentName(req.Name)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	cmd := &commands.CreateApplicationComponent{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create component")
		return
	}

	component, err := h.readModel.GetByID(r.Context(), cmd.ID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created component")
		return
	}

	if component == nil {
		location := fmt.Sprintf("/api/v1/components/%s", cmd.ID)
		w.Header().Set("Location", location)
		sharedAPI.RespondJSON(w, http.StatusCreated, map[string]string{
			"id":      cmd.ID,
			"message": "Component created, processing",
		})
		return
	}

	component.Links = h.hateoas.ComponentLinks(component.ID)

	location := fmt.Sprintf("/api/v1/components/%s", component.ID)
	w.Header().Set("Location", location)
	sharedAPI.RespondJSON(w, http.StatusCreated, component)
}

// GetAllComponents godoc
// @Summary Get all application components
// @Description Retrieves all application components with cursor-based pagination
// @Tags components
// @Produce json
// @Param limit query int false "Number of items per page (max 100)" default(50)
// @Param after query string false "Cursor for pagination (opaque token)"
// @Success 200 {object} easi_backend_internal_shared_api.PaginatedResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /components [get]
func (h *ComponentHandlers) GetAllComponents(w http.ResponseWriter, r *http.Request) {
	params := sharedAPI.ParsePaginationParams(r)

	afterCursor, afterTimestamp, err := h.processPaginationCursor(params.After)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid pagination cursor")
		return
	}

	components, hasMore, err := h.readModel.GetAllPaginated(r.Context(), params.Limit, afterCursor, afterTimestamp)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve components")
		return
	}

	h.addHATEOASLinks(components)

	nextCursor := h.generateNextCursor(components, hasMore)
	selfLink := h.buildSelfLink("/api/v1/components", params)

	sharedAPI.RespondPaginated(w, http.StatusOK, components, hasMore, nextCursor, params.Limit, selfLink, "/api/v1/components")
}

func (h *ComponentHandlers) processPaginationCursor(after string) (string, int64, error) {
	if after == "" {
		return "", 0, nil
	}

	cursor, err := sharedAPI.DecodeCursor(after)
	if err != nil {
		return "", 0, err
	}

	if cursor == nil {
		return "", 0, nil
	}

	return cursor.ID, cursor.Timestamp, nil
}

func (h *ComponentHandlers) addHATEOASLinks(components []readmodels.ApplicationComponentDTO) {
	for i := range components {
		components[i].Links = h.hateoas.ComponentLinks(components[i].ID)
	}
}

func (h *ComponentHandlers) generateNextCursor(components []readmodels.ApplicationComponentDTO, hasMore bool) string {
	if !hasMore || len(components) == 0 {
		return ""
	}

	lastComponent := components[len(components)-1]
	return sharedAPI.EncodeCursor(lastComponent.ID, lastComponent.CreatedAt)
}

func (h *ComponentHandlers) buildSelfLink(basePath string, params sharedAPI.PaginationParams) string {
	if params.After == "" {
		return basePath
	}
	return fmt.Sprintf("%s?after=%s&limit=%d", basePath, params.After, params.Limit)
}

// GetComponentByID godoc
// @Summary Get an application component by ID
// @Description Retrieves a specific application component by its ID
// @Tags components
// @Produce json
// @Param id path string true "Component ID"
// @Success 200 {object} readmodels.ApplicationComponentDTO
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /components/{id} [get]
func (h *ComponentHandlers) GetComponentByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	component, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve component")
		return
	}

	if component == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Component not found")
		return
	}

	component.Links = h.hateoas.ComponentLinks(component.ID)

	sharedAPI.RespondJSON(w, http.StatusOK, component)
}

// UpdateApplicationComponent godoc
// @Summary Update an application component
// @Description Updates an existing application component's name and description
// @Tags components
// @Accept json
// @Produce json
// @Param id path string true "Component ID"
// @Param component body UpdateApplicationComponentRequest true "Updated component data"
// @Success 200 {object} readmodels.ApplicationComponentDTO
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /components/{id} [put]
func (h *ComponentHandlers) UpdateApplicationComponent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateApplicationComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	_, err := valueobjects.NewComponentName(req.Name)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	cmd := &commands.UpdateApplicationComponent{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to update component")
		return
	}

	component, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve updated component")
		return
	}

	if component == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Component not found")
		return
	}

	component.Links = h.hateoas.ComponentLinks(component.ID)

	sharedAPI.RespondJSON(w, http.StatusOK, component)
}

func (h *ComponentHandlers) DeleteApplicationComponent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cmd := &commands.DeleteApplicationComponent{
		ID: id,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		if errors.Is(err, repositories.ErrComponentNotFound) {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Component not found")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to delete component")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

