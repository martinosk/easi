package events

import (
	"time"

	"easi/backend/internal/shared/eventsourcing"
)

type ImportFailed struct {
	domain.BaseEvent
	ID       string
	Reason   string
	FailedAt time.Time
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
