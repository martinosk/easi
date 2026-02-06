package api

import (
	"net/url"

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

func (h *HATEOASLinks) get(path string) types.Link   { return h.l(h.base+path, "GET") }
func (h *HATEOASLinks) put(path string) types.Link   { return h.l(h.base+path, "PUT") }
func (h *HATEOASLinks) post(path string) types.Link  { return h.l(h.base+path, "POST") }
func (h *HATEOASLinks) del(path string) types.Link   { return h.l(h.base+path, "DELETE") }
func (h *HATEOASLinks) patch(path string) types.Link { return h.l(h.base+path, "PATCH") }

func (h *HATEOASLinks) crud(path string) Links {
	return Links{"self": h.get(path), "edit": h.put(path), "delete": h.del(path)}
}

type ResourceConfig struct {
	Path       string
	Collection string
	Permission string
}

func (h *HATEOASLinks) simpleResourceLinks(cfg ResourceConfig, id string, actor sharedctx.Actor) Links {
	p := cfg.Path + "/" + id
	links := Links{"self": h.get(p), "collection": h.get(cfg.Collection)}
	if actor.CanWrite(cfg.Permission) {
		links["edit"] = h.put(p)
	}
	if actor.CanDelete(cfg.Permission) {
		links["delete"] = h.del(p)
	}
	return links
}

type ExpertParams struct {
	ResourcePath string
	ExpertName   string
	ExpertRole   string
	ContactInfo  string
}

func (h *HATEOASLinks) expertRemoveLink(p ExpertParams, actor sharedctx.Actor, permission string) Links {
	links := Links{}
	if actor.CanDelete(permission) {
		links["x-remove"] = h.del(p.ResourcePath + "/experts?name=" + url.QueryEscape(p.ExpertName) + "&role=" + url.QueryEscape(p.ExpertRole) + "&contact=" + url.QueryEscape(p.ContactInfo))
	}
	return links
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
		"self":           h.get(p),
		"describedby":    h.get("/reference/components"),
		"collection":     h.get("/components"),
		"x-expert-roles": h.get("/components/expert-roles"),
	}
	if actor.CanWrite("components") {
		links["edit"] = h.put(p)
		links["x-add-expert"] = h.post(p + "/experts")
	}
	if actor.CanDelete("components") {
		links["delete"] = h.del(p)
	}
	return links
}

