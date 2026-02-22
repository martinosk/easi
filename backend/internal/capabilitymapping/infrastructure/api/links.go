package api

import (
	sharedAPI "easi/backend/internal/shared/api"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type CapabilityMappingLinks struct {
	*sharedAPI.HATEOASLinks
}

func NewCapabilityMappingLinks(h *sharedAPI.HATEOASLinks) *CapabilityMappingLinks {
	return &CapabilityMappingLinks{HATEOASLinks: h}
}

func (h *CapabilityMappingLinks) CapabilityLinksForActor(id, parentID string, actor sharedctx.Actor) sharedAPI.Links {
	links := h.capabilityBaseForActor(id, actor)
	if parentID != "" {
		links["up"] = h.Get("/capabilities/" + parentID)
	}
	return links
}

func (h *CapabilityMappingLinks) capabilityBaseForActor(id string, actor sharedctx.Actor) sharedAPI.Links {
	p := "/capabilities/" + id
	links := sharedAPI.Links{
		"self":                    h.Get(p),
		"x-children":              h.Get(p + "/children"),
		"x-systems":               h.Get(p + "/systems"),
		"x-outgoing-dependencies": h.Get(p + "/dependencies/outgoing"),
		"x-incoming-dependencies": h.Get(p + "/dependencies/incoming"),
		"collection":              h.Get("/capabilities"),
		"x-expert-roles":          h.Get("/capabilities/expert-roles"),
	}
	h.AddEditOrGrantLink(links, actor, "capabilities", "capabilities", id, h.Put(p), map[string]types.Link{
		"x-add-expert": h.Post(p + "/experts"),
	})
	if actor.CanDelete("capabilities") {
		links["delete"] = h.Del(p)
	}
	h.AddEditGrantsLink(links, actor, "capabilities")
	return links
}

func (h *CapabilityMappingLinks) CapabilityExpertLinksForActor(p sharedAPI.ExpertParams, actor sharedctx.Actor) sharedAPI.Links {
	return h.ExpertRemoveLink(p, actor, "capabilities")
}

func (h *CapabilityMappingLinks) DependencyLinks(id, srcCapID, tgtCapID string) sharedAPI.Links {
	p := "/capability-dependencies/" + id
	return sharedAPI.Links{
		"self": h.Get(p), "delete": h.Del(p),
		"x-source-capability": h.Get("/capabilities/" + srcCapID),
		"x-target-capability": h.Get("/capabilities/" + tgtCapID),
		"collection":          h.Get("/capability-dependencies"),
	}
}

func (h *CapabilityMappingLinks) RealizationLinks(id, capID, compID string) sharedAPI.Links {
	p := "/capability-realizations/" + id
	return sharedAPI.Links{
		"self": h.Get(p), "edit": h.Put(p), "delete": h.Del(p),
		"x-capability": h.Get("/capabilities/" + capID),
		"x-component":  h.Get("/components/" + compID),
	}
}

func (h *CapabilityMappingLinks) BusinessDomainLinksForActor(id string, hasCaps bool, actor sharedctx.Actor) sharedAPI.Links {
	p := "/business-domains/" + id
	links := sharedAPI.Links{
		"self":           h.Get(p),
		"x-capabilities": h.Get(p + "/capabilities"),
		"collection":     h.Get("/business-domains"),
	}
	if actor.CanWrite("domains") {
		links["edit"] = h.Put(p)
	}
	if actor.CanDelete("domains") && !hasCaps {
		links["delete"] = h.Del(p)
	}
	h.AddEditGrantsLink(links, actor, "domains")
	return links
}

func (h *CapabilityMappingLinks) BusinessDomainCollectionLinksForActor(actor sharedctx.Actor) sharedAPI.Links {
	links := sharedAPI.Links{"self": h.Get("/business-domains")}
	if actor.CanWrite("domains") {
		links["create"] = h.Post("/business-domains")
	}
	return links
}

func (h *CapabilityMappingLinks) capabilityInDomainBase(capID, _ string) sharedAPI.Links {
	return sharedAPI.Links{
		"self":               h.Get("/capabilities/" + capID),
		"x-children":         h.Get("/capabilities/" + capID + "/children"),
		"x-business-domains": h.Get("/capabilities/" + capID + "/business-domains"),
	}
}

func (h *CapabilityMappingLinks) CapabilityInDomainLinksForActor(capID, domID string, actor sharedctx.Actor) sharedAPI.Links {
	links := h.capabilityInDomainBase(capID, domID)
	if actor.CanDelete("domains") {
		links["x-remove-from-domain"] = h.Del("/business-domains/" + domID + "/capabilities/" + capID)
	}
	return links
}

func (h *CapabilityMappingLinks) domainForCapabilityBase(domID string) sharedAPI.Links {
	p := "/business-domains/" + domID
	return sharedAPI.Links{
		"self":           h.Get(p),
		"x-capabilities": h.Get(p + "/capabilities"),
	}
}

func (h *CapabilityMappingLinks) DomainForCapabilityLinksForActor(domID, capID string, actor sharedctx.Actor) sharedAPI.Links {
	links := h.domainForCapabilityBase(domID)
	if actor.CanDelete("domains") {
		links["x-remove-capability"] = h.Del("/business-domains/" + domID + "/capabilities/" + capID)
	}
	return links
}

func (h *CapabilityMappingLinks) assignmentBase(domID, capID string) sharedAPI.Links {
	return sharedAPI.Links{
		"x-capability":      h.Get("/capabilities/" + capID),
		"x-business-domain": h.Get("/business-domains/" + domID),
	}
}

func (h *CapabilityMappingLinks) AssignmentLinksForActor(domID, capID string, actor sharedctx.Actor) sharedAPI.Links {
	links := h.assignmentBase(domID, capID)
	if actor.CanDelete("domains") {
		links["x-remove"] = h.Del("/business-domains/" + domID + "/capabilities/" + capID)
	}
	return links
}

func (h *CapabilityMappingLinks) strategyImportanceBase(domID, capID, impID string) sharedAPI.Links {
	p := "/business-domains/" + domID + "/capabilities/" + capID + "/importance/" + impID
	return sharedAPI.Links{
		"self":         h.Get(p),
		"x-capability": h.Get("/capabilities/" + capID),
		"x-domain":     h.Get("/business-domains/" + domID),
	}
}

func (h *CapabilityMappingLinks) StrategyImportanceLinksForActor(domID, capID, impID string, actor sharedctx.Actor) sharedAPI.Links {
	links := h.strategyImportanceBase(domID, capID, impID)
	p := "/business-domains/" + domID + "/capabilities/" + capID + "/importance/" + impID
	if actor.CanWrite("domains") {
		links["edit"] = h.Put(p)
	}
	if actor.CanDelete("domains") {
		links["delete"] = h.Del(p)
	}
	return links
}

func (h *CapabilityMappingLinks) StrategyImportanceCollectionLinksForActor(domID, capID string, actor sharedctx.Actor) sharedAPI.Links {
	p := "/business-domains/" + domID + "/capabilities/" + capID + "/importance"
	links := sharedAPI.Links{"self": h.Get(p)}
	if actor.CanWrite("domains") {
		links["create"] = h.Post(p)
	}
	return links
}

func (h *CapabilityMappingLinks) fitScoreBase(componentID, path string) sharedAPI.Links {
	return sharedAPI.Links{
		"self": h.Get(path),
		"up":   h.Get("/components/" + componentID),
	}
}

func (h *CapabilityMappingLinks) FitScoreLinksForActor(componentID, pillarID string, actor sharedctx.Actor) sharedAPI.Links {
	p := "/components/" + componentID + "/fit-scores/" + pillarID
	links := h.fitScoreBase(componentID, p)
	if actor.CanWrite("components") {
		links["edit"] = h.Put(p)
	}
	if actor.CanDelete("components") {
		links["delete"] = h.Del(p)
	}
	return links
}

func (h *CapabilityMappingLinks) FitScoresCollectionLinksForActor(componentID string, actor sharedctx.Actor) sharedAPI.Links {
	p := "/components/" + componentID + "/fit-scores"
	links := h.fitScoreBase(componentID, p)
	if actor.CanWrite("components") {
		links["create"] = h.Put(p + "/{pillarId}")
	}
	return links
}
