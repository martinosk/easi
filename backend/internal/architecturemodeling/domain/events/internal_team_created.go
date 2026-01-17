package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type InternalTeamCreated struct {
	domain.BaseEvent
	ID            string
	Name          string
	Department    string
	ContactPerson string
	Notes         string
	CreatedAt     time.Time
}

func NewInternalTeamCreated(id, name, department, contactPerson, notes string) InternalTeamCreated {
	return InternalTeamCreated{
		BaseEvent:     domain.NewBaseEvent(id),
		ID:            id,
		Name:          name,
		Department:    department,
		ContactPerson: contactPerson,
		Notes:         notes,
		CreatedAt:     time.Now().UTC(),
	}
}

func (e InternalTeamCreated) EventType() string {
	return "InternalTeamCreated"
}

func (e InternalTeamCreated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":            e.ID,
		"name":          e.Name,
		"department":    e.Department,
		"contactPerson": e.ContactPerson,
		"notes":         e.Notes,
		"createdAt":     e.CreatedAt,
	}
}