func (h *HATEOASLinks) ComponentExpertLinksForActor(p ExpertParams, actor sharedctx.Actor) Links {
	return h.expertRemoveLink(p, actor, "components")
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

type ViewInfo struct {
	ID          string
	IsPrivate   bool
	IsDefault   bool
	OwnerUserID *string
}

func (v ViewInfo) isOwnedBy(actorID string) bool {
	return v.OwnerUserID != nil && *v.OwnerUserID == actorID
}

func (v ViewInfo) canBeEditedBy(actor sharedctx.Actor) bool {
	return (!v.IsPrivate || v.isOwnedBy(actor.ID)) && actor.CanWrite("views")
}

func (v ViewInfo) canBeDeletedBy(actor sharedctx.Actor) bool {
	return (!v.IsPrivate || v.isOwnedBy(actor.ID)) && actor.CanDelete("views") && !v.IsDefault
}

func (h *HATEOASLinks) ViewLinksForActor(v ViewInfo, actor sharedctx.Actor) Links {
	p := "/views/" + v.ID
	links := Links{
		"self":         h.get(p),
		"x-components": h.get(p + "/components"),
		"collection":   h.get("/views"),
	}
	if v.canBeEditedBy(actor) {
		links["edit"] = h.patch(p + "/name")
		links["x-change-visibility"] = h.patch(p + "/visibility")
	}
	if v.canBeDeletedBy(actor) {
		links["delete"] = h.del(p)
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
		"x-expert-roles":          h.get("/capabilities/expert-roles"),
	}
	if actor.CanWrite("capabilities") {
		links["edit"] = h.put(p)
		links["x-add-expert"] = h.post(p + "/experts")
	}
	if actor.CanDelete("capabilities") {
		links["delete"] = h.del(p)
	}
	return links
}

func (h *HATEOASLinks) CapabilityExpertLinksForActor(p ExpertParams, actor sharedctx.Actor) Links {
	return h.expertRemoveLink(p, actor, "capabilities")
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

func (h *HATEOASLinks) capabilityInDomainBase(capID, domID string) Links {
	return Links{
		"self":               h.get("/capabilities/" + capID),
		"x-children":         h.get("/capabilities/" + capID + "/children"),
		"x-business-domains": h.get("/capabilities/" + capID + "/business-domains"),
	}
}

func (h *HATEOASLinks) CapabilityInDomainLinks(capID, domID string) Links {
	links := h.capabilityInDomainBase(capID, domID)
	links["x-remove-from-domain"] = h.del("/business-domains/" + domID + "/capabilities/" + capID)
	return links
}

func (h *HATEOASLinks) CapabilityInDomainLinksForActor(capID, domID string, actor sharedctx.Actor) Links {
	links := h.capabilityInDomainBase(capID, domID)
	if actor.CanDelete("domains") {
		links["x-remove-from-domain"] = h.del("/business-domains/" + domID + "/capabilities/" + capID)
	}
	return links
}

func (h *HATEOASLinks) domainForCapabilityBase(domID string) Links {
	p := "/business-domains/" + domID
	return Links{
		"self":           h.get(p),
		"x-capabilities": h.get(p + "/capabilities"),
	}
}

func (h *HATEOASLinks) DomainForCapabilityLinks(domID, capID string) Links {
	links := h.domainForCapabilityBase(domID)
	links["x-remove-capability"] = h.del("/business-domains/" + domID + "/capabilities/" + capID)
	return links
}

func (h *HATEOASLinks) DomainForCapabilityLinksForActor(domID, capID string, actor sharedctx.Actor) Links {
	links := h.domainForCapabilityBase(domID)
	if actor.CanDelete("domains") {
		links["x-remove-capability"] = h.del("/business-domains/" + domID + "/capabilities/" + capID)
	}
	return links
}

func (h *HATEOASLinks) assignmentBase(domID, capID string) Links {
	return Links{
		"x-capability":      h.get("/capabilities/" + capID),
		"x-business-domain": h.get("/business-domains/" + domID),
	}
}

func (h *HATEOASLinks) AssignmentLinks(domID, capID string) Links {
	links := h.assignmentBase(domID, capID)
	links["x-remove"] = h.del("/business-domains/" + domID + "/capabilities/" + capID)
	return links
}

func (h *HATEOASLinks) AssignmentLinksForActor(domID, capID string, actor sharedctx.Actor) Links {
	links := h.assignmentBase(domID, capID)
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

func (h *HATEOASLinks) enterpriseStrategicImportanceBase(ecID, impID string) Links {
	p := "/enterprise-capabilities/" + ecID + "/strategic-importance/" + impID
	return Links{
		"self":                    h.get(p),
		"x-enterprise-capability": h.get("/enterprise-capabilities/" + ecID),
	}
}

func (h *HATEOASLinks) EnterpriseStrategicImportanceLinks(ecID, impID string) Links {
	links := h.enterpriseStrategicImportanceBase(ecID, impID)
	p := "/enterprise-capabilities/" + ecID + "/strategic-importance/" + impID
	links["edit"] = h.put(p)
	links["delete"] = h.del(p)
	return links
}

func (h *HATEOASLinks) EnterpriseStrategicImportanceLinksForActor(ecID, impID string, actor sharedctx.Actor) Links {
	links := h.enterpriseStrategicImportanceBase(ecID, impID)
	p := "/enterprise-capabilities/" + ecID + "/strategic-importance/" + impID
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

func (h *HATEOASLinks) strategyImportanceBase(domID, capID, impID string) Links {
	p := "/business-domains/" + domID + "/capabilities/" + capID + "/importance/" + impID
	return Links{
		"self":         h.get(p),
		"x-capability": h.get("/capabilities/" + capID),
		"x-domain":     h.get("/business-domains/" + domID),
	}
}

func (h *HATEOASLinks) StrategyImportanceLinks(domID, capID, impID string) Links {
	links := h.strategyImportanceBase(domID, capID, impID)
	p := "/business-domains/" + domID + "/capabilities/" + capID + "/importance/" + impID
	links["edit"] = h.put(p)
	links["delete"] = h.del(p)
	return links
}

func (h *HATEOASLinks) StrategyImportanceLinksForActor(domID, capID, impID string, actor sharedctx.Actor) Links {
	links := h.strategyImportanceBase(domID, capID, impID)
	p := "/business-domains/" + domID + "/capabilities/" + capID + "/importance/" + impID
	if actor.CanWrite("domains") {
		links["edit"] = h.put(p)
	}
	if actor.CanDelete("domains") {
		links["delete"] = h.del(p)
	}
	return links
}

type LinkStatusParams struct {
	CapabilityID  string
	Status        string
	LinkedToID    *string
	BlockingCapID *string
	BlockingEcID  *string
}

func (h *HATEOASLinks) CapabilityLinkStatusLinks(p LinkStatusParams) Links {
	links := Links{"self": h.get("/domain-capabilities/" + p.CapabilityID + "/enterprise-link-status")}
	if p.Status == "available" {
		links["x-available-enterprise-capabilities"] = h.get("/enterprise-capabilities")
	}
	if p.LinkedToID != nil {
		links["x-linked-to"] = h.get("/enterprise-capabilities/" + *p.LinkedToID)
		links["x-enterprise-capability"] = h.get("/domain-capabilities/" + p.CapabilityID + "/enterprise-capability")
	}
	if p.BlockingCapID != nil {
		links["x-blocking-capability"] = h.get("/capabilities/" + *p.BlockingCapID)
	}
	if p.BlockingEcID != nil {
		links["x-blocking-enterprise-capability"] = h.get("/enterprise-capabilities/" + *p.BlockingEcID)
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

func (h *HATEOASLinks) TimeSuggestionsCollectionLinks() Links {
	return Links{"self": h.get("/time-suggestions")}
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

func (h *HATEOASLinks) StrategyImportanceCollectionLinksForActor(domID, capID string, actor sharedctx.Actor) Links {
	p := "/business-domains/" + domID + "/capabilities/" + capID + "/importance"
	links := Links{"self": h.get(p)}
	if actor.CanWrite("domains") {
		links["create"] = h.post(p)
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

var (
	acquiredEntityConfig = ResourceConfig{Path: "/acquired-entities", Collection: "/acquired-entities", Permission: "components"}
	vendorConfig         = ResourceConfig{Path: "/vendors", Collection: "/vendors", Permission: "components"}
	internalTeamConfig   = ResourceConfig{Path: "/internal-teams", Collection: "/internal-teams", Permission: "components"}
)

func (h *HATEOASLinks) AcquiredEntityLinksForActor(id string, actor sharedctx.Actor) Links {
	return h.simpleResourceLinks(acquiredEntityConfig, id, actor)
}

func (h *HATEOASLinks) VendorLinksForActor(id string, actor sharedctx.Actor) Links {
	return h.simpleResourceLinks(vendorConfig, id, actor)
}

func (h *HATEOASLinks) InternalTeamLinksForActor(id string, actor sharedctx.Actor) Links {
	return h.simpleResourceLinks(internalTeamConfig, id, actor)
}

func (h *HATEOASLinks) OriginRelationshipLinksForActor(basePath, id, componentID string, extraLinks map[string]types.Link, actor sharedctx.Actor) Links {
	links := Links{
		"self":      h.get(basePath + "/" + id),
		"component": h.get("/components/" + componentID),
	}
	for k, v := range extraLinks {
		links[k] = v
	}
	if actor.CanDelete("components") {
		links["delete"] = h.del(basePath + "/" + id)
	}
	return links
}
