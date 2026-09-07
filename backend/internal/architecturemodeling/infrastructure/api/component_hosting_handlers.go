package api

import (
	"net/http"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
)

type ComponentHostingHandlers struct {
	componentCommandHandlers
}

func NewComponentHostingHandlers(
	commandBus cqrs.CommandBus,
	readModel *readmodels.ApplicationComponentReadModel,
	hateoas *ArchitectureModelingLinks,
) *ComponentHostingHandlers {
	return &ComponentHostingHandlers{
		componentCommandHandlers: componentCommandHandlers{
			commandBus: commandBus,
			readModel:  readModel,
			hateoas:    hateoas,
		},
	}
}

type ClassifyHostingRequest struct {
	Hosting string `json:"hosting"`
}

// ClassifyHosting godoc
// @Summary Classify where an application component is hosted
// @Description Sets the hosting classification (on-premises, cloud, saas, third-party-hosted, or unknown); reclassification is unrestricted
// @Tags components
// @Accept json
// @Produce json
// @Param id path string true "Component ID"
// @Param hosting body ClassifyHostingRequest true "Hosting classification"
// @Success 200 {object} readmodels.ApplicationComponentDTO
// @Failure 400 {object} sharedAPI.ErrorResponse
// @Failure 404 {object} sharedAPI.ErrorResponse
// @Failure 500 {object} sharedAPI.ErrorResponse
// @Router /components/{id}/hosting [put]
func (h *ComponentHostingHandlers) ClassifyHosting(w http.ResponseWriter, r *http.Request) {
	id := sharedAPI.GetPathParam(r, "id")

	req, ok := sharedAPI.DecodeRequestOrFail[ClassifyHostingRequest](w, r)
	if !ok {
		return
	}

	h.dispatchAndRespond(w, r, id, &commands.ClassifyApplicationHosting{ComponentID: id, Hosting: req.Hosting})
}
