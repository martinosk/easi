package api

import (
	"context"
	"net/http"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/application/handlers"
	"easi/backend/internal/architecturedirection/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/cqrs"
)

type RealizationRoleQueries interface {
	GetByPair(ctx context.Context, capabilityID, componentID string) (*readmodels.RealizationRoleDTO, error)
	GetByCapabilityIDs(ctx context.Context, capabilityIDs []string) ([]readmodels.RealizationRoleDTO, error)
}

type RealizationRoleHandlers struct {
	commandBus cqrs.CommandBus
	queries    RealizationRoleQueries
	hateoas    *RealizationRoleLinks
}

func NewRealizationRoleHandlers(commandBus cqrs.CommandBus, queries RealizationRoleQueries, hateoas *RealizationRoleLinks) *RealizationRoleHandlers {
	return &RealizationRoleHandlers{commandBus: commandBus, queries: queries, hateoas: hateoas}
}

type AssignRealizationRoleRequest struct {
	Role string `json:"role"`
}

// GetRealizationRoles godoc
// @Summary Bulk-fetch current realization roles for a capability set
// @Description Returns the current role for every (capability, component) pair among the given capabilities.
// @Tags realization-roles
// @Produce json
// @Security CookieAuth
// @Param capabilityIds query string true "Comma-separated capability IDs"
// @Success 200 {object} sharedAPI.CollectionResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /realization-roles [get]
func (h *RealizationRoleHandlers) GetRealizationRoles(w http.ResponseWriter, r *http.Request) {
	capabilityIDs := parseIDList(r.URL.Query().Get("capabilityIds"))
	roles, err := h.queries.GetByCapabilityIDs(r.Context(), capabilityIDs)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	for i := range roles {
		roles[i].Links = h.hateoas.ItemLinks(roles[i].CapabilityID, roles[i].ComponentID, actor)
	}
	links := h.hateoas.CollectionLinks(string(realizationRolesPath), actor)
	sharedAPI.RespondCollection(w, http.StatusOK, roles, links)
}

// GetRealizationRole godoc
// @Summary Get the current realization role for a realisation
// @Description Returns the current role for the given (capability, component) pair, or 404 when unclassified.
// @Tags realization-roles
// @Produce json
// @Security CookieAuth
// @Param id path string true "Capability ID"
// @Param componentId path string true "Application component ID"
// @Success 200 {object} readmodels.RealizationRoleDTO
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id}/components/{componentId}/realization-role [get]
func (h *RealizationRoleHandlers) GetRealizationRole(w http.ResponseWriter, r *http.Request) {
	capabilityID := sharedAPI.GetPathParam(r, "id")
	componentID := sharedAPI.GetPathParam(r, "componentId")
	dto, err := h.queries.GetByPair(r.Context(), capabilityID, componentID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	if dto == nil {
		sharedAPI.HandleError(w, handlers.ErrRealizationRoleNotFoundForPair)
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	dto.Links = h.hateoas.ItemLinks(capabilityID, componentID, actor)
	sharedAPI.RespondJSON(w, http.StatusOK, dto)
}

// PutRealizationRole godoc
// @Summary Assign or re-assign a realisation's role
// @Description Records the architect's role (standard/legacy) for the (capability, component) pair. Requires a direct realisation. 201 on first assignment, 200 on re-assignment. Assigning standard atomically displaces any previous standard holder for the capability.
// @Tags realization-roles
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param id path string true "Capability ID"
// @Param componentId path string true "Application component ID"
// @Param body body AssignRealizationRoleRequest true "Role data"
// @Success 200 {object} readmodels.RealizationRoleDTO
// @Success 201 {object} readmodels.RealizationRoleDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id}/components/{componentId}/realization-role [put]
func (h *RealizationRoleHandlers) PutRealizationRole(w http.ResponseWriter, r *http.Request) {
	capabilityID := sharedAPI.GetPathParam(r, "id")
	componentID := sharedAPI.GetPathParam(r, "componentId")
	req, ok := sharedAPI.DecodeRequestOrFail[AssignRealizationRoleRequest](w, r)
	if !ok {
		return
	}
	existedBefore, err := h.queries.GetByPair(r.Context(), capabilityID, componentID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	cmd := &commands.AssignRealizationRole{
		CapabilityID: capabilityID,
		ComponentID:  componentID,
		Role:         req.Role,
		AssignedBy:   actor.Email,
	}
	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	statusCode := http.StatusOK
	if existedBefore == nil {
		statusCode = http.StatusCreated
	}
	h.respondWithCurrentRole(w, r, realizationRolePairID{CapabilityID: capabilityID, ComponentID: componentID}, statusCode)
}

// DeleteRealizationRole godoc
// @Summary Clear a realization role
// @Description Clears the current role for the pair; the realisation presents as unclassified. Recorded as a discrete RealizationRoleCleared event.
// @Tags realization-roles
// @Security CookieAuth
// @Param id path string true "Capability ID"
// @Param componentId path string true "Application component ID"
// @Success 204 "No Content"
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /capabilities/{id}/components/{componentId}/realization-role [delete]
func (h *RealizationRoleHandlers) DeleteRealizationRole(w http.ResponseWriter, r *http.Request) {
	capabilityID := sharedAPI.GetPathParam(r, "id")
	componentID := sharedAPI.GetPathParam(r, "componentId")
	actor, _ := sharedctx.GetActor(r.Context())
	cmd := &commands.ClearRealizationRole{
		CapabilityID: capabilityID,
		ComponentID:  componentID,
		ClearedBy:    actor.Email,
	}
	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	sharedAPI.RespondNoContent(w)
}

type realizationRolePairID struct {
	CapabilityID string
	ComponentID  string
}

func (h *RealizationRoleHandlers) respondWithCurrentRole(w http.ResponseWriter, r *http.Request, pair realizationRolePairID, statusCode int) {
	dto, err := h.queries.GetByPair(r.Context(), pair.CapabilityID, pair.ComponentID)
	if err != nil {
		sharedAPI.HandleError(w, err)
		return
	}
	if dto == nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, handlers.ErrRealizationRoleNotFoundForPair, "failed to load realization role")
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	dto.Links = h.hateoas.ItemLinks(pair.CapabilityID, pair.ComponentID, actor)
	if statusCode == http.StatusCreated {
		location := sharedAPI.APIVersionPrefix + realizationRoleItemResourcePath(pair.CapabilityID, pair.ComponentID)
		sharedAPI.RespondCreated(w, location, dto)
		return
	}
	sharedAPI.RespondJSON(w, statusCode, dto)
}
