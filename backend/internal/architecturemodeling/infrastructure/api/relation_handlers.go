package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/easi/backend/internal/architecturemodeling/application/commands"
	"github.com/easi/backend/internal/architecturemodeling/application/readmodels"
	"github.com/easi/backend/internal/architecturemodeling/domain/valueobjects"
	sharedAPI "github.com/easi/backend/internal/shared/api"
	"github.com/easi/backend/internal/shared/cqrs"
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

// CreateComponentRelation godoc
// @Summary Create a new component relation
// @Description Creates a new relation between two application components
// @Tags relations
// @Accept json
// @Produce json
// @Param relation body CreateComponentRelationRequest true "Relation data"
// @Success 201 {object} readmodels.ComponentRelationDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
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
	sharedAPI.RespondCreated(w, location, relation, nil)
}

// GetAllRelations godoc
// @Summary Get all component relations
// @Description Retrieves all component relations with cursor-based pagination
// @Tags relations
// @Produce json
// @Param limit query int false "Number of items per page (max 100)" default(50)
// @Param after query string false "Cursor for pagination (opaque token)"
// @Success 200 {object} sharedAPI.PaginatedResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /relations [get]
func (h *RelationHandlers) GetAllRelations(w http.ResponseWriter, r *http.Request) {
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

	// Get paginated relations
	relations, hasMore, err := h.readModel.GetAllPaginated(r.Context(), params.Limit, afterCursor, afterTimestamp)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve relations")
		return
	}

	// Add HATEOAS links to each relation
	for i := range relations {
		relations[i].Links = h.hateoas.RelationLinks(relations[i].ID)
	}

	// Generate next cursor if there are more results
	var nextCursor string
	if hasMore && len(relations) > 0 {
		lastRelation := relations[len(relations)-1]
		nextCursor = sharedAPI.EncodeCursor(lastRelation.ID, lastRelation.CreatedAt)
	}

	// Build self link
	selfLink := "/api/v1/relations"
	if params.After != "" {
		selfLink = fmt.Sprintf("/api/v1/relations?after=%s&limit=%d", params.After, params.Limit)
	}

	// Respond with paginated data
	sharedAPI.RespondPaginated(w, http.StatusOK, relations, hasMore, nextCursor, params.Limit, selfLink, "/api/v1/relations")
}

// GetRelationByID godoc
// @Summary Get a component relation by ID
// @Description Retrieves a specific component relation by its ID
// @Tags relations
// @Produce json
// @Param id path string true "Relation ID"
// @Success 200 {object} readmodels.ComponentRelationDTO
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
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

	sharedAPI.RespondSuccess(w, http.StatusOK, relation, nil)
}

// GetRelationsFromComponent godoc
// @Summary Get relations from a component
// @Description Retrieves all relations where the specified component is the source
// @Tags relations
// @Produce json
// @Param componentId path string true "Component ID"
// @Success 200 {array} readmodels.ComponentRelationDTO
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /relations/from/{componentId} [get]
func (h *RelationHandlers) GetRelationsFromComponent(w http.ResponseWriter, r *http.Request) {
	componentID := chi.URLParam(r, "componentId")

	relations, err := h.readModel.GetBySourceID(r.Context(), componentID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve relations")
		return
	}

	// Add HATEOAS links to each relation
	for i := range relations {
		relations[i].Links = h.hateoas.RelationLinks(relations[i].ID)
	}

	sharedAPI.RespondSuccess(w, http.StatusOK, relations, nil)
}

// GetRelationsToComponent godoc
// @Summary Get relations to a component
// @Description Retrieves all relations where the specified component is the target
// @Tags relations
// @Produce json
// @Param componentId path string true "Component ID"
// @Success 200 {array} readmodels.ComponentRelationDTO
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /relations/to/{componentId} [get]
func (h *RelationHandlers) GetRelationsToComponent(w http.ResponseWriter, r *http.Request) {
	componentID := chi.URLParam(r, "componentId")

	relations, err := h.readModel.GetByTargetID(r.Context(), componentID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve relations")
		return
	}

	// Add HATEOAS links to each relation
	for i := range relations {
		relations[i].Links = h.hateoas.RelationLinks(relations[i].ID)
	}

	sharedAPI.RespondSuccess(w, http.StatusOK, relations, nil)
}
