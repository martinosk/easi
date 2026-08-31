package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseCapabilityTargetMaturitySet struct {
	domain.BaseEvent
	ID             string    `json:"id"`
	TargetMaturity int       `json:"targetMaturity"`
	SetAt          time.Time `json:"setAt"`
}

func (e EnterpriseCapabilityTargetMaturitySet) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}

func NewEnterpriseCapabilityTargetMaturitySet(id string, targetMaturity int) EnterpriseCapabilityTargetMaturitySet {
	return EnterpriseCapabilityTargetMaturitySet{
		BaseEvent:      domain.NewBaseEvent(id),
		ID:             id,
		TargetMaturity: targetMaturity,
		SetAt:          time.Now().UTC(),
	}
}

func (e EnterpriseCapabilityTargetMaturitySet) EventType() string {
	return "EnterpriseCapabilityTargetMaturitySet"
}

func (e EnterpriseCapabilityTargetMaturitySet) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":             e.ID,
		"targetMaturity": e.TargetMaturity,
		"setAt":          e.SetAt,
	}
}
