package publishedlanguage

type CreateCapability struct {
	Name        string
	Description string
	ParentID    string
	Level       string
}

func (c CreateCapability) CommandName() string {
	return "CreateCapability"
}

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

type LinkSystemToCapability struct {
	CapabilityID     string
	ComponentID      string
	RealizationLevel string
	Notes            string
}

func (c LinkSystemToCapability) CommandName() string {
	return "LinkSystemToCapability"
}

type AssignCapabilityToDomain struct {
	BusinessDomainID string
	CapabilityID     string
}

func (c AssignCapabilityToDomain) CommandName() string {
	return "AssignCapabilityToDomain"
}
