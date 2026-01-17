package commands

type UpdateBusinessDomain struct {
	ID                string
	Name              string
	Description       string
	DomainArchitectID string
}

func (c UpdateBusinessDomain) CommandName() string {
	return "UpdateBusinessDomain"
}
