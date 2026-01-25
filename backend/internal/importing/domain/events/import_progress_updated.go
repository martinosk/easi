package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"time"
)

type ImportProgressUpdated struct {
	domain.BaseEvent
	ID             string    `json:"id"`
	Phase          string    `json:"phase"`
	TotalItems     int       `json:"totalItems"`
	CompletedItems int       `json:"completedItems"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (e ImportProgressUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewImportProgressUpdated(id, phase string, totalItems, completedItems int) ImportProgressUpdated {
	return ImportProgressUpdated{
		BaseEvent:      domain.NewBaseEvent(id),
		ID:             id,
		Phase:          phase,
		TotalItems:     totalItems,
		CompletedItems: completedItems,
		UpdatedAt:      time.Now().UTC(),
	}
}

func (e ImportProgressUpdated) EventType() string {
	return "ImportProgressUpdated"
}

func (e ImportProgressUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":             e.ID,
		"phase":          e.Phase,
		"totalItems":     e.TotalItems,
		"completedItems": e.CompletedItems,
		"updatedAt":      e.UpdatedAt,
	}
}
