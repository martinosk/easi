package api

import (
	"easi/backend/internal/architecturemodeling/application/readmodels"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

func (h *ArchitectureModelingLinks) AddContainmentAffordances(links sharedAPI.Links, component *readmodels.ApplicationComponentDTO, actor sharedctx.Actor) {
	if !actor.CanWrite("components") {
		return
	}
	containment := "/components/" + component.ID + "/containment"
	switch {
	case component.PartOf != nil:
		links["x-detach"] = h.Del(containment)
	case len(component.Parts) == 0:
		links["x-attach-to"] = h.Put(containment)
	}
}
