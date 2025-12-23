package events

import (
	"time"

	"easi/backend/internal/shared/eventsourcing"
)

type CapabilityCreated struct {
	domain.BaseEvent
	ID          string
	Name        string
	Description string
	ParentID    string
	Level       string
	CreatedAt   time.Time
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
