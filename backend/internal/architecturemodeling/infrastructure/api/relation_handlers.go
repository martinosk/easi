package api

import (
	"context"
	"fmt"
	"net/http"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
)

// RelationHandlers handles HTTP requests for component relations
type RelationHandlers struct {
	commandBus cqrs.CommandBus
	readModel  *readmodels.ComponentRelationReadModel
	hateoas    *sharedAPI.HATEOASLinks
}

// NewRelationHandlers creates a new relation handlers instance
func NewRelationHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.ComponentRelationReadModel,
	hateoas *sharedAPI.HATEOASLinks,
) *RelationHandlers {
	return &RelationHandlers{
		commandBus: commandBus,
		readModel:  readModel,
		hateoas:    hateoas,
	}
}

// CreateComponentRelationRequest represents the request body for creating a relation
type CreateComponentRelationRequest struct {
	SourceComponentID string `json:"sourceComponentId"`
	TargetComponentID string `json:"targetComponentId"`
	RelationType      string `json:"relationType"`
	Name              string `json:"name,omitempty"`
	Description       string `json:"description,omitempty"`
}

// UpdateComponentRelationRequest represents the request body for updating a relation
type UpdateComponentRelationRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// CreateComponentRelation godoc
// @Summary Create a new component relation
// @Description Creates a new relation between two application components
// @Tags relations
// @Accept json
// @Produce json
// @Param relation body CreateComponentRelationRequest true "Relation data"
// @Success 201 {object} readmodels.ComponentRelationDTO
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /relations [post]
func (h *RelationHandlers) CreateComponentRelation(w http.ResponseWriter, r *http.Request) {
	req, ok := sharedAPI.DecodeRequestOrFail[CreateComponentRelationRequest](w, r)
	if !ok {
		return
	}

	if _, err := valueobjects.NewComponentIDFromString(req.SourceComponentID); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid source component ID")
		return
	}

	if _, err := valueobjects.NewComponentIDFromString(req.TargetComponentID); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid target component ID")
		return
	}

	if _, err := valueobjects.NewRelationType(req.RelationType); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	cmd := &commands.CreateComponentRelation{
		SourceComponentID: req.SourceComponentID,
		TargetComponentID: req.TargetComponentID,
		RelationType:      req.RelationType,
		Name:              req.Name,
		Description:       req.Description,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	location := sharedAPI.BuildResourceLink(sharedAPI.ResourcePath("/relations"), sharedAPI.ResourceID(result.CreatedID))
	relation, err := h.readModel.GetByID(r.Context(), result.CreatedID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created relation")
		return
	}

	if relation == nil {
		sharedAPI.RespondCreated(w, location, map[string]string{
			"id":      result.CreatedID,
			"message": "Relation created, processing",
		})
		return
	}

	relation.Links = h.hateoas.RelationLinks(relation.ID)
	sharedAPI.RespondCreated(w, location, relation)
}

// GetAllRelations godoc
// @Summary Get all component relations
// @Description Retrieves all component relations with cursor-based pagination
// @Tags relations
// @Produce json
// @Param limit query int false "Number of items per page (max 100)" default(50)
// @Param after query string false "Cursor for pagination (opaque token)"
// @Success 200 {object} easi_backend_internal_shared_api.PaginatedResponse{data=[]readmodels.ComponentRelationDTO}
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /relations [get]
func (h *RelationHandlers) GetAllRelations(w http.ResponseWriter, r *http.Request) {
	params := sharedAPI.ParsePaginationParams(r)

	afterCursor, afterTimestamp, err := h.decodePaginationCursor(params.After)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid pagination cursor")
		return
	}

	relations, hasMore, err := h.readModel.GetAllPaginated(r.Context(), params.Limit, afterCursor, afterTimestamp)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve relations")
		return
	}

	h.addLinksToRelations(relations)

	nextCursor := h.buildNextCursor(relations, hasMore)
	selfLink := h.buildSelfLink("/api/v1/relations", params)

	sharedAPI.RespondPaginated(w, sharedAPI.PaginatedResponseParams{
		StatusCode: http.StatusOK,
		Data:       relations,
		HasMore:    hasMore,
		NextCursor: nextCursor,
		Limit:      params.Limit,
		SelfLink:   selfLink,
		BaseLink:   "/api/v1/relations",
	})
}

