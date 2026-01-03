package api

import "fmt"

// HATEOASLinks generates common HATEOAS links
type HATEOASLinks struct {
	baseURL string
}

// NewHATEOASLinks creates a new HATEOAS link generator
func NewHATEOASLinks(baseURL string) *HATEOASLinks {
	if baseURL == "" {
		baseURL = "/api/v1"
	}
	return &HATEOASLinks{baseURL: baseURL}
}

// ComponentLinks generates links for a component resource
func (h *HATEOASLinks) ComponentLinks(componentID string) map[string]string {
	return map[string]string{
		"self":       fmt.Sprintf("%s/components/%s", h.baseURL, componentID),
		"delete":     fmt.Sprintf("%s/components/%s", h.baseURL, componentID),
		"reference":  h.ReferenceDocLink("components"),
		"collection": fmt.Sprintf("%s/components", h.baseURL),
	}
}

// RelationLinks generates links for a relation resource
func (h *HATEOASLinks) RelationLinks(relationID string) map[string]string {
	return map[string]string{
		"self":       fmt.Sprintf("%s/relations/%s", h.baseURL, relationID),
		"delete":     fmt.Sprintf("%s/relations/%s", h.baseURL, relationID),
		"reference":  h.ReferenceDocLink("relations/generic"),
		"collection": fmt.Sprintf("%s/relations", h.baseURL),
	}
}

// RelationTypeLinks generates reference documentation links for relation types
func (h *HATEOASLinks) RelationTypeLinks(relationType string) map[string]string {
	links := make(map[string]string)

	switch relationType {
	case "Triggers":
		links["reference"] = h.ReferenceDocLink("relations/triggering")
	case "Serves":
		links["reference"] = h.ReferenceDocLink("relations/serving")
	default:
		links["reference"] = h.ReferenceDocLink("relations/generic")
	}

	return links
}

// ViewLinks generates links for a view resource
func (h *HATEOASLinks) ViewLinks(viewID string) map[string]string {
	return map[string]string{
		"self":       fmt.Sprintf("%s/views/%s", h.baseURL, viewID),
		"components": fmt.Sprintf("%s/views/%s/components", h.baseURL, viewID),
		"collection": fmt.Sprintf("%s/views", h.baseURL),
	}
}

func (h *HATEOASLinks) CapabilityLinks(capabilityID, parentID string) map[string]string {
	links := h.buildCapabilityBaseLinks(capabilityID)
	if parentID != "" {
		links["parent"] = fmt.Sprintf("%s/capabilities/%s", h.baseURL, parentID)
	}
	return links
}

func (h *HATEOASLinks) buildCapabilityBaseLinks(capabilityID string) map[string]string {
	return map[string]string{
		"self":                 fmt.Sprintf("%s/capabilities/%s", h.baseURL, capabilityID),
		"children":             fmt.Sprintf("%s/capabilities/%s/children", h.baseURL, capabilityID),
		"systems":              fmt.Sprintf("%s/capabilities/%s/systems", h.baseURL, capabilityID),
		"outgoingDependencies": fmt.Sprintf("%s/capabilities/%s/dependencies/outgoing", h.baseURL, capabilityID),
		"incomingDependencies": fmt.Sprintf("%s/capabilities/%s/dependencies/incoming", h.baseURL, capabilityID),
		"collection":           fmt.Sprintf("%s/capabilities", h.baseURL),
	}
}

// DependencyLinks generates links for a capability dependency resource
func (h *HATEOASLinks) DependencyLinks(dependencyID, sourceCapabilityID, targetCapabilityID string) map[string]string {
	return map[string]string{
		"self":             fmt.Sprintf("%s/capability-dependencies/%s", h.baseURL, dependencyID),
		"sourceCapability": fmt.Sprintf("%s/capabilities/%s", h.baseURL, sourceCapabilityID),
		"targetCapability": fmt.Sprintf("%s/capabilities/%s", h.baseURL, targetCapabilityID),
		"delete":           fmt.Sprintf("%s/capability-dependencies/%s", h.baseURL, dependencyID),
		"collection":       fmt.Sprintf("%s/capability-dependencies", h.baseURL),
	}
}

// RealizationLinks generates links for a capability realization resource
func (h *HATEOASLinks) RealizationLinks(realizationID, capabilityID, componentID string) map[string]string {
	return map[string]string{
		"self":       fmt.Sprintf("%s/capability-realizations/%s", h.baseURL, realizationID),
		"capability": fmt.Sprintf("%s/capabilities/%s", h.baseURL, capabilityID),
		"component":  fmt.Sprintf("%s/components/%s", h.baseURL, componentID),
		"update":     fmt.Sprintf("%s/capability-realizations/%s", h.baseURL, realizationID),
		"delete":     fmt.Sprintf("%s/capability-realizations/%s", h.baseURL, realizationID),
	}
}

