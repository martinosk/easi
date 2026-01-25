package events

import domain "easi/backend/internal/shared/eventsourcing"

type CapabilityMetadataUpdated struct {
	domain.BaseEvent
	ID             string `json:"id"`
	StrategyPillar string `json:"strategyPillar"`
	PillarWeight   int    `json:"pillarWeight"`
	MaturityValue  int    `json:"maturityValue"`
	OwnershipModel string `json:"ownershipModel"`
	PrimaryOwner   string `json:"primaryOwner"`
	EAOwner        string `json:"eaOwner"`
	Status         string `json:"status"`
}

func NewCapabilityMetadataUpdated(
	id, strategyPillar string,
	pillarWeight int,
	maturityValue int,
	ownershipModel, primaryOwner, eaOwner, status string,
) CapabilityMetadataUpdated {
	return CapabilityMetadataUpdated{
		BaseEvent:      domain.NewBaseEvent(id),
		ID:             id,
		StrategyPillar: strategyPillar,
		PillarWeight:   pillarWeight,
		MaturityValue:  maturityValue,
		OwnershipModel: ownershipModel,
		PrimaryOwner:   primaryOwner,
		EAOwner:        eaOwner,
		Status:         status,
	}
}

func (e CapabilityMetadataUpdated) EventType() string {
	return "CapabilityMetadataUpdated"
}

func (e CapabilityMetadataUpdated) EventData() map[string]interface{} {
	return map[string]interface{}{
		"id":             e.ID,
		"strategyPillar": e.StrategyPillar,
		"pillarWeight":   e.PillarWeight,
		"maturityValue":  e.MaturityValue,
		"ownershipModel": e.OwnershipModel,
		"primaryOwner":   e.PrimaryOwner,
		"eaOwner":        e.EAOwner,
		"status":         e.Status,
	}
}

func (e CapabilityMetadataUpdated) AggregateID() string {
	if baseID := e.BaseEvent.AggregateID(); baseID != "" {
		return baseID
	}
	return e.ID
}
