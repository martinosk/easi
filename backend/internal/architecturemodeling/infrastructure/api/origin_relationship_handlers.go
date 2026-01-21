package api

import (
	"net/http"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/domain"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/types"
)

type OriginReadModels struct {
	AcquiredVia   *readmodels.AcquiredViaRelationshipReadModel
	PurchasedFrom *readmodels.PurchasedFromRelationshipReadModel
	BuiltBy       *readmodels.BuiltByRelationshipReadModel
}

type OriginRelationshipHandlersConfig struct {
	CommandBus cqrs.CommandBus
	ReadModels OriginReadModels
	HATEOAS    *sharedAPI.HATEOASLinks
}

type OriginRelationshipHandlers struct {
	commandBus cqrs.CommandBus
	readModels OriginReadModels
	hateoas    *sharedAPI.HATEOASLinks
}

func NewOriginRelationshipHandlers(
	commandBus cqrs.CommandBus,
	acquiredViaReadModel *readmodels.AcquiredViaRelationshipReadModel,
	purchasedFromReadModel *readmodels.PurchasedFromRelationshipReadModel,
	builtByReadModel *readmodels.BuiltByRelationshipReadModel,
	hateoas *sharedAPI.HATEOASLinks,
) *OriginRelationshipHandlers {
	return NewOriginRelationshipHandlersFromConfig(OriginRelationshipHandlersConfig{
		CommandBus: commandBus,
		ReadModels: OriginReadModels{
			AcquiredVia:   acquiredViaReadModel,
			PurchasedFrom: purchasedFromReadModel,
			BuiltBy:       builtByReadModel,
		},
		HATEOAS: hateoas,
	})
}

func NewOriginRelationshipHandlersFromConfig(cfg OriginRelationshipHandlersConfig) *OriginRelationshipHandlers {
	return &OriginRelationshipHandlers{
		commandBus: cfg.CommandBus,
		readModels: cfg.ReadModels,
		hateoas:    cfg.HATEOAS,
	}
}

type CreateAcquiredViaRelationshipRequest struct {
	AcquiredEntityID string `json:"acquiredEntityId"`
	ComponentID      string `json:"componentId"`
	Notes            string `json:"notes,omitempty"`
	ReplaceExisting  bool   `json:"replaceExisting,omitempty"`
}

type CreatePurchasedFromRelationshipRequest struct {
	VendorID        string `json:"vendorId"`
	ComponentID     string `json:"componentId"`
	Notes           string `json:"notes,omitempty"`
	ReplaceExisting bool   `json:"replaceExisting,omitempty"`
}

type CreateBuiltByRelationshipRequest struct {
	InternalTeamID  string `json:"internalTeamId"`
	ComponentID     string `json:"componentId"`
	Notes           string `json:"notes,omitempty"`
	ReplaceExisting bool   `json:"replaceExisting,omitempty"`
}

type ComponentOriginsDTO struct {
	ComponentID   string                                    `json:"componentId"`
	AcquiredVia   []readmodels.AcquiredViaRelationshipDTO   `json:"acquiredVia"`
	PurchasedFrom []readmodels.PurchasedFromRelationshipDTO `json:"purchasedFrom"`
	BuiltBy       []readmodels.BuiltByRelationshipDTO       `json:"builtBy"`
	Links         types.Links                               `json:"_links,omitempty"`
}

type AllOriginRelationshipsDTO struct {
	AcquiredVia   []readmodels.AcquiredViaRelationshipDTO   `json:"acquiredVia"`
	PurchasedFrom []readmodels.PurchasedFromRelationshipDTO `json:"purchasedFrom"`
	BuiltBy       []readmodels.BuiltByRelationshipDTO       `json:"builtBy"`
	Links         types.Links                               `json:"_links,omitempty"`
}

