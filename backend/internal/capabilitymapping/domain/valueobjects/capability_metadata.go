package valueobjects

type CapabilityMetadata struct {
	maturityLevel  MaturityLevel
	ownershipModel OwnershipModel
	primaryOwner   Owner
	eaOwner        Owner
	status         CapabilityStatus
}

func NewCapabilityMetadata(
	maturityLevel MaturityLevel,
	ownershipModel OwnershipModel,
	primaryOwner Owner,
	eaOwner Owner,
	status CapabilityStatus,
) CapabilityMetadata {
	return CapabilityMetadata{
		maturityLevel:  maturityLevel,
		ownershipModel: ownershipModel,
		primaryOwner:   primaryOwner,
		eaOwner:        eaOwner,
		status:         status,
	}
}

func (m CapabilityMetadata) MaturityLevel() MaturityLevel {
	return m.maturityLevel
}

func (m CapabilityMetadata) OwnershipModel() OwnershipModel {
	return m.ownershipModel
}

func (m CapabilityMetadata) PrimaryOwner() Owner {
	return m.primaryOwner
}

func (m CapabilityMetadata) EAOwner() Owner {
	return m.eaOwner
}

func (m CapabilityMetadata) Status() CapabilityStatus {
	return m.status
}
