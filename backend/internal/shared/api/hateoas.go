package api

import (
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/types"
)

type Links = types.Links

type HATEOASLinks struct {
	base string
}

func NewHATEOASLinks(baseURL string) *HATEOASLinks {
	if baseURL == "" {
		baseURL = APIVersionPrefix
	}
	return &HATEOASLinks{base: baseURL}
}

func (h *HATEOASLinks) l(href, method string) types.Link {
	return types.Link{Href: href, Method: method}
}

func (h *HATEOASLinks) get(path string) types.Link  { return h.l(h.base+path, "GET") }
func (h *HATEOASLinks) put(path string) types.Link  { return h.l(h.base+path, "PUT") }
func (h *HATEOASLinks) post(path string) types.Link { return h.l(h.base+path, "POST") }
func (h *HATEOASLinks) del(path string) types.Link  { return h.l(h.base+path, "DELETE") }
func (h *HATEOASLinks) patch(path string) types.Link { return h.l(h.base+path, "PATCH") }

func (h *HATEOASLinks) crud(path string) Links {
	return Links{"self": h.get(path), "edit": h.put(path), "delete": h.del(path)}
}

func (h *HATEOASLinks) ComponentLinks(id string) Links {
	links := h.crud("/components/" + id)
	links["describedby"] = h.get("/reference/components")
	links["collection"] = h.get("/components")
	return links
}

func (h *HATEOASLinks) ComponentLinksForActor(id string, actor sharedctx.Actor) Links {
	p := "/components/" + id
	links := Links{
		"self":        h.get(p),
		"describedby": h.get("/reference/components"),
		"collection":  h.get("/components"),
	}
	if actor.CanWrite("components") {
		links["edit"] = h.put(p)
	}
	if actor.CanDelete("components") {
		links["delete"] = h.del(p)
	}
	return links
}

func (h *HATEOASLinks) RelationLinks(id string) Links {
	links := h.crud("/relations/" + id)
	links["describedby"] = h.get("/reference/relations/generic")
	links["collection"] = h.get("/relations")
	return links
}

func (h *HATEOASLinks) RelationTypeLinks(relationType string) Links {
	doc := "relations/generic"
	if relationType == "Triggers" {
		doc = "relations/triggering"
	} else if relationType == "Serves" {
		doc = "relations/serving"
	}
	return Links{"describedby": h.get("/reference/" + doc)}
}

func (h *HATEOASLinks) ViewLinks(id string) Links {
	return Links{
		"self":         h.get("/views/" + id),
		"x-components": h.get("/views/" + id + "/components"),
		"collection":   h.get("/views"),
	}
}

type ViewPermissions struct {
	IsPrivate   bool
	IsDefault   bool
	OwnerUserID *string
	CurrentUser string
}

type ViewInfo struct {
	ID          string
	IsPrivate   bool
	IsDefault   bool
	OwnerUserID *string
}

func (h *HATEOASLinks) ViewLinksForActor(v ViewInfo, actor sharedctx.Actor) Links {
	links := Links{
		"self":         h.get("/views/" + v.ID),
		"x-components": h.get("/views/" + v.ID + "/components"),
		"collection":   h.get("/views"),
	}
	isOwner := v.OwnerUserID != nil && *v.OwnerUserID == actor.ID
	canEdit := (!v.IsPrivate || isOwner) && actor.CanWrite("views")
	if canEdit {
		links["edit"] = h.patch("/views/" + v.ID + "/name")
		links["x-change-visibility"] = h.patch("/views/" + v.ID + "/visibility")
	}
	canDelete := (!v.IsPrivate || isOwner) && actor.CanDelete("views") && !v.IsDefault
	if canDelete {
		links["delete"] = h.del("/views/" + v.ID)
	}
	return links
}

