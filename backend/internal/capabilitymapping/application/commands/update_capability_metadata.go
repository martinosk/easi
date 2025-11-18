package commands

type UpdateCapabilityMetadata struct {
	ID             string
	StrategyPillar string
	PillarWeight   int
	MaturityLevel  string
	OwnershipModel string
	PrimaryOwner   string
	EAOwner        string
	Status         string
}

func (c UpdateCapabilityMetadata) CommandName() string {
	return "UpdateCapabilityMetadata"
}
