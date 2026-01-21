package api

import (
	"net/http"
	"time"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type AcquiredEntityHandlers struct {
	commandBus       cqrs.CommandBus
	readModel        *readmodels.AcquiredEntityReadModel
	paginationHelper *sharedAPI.PaginationHelper
	hateoas          *sharedAPI.HATEOASLinks
}

func NewAcquiredEntityHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.AcquiredEntityReadModel,
	hateoas *sharedAPI.HATEOASLinks,
) *AcquiredEntityHandlers {
	return &AcquiredEntityHandlers{
		commandBus:       commandBus,
		readModel:        readModel,
		paginationHelper: sharedAPI.NewPaginationHelper("/api/v1/acquired-entities"),
		hateoas:          hateoas,
	}
}

type CreateAcquiredEntityRequest struct {
	Name              string  `json:"name"`
	AcquisitionDate   *string `json:"acquisitionDate,omitempty"`
	IntegrationStatus string  `json:"integrationStatus,omitempty"`
	Notes             string  `json:"notes,omitempty"`
}

type UpdateAcquiredEntityRequest struct {
	Name              string  `json:"name"`
	AcquisitionDate   *string `json:"acquisitionDate,omitempty"`
	IntegrationStatus string  `json:"integrationStatus,omitempty"`
	Notes             string  `json:"notes,omitempty"`
}

// CreateAcquiredEntity godoc
// @Summary Create a new acquired entity
// @Description Creates a new acquired entity (company/product acquired through M&A)
// @Tags acquired-entities
// @Accept json
// @Produce json
// @Param entity body CreateAcquiredEntityRequest true "Acquired entity data"
// @Success 201 {object} readmodels.AcquiredEntityDTO
// @Failure 400 {object} sharedAPI.ErrorResponse "Validation error"
// @Failure 401 {object} sharedAPI.ErrorResponse "Unauthorized - authentication required"
// @Failure 403 {object} sharedAPI.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /acquired-entities [post]
func (h *AcquiredEntityHandlers) CreateAcquiredEntity(w http.ResponseWriter, r *http.Request) {
	req, ok := sharedAPI.DecodeRequestOrFail[CreateAcquiredEntityRequest](w, r)
	if !ok {
		return
	}

	acquisitionDate, err := parseAcquisitionDate(req.AcquisitionDate)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid acquisition date format (expected YYYY-MM-DD)")
		return
	}

	cmd := &commands.CreateAcquiredEntity{
		Name:              req.Name,
		AcquisitionDate:   acquisitionDate,
		IntegrationStatus: req.IntegrationStatus,
		Notes:             req.Notes,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	location := sharedAPI.BuildResourceLink(sharedAPI.ResourcePath("/acquired-entities"), sharedAPI.ResourceID(result.CreatedID))
	entity, err := h.readModel.GetByID(r.Context(), result.CreatedID)
	if err != nil {
		sharedAPI.HandleErrorWithDefault(w, err, "Failed to retrieve created entity")
		return
	}

	if entity == nil {
		sharedAPI.RespondCreated(w, location, map[string]string{
			"id":      result.CreatedID,
			"message": "Entity created, processing",
		})
		return
	}

	h.enrichWithLinks(r, entity)
	sharedAPI.RespondCreated(w, location, entity)
}

// GetAllAcquiredEntities godoc
// @Summary Get all acquired entities
// @Description Retrieves all acquired entities with cursor-based pagination
// @Tags acquired-entities
// @Produce json
// @Param limit query int false "Number of items per page (max 100)" default(50)
// @Param after query string false "Cursor for pagination"
// @Success 200 {object} sharedAPI.PaginatedResponse{data=[]readmodels.AcquiredEntityDTO}
// @Failure 401 {object} sharedAPI.ErrorResponse "Unauthorized - authentication required"
// @Failure 403 {object} sharedAPI.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /acquired-entities [get]
func (h *AcquiredEntityHandlers) GetAllAcquiredEntities(w http.ResponseWriter, r *http.Request) {
	params := sharedAPI.ParsePaginationParams(r)

	afterID, afterName, err := h.paginationHelper.ProcessNameCursor(params.After)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid pagination cursor")
		return
	}

	entities, hasMore, err := h.readModel.GetAllPaginated(r.Context(), params.Limit, afterID, afterName)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve acquired entities")
		return
	}

	for i := range entities {
		h.enrichWithLinks(r, &entities[i])
	}

	pageables := ConvertAcquiredEntitiesToNamePageable(entities)
	nextCursor := h.paginationHelper.GenerateNextNameCursor(pageables, hasMore)
	selfLink := h.paginationHelper.BuildSelfLink(params)

	sharedAPI.RespondPaginated(w, sharedAPI.PaginatedResponseParams{
		StatusCode: http.StatusOK,
		Data:       entities,
		HasMore:    hasMore,
		NextCursor: nextCursor,
		Limit:      params.Limit,
		SelfLink:   selfLink,
		BaseLink:   "/api/v1/acquired-entities",
	})
}

