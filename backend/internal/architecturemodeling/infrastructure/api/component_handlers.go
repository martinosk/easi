package api

import (
	"net/http"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type ComponentHandlers struct {
	commandBus       cqrs.CommandBus
	readModel        *readmodels.ApplicationComponentReadModel
	hateoas          *sharedAPI.HATEOASLinks
	paginationHelper *sharedAPI.PaginationHelper
}

func NewComponentHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.ApplicationComponentReadModel,
	hateoas *sharedAPI.HATEOASLinks,
) *ComponentHandlers {
	return &ComponentHandlers{
		commandBus:       commandBus,
		readModel:        readModel,
		hateoas:          hateoas,
		paginationHelper: sharedAPI.NewPaginationHelper("/api/v1/components"),
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
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components [post]
func (h *ComponentHandlers) CreateApplicationComponent(w http.ResponseWriter, r *http.Request) {
	req, ok := sharedAPI.DecodeRequestOrFail[CreateApplicationComponentRequest](w, r)
	if !ok {
		return
	}

	if !h.validateComponentName(w, req.Name) {
		return
	}

	cmd := &commands.CreateApplicationComponent{
		Name:        req.Name,
		Description: req.Description,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to create component")
		return
	}

	location := sharedAPI.BuildResourceLink(sharedAPI.ResourcePath("/components"), sharedAPI.ResourceID(result.CreatedID))
	component, err := h.readModel.GetByID(r.Context(), result.CreatedID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created component")
		return
	}

	if component == nil {
		sharedAPI.RespondCreated(w, location, map[string]string{
			"id":      result.CreatedID,
			"message": "Component created, processing",
		})
		return
	}

	h.enrichWithLinks(r, component)
	sharedAPI.RespondCreated(w, location, component)
}

// GetAllComponents godoc
// @Summary Get all application components
// @Description Retrieves all application components with cursor-based pagination
// @Tags components
// @Produce json
// @Param limit query int false "Number of items per page (max 100)" default(50)
// @Param after query string false "Cursor for pagination (opaque token)"
// @Success 200 {object} easi_backend_internal_shared_api.PaginatedResponse{data=[]readmodels.ApplicationComponentDTO}
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components [get]
func (h *ComponentHandlers) GetAllComponents(w http.ResponseWriter, r *http.Request) {
	params := sharedAPI.ParsePaginationParams(r)

	afterID, afterName, err := h.paginationHelper.ProcessNameCursor(params.After)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid pagination cursor")
		return
	}

	components, hasMore, err := h.readModel.GetAllPaginated(r.Context(), params.Limit, afterID, afterName)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve components")
		return
	}

	for i := range components {
		h.enrichWithLinks(r, &components[i])
	}

	pageables := ConvertToNamePageable(components)
	nextCursor := h.paginationHelper.GenerateNextNameCursor(pageables, hasMore)
	selfLink := h.paginationHelper.BuildSelfLink(params)

	sharedAPI.RespondPaginated(w, sharedAPI.PaginatedResponseParams{
		StatusCode: http.StatusOK,
		Data:       components,
		HasMore:    hasMore,
		NextCursor: nextCursor,
		Limit:      params.Limit,
		SelfLink:   selfLink,
		BaseLink:   "/api/v1/components",
	})
}

// GetComponentByID godoc
// @Summary Get an application component by ID
// @Description Retrieves a specific application component by its ID
// @Tags components
// @Produce json
// @Param id path string true "Component ID"
// @Success 200 {object} readmodels.ApplicationComponentDTO
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{id} [get]
func (h *ComponentHandlers) GetComponentByID(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	component, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve component")
		return
	}

	if component == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Component not found")
		return
	}

	h.enrichWithLinks(r, component)
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
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{id} [put]
func (h *ComponentHandlers) UpdateApplicationComponent(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[UpdateApplicationComponentRequest](w, r)
	if !ok {
		return
	}

	if !h.validateComponentName(w, req.Name) {
		return
	}

	cmd := &commands.UpdateApplicationComponent{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
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

	h.enrichWithLinks(r, component)
	sharedAPI.RespondJSON(w, http.StatusOK, component)
}

// DeleteApplicationComponent godoc
// @Summary Delete an application component
// @Description Permanently deletes an application component from the model
// @Tags components
// @Produce json
// @Param id path string true "Component ID"
// @Success 204
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{id} [delete]
func (h *ComponentHandlers) DeleteApplicationComponent(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	cmd := &commands.DeleteApplicationComponent{
		ID: id,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		sharedAPI.RespondDeleted(w)
	})
}

func (h *ComponentHandlers) validateComponentName(w http.ResponseWriter, name string) bool {
	if _, err := valueobjects.NewComponentName(name); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return false
	}
	return true
}

func (h *ComponentHandlers) enrichWithLinks(r *http.Request, component *readmodels.ApplicationComponentDTO) {
	actor, _ := sharedctx.GetActor(r.Context())
	component.Links = h.hateoas.ComponentLinksForActor(component.ID, actor)
	for i := range component.Experts {
		e := component.Experts[i]
		component.Experts[i].Links = h.hateoas.ComponentExpertLinksForActor(sharedAPI.ExpertParams{
			ResourcePath: "/components/" + component.ID,
			ExpertName:   e.Name,
			ExpertRole:   e.Role,
			ContactInfo:  e.Contact,
		}, actor)
	}
}