// GetAllOriginsByComponent godoc
// @Summary Get all origin relationships for a component
// @Description Retrieves all origin relationships (acquired-via, purchased-from, built-by) for a component
// @Tags origin-relationships
// @Produce json
// @Param componentId path string true "Component ID"
// @Success 200 {object} ComponentOriginsDTO
// @Failure 401 {object} sharedAPI.ErrorResponse "Unauthorized - authentication required"
// @Failure 403 {object} sharedAPI.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{componentId}/origins [get]
func (h *OriginRelationshipHandlers) GetAllOriginsByComponent(w http.ResponseWriter, r *http.Request) {
	componentID := sharedAPI.GetPathParam(r, "componentId")
	actor, _ := sharedctx.GetActor(r.Context())

	acquiredVia, err := h.readModels.AcquiredVia.GetByComponentID(r.Context(), componentID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve acquired-via relationships")
		return
	}

	purchasedFrom, err := h.readModels.PurchasedFrom.GetByComponentID(r.Context(), componentID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve purchased-from relationships")
		return
	}

	builtBy, err := h.readModels.BuiltBy.GetByComponentID(r.Context(), componentID)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve built-by relationships")
		return
	}

	h.enrichAllRelationships(actor, acquiredVia, purchasedFrom, builtBy)

	origins := ComponentOriginsDTO{
		ComponentID:   componentID,
		AcquiredVia:   acquiredVia,
		PurchasedFrom: purchasedFrom,
		BuiltBy:       builtBy,
		Links: types.Links{
			"self":           {Href: sharedAPI.BuildSubResourceLink("/components", sharedAPI.ResourceID(componentID), "/origins"), Method: "GET"},
			"component":      {Href: sharedAPI.BuildResourceLink("/components", sharedAPI.ResourceID(componentID)), Method: "GET"},
			"acquired-via":   {Href: sharedAPI.BuildSubResourceLink("/components", sharedAPI.ResourceID(componentID), "/origin/acquired-via"), Method: "GET"},
			"purchased-from": {Href: sharedAPI.BuildSubResourceLink("/components", sharedAPI.ResourceID(componentID), "/origin/purchased-from"), Method: "GET"},
			"built-by":       {Href: sharedAPI.BuildSubResourceLink("/components", sharedAPI.ResourceID(componentID), "/origin/built-by"), Method: "GET"},
		},
	}

	sharedAPI.RespondJSON(w, http.StatusOK, origins)
}

// GetAllOriginRelationships godoc
// @Summary Get all origin relationships
// @Description Retrieves all origin relationships (acquired-via, purchased-from, built-by) across all components
// @Tags origin-relationships
// @Produce json
// @Success 200 {object} AllOriginRelationshipsDTO
// @Failure 401 {object} sharedAPI.ErrorResponse "Unauthorized - authentication required"
// @Failure 403 {object} sharedAPI.ErrorResponse "Forbidden - insufficient permissions"
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /origin-relationships [get]
func (h *OriginRelationshipHandlers) GetAllOriginRelationships(w http.ResponseWriter, r *http.Request) {
	actor, _ := sharedctx.GetActor(r.Context())

	acquiredVia, err := h.readModels.AcquiredVia.GetAll(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve acquired-via relationships")
		return
	}

	purchasedFrom, err := h.readModels.PurchasedFrom.GetAll(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve purchased-from relationships")
		return
	}

	builtBy, err := h.readModels.BuiltBy.GetAll(r.Context())
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve built-by relationships")
		return
	}

	h.enrichAllRelationships(actor, acquiredVia, purchasedFrom, builtBy)

	result := AllOriginRelationshipsDTO{
		AcquiredVia:   acquiredVia,
		PurchasedFrom: purchasedFrom,
		BuiltBy:       builtBy,
		Links: types.Links{
			"self": {Href: sharedAPI.BuildResourceLink("/origin-relationships", ""), Method: "GET"},
		},
	}

	sharedAPI.RespondJSON(w, http.StatusOK, result)
}

