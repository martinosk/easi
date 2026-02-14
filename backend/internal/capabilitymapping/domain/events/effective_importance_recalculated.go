package events

import (
	domain "easi/backend/internal/shared/eventsourcing"
)

type EffectiveImportanceRecalculated struct {
	domain.BaseEvent
	CapabilityID     string `json:"capabilityId"`
	BusinessDomainID string `json:"businessDomainId"`
	PillarID         string `json:"pillarId"`
	Importance       int    `json:"importance"`
}

func NewEffectiveImportanceRecalculated(capabilityID, businessDomainID, pillarID string, importance int) EffectiveImportanceRecalculated {
	return EffectiveImportanceRecalculated{
		BaseEvent:        domain.NewBaseEvent(capabilityID),
		CapabilityID:     capabilityID,
		BusinessDomainID: businessDomainID,
		PillarID:         pillarID,
		Importance:       importance,
	}
}

func (e EffectiveImportanceRecalculated) EventType() string {
	return "EffectiveImportanceRecalculated"
}

func (e EffectiveImportanceRecalculated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"capabilityId":     e.CapabilityID,
		"businessDomainId": e.BusinessDomainID,
		"pillarId":         e.PillarID,
		"importance":       e.Importance,
	}
}
