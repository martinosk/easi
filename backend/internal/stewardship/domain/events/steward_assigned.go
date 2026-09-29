package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
	pl "easi/backend/internal/stewardship/publishedlanguage"
)

type StewardAssigned struct {
	domain.BaseEvent
	ID         string    `json:"id"`
	DomainID   string    `json:"domainId"`
	Concern    string    `json:"concern"`
	StewardID  string    `json:"stewardId"`
	AssignedBy string    `json:"assignedBy"`
	AssignedAt time.Time `json:"assignedAt"`
}

type StewardAssignedFields struct {
	ID         string
	DomainID   string
	Concern    string
	StewardID  string
	AssignedBy string
}

func NewStewardAssigned(f StewardAssignedFields) StewardAssigned {
	return StewardAssigned{
		BaseEvent:  domain.NewBaseEvent(f.ID),
		ID:         f.ID,
		DomainID:   f.DomainID,
		Concern:    f.Concern,
		StewardID:  f.StewardID,
		AssignedBy: f.AssignedBy,
		AssignedAt: time.Now().UTC(),
	}
}

func (e StewardAssigned) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func (e StewardAssigned) EventType() string { return pl.StewardAssigned }

func (e StewardAssigned) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":         e.ID,
		"domainId":   e.DomainID,
		"concern":    e.Concern,
		"stewardId":  e.StewardID,
		"assignedBy": e.AssignedBy,
		"assignedAt": e.AssignedAt.Format(time.RFC3339Nano),
	}
}
