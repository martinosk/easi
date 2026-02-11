package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
)

type ValueStreamStageUpdated struct {
	domain.BaseEvent
	ID          string `json:"id"`
	StageID     string `json:"stageId"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewValueStreamStageUpdated(id, stageID, name, description string) ValueStreamStageUpdated {
	return ValueStreamStageUpdated{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		StageID:     stageID,
		Name:        name,
		Description: description,
	}
}

func (e ValueStreamStageUpdated) EventType() string {
	return "ValueStreamStageUpdated"
}

func (e ValueStreamStageUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"stageId":     e.StageID,
		"name":        e.Name,
		"description": e.Description,
	}
}

func (e ValueStreamStageUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
