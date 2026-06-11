package api

import (
	"errors"
	"net/http"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/services"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

type AddDirectionSourceRequest struct {
	CapabilityID string `json:"capabilityId"`
}

// AddDirectionSource godoc
// @Summary Add a domain capability to the active direction's source set
// @Description Adds a source to the active direction (R1 same-node exclusivity checked; agreed directions are immutable). Idempotent: re-adding an existing source returns the unchanged direction.
// @Tags architecturedirection
// @Accept json
// @Produce json
// @Security CookieAuth
// @Param id path string true "Enterprise capability ID"
// @Param body body AddDirectionSourceRequest true "Capability to add"
// @Success 200 {object} easi_backend_internal_architecturedirection_application_readmodels.DirectionDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/direction/sources [post]
func (h *DirectionHandlers) AddDirectionSource(w http.ResponseWriter, r *http.Request) {
	ecID := sharedAPI.GetPathParam(r, "id")
	req, ok := sharedAPI.DecodeRequestOrFail[AddDirectionSourceRequest](w, r)
	if !ok {
		return
	}
	direction, ok := h.resolveActiveDirection(w, r, ecID)
	if !ok {
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	cmd := &commands.AddDirectionSource{
		DirectionID:  direction.ID,
		CapabilityID: req.CapabilityID,
		Actor:        actor.Email,
	}
	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		h.respondSourceMutationError(w, err, ecID)
		return
	}
	h.respondWithActiveDirection(w, r, ecID, http.StatusOK)
}

// RemoveDirectionSource godoc
// @Summary Remove a domain capability from the active direction's source set
// @Description Excludes a source from the active direction; the capability and its subtree leave the EC's composition. Agreed directions are immutable.
// @Tags architecturedirection
// @Produce json
// @Security CookieAuth
// @Param id path string true "Enterprise capability ID"
// @Param capabilityId path string true "Domain capability ID to remove"
// @Success 204 "No Content"
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /enterprise-capabilities/{id}/direction/sources/{capabilityId} [delete]
func (h *DirectionHandlers) RemoveDirectionSource(w http.ResponseWriter, r *http.Request) {
	ecID := sharedAPI.GetPathParam(r, "id")
	capabilityID := sharedAPI.GetPathParam(r, "capabilityId")
	direction, ok := h.resolveActiveDirection(w, r, ecID)
	if !ok {
		return
	}
	actor, _ := sharedctx.GetActor(r.Context())
	cmd := &commands.RemoveDirectionSource{
		DirectionID:  direction.ID,
		CapabilityID: capabilityID,
		Actor:        actor.Email,
	}
	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		h.respondSourceMutationError(w, err, ecID)
		return
	}
	sharedAPI.RespondNoContent(w)
}

func (h *DirectionHandlers) respondSourceMutationError(w http.ResponseWriter, err error, ecID string) {
	var conflict *services.SourceConflictError
	if errors.As(err, &conflict) {
		h.respondSourceConflict(w, conflict)
		return
	}
	if errors.Is(err, aggregates.ErrDirectionAgreedImmutable) {
		h.respondAgreedImmutable(w, ecID)
		return
	}
	sharedAPI.HandleError(w, err)
}

func (h *DirectionHandlers) respondSourceConflict(w http.ResponseWriter, conflict *services.SourceConflictError) {
	sharedAPI.RespondErrorWithLinks(w, sharedAPI.ErrorWithLinksParams{
		StatusCode: http.StatusConflict,
		Message:    conflict.Error(),
		Details: map[string]string{
			"capabilityId":                        conflict.Conflict.CapabilityID,
			"capabilityName":                      conflict.Conflict.CapabilityName,
			"conflictingEnterpriseCapabilityId":   conflict.Conflict.EnterpriseCapabilityID,
			"conflictingEnterpriseCapabilityName": conflict.Conflict.EnterpriseCapabilityName,
		},
		Links: sharedAPI.Links{
			"x-conflicting-ec": h.hateoas.Get(enterpriseCapabilityResourcePath(conflict.Conflict.EnterpriseCapabilityID)),
		},
	})
}

func (h *DirectionHandlers) respondAgreedImmutable(w http.ResponseWriter, ecID string) {
	sharedAPI.RespondErrorWithLinks(w, sharedAPI.ErrorWithLinksParams{
		StatusCode: http.StatusConflict,
		Message:    "This direction is agreed and its source set is frozen. To recompose, reject the direction and capture a new one.",
		Details:    map[string]string{"directionStatus": "agreed"},
		Links: sharedAPI.Links{
			"x-reject": h.hateoas.Post(directionResourcePath(ecID) + "/reject"),
		},
	})
}