func (h *HATEOASLinks) ViewLinksWithPermissions(id string, p ViewPermissions) Links {
	links := Links{
		"self":         h.get("/views/" + id),
		"x-components": h.get("/views/" + id + "/components"),
		"collection":   h.get("/views"),
	}
	isOwner := p.OwnerUserID != nil && *p.OwnerUserID == p.CurrentUser
	if canEdit := !p.IsPrivate || isOwner; canEdit {
		links["edit"] = h.patch("/views/" + id + "/name")
		links["x-change-visibility"] = h.patch("/views/" + id + "/visibility")
		if !p.IsDefault {
			links["delete"] = h.del("/views/" + id)
		}
	}
	return links
}

func (h *HATEOASLinks) CapabilityLinks(id, parentID string) Links {
	links := h.capabilityBase(id)
	if parentID != "" {
		links["up"] = h.get("/capabilities/" + parentID)
	}
	return links
}

func (h *HATEOASLinks) CapabilityLinksForActor(id, parentID string, actor sharedctx.Actor) Links {
	links := h.capabilityBaseForActor(id, actor)
	if parentID != "" {
		links["up"] = h.get("/capabilities/" + parentID)
	}
	return links
}

func (h *HATEOASLinks) capabilityBase(id string) Links {
	p := "/capabilities/" + id
	return Links{
		"self": h.get(p), "edit": h.put(p), "delete": h.del(p),
		"x-children":              h.get(p + "/children"),
		"x-systems":               h.get(p + "/systems"),
		"x-outgoing-dependencies": h.get(p + "/dependencies/outgoing"),
		"x-incoming-dependencies": h.get(p + "/dependencies/incoming"),
		"collection":              h.get("/capabilities"),
	}
}

func (h *HATEOASLinks) capabilityBaseForActor(id string, actor sharedctx.Actor) Links {
	p := "/capabilities/" + id
	links := Links{
		"self":                    h.get(p),
		"x-children":              h.get(p + "/children"),
		"x-systems":               h.get(p + "/systems"),
		"x-outgoing-dependencies": h.get(p + "/dependencies/outgoing"),
		"x-incoming-dependencies": h.get(p + "/dependencies/incoming"),
		"collection":              h.get("/capabilities"),
	}
	if actor.CanWrite("capabilities") {
		links["edit"] = h.put(p)
	}
	if actor.CanDelete("capabilities") {
		links["delete"] = h.del(p)
	}
	return links
}

func (h *HATEOASLinks) DependencyLinks(id, srcCapID, tgtCapID string) Links {
	p := "/capability-dependencies/" + id
	return Links{
		"self": h.get(p), "delete": h.del(p),
		"x-source-capability": h.get("/capabilities/" + srcCapID),
		"x-target-capability": h.get("/capabilities/" + tgtCapID),
		"collection":          h.get("/capability-dependencies"),
	}
}

func (h *HATEOASLinks) RealizationLinks(id, capID, compID string) Links {
	p := "/capability-realizations/" + id
	return Links{
		"self": h.get(p), "edit": h.put(p), "delete": h.del(p),
		"x-capability": h.get("/capabilities/" + capID),
		"x-component":  h.get("/components/" + compID),
	}
}

func (h *HATEOASLinks) BusinessDomainLinks(id string, hasCaps bool) Links {
	p := "/business-domains/" + id
	links := Links{
		"self": h.get(p), "edit": h.put(p),
		"x-capabilities": h.get(p + "/capabilities"),
		"collection":     h.get("/business-domains"),
	}
	if !hasCaps {
		links["delete"] = h.del(p)
	}
	return links
}

func (h *HATEOASLinks) BusinessDomainLinksForActor(id string, hasCaps bool, actor sharedctx.Actor) Links {
	p := "/business-domains/" + id
	links := Links{
		"self":           h.get(p),
		"x-capabilities": h.get(p + "/capabilities"),
		"collection":     h.get("/business-domains"),
	}
	if actor.CanWrite("domains") {
		links["edit"] = h.put(p)
	}
	if actor.CanDelete("domains") && !hasCaps {
		links["delete"] = h.del(p)
	}
	return links
}

