package api

import (
	"net/http"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
)

type ComponentContainmentHandlers struct {
	componentCommandHandlers
}

func NewComponentContainmentHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.ApplicationComponentReadModel,
	hateoas *ArchitectureModelingLinks,
) *ComponentContainmentHandlers {
	return &ComponentContainmentHandlers{
		componentCommandHandlers: componentCommandHandlers{
			commandBus: commandBus,
			readModel:  readModel,
			hateoas:    hateoas,
		},
	}
}

type AttachComponentRequest struct {
	ParentID string `json:"parentId"`
	Kind     string `json:"kind"`
}

// AttachComponent godoc
// @Summary Attach an application component to a parent as a part
// @Description Makes the component a part of the parent by composition (the part exists only within its parent) or aggregation (the part is autonomous). Containment is capped at two levels: a part cannot accept parts and a parent cannot become a part.
// @Tags components
// @Accept json
// @Produce json
// @Param id path string true "Component ID"
// @Param containment body AttachComponentRequest true "Parent reference and containment kind"
// @Success 200 {object} readmodels.ApplicationComponentDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /components/{id}/containment [put]
func (h *ComponentContainmentHandlers) AttachComponent(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[AttachComponentRequest](w, r)
	if !ok {
		return
	}

	h.dispatchAndRespond(w, r, id, &commands.AttachComponent{ComponentID: id, ParentID: req.ParentID, Kind: req.Kind})
}

// DetachComponent godoc
// @Summary Detach an application component from its parent
// @Description Makes the part a standalone component again; its relations, realisations, experts, ownership, hosting and one-pager are untouched
// @Tags components
// @Produce json
// @Param id path string true "Component ID"
// @Success 200 {object} readmodels.ApplicationComponentDTO
// @Failure 401 {object} sharedAPI.ErrorResponse
// @Failure 403 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 409 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Security CookieAuth
// @Router /components/{id}/containment [delete]
func (h *ComponentContainmentHandlers) DetachComponent(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	h.dispatchAndRespond(w, r, id, &commands.DetachComponent{ComponentID: id})
}
