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
		"archimate":  h.ArchimateDocLink("application-component"),
		"collection": fmt.Sprintf("%s/components", h.baseURL),
	}
}

// RelationLinks generates links for a relation resource
func (h *HATEOASLinks) RelationLinks(relationID string) map[string]string {
	return map[string]string{
		"self":       fmt.Sprintf("%s/relations/%s", h.baseURL, relationID),
		"delete":     fmt.Sprintf("%s/relations/%s", h.baseURL, relationID),
		"archimate":  h.ArchimateDocLink("relationship"),
		"collection": fmt.Sprintf("%s/relations", h.baseURL),
	}
}

// RelationTypeLinks generates archimate documentation links for relation types
func (h *HATEOASLinks) RelationTypeLinks(relationType string) map[string]string {
	links := make(map[string]string)

	switch relationType {
	case "Triggers":
		links["archimate"] = h.ArchimateDocLink("triggering-relationship")
	case "Serves":
		links["archimate"] = h.ArchimateDocLink("serving-relationship")
	default:
		links["archimate"] = h.ArchimateDocLink("relationship")
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

// CapabilityLinks generates links for a capability resource
func (h *HATEOASLinks) CapabilityLinks(capabilityID, parentID string) map[string]string {
	links := map[string]string{
		"self":                  fmt.Sprintf("%s/capabilities/%s", h.baseURL, capabilityID),
		"children":              fmt.Sprintf("%s/capabilities/%s/children", h.baseURL, capabilityID),
		"systems":               fmt.Sprintf("%s/capabilities/%s/systems", h.baseURL, capabilityID),
		"outgoingDependencies":  fmt.Sprintf("%s/capabilities/%s/dependencies/outgoing", h.baseURL, capabilityID),
		"incomingDependencies":  fmt.Sprintf("%s/capabilities/%s/dependencies/incoming", h.baseURL, capabilityID),
		"collection":            fmt.Sprintf("%s/capabilities", h.baseURL),
	}

	if parentID != "" {
		links["parent"] = fmt.Sprintf("%s/capabilities/%s", h.baseURL, parentID)
	}

	return links
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

// ArchimateDocLink generates a link to ArchiMate documentation
func (h *HATEOASLinks) ArchimateDocLink(elementType string) string {
	baseURL := "https://pubs.opengroup.org/architecture/archimate3-doc/chap"

	switch elementType {
	case "application-component":
		return baseURL + "09.html#_Toc489946090"
	case "triggering-relationship":
		return baseURL + "05.html#_Toc489945994"
	case "serving-relationship":
		return baseURL + "05.html#_Toc489945993"
	case "relationship":
		return baseURL + "05.html"
	default:
		return baseURL + "01.html"
	}
}