func (h *OriginRelationshipHandlers) enrichAllRelationships(
	actor sharedctx.Actor,
	acquiredVia []readmodels.AcquiredViaRelationshipDTO,
	purchasedFrom []readmodels.PurchasedFromRelationshipDTO,
	builtBy []readmodels.BuiltByRelationshipDTO,
) {
	for i := range acquiredVia {
		h.enrichAcquiredViaWithLinks(actor, &acquiredVia[i])
	}
	for i := range purchasedFrom {
		h.enrichPurchasedFromWithLinks(actor, &purchasedFrom[i])
	}
	for i := range builtBy {
		h.enrichBuiltByWithLinks(actor, &builtBy[i])
	}
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
		ReplaceExisting:  req.ReplaceExisting,
	}

	h.dispatchCreateAndRespond(w, r, cmd, componentID, "acquired-via", func(createdID string) (interface{}, error) {
		return h.readModels.AcquiredVia.GetByID(r.Context(), createdID)
	}, func(rel interface{}) {
		if dto, ok := rel.(*readmodels.AcquiredViaRelationshipDTO); ok {
			actor, _ := sharedctx.GetActor(r.Context())
			h.enrichAcquiredViaWithLinks(actor, dto)
		}
	})
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
	actor, _ := sharedctx.GetActor(r.Context())
	relationships, err := h.readModels.AcquiredVia.GetByComponentID(r.Context(), componentID)
	respondRelationshipsByComponent(w, componentID, "acquired-via", relationships, err, func(rel *readmodels.AcquiredViaRelationshipDTO) {
		h.enrichAcquiredViaWithLinks(actor, rel)
	})
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
		VendorID:        req.VendorID,
		ComponentID:     componentID,
		Notes:           req.Notes,
		ReplaceExisting: req.ReplaceExisting,
	}

	h.dispatchCreateAndRespond(w, r, cmd, componentID, "purchased-from", func(createdID string) (interface{}, error) {
		return h.readModels.PurchasedFrom.GetByID(r.Context(), createdID)
	}, func(rel interface{}) {
		if dto, ok := rel.(*readmodels.PurchasedFromRelationshipDTO); ok {
			actor, _ := sharedctx.GetActor(r.Context())
			h.enrichPurchasedFromWithLinks(actor, dto)
		}
	})
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
	actor, _ := sharedctx.GetActor(r.Context())
	relationships, err := h.readModels.PurchasedFrom.GetByComponentID(r.Context(), componentID)
	respondRelationshipsByComponent(w, componentID, "purchased-from", relationships, err, func(rel *readmodels.PurchasedFromRelationshipDTO) {
		h.enrichPurchasedFromWithLinks(actor, rel)
	})
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
		InternalTeamID:  req.InternalTeamID,
		ComponentID:     componentID,
		Notes:           req.Notes,
		ReplaceExisting: req.ReplaceExisting,
	}

	h.dispatchCreateAndRespond(w, r, cmd, componentID, "built-by", func(createdID string) (interface{}, error) {
		return h.readModels.BuiltBy.GetByID(r.Context(), createdID)
	}, func(rel interface{}) {
		if dto, ok := rel.(*readmodels.BuiltByRelationshipDTO); ok {
			actor, _ := sharedctx.GetActor(r.Context())
			h.enrichBuiltByWithLinks(actor, dto)
		}
	})
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
	actor, _ := sharedctx.GetActor(r.Context())
	relationships, err := h.readModels.BuiltBy.GetByComponentID(r.Context(), componentID)
	respondRelationshipsByComponent(w, componentID, "built-by", relationships, err, func(rel *readmodels.BuiltByRelationshipDTO) {
		h.enrichBuiltByWithLinks(actor, rel)
	})
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

