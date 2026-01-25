package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type ImportSessionCancelled struct {
	domain.BaseEvent
	ID          string    `json:"id"`
	CancelledAt time.Time `json:"cancelledAt"`
}

func (e ImportSessionCancelled) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewImportSessionCancelled(id string) ImportSessionCancelled {
	return ImportSessionCancelled{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		CancelledAt: time.Now().UTC(),
	}
}

func (e ImportSessionCancelled) EventType() string {
	return "ImportSessionCancelled"
}

func (e ImportSessionCancelled) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"cancelledAt": e.CancelledAt,
	}
}
