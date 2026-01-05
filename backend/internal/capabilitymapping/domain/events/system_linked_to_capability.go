package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type SystemLinkedToCapability struct {
	domain.BaseEvent
	ID               string
	CapabilityID     string
	ComponentID      string
	ComponentName    string
	RealizationLevel string
	Notes            string
	LinkedAt         time.Time
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
