package api

import (
	"net/http"

	"easi/backend/internal/architecturemodeling/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
)

type componentCommandHandlers struct {
	commandBus cqrs.CommandBus
	readModel  *readmodels.ApplicationComponentReadModel
	hateoas    *ArchitectureModelingLinks
}

func (h componentCommandHandlers) dispatchAndRespond(w http.ResponseWriter, r *http.Request, id string, cmd cqrs.Command) {
	if _, err := h.commandBus.Dispatch(r.Context(), cmd); err != nil {
		sharedAPI.HandleError(w, err)
		return
	}

	component, err := h.readModel.GetByID(r.Context(), id)
	if err != nil {
		sharedAPI.RespondError(w, http.StatusInternalServerError, err, "Failed to retrieve component")
		return
	}
	if component == nil {
		sharedAPI.RespondError(w, http.StatusNotFound, nil, "Component not found")
		return
	}

	enrichComponentDTO(r, h.hateoas, component)
	sharedAPI.RespondJSON(w, http.StatusOK, component)
}
