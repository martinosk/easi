package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type SystemLinkedToCapability struct {
	domain.BaseEvent
	ID               string    `json:"id"`
	CapabilityID     string    `json:"capabilityId"`
	ComponentID      string    `json:"componentId"`
	ComponentName    string    `json:"componentName"`
	RealizationLevel string    `json:"realizationLevel"`
	Notes            string    `json:"notes"`
	LinkedAt         time.Time `json:"linkedAt"`
}

func NewSystemLinkedToCapability(id, capabilityID, componentID, componentName, realizationLevel, notes string) SystemLinkedToCapability {
	return SystemLinkedToCapability{
		BaseEvent:        domain.NewBaseEvent(id),
		ID:               id,
		CapabilityID:     capabilityID,
		ComponentID:      componentID,
		ComponentName:    componentName,
		RealizationLevel: realizationLevel,
		Notes:            notes,
		LinkedAt:         time.Now().UTC(),
	}
}

func (e SystemLinkedToCapability) EventType() string {
	return "SystemLinkedToCapability"
}

func (e SystemLinkedToCapability) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":               e.ID,
		"capabilityId":     e.CapabilityID,
		"componentId":      e.ComponentID,
		"componentName":    e.ComponentName,
		"realizationLevel": e.RealizationLevel,
		"notes":            e.Notes,
		"linkedAt":         e.LinkedAt,
	}
}

func (e SystemLinkedToCapability) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
