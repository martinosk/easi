package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type CapabilityCreated struct {
	domain.BaseEvent
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ParentID    string    `json:"parentId"`
	Level       string    `json:"level"`
	CreatedAt   time.Time `json:"createdAt"`
}

func NewCapabilityCreated(id, name, description, parentID, level string) CapabilityCreated {
	return CapabilityCreated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		Name:        name,
		Description: description,
		ParentID:    parentID,
		Level:       level,
		CreatedAt:   time.Now().UTC(),
	}
}

func (e CapabilityCreated) EventType() string {
	return "CapabilityCreated"
}

func (e CapabilityCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"name":        e.Name,
		"description": e.Description,
		"parentId":    e.ParentID,
		"level":       e.Level,
		"createdAt":   e.CreatedAt,
	}
}

func (e CapabilityCreated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
