package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
)

type ValueStreamStageAdded struct {
	domain.BaseEvent
	ID          string `json:"id"`
	StageID     string `json:"stageId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Position    int    `json:"position"`
}

func NewValueStreamStageAdded(id, stageID, name, description string, position int) ValueStreamStageAdded {
	return ValueStreamStageAdded{
		BaseEvent:   domain.NewBaseEvent(id),
		ID:          id,
		StageID:     stageID,
		Name:        name,
		Description: description,
		Position:    position,
	}
}

func (e ValueStreamStageAdded) EventType() string {
	return "ValueStreamStageAdded"
}

func (e ValueStreamStageAdded) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":          e.ID,
		"stageId":     e.StageID,
		"name":        e.Name,
		"description": e.Description,
		"position":    e.Position,
	}
}

func (e ValueStreamStageAdded) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
