package projectors

import (
	"encoding/json"
	"fmt"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/onepagers/application/readmodels"
)

var envelopeAttributeKeys = map[string]struct{}{
	"id": {}, "tenantId": {}, "name": {}, "version": {},
	"actor": {}, "actorId": {}, "actorEmail": {},
	"occurredAt": {}, "occurredOn": {},
	"createdAt": {}, "updatedAt": {}, "deletedAt": {},
	"capabilityId": {}, "componentId": {},
}

func publishedAttributes(eventType string, eventData []byte) (readmodels.SubjectAttributes, error) {
	attributes := readmodels.SubjectAttributes{}
	if err := json.Unmarshal(eventData, &attributes); err != nil {
		return nil, fmt.Errorf("unmarshal published attributes of %s: %w", eventType, err)
	}
	for key := range envelopeAttributeKeys {
		delete(attributes, key)
	}
	return attributes, nil
}

type attributeOnlyEvent struct {
	subjectType string
	keys        map[string]string
}

var attributeOnlyEvents = map[string]attributeOnlyEvent{
	capPL.CapabilityParentChanged: {
		subjectType: subjectTypeCapability,
		keys:        map[string]string{"newParentId": "parentId", "newLevel": "level"},
	},
	capPL.CapabilityLevelChanged: {
		subjectType: subjectTypeCapability,
		keys:        map[string]string{"newLevel": "level"},
	},
}

func attributeOnlyEventTypes() []string {
	types := make([]string, 0, len(attributeOnlyEvents))
	for eventType := range attributeOnlyEvents {
		types = append(types, eventType)
	}
	return types
}

func attributeOnlyChange(eventType string, eventData []byte) (string, readmodels.SubjectAttributes, error) {
	event, found := attributeOnlyEvents[eventType]
	if !found {
		return "", nil, nil
	}
	var payload readmodels.SubjectAttributes
	if err := json.Unmarshal(eventData, &payload); err != nil {
		return "", nil, fmt.Errorf("unmarshal attributes of %s: %w", eventType, err)
	}
	attributes := readmodels.SubjectAttributes{}
	for published, attribute := range event.keys {
		if raw, carried := payload.Raw(published); carried {
			attributes[attribute] = raw
		}
	}
	return event.subjectType, attributes, nil
}

var expertEventAdds = map[string]bool{
	capPL.CapabilityExpertAdded:            true,
	capPL.CapabilityExpertRemoved:          false,
	amPL.ApplicationComponentExpertAdded:   true,
	amPL.ApplicationComponentExpertRemoved: false,
}

type cachedExpertChange struct {
	expert readmodels.SubjectExpert
	added  bool
}

func expertChange(eventType string, eventData []byte) (*cachedExpertChange, error) {
	added, found := expertEventAdds[eventType]
	if !found {
		return nil, nil
	}
	var expert readmodels.SubjectExpert
	if err := json.Unmarshal(eventData, &expert); err != nil {
		return nil, fmt.Errorf("unmarshal expert of %s: %w", eventType, err)
	}
	return &cachedExpertChange{expert: expert, added: added}, nil
}
