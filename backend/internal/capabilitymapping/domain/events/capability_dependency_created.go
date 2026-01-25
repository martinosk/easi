package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type CapabilityDependencyCreated struct {
	domain.BaseEvent
	ID                 string    `json:"id"`
	SourceCapabilityID string    `json:"sourceCapabilityId"`
	TargetCapabilityID string    `json:"targetCapabilityId"`
	DependencyType     string    `json:"dependencyType"`
	Description        string    `json:"description"`
	CreatedAt          time.Time `json:"createdAt"`
}

func NewCapabilityDependencyCreated(id, sourceCapabilityID, targetCapabilityID, dependencyType, description string) CapabilityDependencyCreated {
	return CapabilityDependencyCreated{
		BaseEvent:          domain.NewBaseEvent(id),
		ID:                 id,
		SourceCapabilityID: sourceCapabilityID,
		TargetCapabilityID: targetCapabilityID,
		DependencyType:     dependencyType,
		Description:        description,
		CreatedAt:          time.Now().UTC(),
	}
}

func (e CapabilityDependencyCreated) EventType() string {
	return "CapabilityDependencyCreated"
}

func (e CapabilityDependencyCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":                 e.ID,
		"sourceCapabilityId": e.SourceCapabilityID,
		"targetCapabilityId": e.TargetCapabilityID,
		"dependencyType":     e.DependencyType,
		"description":        e.Description,
		"createdAt":          e.CreatedAt,
	}
}

func (e CapabilityDependencyCreated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
