package api

import (
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
)

type EnterpriseArchLinks struct {
	*sharedAPI.HATEOASLinks
}

func NewEnterpriseArchLinks(h *sharedAPI.HATEOASLinks) *EnterpriseArchLinks {
	return &EnterpriseArchLinks{HATEOASLinks: h}
}

func (h *EnterpriseArchLinks) EnterpriseCapabilityLinksForActor(id string, actor sharedctx.Actor) sharedAPI.Links {
	p := "/enterprise-capabilities/" + id
	links := sharedAPI.Links{
		"self":                   h.Get(p),
		"x-strategic-importance": h.Get(p + "/strategic-importance"),
		"x-one-pager":            h.Get("/one-pagers/enterprise-capability/" + id),
	}
	if actor.CanWrite("enterprise-arch") {
		links["edit"] = h.Put(p)
	}
	if actor.CanDelete("enterprise-arch") {
		links["delete"] = h.Del(p)
	}
	return links
}

func (h *EnterpriseArchLinks) EnterpriseCapabilityCollectionLinks() sharedAPI.Links {
	return sharedAPI.Links{"self": h.Get("/enterprise-capabilities")}
}

func (h *EnterpriseArchLinks) enterpriseStrategicImportanceBase(ecID, impID string) sharedAPI.Links {
	p := "/enterprise-capabilities/" + ecID + "/strategic-importance/" + impID
	return sharedAPI.Links{
		"self":                    h.Get(p),
		"x-enterprise-capability": h.Get("/enterprise-capabilities/" + ecID),
	}
}

func (h *EnterpriseArchLinks) EnterpriseStrategicImportanceLinksForActor(ecID, impID string, actor sharedctx.Actor) sharedAPI.Links {
	links := h.enterpriseStrategicImportanceBase(ecID, impID)
	p := "/enterprise-capabilities/" + ecID + "/strategic-importance/" + impID
	if actor.CanWrite("enterprise-arch") {
		links["edit"] = h.Put(p)
	}
	if actor.CanDelete("enterprise-arch") {
		links["delete"] = h.Del(p)
	}
	return links
}

func (h *EnterpriseArchLinks) EnterpriseStrategicImportanceCollectionLinks(ecID string) sharedAPI.Links {
	return sharedAPI.Links{
		"self":                    h.Get("/enterprise-capabilities/" + ecID + "/strategic-importance"),
		"x-enterprise-capability": h.Get("/enterprise-capabilities/" + ecID),
	}
}

func (h *EnterpriseArchLinks) TimeSuggestionsCollectionLinks() sharedAPI.Links {
	return sharedAPI.Links{"self": h.Get("/time-suggestions")}
}
