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
		"archimate":  h.ArchimateDocLink("application-component"),
		"collection": fmt.Sprintf("%s/components", h.baseURL),
	}
}

// RelationLinks generates links for a relation resource
func (h *HATEOASLinks) RelationLinks(relationID string) map[string]string {
	return map[string]string{
		"self":       fmt.Sprintf("%s/relations/%s", h.baseURL, relationID),
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
