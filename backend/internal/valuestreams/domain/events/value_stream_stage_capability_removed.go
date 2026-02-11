package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
)

type ValueStreamStageCapabilityRemoved struct {
	domain.BaseEvent
	ID           string `json:"id"`
	StageID      string `json:"stageId"`
	CapabilityID string `json:"capabilityId"`
}

func NewValueStreamStageCapabilityRemoved(id, stageID, capabilityID string) ValueStreamStageCapabilityRemoved {
	return ValueStreamStageCapabilityRemoved{
		BaseEvent:    domain.NewBaseEvent(id),
		ID:           id,
		StageID:      stageID,
		CapabilityID: capabilityID,
	}
}

func (e ValueStreamStageCapabilityRemoved) EventType() string {
	return "ValueStreamStageCapabilityRemoved"
}

func (e ValueStreamStageCapabilityRemoved) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":           e.ID,
		"stageId":      e.StageID,
		"capabilityId": e.CapabilityID,
	}
}

func (e ValueStreamStageCapabilityRemoved) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
