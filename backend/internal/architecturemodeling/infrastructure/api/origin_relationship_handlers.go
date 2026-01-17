package api

import (
	"net/http"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/types"
)

type OriginRelationshipHandlers struct {
	commandBus             cqrs.CommandBus
	acquiredViaReadModel   *readmodels.AcquiredViaRelationshipReadModel
	purchasedFromReadModel *readmodels.PurchasedFromRelationshipReadModel
	builtByReadModel       *readmodels.BuiltByRelationshipReadModel
}

func NewOriginRelationshipHandlers(
	commandBus cqrs.CommandBus,
	acquiredViaReadModel *readmodels.AcquiredViaRelationshipReadModel,
	purchasedFromReadModel *readmodels.PurchasedFromRelationshipReadModel,
	builtByReadModel *readmodels.BuiltByRelationshipReadModel,
) *OriginRelationshipHandlers {
	return &OriginRelationshipHandlers{
		commandBus:             commandBus,
		acquiredViaReadModel:   acquiredViaReadModel,
		purchasedFromReadModel: purchasedFromReadModel,
		builtByReadModel:       builtByReadModel,
	}
}

type CreateAcquiredViaRelationshipRequest struct {
	AcquiredEntityID string `json:"acquiredEntityId"`
	ComponentID      string `json:"componentId"`
	Notes            string `json:"notes,omitempty"`
}

type CreatePurchasedFromRelationshipRequest struct {
	VendorID    string `json:"vendorId"`
	ComponentID string `json:"componentId"`
	Notes       string `json:"notes,omitempty"`
}

type CreateBuiltByRelationshipRequest struct {
	InternalTeamID string `json:"internalTeamId"`
	ComponentID    string `json:"componentId"`
	Notes          string `json:"notes,omitempty"`
}

