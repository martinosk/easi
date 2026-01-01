package events

import (
	"time"

	domain "easi/backend/internal/shared/eventsourcing"
)

type EnterpriseCapabilityTargetMaturitySet struct {
	domain.BaseEvent
	ID             string
	TargetMaturity int
	SetAt          time.Time
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