func (h *HATEOASLinks) BusinessDomainCollectionLinksForActor(actor sharedctx.Actor) Links {
	links := Links{"self": h.get("/business-domains")}
	if actor.CanWrite("domains") {
		links["create"] = h.post("/business-domains")
	}
	return links
}

func (h *HATEOASLinks) CapabilityInDomainLinks(capID, domID string) Links {
	return Links{
		"self":                 h.get("/capabilities/" + capID),
		"x-children":           h.get("/capabilities/" + capID + "/children"),
		"x-business-domains":   h.get("/capabilities/" + capID + "/business-domains"),
		"x-remove-from-domain": h.del("/business-domains/" + domID + "/capabilities/" + capID),
	}
}

func (h *HATEOASLinks) CapabilityInDomainLinksForActor(capID, domID string, actor sharedctx.Actor) Links {
	links := Links{
		"self":               h.get("/capabilities/" + capID),
		"x-children":         h.get("/capabilities/" + capID + "/children"),
		"x-business-domains": h.get("/capabilities/" + capID + "/business-domains"),
	}
	if actor.CanDelete("domains") {
		links["x-remove-from-domain"] = h.del("/business-domains/" + domID + "/capabilities/" + capID)
	}
	return links
}

func (h *HATEOASLinks) DomainForCapabilityLinks(domID, capID string) Links {
	p := "/business-domains/" + domID
	return Links{
		"self":                h.get(p),
		"x-capabilities":      h.get(p + "/capabilities"),
		"x-remove-capability": h.del(p + "/capabilities/" + capID),
	}
}

func (h *HATEOASLinks) DomainForCapabilityLinksForActor(domID, capID string, actor sharedctx.Actor) Links {
	p := "/business-domains/" + domID
	links := Links{
		"self":           h.get(p),
		"x-capabilities": h.get(p + "/capabilities"),
	}
	if actor.CanDelete("domains") {
		links["x-remove-capability"] = h.del(p + "/capabilities/" + capID)
	}
	return links
}

func (h *HATEOASLinks) AssignmentLinks(domID, capID string) Links {
	return Links{
		"x-capability":      h.get("/capabilities/" + capID),
		"x-business-domain": h.get("/business-domains/" + domID),
		"x-remove":          h.del("/business-domains/" + domID + "/capabilities/" + capID),
	}
}

func (h *HATEOASLinks) AssignmentLinksForActor(domID, capID string, actor sharedctx.Actor) Links {
	links := Links{
		"x-capability":      h.get("/capabilities/" + capID),
		"x-business-domain": h.get("/business-domains/" + domID),
	}
	if actor.CanDelete("domains") {
		links["x-remove"] = h.del("/business-domains/" + domID + "/capabilities/" + capID)
	}
	return links
}

func (h *HATEOASLinks) ReferenceDocLink(resourceType string) string {
	return h.base + "/reference/" + resourceType
}

func (h *HATEOASLinks) MaturityScaleLinks(isDefault bool) Links {
	return h.maturityLinks("/meta-model/maturity-scale", isDefault)
}

func (h *HATEOASLinks) MetaModelConfigLinks(id string, isDefault bool) Links {
	return h.maturityLinks("/meta-model/configurations/"+id, isDefault)
}

func (h *HATEOASLinks) maturityLinks(selfPath string, isDefault bool) Links {
	links := Links{
		"self":              h.get(selfPath),
		"edit":              h.put("/meta-model/maturity-scale"),
		"x-maturity-levels": h.get("/capabilities/metadata/maturity-levels"),
	}
	if !isDefault {
		links["x-reset"] = h.post("/meta-model/maturity-scale/reset")
	}
	return links
}

func (h *HATEOASLinks) StrategyPillarLinks(id string, isActive bool) Links {
	p := "/meta-model/strategy-pillars/" + id
	links := Links{"self": h.get(p), "collection": h.get("/meta-model/strategy-pillars")}
	if isActive {
		links["edit"] = h.put(p)
		links["delete"] = h.del(p)
	}
	return links
}

