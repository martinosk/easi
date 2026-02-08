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

type LinkStatusParams struct {
	CapabilityID  string
	Status        string
	LinkedToID    *string
	BlockingCapID *string
	BlockingEcID  *string
}

func (h *EnterpriseArchLinks) EnterpriseCapabilityLinksForActor(id string, actor sharedctx.Actor) sharedAPI.Links {
	p := "/enterprise-capabilities/" + id
	links := sharedAPI.Links{
		"self":                   h.Get(p),
		"x-links":                h.Get(p + "/links"),
		"x-strategic-importance": h.Get(p + "/strategic-importance"),
	}
	if actor.CanWrite("enterprise-arch") {
		links["edit"] = h.Put(p)
		links["x-create-link"] = h.Post(p + "/links")
	}
	if actor.CanDelete("enterprise-arch") {
		links["delete"] = h.Del(p)
	}
	return links
}

func (h *EnterpriseArchLinks) EnterpriseCapabilityCollectionLinks() sharedAPI.Links {
	return sharedAPI.Links{"self": h.Get("/enterprise-capabilities")}
}

func (h *EnterpriseArchLinks) EnterpriseCapabilityLinkLinks(ecID, linkID string) sharedAPI.Links {
	p := "/enterprise-capabilities/" + ecID + "/links/" + linkID
	return sharedAPI.Links{
		"self": h.Get(p), "delete": h.Del(p),
		"x-enterprise-capability": h.Get("/enterprise-capabilities/" + ecID),
	}
}

func (h *EnterpriseArchLinks) EnterpriseCapabilityLinksCollectionLinks(ecID string) sharedAPI.Links {
	return sharedAPI.Links{
		"self":                    h.Get("/enterprise-capabilities/" + ecID + "/links"),
		"x-enterprise-capability": h.Get("/enterprise-capabilities/" + ecID),
	}
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

func (h *EnterpriseArchLinks) DomainCapabilityEnterpriseLinks(dcID string) sharedAPI.Links {
	return sharedAPI.Links{"self": h.Get("/domain-capabilities/" + dcID + "/enterprise-capability")}
}

func (h *EnterpriseArchLinks) DomainCapabilityEnterpriseLinkedLinks(dcID, ecID, linkID string) sharedAPI.Links {
	return sharedAPI.Links{
		"self":                    h.Get("/domain-capabilities/" + dcID + "/enterprise-capability"),
		"x-enterprise-capability": h.Get("/enterprise-capabilities/" + ecID),
		"x-unlink":                h.Del("/enterprise-capabilities/" + ecID + "/links/" + linkID),
	}
}

func (h *EnterpriseArchLinks) CapabilityLinkStatusLinks(p LinkStatusParams) sharedAPI.Links {
	links := sharedAPI.Links{"self": h.Get("/domain-capabilities/" + p.CapabilityID + "/enterprise-link-status")}
	if p.Status == "available" {
		links["x-available-enterprise-capabilities"] = h.Get("/enterprise-capabilities")
	}
	if p.LinkedToID != nil {
		links["x-linked-to"] = h.Get("/enterprise-capabilities/" + *p.LinkedToID)
		links["x-enterprise-capability"] = h.Get("/domain-capabilities/" + p.CapabilityID + "/enterprise-capability")
	}
	if p.BlockingCapID != nil {
		links["x-blocking-capability"] = h.Get("/capabilities/" + *p.BlockingCapID)
	}
	if p.BlockingEcID != nil {
		links["x-blocking-enterprise-capability"] = h.Get("/enterprise-capabilities/" + *p.BlockingEcID)
	}
	return links
}

func (h *EnterpriseArchLinks) MaturityAnalysisCandidateLinks(ecID string) sharedAPI.Links {
	return sharedAPI.Links{
		"self":           h.Get("/enterprise-capabilities/" + ecID),
		"x-maturity-gap": h.Get("/enterprise-capabilities/" + ecID + "/maturity-gap"),
	}
}

func (h *EnterpriseArchLinks) MaturityAnalysisCollectionLinks() sharedAPI.Links {
	return sharedAPI.Links{"self": h.Get("/enterprise-capabilities/maturity-analysis")}
}

func (h *EnterpriseArchLinks) MaturityGapDetailLinks(ecID string) sharedAPI.Links {
	p := "/enterprise-capabilities/" + ecID
	return sharedAPI.Links{
		"self":                  h.Get(p + "/maturity-gap"),
		"up":                    h.Get(p),
		"x-set-target-maturity": h.Put(p + "/target-maturity"),
	}
}

func (h *EnterpriseArchLinks) TimeSuggestionsCollectionLinks() sharedAPI.Links {
	return sharedAPI.Links{"self": h.Get("/time-suggestions")}
}

func (h *EnterpriseArchLinks) UnlinkedCapabilitiesLinks() sharedAPI.Links {
	return sharedAPI.Links{"self": h.Get("/domain-capabilities/unlinked")}
}
