package api

import (
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

func (h *ArchitectureModelingLinks) AddOwnershipAffordances(links sharedAPI.Links, id, state string, actor sharedctx.Actor) {
	if !actor.CanWrite("components") {
		return
	}
	p := "/components/" + id + "/ownership"
	switch state {
	case valueobjects.OwnershipStateUnknown:
		links["x-nominate-owner"] = h.Post(p + "/nomination")
		links["x-assign-owner"] = h.Put(p)
	case valueobjects.OwnershipStateNominated:
		links["x-confirm-owner"] = h.Post(p + "/confirmation")
		links["x-clear-owner"] = h.Del(p)
	case valueobjects.OwnershipStateOwned, valueobjects.OwnershipStateManaged:
		links["x-clear-owner"] = h.Del(p)
	}
}

func (h *ArchitectureModelingLinks) OwnershipStatisticsLinks() sharedAPI.Links {
	return sharedAPI.Links{
		"self": h.Get("/components/ownership-statistics"),
	}
}
