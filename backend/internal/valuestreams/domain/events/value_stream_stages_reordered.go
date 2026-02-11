package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
)

type StagePositionEntry struct {
	StageID  string `json:"stageId"`
	Position int    `json:"position"`
}

type ValueStreamStagesReordered struct {
	domain.BaseEvent
	ID        string               `json:"id"`
	Positions []StagePositionEntry `json:"positions"`
}

func NewValueStreamStagesReordered(id string, positions []StagePositionEntry) ValueStreamStagesReordered {
	return ValueStreamStagesReordered{
		BaseEvent: domain.NewBaseEvent(id),
		ID:        id,
		Positions: positions,
	}
}

func (e ValueStreamStagesReordered) EventType() string {
	return "ValueStreamStagesReordered"
}

func (e ValueStreamStagesReordered) EventData() map[string]interface{} {
	posData := make([]interface{}, len(e.Positions))
	for i, p := range e.Positions {
		posData[i] = map[string]interface{}{
			"stageId":  p.StageID,
			"position": p.Position,
		}
	}
	return map[string]interface{}{
		"id":        e.ID,
		"positions": posData,
	}
}

func (e ValueStreamStagesReordered) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
