package api

import (
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

func (h *ArchitectureModelingLinks) AddHostingAffordances(links sharedAPI.Links, id string, actor sharedctx.Actor) {
	if !actor.CanWrite("components") {
		return
	}
	links["x-classify-hosting"] = h.Put("/components/" + id + "/hosting")
}
