package events

import (
	"time"

	"easi/backend/internal/shared/eventsourcing"
)

type ImportSessionCancelled struct {
	domain.BaseEvent
	ID          string
	CancelledAt time.Time
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