func (h *OriginRelationshipHandlers) dispatchCreateAndRespond(
	w http.ResponseWriter,
	r *http.Request,
	cmd cqrs.Command,
	componentID string,
	originType string,
	fetchFn func(createdID string) (interface{}, error),
	enrichFn func(interface{}),
) {
	result, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		if existsErr, ok := err.(*domain.RelationshipExistsError); ok {
			respondRelationshipConflict(w, existsErr)
			return
		}
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	location := sharedAPI.BuildSubResourceLink("/components", sharedAPI.ResourceID(componentID), sharedAPI.ResourcePath("/origin/"+originType))
	relationship, fetchErr := fetchFn(result.CreatedID)
	if fetchErr != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, fetchErr, "Failed to retrieve created relationship")
		return
	}
	if relationship == nil {
		sharedAPI.RespondCreated(w, location, map[string]string{
			"id":      result.CreatedID,
			"message": "Relationship created, processing",
		})
		return
	}
	enrichFn(relationship)
	sharedAPI.RespondCreated(w, location, relationship)
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

func buildComponentOriginLinks(componentID, originType string) types.Links {
	return types.Links{
		"self":      {Href: sharedAPI.BuildSubResourceLink("/components", sharedAPI.ResourceID(componentID), sharedAPI.ResourcePath("/origin/"+originType)), Method: "GET"},
		"component": {Href: sharedAPI.BuildResourceLink("/components", sharedAPI.ResourceID(componentID)), Method: "GET"},
	}
}

func (h *OriginRelationshipHandlers) enrichAcquiredViaWithLinks(actor sharedctx.Actor, rel *readmodels.AcquiredViaRelationshipDTO) {
	extraLinks := map[string]types.Link{
		"acquiredEntity": {Href: sharedAPI.BuildResourceLink("/acquired-entities", sharedAPI.ResourceID(rel.AcquiredEntityID)), Method: "GET"},
	}
	rel.Links = h.hateoas.OriginRelationshipLinksForActor("/origin-relationships/acquired-via", rel.ID, rel.ComponentID, extraLinks, actor)
}

func (h *OriginRelationshipHandlers) enrichPurchasedFromWithLinks(actor sharedctx.Actor, rel *readmodels.PurchasedFromRelationshipDTO) {
	extraLinks := map[string]types.Link{
		"vendor": {Href: sharedAPI.BuildResourceLink("/vendors", sharedAPI.ResourceID(rel.VendorID)), Method: "GET"},
	}
	rel.Links = h.hateoas.OriginRelationshipLinksForActor("/origin-relationships/purchased-from", rel.ID, rel.ComponentID, extraLinks, actor)
}

func (h *OriginRelationshipHandlers) enrichBuiltByWithLinks(actor sharedctx.Actor, rel *readmodels.BuiltByRelationshipDTO) {
	extraLinks := map[string]types.Link{
		"internalTeam": {Href: sharedAPI.BuildResourceLink("/internal-teams", sharedAPI.ResourceID(rel.InternalTeamID)), Method: "GET"},
	}
	rel.Links = h.hateoas.OriginRelationshipLinksForActor("/origin-relationships/built-by", rel.ID, rel.ComponentID, extraLinks, actor)
}

type RelationshipConflictResponse struct {
	Error                  string `json:"error"`
	ExistingRelationshipID string `json:"existingRelationshipId"`
	ComponentID            string `json:"componentId"`
	OriginEntityID         string `json:"originEntityId"`
	OriginEntityName       string `json:"originEntityName"`
	RelationshipType       string `json:"relationshipType"`
}

func respondRelationshipConflict(w http.ResponseWriter, err *domain.RelationshipExistsError) {
	response := RelationshipConflictResponse{
		Error:                  err.Error(),
		ExistingRelationshipID: err.ExistingRelationshipID,
		ComponentID:            err.ComponentID,
		OriginEntityID:         err.OriginEntityID,
		OriginEntityName:       err.OriginEntityName,
		RelationshipType:       err.RelationshipType,
	}
	sharedAPI.RespondJSON(w, http.StatusConflict, response)
}
