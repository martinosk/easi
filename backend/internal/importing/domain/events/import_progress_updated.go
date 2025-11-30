package events

import (
	"time"

	"easi/backend/internal/shared/domain"
)

type ImportProgressUpdated struct {
	domain.BaseEvent
	ID             string
	Phase          string
	TotalItems     int
	CompletedItems int
	UpdatedAt      time.Time
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
