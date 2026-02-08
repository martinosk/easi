package commands

type CreateEditGrant struct {
	GrantorID    string
	GrantorEmail string
	GranteeEmail string
	ArtifactType string
	ArtifactID   string
	Scope        string
	Reason       string
}

func (c CreateEditGrant) CommandName() string {
	return "CreateEditGrant"
}
