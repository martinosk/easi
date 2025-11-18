package events

import (
	"easi/backend/internal/shared/domain"
)

type CapabilityMetadataUpdated struct {
	domain.BaseEvent
	ID             string
	StrategyPillar string
	PillarWeight   int
	MaturityLevel  string
	OwnershipModel string
	PrimaryOwner   string
	EAOwner        string
	Status         string
}

func NewCapabilityMetadataUpdated(
	id, strategyPillar string,
	pillarWeight int,
	maturityLevel, ownershipModel, primaryOwner, eaOwner, status string,
) CapabilityMetadataUpdated {
	return CapabilityMetadataUpdated{
		BaseEvent:      domain.NewBaseEvent(id),
		ID:             id,
		StrategyPillar: strategyPillar,
		PillarWeight:   pillarWeight,
		MaturityLevel:  maturityLevel,
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
		"maturityLevel":  e.MaturityLevel,
		"ownershipModel": e.OwnershipModel,
		"primaryOwner":   e.PrimaryOwner,
		"eaOwner":        e.EAOwner,
		"status":         e.Status,
	}
}
