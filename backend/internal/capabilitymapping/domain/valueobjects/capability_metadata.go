package valueobjects

// CapabilityMetadata encapsulates all metadata fields for a capability
type CapabilityMetadata struct {
	strategyPillar StrategyPillar
	pillarWeight   PillarWeight
	maturityLevel  MaturityLevel
	ownershipModel OwnershipModel
	primaryOwner   Owner
	eaOwner        Owner
	status         CapabilityStatus
}

// NewCapabilityMetadata creates a new capability metadata value object
func NewCapabilityMetadata(
	strategyPillar StrategyPillar,
	pillarWeight PillarWeight,
	maturityLevel MaturityLevel,
	ownershipModel OwnershipModel,
	primaryOwner Owner,
	eaOwner Owner,
	status CapabilityStatus,
) CapabilityMetadata {
	return CapabilityMetadata{
		strategyPillar: strategyPillar,
		pillarWeight:   pillarWeight,
		maturityLevel:  maturityLevel,
		ownershipModel: ownershipModel,
		primaryOwner:   primaryOwner,
		eaOwner:        eaOwner,
		status:         status,
	}
}

// StrategyPillar returns the strategy pillar
func (m CapabilityMetadata) StrategyPillar() StrategyPillar {
	return m.strategyPillar
}

// PillarWeight returns the pillar weight
func (m CapabilityMetadata) PillarWeight() PillarWeight {
	return m.pillarWeight
}

// MaturityLevel returns the maturity level
func (m CapabilityMetadata) MaturityLevel() MaturityLevel {
	return m.maturityLevel
}

// OwnershipModel returns the ownership model
func (m CapabilityMetadata) OwnershipModel() OwnershipModel {
	return m.ownershipModel
}

// PrimaryOwner returns the primary owner
func (m CapabilityMetadata) PrimaryOwner() Owner {
	return m.primaryOwner
}

// EAOwner returns the EA owner
func (m CapabilityMetadata) EAOwner() Owner {
	return m.eaOwner
}

// Status returns the capability status
func (m CapabilityMetadata) Status() CapabilityStatus {
	return m.status
}