func (h *HATEOASLinks) StrategyPillarsCollectionLinks() Links {
	return Links{"self": h.get("/meta-model/strategy-pillars"), "create": h.post("/meta-model/strategy-pillars")}
}

func (h *HATEOASLinks) EnterpriseCapabilityLinks(id string) Links {
	p := "/enterprise-capabilities/" + id
	return Links{
		"self": h.get(p), "edit": h.put(p), "delete": h.del(p),
		"x-links":                h.get(p + "/links"),
		"x-create-link":          h.post(p + "/links"),
		"x-strategic-importance": h.get(p + "/strategic-importance"),
	}
}

func (h *HATEOASLinks) EnterpriseCapabilityLinksForActor(id string, actor sharedctx.Actor) Links {
	p := "/enterprise-capabilities/" + id
	links := Links{
		"self":                   h.get(p),
		"x-links":                h.get(p + "/links"),
		"x-strategic-importance": h.get(p + "/strategic-importance"),
	}
	if actor.CanWrite("enterprise-arch") {
		links["edit"] = h.put(p)
		links["x-create-link"] = h.post(p + "/links")
	}
	if actor.CanDelete("enterprise-arch") {
		links["delete"] = h.del(p)
	}
	return links
}

func (h *HATEOASLinks) EnterpriseCapabilityCollectionLinks() Links {
	return Links{"self": h.get("/enterprise-capabilities")}
}

func (h *HATEOASLinks) EnterpriseCapabilityLinkLinks(ecID, linkID string) Links {
	p := "/enterprise-capabilities/" + ecID + "/links/" + linkID
	return Links{
		"self": h.get(p), "delete": h.del(p),
		"x-enterprise-capability": h.get("/enterprise-capabilities/" + ecID),
	}
}

func (h *HATEOASLinks) EnterpriseCapabilityLinksCollectionLinks(ecID string) Links {
	return Links{
		"self":                    h.get("/enterprise-capabilities/" + ecID + "/links"),
		"x-enterprise-capability": h.get("/enterprise-capabilities/" + ecID),
	}
}

func (h *HATEOASLinks) EnterpriseStrategicImportanceLinks(ecID, impID string) Links {
	p := "/enterprise-capabilities/" + ecID + "/strategic-importance/" + impID
	return Links{
		"self": h.get(p), "edit": h.put(p), "delete": h.del(p),
		"x-enterprise-capability": h.get("/enterprise-capabilities/" + ecID),
	}
}

func (h *HATEOASLinks) EnterpriseStrategicImportanceLinksForActor(ecID, impID string, actor sharedctx.Actor) Links {
	p := "/enterprise-capabilities/" + ecID + "/strategic-importance/" + impID
	links := Links{
		"self":                    h.get(p),
		"x-enterprise-capability": h.get("/enterprise-capabilities/" + ecID),
	}
	if actor.CanWrite("enterprise-arch") {
		links["edit"] = h.put(p)
	}
	if actor.CanDelete("enterprise-arch") {
		links["delete"] = h.del(p)
	}
	return links
}

func (h *HATEOASLinks) EnterpriseStrategicImportanceCollectionLinks(ecID string) Links {
	return Links{
		"self":                    h.get("/enterprise-capabilities/" + ecID + "/strategic-importance"),
		"x-enterprise-capability": h.get("/enterprise-capabilities/" + ecID),
	}
}

func (h *HATEOASLinks) DomainCapabilityEnterpriseLinks(dcID string) Links {
	return Links{"self": h.get("/domain-capabilities/" + dcID + "/enterprise-capability")}
}

func (h *HATEOASLinks) DomainCapabilityEnterpriseLinkedLinks(dcID, ecID, linkID string) Links {
	return Links{
		"self":                    h.get("/domain-capabilities/" + dcID + "/enterprise-capability"),
		"x-enterprise-capability": h.get("/enterprise-capabilities/" + ecID),
		"x-unlink":                h.del("/enterprise-capabilities/" + ecID + "/links/" + linkID),
	}
}