func (h *RelationHandlers) decodePaginationCursor(after string) (string, int64, error) {
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

func (h *RelationHandlers) buildNextCursor(relations []readmodels.ComponentRelationDTO, hasMore bool) string {
	if !hasMore || len(relations) == 0 {
		return ""
	}
	lastRelation := relations[len(relations)-1]
	return sharedAPI.EncodeCursor(lastRelation.ID, lastRelation.CreatedAt)
}

func (h *RelationHandlers) buildSelfLink(basePath string, params sharedAPI.PaginationParams) string {
	if params.After == "" {
		return basePath
	}
	return fmt.Sprintf("%s?after=%s&limit=%d", basePath, params.After, params.Limit)
}

func (h *RelationHandlers) addLinksToRelations(relations []readmodels.ComponentRelationDTO) {
	for i := range relations {
		relations[i].Links = h.hateoas.RelationLinks(relations[i].ID)
	}
}

// GetRelationByID godoc
// @Summary Get a component relation by ID
// @Description Retrieves a specific component relation by its ID
// @Tags relations
// @Produce json
// @Param id path string true "Relation ID"
// @Success 200 {object} readmodels.ComponentRelationDTO
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /relations/{id} [get]
func (h *RelationHandlers) GetRelationByID(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	relation, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve relation")
		return
	}

	if relation == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Relation not found")
		return
	}

	relation.Links = h.hateoas.RelationLinks(relation.ID)
	sharedAPI.RespondJSON(w, http.StatusOK, relation)
}

// GetRelationsFromComponent godoc
// @Summary Get relations from a component
// @Description Retrieves all relations where the specified component is the source
// @Tags relations
// @Produce json
// @Param componentId path string true "Component ID"
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]readmodels.ComponentRelationDTO}
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /relations/from/{componentId} [get]
func (h *RelationHandlers) GetRelationsFromComponent(w http.ResponseWriter, r *http.Request) {
	h.getRelationsByComponent(w, r, "from", h.readModel.GetBySourceID)
}

// GetRelationsToComponent godoc
// @Summary Get relations to a component
// @Description Retrieves all relations where the specified component is the target
// @Tags relations
// @Produce json
// @Param componentId path string true "Component ID"
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse{data=[]readmodels.ComponentRelationDTO}
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /relations/to/{componentId} [get]
func (h *RelationHandlers) GetRelationsToComponent(w http.ResponseWriter, r *http.Request) {
	h.getRelationsByComponent(w, r, "to", h.readModel.GetByTargetID)
}

type relationFetcher func(ctx context.Context, componentID string) ([]readmodels.ComponentRelationDTO, error)

func (h *RelationHandlers) getRelationsByComponent(w http.ResponseWriter, r *http.Request, direction string, fetch relationFetcher) {
	componentID := sharedAPI.GetPathParam(r, "componentId")

	relations, err := fetch(r.Context(), componentID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve relations")
		return
	}

	h.addLinksToRelations(relations)

	links := sharedAPI.NewResourceLinks().
		SelfWithID(sharedAPI.ResourcePath("/relations/"+direction), sharedAPI.ResourceID(componentID)).
		Related(sharedAPI.LinkRelation("component"), sharedAPI.ResourcePath("/components"), sharedAPI.ResourceID(componentID)).
		Build()

	sharedAPI.RespondCollection(w, http.StatusOK, relations, links)
}

// UpdateComponentRelation godoc
// @Summary Update a component relation
// @Description Updates an existing component relation's name and description
// @Tags relations
// @Accept json
// @Produce json
// @Param id path string true "Relation ID"
// @Param relation body UpdateComponentRelationRequest true "Updated relation data"
// @Success 200 {object} readmodels.ComponentRelationDTO
// @Failure 400 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 404 {object} easi_backend_internal_shared_api.ErrorResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /relations/{id} [put]
func (h *RelationHandlers) UpdateComponentRelation(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[UpdateComponentRelationRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.UpdateComponentRelation{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}

	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to update relation")
		return
	}

	relation, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve updated relation")
		return
	}

	if relation == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Relation not found")
		return
	}

	relation.Links = h.hateoas.RelationLinks(relation.ID)
	sharedAPI.RespondJSON(w, http.StatusOK, relation)
}

func (h *RelationHandlers) DeleteComponentRelation(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	cmd := &commands.DeleteComponentRelation{
		ID: id,
	}

	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		sharedAPI.RespondDeleted(w)
	})
}
