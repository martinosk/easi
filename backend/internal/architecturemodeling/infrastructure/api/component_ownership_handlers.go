package api

import (
	"net/http"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
)

type ComponentOwnershipHandlers struct {
	componentCommandHandlers
}

func NewComponentOwnershipHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.ApplicationComponentReadModel,
	hateoas *ArchitectureModelingLinks,
) *ComponentOwnershipHandlers {
	return &ComponentOwnershipHandlers{
		componentCommandHandlers: componentCommandHandlers{
			commandBus: commandBus,
			readModel:  readModel,
			hateoas:    hateoas,
		},
	}
}

type OwnerReferenceRequest struct {
	OwnerKind string `json:"ownerKind"`
	OwnerID   string `json:"ownerId"`
}

// NominateOwner godoc
// @Summary Nominate an owner for an application component
// @Description Records a user or internal team as the nominated owner candidate; the component moves to ownership state "nominated"
// @Tags components
// @Accept json
// @Produce json
// @Param id path string true "Component ID"
// @Param owner body OwnerReferenceRequest true "Owner reference (kind: user or team)"
// @Success 200 {object} readmodels.ApplicationComponentDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{id}/ownership/nomination [post]
func (h *ComponentOwnershipHandlers) NominateOwner(w http.ResponseWriter, r *http.Request) {
	h.transitionWithOwner(w, r, func(id string, req OwnerReferenceRequest) cqrs.Command {
		return &commands.NominateApplicationComponentOwner{ComponentID: id, OwnerKind: req.OwnerKind, OwnerID: req.OwnerID}
	})
}

// ConfirmOwnership godoc
// @Summary Confirm a nominated owner
// @Description Confirms the nominated owner; the component resolves to "owned" (user) or "managed" (team)
// @Tags components
// @Produce json
// @Param id path string true "Component ID"
// @Success 200 {object} readmodels.ApplicationComponentDTO
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{id}/ownership/confirmation [post]
func (h *ComponentOwnershipHandlers) ConfirmOwnership(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	h.dispatchAndRespond(w, r, id, &commands.ConfirmApplicationComponentOwnership{ComponentID: id})
}

// AssignOwner godoc
// @Summary Assign an owner directly
// @Description Assigns a user or internal team as owner without nomination; the component resolves to "owned" (user) or "managed" (team)
// @Tags components
// @Accept json
// @Produce json
// @Param id path string true "Component ID"
// @Param owner body OwnerReferenceRequest true "Owner reference (kind: user or team)"
// @Success 200 {object} readmodels.ApplicationComponentDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{id}/ownership [put]
func (h *ComponentOwnershipHandlers) AssignOwner(w http.ResponseWriter, r *http.Request) {
	h.transitionWithOwner(w, r, func(id string, req OwnerReferenceRequest) cqrs.Command {
		return &commands.AssignApplicationComponentOwner{ComponentID: id, OwnerKind: req.OwnerKind, OwnerID: req.OwnerID}
	})
}

// ClearOwnership godoc
// @Summary Clear ownership of an application component
// @Description Removes the owner reference and returns the component to ownership state "unknown"
// @Tags components
// @Produce json
// @Param id path string true "Component ID"
// @Success 204
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{id}/ownership [delete]
func (h *ComponentOwnershipHandlers) ClearOwnership(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	if _, err := h.commandBus.Dispatch(r.Context(), &commands.ClearApplicationComponentOwnership{ComponentID: id}); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ComponentOwnershipHandlers) transitionWithOwner(w http.ResponseWriter, r *http.Request, buildCommand func(id string, req OwnerReferenceRequest) cqrs.Command) {
	id := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[OwnerReferenceRequest](w, r)
	if !ok {
		return
	}

	h.dispatchAndRespond(w, r, id, buildCommand(id, req))
}
