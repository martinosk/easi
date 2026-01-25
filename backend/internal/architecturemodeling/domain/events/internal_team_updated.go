package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type InternalTeamUpdated struct {
	domain.BaseEvent
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Department    string    `json:"department"`
	ContactPerson string    `json:"contactPerson"`
	Notes         string    `json:"notes"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (e InternalTeamUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewInternalTeamUpdated(id, name, department, contactPerson, notes string) InternalTeamUpdated {
	return InternalTeamUpdated{
		BaseEvent:     domain.NewBaseEvent(id),
		ID:            id,
		Name:          name,
		Department:    department,
		ContactPerson: contactPerson,
		Notes:         notes,
		UpdatedAt:     time.Now().UTC(),
	}
}

func (e InternalTeamUpdated) EventType() string {
	return "InternalTeamUpdated"
}

func (e InternalTeamUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":            e.ID,
		"name":          e.Name,
		"department":    e.Department,
		"contactPerson": e.ContactPerson,
		"notes":         e.Notes,
		"updatedAt":     e.UpdatedAt,
	}
}
