package api

import (
	"context"
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
	var req CreateComponentRelationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// Validate using domain value objects
	_, err := valueobjects.NewComponentIDFromString(req.SourceComponentID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid source component ID")
		return
	}

	_, err = valueobjects.NewComponentIDFromString(req.TargetComponentID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid target component ID")
		return
	}

	_, err = valueobjects.NewRelationType(req.RelationType)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	// Create command (pass by reference so handler can set ID)
	cmd := &commands.CreateComponentRelation{
		SourceComponentID: req.SourceComponentID,
		TargetComponentID: req.TargetComponentID,
		RelationType:      req.RelationType,
		Name:              req.Name,
		Description:       req.Description,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	// Retrieve the created relation from read model
	relation, err := h.readModel.GetByID(r.Context(), cmd.ID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created relation")
		return
	}

	if relation == nil {
		// If not yet in read model, return minimal response with Location header
		location := fmt.Sprintf("/api/v1/relations/%s", cmd.ID)
		w.Header().Set("Location", location)
		sharedAPI.RespondJSON(w, http.StatusCreated, map[string]string{
			"id":      cmd.ID,
			"message": "Relation created, processing",
		})
		return
	}

	// Add HATEOAS links
	relation.Links = h.hateoas.RelationLinks(relation.ID)

	// Return created resource with Location header
	location := fmt.Sprintf("/api/v1/relations/%s", relation.ID)
	w.Header().Set("Location", location)
	sharedAPI.RespondJSON(w, http.StatusCreated, relation)
}

// GetAllRelations godoc
// @Summary Get all component relations
// @Description Retrieves all component relations with cursor-based pagination
// @Tags relations
// @Produce json
// @Param limit query int false "Number of items per page (max 100)" default(50)
// @Param after query string false "Cursor for pagination (opaque token)"
// @Success 200 {object} easi_backend_internal_shared_api.PaginatedResponse
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
	id := chi.URLParam(r, "id")

	relation, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve relation")
		return
	}

	if relation == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Relation not found")
		return
	}

	// Add HATEOAS links
	relation.Links = h.hateoas.RelationLinks(relation.ID)

	sharedAPI.RespondJSON(w, http.StatusOK, relation)
}

// GetRelationsFromComponent godoc
// @Summary Get relations from a component
// @Description Retrieves all relations where the specified component is the source
// @Tags relations
// @Produce json
// @Param componentId path string true "Component ID"
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse
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
// @Success 200 {object} easi_backend_internal_shared_api.CollectionResponse
// @Failure 500 {object} easi_backend_internal_shared_api.ErrorResponse
// @Router /relations/to/{componentId} [get]
func (h *RelationHandlers) GetRelationsToComponent(w http.ResponseWriter, r *http.Request) {
	h.getRelationsByComponent(w, r, "to", h.readModel.GetByTargetID)
}

type relationFetcher func(ctx context.Context, componentID string) ([]readmodels.ComponentRelationDTO, error)

func (h *RelationHandlers) getRelationsByComponent(w http.ResponseWriter, r *http.Request, direction string, fetch relationFetcher) {
	componentID := chi.URLParam(r, "componentId")

	relations, err := fetch(r.Context(), componentID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve relations")
		return
	}

	h.addLinksToRelations(relations)

	links := map[string]string{
		"self":      fmt.Sprintf("/api/v1/relations/%s/%s", direction, componentID),
		"component": "/api/v1/components/" + componentID,
	}

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
	id := chi.URLParam(r, "id")

	var req UpdateComponentRelationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "Invalid request body")
		return
	}

	// Create command
	cmd := &commands.UpdateComponentRelation{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}

	// Dispatch command
	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to update relation")
		return
	}

	// Retrieve the updated relation from read model
	relation, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve updated relation")
		return
	}

	if relation == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Relation not found")
		return
	}

	// Add HATEOAS links
	relation.Links = h.hateoas.RelationLinks(relation.ID)

	sharedAPI.RespondJSON(w, http.StatusOK, relation)
}

func (h *RelationHandlers) DeleteComponentRelation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cmd := &commands.DeleteComponentRelation{
		ID: id,
	}

	if err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		if err.Error() == "relation not found" {
			sharedAPI.RespondError(w, http.StatusNotFound, err, "Relation not found")
			return
		}
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to delete relation")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

