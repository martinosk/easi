package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"github.com/go-chi/chi/v5"
)

// ComponentHandlers handles HTTP requests for application components
type ComponentHandlers struct {
	commandBus cqrs.CommandBus
	readModel  *readmodels.ApplicationComponentReadModel
	hateoas    *sharedAPI.HATEOASLinks
}

// NewComponentHandlers creates a new component handlers instance
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

// CreateApplicationComponentRequest represents the request body for creating a component
type CreateApplicationComponentRequest struct {
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
// @Failure 400 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /components [post]
func (h *ComponentHandlers) CreateApplicationComponent(w http.ResponseWriter, r *http.Request) {
	var req CreateApplicationComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// Validate using domain value objects
	_, err := valueobjects.NewComponentName(req.Name)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	// Create command (pass by reference so handler can set ID)
	cmd := &commands.CreateApplicationComponent{
		Name:        req.Name,
		Description: req.Description,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create component")
		return
	}

	// Retrieve the created component from read model
	// Note: Due to eventual consistency, there might be a slight delay
	component, err := h.readModel.GetByID(r.Context(), cmd.ID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created component")
		return
	}

	if component == nil {
		// If not yet in read model, return minimal response with Location header
		location := fmt.Sprintf("/api/v1/components/%s", cmd.ID)
		w.Header().Set("Location", location)
		sharedAPI.RespondJSON(w, http.StatusCreated, map[string]string{
			"id":      cmd.ID,
			"message": "Component created, processing",
		})
		return
	}

	// Add HATEOAS links
	component.Links = h.hateoas.ComponentLinks(component.ID)

	// Return created resource with Location header
	location := fmt.Sprintf("/api/v1/components/%s", component.ID)
	sharedAPI.RespondCreated(w, location, component, nil)
}

// GetAllComponents godoc
// @Summary Get all application components
// @Description Retrieves all application components with cursor-based pagination
// @Tags components
// @Produce json
// @Param limit query int false "Number of items per page (max 100)" default(50)
// @Param after query string false "Cursor for pagination (opaque token)"
// @Success 200 {object} api.PaginatedResponse
// @Failure 500 {object} api.ErrorResponse
// @Router /components [get]
func (h *ComponentHandlers) GetAllComponents(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	params := sharedAPI.ParsePaginationParams(r)

	// Decode cursor if present
	var afterCursor string
	var afterTimestamp int64
	if params.After != "" {
		cursor, err := sharedAPI.DecodeCursor(params.After)
		if err != nil {
			sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid pagination cursor")
			return
		}
		if cursor != nil {
			afterCursor = cursor.ID
			afterTimestamp = cursor.Timestamp
		}
	}

	// Get paginated components
	components, hasMore, err := h.readModel.GetAllPaginated(r.Context(), params.Limit, afterCursor, afterTimestamp)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve components")
		return
	}

	// Add HATEOAS links to each component
	for i := range components {
		components[i].Links = h.hateoas.ComponentLinks(components[i].ID)
	}

	// Generate next cursor if there are more results
	var nextCursor string
	if hasMore && len(components) > 0 {
		lastComponent := components[len(components)-1]
		nextCursor = sharedAPI.EncodeCursor(lastComponent.ID, lastComponent.CreatedAt)
	}

	// Build self link
	selfLink := fmt.Sprintf("/api/v1/components")
	if params.After != "" {
		selfLink = fmt.Sprintf("/api/v1/components?after=%s&limit=%d", params.After, params.Limit)
	}

	// Respond with paginated data
	sharedAPI.RespondPaginated(w, http.StatusOK, components, hasMore, nextCursor, params.Limit, selfLink, "/api/v1/components")
}

// GetComponentByID godoc
// @Summary Get an application component by ID
// @Description Retrieves a specific application component by its ID
// @Tags components
// @Produce json
// @Param id path string true "Component ID"
// @Success 200 {object} readmodels.ApplicationComponentDTO
// @Failure 404 {object} api.ErrorResponse
// @Failure 500 {object} api.ErrorResponse
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

	// Add HATEOAS links
	component.Links = h.hateoas.ComponentLinks(component.ID)

	sharedAPI.RespondSuccess(w, http.StatusOK, component, nil)
}
