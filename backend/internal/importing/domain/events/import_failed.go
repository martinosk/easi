package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type ImportFailed struct {
	domain.BaseEvent
	ID       string    `json:"id"`
	Reason   string    `json:"reason"`
	FailedAt time.Time `json:"failedAt"`
}

func (e ImportFailed) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewImportFailed(id, reason string) ImportFailed {
	return ImportFailed{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		Reason:    reason,
		FailedAt:  time.Now().UTC(),
	}
}

func (e ImportFailed) EventType() string {
	return "ImportFailed"
}

func (e ImportFailed) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":       e.ID,
		"reason":   e.Reason,
		"failedAt": e.FailedAt,
	}
}
