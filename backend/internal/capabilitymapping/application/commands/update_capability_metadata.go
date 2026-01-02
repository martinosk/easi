package commands

type UpdateCapabilityMetadata struct {
	ID             string
	MaturityValue  int
	MaturityLevel  string
	OwnershipModel string
	PrimaryOwner   string
	EAOwner        string
	Status         string
}

func (c UpdateCapabilityMetadata) CommandName() string {
	return "UpdateCapabilityMetadata"
}