// GetAcquiredEntityByID godoc
// @Summary Get an acquired entity by ID
// @Description Retrieves a specific acquired entity by its ID
// @Tags acquired-entities
// @Produce json
// @Param id path string true "Acquired entity ID"
// @Success 200 {object} readmodels.AcquiredEntityDTO
// @Failure 401 {object} sharedAPI.ErrorResponse "Unauthorized - authentication required"
// @Failure 403 {object} sharedAPI.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 404 {object} sharedAPI.ErrorResponse "Acquired entity not found"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /acquired-entities/{id} [get]
func (h *AcquiredEntityHandlers) GetAcquiredEntityByID(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	entity, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve acquired entity")
		return
	}

	if entity == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Acquired entity not found")
		return
	}

	h.enrichWithLinks(r, entity)
	sharedAPI.RespondJSON(w, http.StatusOK, entity)
}

// UpdateAcquiredEntity godoc
// @Summary Update an acquired entity
// @Description Updates an existing acquired entity
// @Tags acquired-entities
// @Accept json
// @Produce json
// @Param id path string true "Acquired entity ID"
// @Param entity body UpdateAcquiredEntityRequest true "Updated entity data"
// @Success 200 {object} readmodels.AcquiredEntityDTO
// @Failure 400 {object} sharedAPI.ErrorResponse "Validation error"
// @Failure 401 {object} sharedAPI.ErrorResponse "Unauthorized - authentication required"
// @Failure 403 {object} sharedAPI.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 404 {object} sharedAPI.ErrorResponse "Acquired entity not found"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /acquired-entities/{id} [put]
func (h *AcquiredEntityHandlers) UpdateAcquiredEntity(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[UpdateAcquiredEntityRequest](w, r)
	if !ok {
		return
	}

	acquisitionDate, err := parseAcquisitionDate(req.AcquisitionDate)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid acquisition date format (expected YYYY-MM-DD)")
		return
	}

	cmd := &commands.UpdateAcquiredEntity{
		ID:                id,
		Name:              req.Name,
		AcquisitionDate:   acquisitionDate,
		IntegrationStatus: req.IntegrationStatus,
		Notes:             req.Notes,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	entity, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.HandleErrorWithDefault(w, err, "Failed to retrieve updated entity")
		return
	}

	if entity == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Acquired entity not found")
		return
	}

	h.enrichWithLinks(r, entity)
	sharedAPI.RespondJSON(w, http.StatusOK, entity)
}

// DeleteAcquiredEntity godoc
// @Summary Delete an acquired entity
// @Description Permanently deletes an acquired entity
// @Tags acquired-entities
// @Produce json
// @Param id path string true "Acquired entity ID"
// @Success 204 "Successfully deleted"
// @Failure 401 {object} sharedAPI.ErrorResponse "Unauthorized - authentication required"
// @Failure 403 {object} sharedAPI.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 404 {object} sharedAPI.ErrorResponse "Acquired entity not found"
// @Failure 500 {object} sharedAPI.ErrorResponse "Internal server error"
// @Router /acquired-entities/{id} [delete]
func (h *AcquiredEntityHandlers) DeleteAcquiredEntity(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	cmd := &commands.DeleteAcquiredEntity{
		ID: id,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		sharedAPI.RespondDeleted(w)
	})
}

func parseAcquisitionDate(dateStr *string) (*time.Time, error) {
	if dateStr == nil || *dateStr == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", *dateStr)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (h *AcquiredEntityHandlers) enrichWithLinks(r *http.Request, entity *readmodels.AcquiredEntityDTO) {
	actor, _ := sharedctx.GetActor(r.Context())
	entity.Links = h.hateoas.AcquiredEntityLinksForActor(entity.ID, actor)
}
