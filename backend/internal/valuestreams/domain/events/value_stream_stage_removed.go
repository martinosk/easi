package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
)

type ValueStreamStageRemoved struct {
	domain.BaseEvent
	ID      string `json:"id"`
	StageID string `json:"stageId"`
}

func NewValueStreamStageRemoved(id, stageID string) ValueStreamStageRemoved {
	return ValueStreamStageRemoved{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		StageID:   stageID,
	}
}

func (e ValueStreamStageRemoved) EventType() string {
	return "ValueStreamStageRemoved"
}

func (e ValueStreamStageRemoved) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":      e.ID,
		"stageId": e.StageID,
	}
}

func (e ValueStreamStageRemoved) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
