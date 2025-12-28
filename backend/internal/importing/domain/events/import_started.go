package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type ImportStarted struct {
	domain.BaseEvent
	ID         string
	TotalItems int
	StartedAt  time.Time
}

func NewImportStarted(id string, totalItems int) ImportStarted {
	return ImportStarted{
		BaseEvent:  domain.NewBaseEvent(id),
		ID:         id,
		TotalItems: totalItems,
		StartedAt:  time.Now().UTC(),
	}
}

func (e ImportStarted) EventType() string {
	return "ImportStarted"
}

func (e ImportStarted) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":         e.ID,
		"totalItems": e.TotalItems,
		"startedAt":  e.StartedAt,
	}
}
