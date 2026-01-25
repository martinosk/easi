package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type ImportStarted struct {
	domain.BaseEvent
	ID         string    `json:"id"`
	TotalItems int       `json:"totalItems"`
	StartedAt  time.Time `json:"startedAt"`
}

func (e ImportStarted) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
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