func (h *HATEOASLinks) BusinessDomainLinks(domainID string, hasCapabilities bool) map[string]string {
	links := map[string]string{
		"self":         fmt.Sprintf("%s/business-domains/%s", h.baseURL, domainID),
		"capabilities": fmt.Sprintf("%s/business-domains/%s/capabilities", h.baseURL, domainID),
		"update":       fmt.Sprintf("%s/business-domains/%s", h.baseURL, domainID),
		"collection":   fmt.Sprintf("%s/business-domains", h.baseURL),
	}
	if !hasCapabilities {
		links["delete"] = fmt.Sprintf("%s/business-domains/%s", h.baseURL, domainID)
	}
	return links
}

func (h *HATEOASLinks) CapabilityInDomainLinks(capabilityID, domainID string) map[string]string {
	return map[string]string{
		"self":             fmt.Sprintf("%s/capabilities/%s", h.baseURL, capabilityID),
		"children":         fmt.Sprintf("%s/capabilities/%s/children", h.baseURL, capabilityID),
		"businessDomains":  fmt.Sprintf("%s/capabilities/%s/business-domains", h.baseURL, capabilityID),
		"removeFromDomain": fmt.Sprintf("%s/business-domains/%s/capabilities/%s", h.baseURL, domainID, capabilityID),
	}
}

func (h *HATEOASLinks) DomainForCapabilityLinks(domainID, capabilityID string) map[string]string {
	return map[string]string{
		"self":             fmt.Sprintf("%s/business-domains/%s", h.baseURL, domainID),
		"capabilities":     fmt.Sprintf("%s/business-domains/%s/capabilities", h.baseURL, domainID),
		"removeCapability": fmt.Sprintf("%s/business-domains/%s/capabilities/%s", h.baseURL, domainID, capabilityID),
	}
}

func (h *HATEOASLinks) AssignmentLinks(domainID, capabilityID string) map[string]string {
	return map[string]string{
		"capability":     fmt.Sprintf("%s/capabilities/%s", h.baseURL, capabilityID),
		"businessDomain": fmt.Sprintf("%s/business-domains/%s", h.baseURL, domainID),
		"remove":         fmt.Sprintf("%s/business-domains/%s/capabilities/%s", h.baseURL, domainID, capabilityID),
	}
}

// ReferenceDocLink generates a link to reference documentation
func (h *HATEOASLinks) ReferenceDocLink(resourceType string) string {
	return fmt.Sprintf("%s/reference/%s", h.baseURL, resourceType)
}

func (h *HATEOASLinks) buildMaturityLinks(selfURL string, isDefault bool) map[string]string {
	links := map[string]string{
		"self":           selfURL,
		"update":         fmt.Sprintf("%s/meta-model/maturity-scale", h.baseURL),
		"maturityLevels": fmt.Sprintf("%s/capabilities/metadata/maturity-levels", h.baseURL),
	}
	if !isDefault {
		links["reset"] = fmt.Sprintf("%s/meta-model/maturity-scale/reset", h.baseURL)
	}
	return links
}

func (h *HATEOASLinks) MaturityScaleLinks(isDefault bool) map[string]string {
	return h.buildMaturityLinks(fmt.Sprintf("%s/meta-model/maturity-scale", h.baseURL), isDefault)
}

func (h *HATEOASLinks) MetaModelConfigLinks(configID string, isDefault bool) map[string]string {
	return h.buildMaturityLinks(fmt.Sprintf("%s/meta-model/configurations/%s", h.baseURL, configID), isDefault)
}

func (h *HATEOASLinks) StrategyPillarLinks(pillarID string, isActive bool) map[string]string {
	links := map[string]string{
		"self":       fmt.Sprintf("%s/meta-model/strategy-pillars/%s", h.baseURL, pillarID),
		"collection": fmt.Sprintf("%s/meta-model/strategy-pillars", h.baseURL),
	}
	if isActive {
		links["update"] = fmt.Sprintf("%s/meta-model/strategy-pillars/%s", h.baseURL, pillarID)
		links["delete"] = fmt.Sprintf("%s/meta-model/strategy-pillars/%s", h.baseURL, pillarID)
	}
	return links
}

func (h *HATEOASLinks) StrategyPillarsCollectionLinks() map[string]string {
	return map[string]string{
		"self":   fmt.Sprintf("%s/meta-model/strategy-pillars", h.baseURL),
		"create": fmt.Sprintf("%s/meta-model/strategy-pillars", h.baseURL),
	}
}