// CreateAcquiredViaRelationship godoc
// @Summary Create an acquired-via relationship
// @Description Links a component to an acquired entity
// @Tags origin-relationships
// @Accept json
// @Produce json
// @Param relationship body CreateAcquiredViaRelationshipRequest true "Relationship data"
// @Success 201 {object} readmodels.AcquiredViaRelationshipDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{componentId}/origin/acquired-via [post]
func (h *OriginRelationshipHandlers) CreateAcquiredViaRelationship(w http.ResponseWriter, r *http.Request) {
	componentID := sharedAPI.GetPathParam(r, "componentId")
	req, ok := sharedAPI.DecodeRequestOrFail[CreateAcquiredViaRelationshipRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.CreateAcquiredViaRelationship{
		AcquiredEntityID: req.AcquiredEntityID,
		ComponentID:      componentID,
		Notes:            req.Notes,
	}
	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	relationship, fetchErr := h.acquiredViaReadModel.GetByID(r.Context(), result.CreatedID)
	respondCreatedRelationship(w, componentID, "acquired-via", result.CreatedID, relationship, fetchErr, h.enrichAcquiredViaWithLinks)
}

// GetAcquiredViaByComponent godoc
// @Summary Get acquired-via relationships for a component
// @Description Retrieves all acquired-via relationships for a component
// @Tags origin-relationships
// @Produce json
// @Param componentId path string true "Component ID"
// @Success 200 {object} []readmodels.AcquiredViaRelationshipDTO
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{componentId}/origin/acquired-via [get]
func (h *OriginRelationshipHandlers) GetAcquiredViaByComponent(w http.ResponseWriter, r *http.Request) {
	componentID := sharedAPI.GetPathParam(r, "componentId")
	relationships, err := h.acquiredViaReadModel.GetByComponentID(r.Context(), componentID)
	respondRelationshipsByComponent(w, componentID, "acquired-via", relationships, err, h.enrichAcquiredViaWithLinks)
}

// DeleteAcquiredViaRelationship godoc
// @Summary Delete an acquired-via relationship
// @Description Removes an acquired-via relationship
// @Tags origin-relationships
// @Produce json
// @Param id path string true "Relationship ID"
// @Success 204
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /origin-relationships/acquired-via/{id} [delete]
func (h *OriginRelationshipHandlers) DeleteAcquiredViaRelationship(w http.ResponseWriter, r *http.Request) {
	h.dispatchDeleteAndRespond(w, r, &commands.DeleteAcquiredViaRelationship{ID: sharedAPI.GetPathParam(r, "id")})
}

// CreatePurchasedFromRelationship godoc
// @Summary Create a purchased-from relationship
// @Description Links a component to a vendor
// @Tags origin-relationships
// @Accept json
// @Produce json
// @Param relationship body CreatePurchasedFromRelationshipRequest true "Relationship data"
// @Success 201 {object} readmodels.PurchasedFromRelationshipDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{componentId}/origin/purchased-from [post]
func (h *OriginRelationshipHandlers) CreatePurchasedFromRelationship(w http.ResponseWriter, r *http.Request) {
	componentID := sharedAPI.GetPathParam(r, "componentId")
	req, ok := sharedAPI.DecodeRequestOrFail[CreatePurchasedFromRelationshipRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.CreatePurchasedFromRelationship{
		VendorID:    req.VendorID,
		ComponentID: componentID,
		Notes:       req.Notes,
	}
	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	relationship, fetchErr := h.purchasedFromReadModel.GetByID(r.Context(), result.CreatedID)
	respondCreatedRelationship(w, componentID, "purchased-from", result.CreatedID, relationship, fetchErr, h.enrichPurchasedFromWithLinks)
}

// GetPurchasedFromByComponent godoc
// @Summary Get purchased-from relationships for a component
// @Description Retrieves all purchased-from relationships for a component
// @Tags origin-relationships
// @Produce json
// @Param componentId path string true "Component ID"
// @Success 200 {object} []readmodels.PurchasedFromRelationshipDTO
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{componentId}/origin/purchased-from [get]
func (h *OriginRelationshipHandlers) GetPurchasedFromByComponent(w http.ResponseWriter, r *http.Request) {
	componentID := sharedAPI.GetPathParam(r, "componentId")
	relationships, err := h.purchasedFromReadModel.GetByComponentID(r.Context(), componentID)
	respondRelationshipsByComponent(w, componentID, "purchased-from", relationships, err, h.enrichPurchasedFromWithLinks)
}

// DeletePurchasedFromRelationship godoc
// @Summary Delete a purchased-from relationship
// @Description Removes a purchased-from relationship
// @Tags origin-relationships
// @Produce json
// @Param id path string true "Relationship ID"
// @Success 204
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /origin-relationships/purchased-from/{id} [delete]
func (h *OriginRelationshipHandlers) DeletePurchasedFromRelationship(w http.ResponseWriter, r *http.Request) {
	h.dispatchDeleteAndRespond(w, r, &commands.DeletePurchasedFromRelationship{ID: sharedAPI.GetPathParam(r, "id")})
}

// CreateBuiltByRelationship godoc
// @Summary Create a built-by relationship
// @Description Links a component to an internal team
// @Tags origin-relationships
// @Accept json
// @Produce json
// @Param relationship body CreateBuiltByRelationshipRequest true "Relationship data"
// @Success 201 {object} readmodels.BuiltByRelationshipDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{componentId}/origin/built-by [post]
func (h *OriginRelationshipHandlers) CreateBuiltByRelationship(w http.ResponseWriter, r *http.Request) {
	componentID := sharedAPI.GetPathParam(r, "componentId")
	req, ok := sharedAPI.DecodeRequestOrFail[CreateBuiltByRelationshipRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.CreateBuiltByRelationship{
		InternalTeamID: req.InternalTeamID,
		ComponentID:    componentID,
		Notes:          req.Notes,
	}
	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	relationship, fetchErr := h.builtByReadModel.GetByID(r.Context(), result.CreatedID)
	respondCreatedRelationship(w, componentID, "built-by", result.CreatedID, relationship, fetchErr, h.enrichBuiltByWithLinks)
}

// GetBuiltByByComponent godoc
// @Summary Get built-by relationships for a component
// @Description Retrieves all built-by relationships for a component
// @Tags origin-relationships
// @Produce json
// @Param componentId path string true "Component ID"
// @Success 200 {object} []readmodels.BuiltByRelationshipDTO
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{componentId}/origin/built-by [get]
func (h *OriginRelationshipHandlers) GetBuiltByByComponent(w http.ResponseWriter, r *http.Request) {
	componentID := sharedAPI.GetPathParam(r, "componentId")
	relationships, err := h.builtByReadModel.GetByComponentID(r.Context(), componentID)
	respondRelationshipsByComponent(w, componentID, "built-by", relationships, err, h.enrichBuiltByWithLinks)
}

// DeleteBuiltByRelationship godoc
// @Summary Delete a built-by relationship
// @Description Removes a built-by relationship
// @Tags origin-relationships
// @Produce json
// @Param id path string true "Relationship ID"
// @Success 204
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /origin-relationships/built-by/{id} [delete]
func (h *OriginRelationshipHandlers) DeleteBuiltByRelationship(w http.ResponseWriter, r *http.Request) {
	h.dispatchDeleteAndRespond(w, r, &commands.DeleteBuiltByRelationship{ID: sharedAPI.GetPathParam(r, "id")})
}

func (h *OriginRelationshipHandlers) dispatchDeleteAndRespond(w http.ResponseWriter, r *http.Request, cmd cqrs.Command) {
	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	sharedAPI.HandleCommandResult(w, result, err, func(_ string) {
		sharedAPI.RespondDeleted(w)
	})
}

func respondRelationshipsByComponent[T any](
	w http.ResponseWriter,
	componentID string,
	originType string,
	relationships []T,
	err error,
	enrichFn func(*T),
) {
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve relationships")
		return
	}
	for i := range relationships {
		enrichFn(&relationships[i])
	}
	sharedAPI.RespondCollection(w, http.StatusOK, relationships, buildComponentOriginLinks(componentID, originType))
}

func respondCreatedRelationship[T any](
	w http.ResponseWriter,
	componentID string,
	originType string,
	createdID string,
	relationship *T,
	err error,
	enrichFn func(*T),
) {
	location := sharedAPI.BuildSubResourceLink("/components", sharedAPI.ResourceID(componentID), sharedAPI.ResourcePath("/origin/"+originType))
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve created relationship")
		return
	}
	if relationship == nil {
		sharedAPI.RespondCreated(w, location, map[string]string{
			"id":      createdID,
			"message": "Relationship created, processing",
		})
		return
	}
	enrichFn(relationship)
	sharedAPI.RespondCreated(w, location, relationship)
}

func buildComponentOriginLinks(componentID, originType string) types.Links {
	return types.Links{
		"self":      {Href: sharedAPI.BuildSubResourceLink("/components", sharedAPI.ResourceID(componentID), sharedAPI.ResourcePath("/origin/"+originType)), Method: "GET"},
		"component": {Href: sharedAPI.BuildResourceLink("/components", sharedAPI.ResourceID(componentID)), Method: "GET"},
	}
}

func buildRelationshipLinks(basePath, id, componentID string, extraLinks map[string]types.Link) types.Links {
	links := types.Links{
		"self":      {Href: sharedAPI.BuildResourceLink(sharedAPI.ResourcePath(basePath), sharedAPI.ResourceID(id)), Method: "GET"},
		"delete":    {Href: sharedAPI.BuildResourceLink(sharedAPI.ResourcePath(basePath), sharedAPI.ResourceID(id)), Method: "DELETE"},
		"component": {Href: sharedAPI.BuildResourceLink("/components", sharedAPI.ResourceID(componentID)), Method: "GET"},
	}
	for k, v := range extraLinks {
		links[k] = v
	}
	return links
}

func (h *OriginRelationshipHandlers) enrichAcquiredViaWithLinks(rel *readmodels.AcquiredViaRelationshipDTO) {
	rel.Links = buildRelationshipLinks("/origin-relationships/acquired-via", rel.ID, rel.ComponentID, map[string]types.Link{
		"acquiredEntity": {Href: sharedAPI.BuildResourceLink("/acquired-entities", sharedAPI.ResourceID(rel.AcquiredEntityID)), Method: "GET"},
	})
}

func (h *OriginRelationshipHandlers) enrichPurchasedFromWithLinks(rel *readmodels.PurchasedFromRelationshipDTO) {
	rel.Links = buildRelationshipLinks("/origin-relationships/purchased-from", rel.ID, rel.ComponentID, map[string]types.Link{
		"vendor": {Href: sharedAPI.BuildResourceLink("/vendors", sharedAPI.ResourceID(rel.VendorID)), Method: "GET"},
	})
}

func (h *OriginRelationshipHandlers) enrichBuiltByWithLinks(rel *readmodels.BuiltByRelationshipDTO) {
	rel.Links = buildRelationshipLinks("/origin-relationships/built-by", rel.ID, rel.ComponentID, map[string]types.Link{
		"internalTeam": {Href: sharedAPI.BuildResourceLink("/internal-teams", sharedAPI.ResourceID(rel.InternalTeamID)), Method: "GET"},
	})
}
