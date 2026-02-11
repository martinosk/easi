package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
)

type ValueStreamStageCapabilityAdded struct {
	domain.BaseEvent
	ID           string `json:"id"`
	StageID      string `json:"stageId"`
	CapabilityID string `json:"capabilityId"`
}

func NewValueStreamStageCapabilityAdded(id, stageID, capabilityID string) ValueStreamStageCapabilityAdded {
	return ValueStreamStageCapabilityAdded{
		BaseEvent:    domain.NewBaseEvent(id),
		ID:           id,
		StageID:      stageID,
		CapabilityID: capabilityID,
	}
}

func (e ValueStreamStageCapabilityAdded) EventType() string {
	return "ValueStreamStageCapabilityAdded"
}

func (e ValueStreamStageCapabilityAdded) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":           e.ID,
		"stageId":      e.StageID,
		"capabilityId": e.CapabilityID,
	}
}

func (e ValueStreamStageCapabilityAdded) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
