package events

import domain "easi/backend/internal/shared/eventsourcing"

type SystemRealizationUpdated struct {
	domain.BaseEvent
	ID               string `json:"id"`
	RealizationLevel string `json:"realizationLevel"`
	Notes            string `json:"notes"`
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

func (e SystemRealizationUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
