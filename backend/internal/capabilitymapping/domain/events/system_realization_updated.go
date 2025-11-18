package events

import (
	"easi/backend/internal/shared/domain"
)

type SystemRealizationUpdated struct {
	domain.BaseEvent
	ID               string
	RealizationLevel string
	Notes            string
}

func NewSystemRealizationUpdated(id, realizationLevel, notes string) SystemRealizationUpdated {
	return SystemRealizationUpdated{
		BaseEvent:        domain.NewBaseEvent(id),
		ID:               id,
		RealizationLevel: realizationLevel,
		Notes:            notes,
	}
}

func (e SystemRealizationUpdated) EventType() string {
	return "SystemRealizationUpdated"
}

func (e SystemRealizationUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":               e.ID,
		"realizationLevel": e.RealizationLevel,
		"notes":            e.Notes,
	}
}
