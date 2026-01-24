package api

import (
	"net/http"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	authValueObjects "easi/backend/internal/auth/domain/valueobjects"
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
// @Summary Set or update an acquired-via relationship
// @Description Sets or updates the acquired-via relationship for a component (idempotent)
// @Tags origin-relationships
// @Accept json
// @Produce json
// @Param componentId path string true "Component ID"
// @Param relationship body CreateAcquiredViaRelationshipRequest true "Relationship data"
// @Success 200 {object} readmodels.AcquiredViaRelationshipDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{componentId}/origin/acquired-via [put]
func (h *OriginRelationshipHandlers) CreateAcquiredViaRelationship(w http.ResponseWriter, r *http.Request) {
	componentID := sharedAPI.GetPathParam(r, "componentId")
	req, ok := sharedAPI.DecodeRequestOrFail[CreateAcquiredViaRelationshipRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.SetAcquiredVia{
		ComponentID: componentID,
		EntityID:    req.AcquiredEntityID,
		Notes:       req.Notes,
	}

	h.dispatchSetAndRespond(w, r, cmd, componentID, "acquired-via", func(componentID string) (interface{}, error) {
		relationships, err := h.readModels.AcquiredVia.GetByComponentID(r.Context(), componentID)
		if err != nil || len(relationships) == 0 {
			return nil, err
		}
		return &relationships[0], nil
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
// @Param componentId path string true "Component ID"
// @Success 204
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{componentId}/origin/acquired-via [delete]
func (h *OriginRelationshipHandlers) DeleteAcquiredViaRelationship(w http.ResponseWriter, r *http.Request) {
	h.dispatchClearAndRespond(w, r, &commands.ClearAcquiredVia{ComponentID: sharedAPI.GetPathParam(r, "componentId")})
}

// CreatePurchasedFromRelationship godoc
// @Summary Set or update a purchased-from relationship
// @Description Sets or updates the purchased-from relationship for a component (idempotent)
// @Tags origin-relationships
// @Accept json
// @Produce json
// @Param componentId path string true "Component ID"
// @Param relationship body CreatePurchasedFromRelationshipRequest true "Relationship data"
// @Success 200 {object} readmodels.PurchasedFromRelationshipDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{componentId}/origin/purchased-from [put]
func (h *OriginRelationshipHandlers) CreatePurchasedFromRelationship(w http.ResponseWriter, r *http.Request) {
	componentID := sharedAPI.GetPathParam(r, "componentId")
	req, ok := sharedAPI.DecodeRequestOrFail[CreatePurchasedFromRelationshipRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.SetPurchasedFrom{
		ComponentID: componentID,
		VendorID:    req.VendorID,
		Notes:       req.Notes,
	}

	h.dispatchSetAndRespond(w, r, cmd, componentID, "purchased-from", func(componentID string) (interface{}, error) {
		relationships, err := h.readModels.PurchasedFrom.GetByComponentID(r.Context(), componentID)
		if err != nil || len(relationships) == 0 {
			return nil, err
		}
		return &relationships[0], nil
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
// @Param componentId path string true "Component ID"
// @Success 204
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{componentId}/origin/purchased-from [delete]
func (h *OriginRelationshipHandlers) DeletePurchasedFromRelationship(w http.ResponseWriter, r *http.Request) {
	h.dispatchClearAndRespond(w, r, &commands.ClearPurchasedFrom{ComponentID: sharedAPI.GetPathParam(r, "componentId")})
}

// CreateBuiltByRelationship godoc
// @Summary Set or update a built-by relationship
// @Description Sets or updates the built-by relationship for a component (idempotent)
// @Tags origin-relationships
// @Accept json
// @Produce json
// @Param componentId path string true "Component ID"
// @Param relationship body CreateBuiltByRelationshipRequest true "Relationship data"
// @Success 200 {object} readmodels.BuiltByRelationshipDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{componentId}/origin/built-by [put]
func (h *OriginRelationshipHandlers) CreateBuiltByRelationship(w http.ResponseWriter, r *http.Request) {
	componentID := sharedAPI.GetPathParam(r, "componentId")
	req, ok := sharedAPI.DecodeRequestOrFail[CreateBuiltByRelationshipRequest](w, r)
	if !ok {
		return
	}

	cmd := &commands.SetBuiltBy{
		ComponentID: componentID,
		TeamID:      req.InternalTeamID,
		Notes:       req.Notes,
	}

	h.dispatchSetAndRespond(w, r, cmd, componentID, "built-by", func(componentID string) (interface{}, error) {
		relationships, err := h.readModels.BuiltBy.GetByComponentID(r.Context(), componentID)
		if err != nil || len(relationships) == 0 {
			return nil, err
		}
		return &relationships[0], nil
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
// @Param componentId path string true "Component ID"
// @Success 204
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{componentId}/origin/built-by [delete]
func (h *OriginRelationshipHandlers) DeleteBuiltByRelationship(w http.ResponseWriter, r *http.Request) {
	h.dispatchClearAndRespond(w, r, &commands.ClearBuiltBy{ComponentID: sharedAPI.GetPathParam(r, "componentId")})
}

func (h *OriginRelationshipHandlers) dispatchSetAndRespond(
	w http.ResponseWriter,
	r *http.Request,
	cmd cqrs.Command,
	componentID string,
	originType string,
	fetchFn func(componentID string) (interface{}, error),
	enrichFn func(interface{}),
) {
	_, err := h.commandBus.Dispatch(r.Context(), cmd)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusBadRequest, err, "")
		return
	}

	relationship, fetchErr := fetchFn(componentID)
	if fetchErr != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, fetchErr, "Failed to retrieve relationship")
		return
	}
	if relationship == nil {
		sharedAPI.RespondJSON(w, http.StatusOK, map[string]string{
			"componentId": componentID,
			"message":     "Relationship set, processing",
		})
		return
	}
	enrichFn(relationship)
	sharedAPI.RespondJSON(w, http.StatusOK, relationship)
}

func (h *OriginRelationshipHandlers) dispatchClearAndRespond(w http.ResponseWriter, r *http.Request, cmd cqrs.Command) {
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

type originLinkParams struct {
	componentID    string
	originType     string
	entityLinkName string
	entityResource string
	entityID       string
}

func buildOriginLinks(actor sharedctx.Actor, p originLinkParams) types.Links {
	baseURL := sharedAPI.BuildSubResourceLink("/components", sharedAPI.ResourceID(p.componentID), sharedAPI.ResourcePath("/origin/"+p.originType))
	links := types.Links{
		"self":          {Href: baseURL, Method: "GET"},
		"component":     {Href: sharedAPI.BuildResourceLink("/components", sharedAPI.ResourceID(p.componentID)), Method: "GET"},
		p.entityLinkName: {Href: sharedAPI.BuildResourceLink(sharedAPI.ResourcePath(p.entityResource), sharedAPI.ResourceID(p.entityID)), Method: "GET"},
	}
	if actor.CanWrite(authValueObjects.PermComponentsWrite.String()) {
		links["update"] = types.Link{Href: baseURL, Method: "PUT"}
	}
	if actor.CanDelete(authValueObjects.PermComponentsDelete.String()) {
		links["delete"] = types.Link{Href: baseURL, Method: "DELETE"}
	}
	return links
}

func (h *OriginRelationshipHandlers) enrichAcquiredViaWithLinks(actor sharedctx.Actor, rel *readmodels.AcquiredViaRelationshipDTO) {
	rel.Links = buildOriginLinks(actor, originLinkParams{rel.ComponentID, "acquired-via", "acquiredEntity", "/acquired-entities", rel.AcquiredEntityID})
}

func (h *OriginRelationshipHandlers) enrichPurchasedFromWithLinks(actor sharedctx.Actor, rel *readmodels.PurchasedFromRelationshipDTO) {
	rel.Links = buildOriginLinks(actor, originLinkParams{rel.ComponentID, "purchased-from", "vendor", "/vendors", rel.VendorID})
}

func (h *OriginRelationshipHandlers) enrichBuiltByWithLinks(actor sharedctx.Actor, rel *readmodels.BuiltByRelationshipDTO) {
	rel.Links = buildOriginLinks(actor, originLinkParams{rel.ComponentID, "built-by", "internalTeam", "/internal-teams", rel.InternalTeamID})
}

