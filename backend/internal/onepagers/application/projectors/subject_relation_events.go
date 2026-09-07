package projectors

import "easi/backend/internal/onepagers/application/readmodels"

const (
	subjectTypeCapability  = "capability"
	subjectTypeApplication = "application"
)

const (
	entryRealizingApplications = "realizing-applications"
	entryRealizedCapabilities  = "realized-capabilities"
	entryParentCapability      = "parent-capability"
	entryChildCapabilities     = "child-capabilities"
	entryDependsOn             = "depends-on"
	entryBusinessDomains       = "business-domains"
	entryComponentRelations    = "component-relations"
)

type originLinkEntries struct {
	forward     string
	mirror      string
	relatedType string
}

func (e originLinkEntries) forwardEntries(entityID string) []readmodels.RelationEntry {
	if entityID == "" {
		return nil
	}
	return []readmodels.RelationEntry{{EntryID: e.forward, RelatedType: e.relatedType, RelatedID: entityID}}
}

var originEntriesByType = map[string]originLinkEntries{
	"built-by":       {forward: "built-by", mirror: "built-applications", relatedType: "internal-team"},
	"purchased-from": {forward: "purchased-from", mirror: "purchased-applications", relatedType: "vendor"},
	"acquired-via":   {forward: "acquired-via", mirror: "acquired-applications", relatedType: "acquired-entity"},
}

type inheritedRealizationPayload struct {
	CapabilityID        string `json:"capabilityId"`
	ComponentID         string `json:"componentId"`
	SourceRealizationID string `json:"sourceRealizationId"`
}

type inheritanceRemovalPayload struct {
	SourceRealizationID string   `json:"sourceRealizationId"`
	CapabilityIDs       []string `json:"capabilityIds"`
}

type relationEventPayload struct {
	ID                    string                        `json:"id"`
	Name                  string                        `json:"name"`
	ParentID              string                        `json:"parentId"`
	CapabilityID          string                        `json:"capabilityId"`
	ComponentID           string                        `json:"componentId"`
	SourceCapabilityID    string                        `json:"sourceCapabilityId"`
	TargetCapabilityID    string                        `json:"targetCapabilityId"`
	SourceComponentID     string                        `json:"sourceComponentId"`
	TargetComponentID     string                        `json:"targetComponentId"`
	BusinessDomainID      string                        `json:"businessDomainId"`
	NewParentID           string                        `json:"newParentId"`
	OriginType            string                        `json:"originType"`
	EntityID              string                        `json:"entityId"`
	NewEntityID           string                        `json:"newEntityId"`
	InheritedRealizations []inheritedRealizationPayload `json:"inheritedRealizations"`
	Removals              []inheritanceRemovalPayload   `json:"removals"`
}
