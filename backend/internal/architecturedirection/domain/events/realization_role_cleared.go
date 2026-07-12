package events

import (
	"time"

	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"
)

type RealizationRoleCleared struct {
	domain.BaseEvent
	ID           string    `json:"id"`
	CapabilityID string    `json:"capabilityId"`
	ComponentID  string    `json:"componentId"`
	ClearedBy    string    `json:"clearedBy"`
	OccurredOn   time.Time `json:"occurredOn"`
}

type RealizationRoleClearedFields struct {
	ID           string
	CapabilityID string
	ComponentID  string
	ClearedBy    string
}

func NewRealizationRoleCleared(f RealizationRoleClearedFields) RealizationRoleCleared {
	return RealizationRoleCleared{
		BaseEvent:    domain.NewBaseEvent(f.ID),
		ID:           f.ID,
		CapabilityID: f.CapabilityID,
		ComponentID:  f.ComponentID,
		ClearedBy:    f.ClearedBy,
		OccurredOn:   time.Now().UTC(),
	}
}

func (e RealizationRoleCleared) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e RealizationRoleCleared) EventType() string { return pl.RealizationRoleCleared }

func (e RealizationRoleCleared) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":           e.ID,
		"capabilityId": e.CapabilityID,
		"componentId":  e.ComponentID,
		"clearedBy":    e.ClearedBy,
		"occurredOn":   e.OccurredOn,
	}
}