func (h *HATEOASLinks) AuditHistory(id string) string {
	return h.base + "/audit/" + id
}

func (h *HATEOASLinks) StrategyImportanceLinks(domID, capID, impID string) Links {
	p := "/business-domains/" + domID + "/capabilities/" + capID + "/importance/" + impID
	return Links{
		"self": h.get(p), "edit": h.put(p), "delete": h.del(p),
		"x-capability": h.get("/capabilities/" + capID),
		"x-domain":     h.get("/business-domains/" + domID),
	}
}

func (h *HATEOASLinks) StrategyImportanceLinksForActor(domID, capID, impID string, actor sharedctx.Actor) Links {
	p := "/business-domains/" + domID + "/capabilities/" + capID + "/importance/" + impID
	links := Links{
		"self":         h.get(p),
		"x-capability": h.get("/capabilities/" + capID),
		"x-domain":     h.get("/business-domains/" + domID),
	}
	if actor.CanWrite("domains") {
		links["edit"] = h.put(p)
	}
	if actor.CanDelete("domains") {
		links["delete"] = h.del(p)
	}
	return links
}

func (h *HATEOASLinks) CapabilityLinkStatusLinks(capID, status string, linkedToID, blockingCapID, blockingEcID *string) Links {
	links := Links{"self": h.get("/domain-capabilities/" + capID + "/enterprise-link-status")}
	if status == "available" {
		links["x-available-enterprise-capabilities"] = h.get("/enterprise-capabilities")
	}
	if linkedToID != nil {
		links["x-linked-to"] = h.get("/enterprise-capabilities/" + *linkedToID)
		links["x-enterprise-capability"] = h.get("/domain-capabilities/" + capID + "/enterprise-capability")
	}
	if blockingCapID != nil {
		links["x-blocking-capability"] = h.get("/capabilities/" + *blockingCapID)
	}
	if blockingEcID != nil {
		links["x-blocking-enterprise-capability"] = h.get("/enterprise-capabilities/" + *blockingEcID)
	}
	return links
}

func (h *HATEOASLinks) MaturityAnalysisCandidateLinks(ecID string) Links {
	return Links{
		"self":           h.get("/enterprise-capabilities/" + ecID),
		"x-maturity-gap": h.get("/enterprise-capabilities/" + ecID + "/maturity-gap"),
	}
}

func (h *HATEOASLinks) MaturityAnalysisCollectionLinks() Links {
	return Links{"self": h.get("/enterprise-capabilities/maturity-analysis")}
}

func (h *HATEOASLinks) MaturityGapDetailLinks(ecID string) Links {
	p := "/enterprise-capabilities/" + ecID
	return Links{
		"self":                  h.get(p + "/maturity-gap"),
		"up":                    h.get(p),
		"x-set-target-maturity": h.put(p + "/target-maturity"),
	}
}

func (h *HATEOASLinks) UnlinkedCapabilitiesLinks() Links {
	return Links{"self": h.get("/domain-capabilities/unlinked")}
}

func NewLink(href, method string) types.Link {
	return types.Link{Href: href, Method: method}
}

func (h *HATEOASLinks) FitScoreLinksForActor(componentID, pillarID string, actor sharedctx.Actor) Links {
	p := "/components/" + componentID + "/fit-scores/" + pillarID
	links := Links{
		"self": h.get(p),
		"up":   h.get("/components/" + componentID),
	}
	if actor.CanWrite("components") {
		links["edit"] = h.put(p)
	}
	if actor.CanDelete("components") {
		links["delete"] = h.del(p)
	}
	return links
}

func (h *HATEOASLinks) FitScoresCollectionLinksForActor(componentID string, actor sharedctx.Actor) Links {
	p := "/components/" + componentID + "/fit-scores"
	links := Links{
		"self": h.get(p),
		"up":   h.get("/components/" + componentID),
	}
	if actor.CanWrite("components") {
		links["create"] = h.put(p + "/{pillarId}")
	}
	return links
}