func (h *HATEOASLinks) EnterpriseCapabilityLinks(id string) map[string]string {
	return map[string]string{
		"self":                fmt.Sprintf("%s/enterprise-capabilities/%s", h.baseURL, id),
		"links":               fmt.Sprintf("%s/enterprise-capabilities/%s/links", h.baseURL, id),
		"strategicImportance": fmt.Sprintf("%s/enterprise-capabilities/%s/strategic-importance", h.baseURL, id),
	}
}

func (h *HATEOASLinks) EnterpriseCapabilityCollectionLinks() map[string]string {
	return map[string]string{
		"self": fmt.Sprintf("%s/enterprise-capabilities", h.baseURL),
	}
}

func (h *HATEOASLinks) EnterpriseCapabilityLinkLinks(enterpriseCapabilityID, linkID string) map[string]string {
	return map[string]string{
		"self":                 fmt.Sprintf("%s/enterprise-capabilities/%s/links/%s", h.baseURL, enterpriseCapabilityID, linkID),
		"enterpriseCapability": fmt.Sprintf("%s/enterprise-capabilities/%s", h.baseURL, enterpriseCapabilityID),
	}
}

func (h *HATEOASLinks) EnterpriseCapabilityLinksCollectionLinks(enterpriseCapabilityID string) map[string]string {
	return map[string]string{
		"self":                 fmt.Sprintf("%s/enterprise-capabilities/%s/links", h.baseURL, enterpriseCapabilityID),
		"enterpriseCapability": fmt.Sprintf("%s/enterprise-capabilities/%s", h.baseURL, enterpriseCapabilityID),
	}
}

func (h *HATEOASLinks) EnterpriseStrategicImportanceLinks(enterpriseCapabilityID, importanceID string) map[string]string {
	return map[string]string{
		"self":                 fmt.Sprintf("%s/enterprise-capabilities/%s/strategic-importance/%s", h.baseURL, enterpriseCapabilityID, importanceID),
		"enterpriseCapability": fmt.Sprintf("%s/enterprise-capabilities/%s", h.baseURL, enterpriseCapabilityID),
	}
}

func (h *HATEOASLinks) EnterpriseStrategicImportanceCollectionLinks(enterpriseCapabilityID string) map[string]string {
	return map[string]string{
		"self":                 fmt.Sprintf("%s/enterprise-capabilities/%s/strategic-importance", h.baseURL, enterpriseCapabilityID),
		"enterpriseCapability": fmt.Sprintf("%s/enterprise-capabilities/%s", h.baseURL, enterpriseCapabilityID),
	}
}

func (h *HATEOASLinks) DomainCapabilityEnterpriseLinks(domainCapabilityID string) map[string]string {
	return map[string]string{
		"self": fmt.Sprintf("%s/domain-capabilities/%s/enterprise-capability", h.baseURL, domainCapabilityID),
	}
}

func (h *HATEOASLinks) DomainCapabilityEnterpriseLinkedLinks(domainCapabilityID, enterpriseCapabilityID, linkID string) map[string]string {
	return map[string]string{
		"self":                 fmt.Sprintf("%s/domain-capabilities/%s/enterprise-capability", h.baseURL, domainCapabilityID),
		"enterpriseCapability": fmt.Sprintf("%s/enterprise-capabilities/%s", h.baseURL, enterpriseCapabilityID),
		"unlink":               fmt.Sprintf("%s/enterprise-capabilities/%s/links/%s", h.baseURL, enterpriseCapabilityID, linkID),
	}
}

func (h *HATEOASLinks) AuditHistory(aggregateID string) string {
	return fmt.Sprintf("%s/audit/%s", h.baseURL, aggregateID)
}

func (h *HATEOASLinks) CapabilityLinkStatusLinks(capabilityID string, status string, linkedToID *string, blockingCapabilityID *string, blockingEnterpriseCapID *string) map[string]string {
	links := map[string]string{
		"self": fmt.Sprintf("%s/domain-capabilities/%s/enterprise-link-status", h.baseURL, capabilityID),
	}

	if status == "available" {
		links["availableEnterpriseCapabilities"] = fmt.Sprintf("%s/enterprise-capabilities", h.baseURL)
	}

	if linkedToID != nil {
		links["linkedTo"] = fmt.Sprintf("%s/enterprise-capabilities/%s", h.baseURL, *linkedToID)
		links["enterpriseCapability"] = fmt.Sprintf("%s/domain-capabilities/%s/enterprise-capability", h.baseURL, capabilityID)
	}

	if blockingCapabilityID != nil {
		links["blockingCapability"] = fmt.Sprintf("%s/capabilities/%s", h.baseURL, *blockingCapabilityID)
	}

	if blockingEnterpriseCapID != nil {
		links["blockingEnterpriseCapability"] = fmt.Sprintf("%s/enterprise-capabilities/%s", h.baseURL, *blockingEnterpriseCapID)
	}

	return links
}
