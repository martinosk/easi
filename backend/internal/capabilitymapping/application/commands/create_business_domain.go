package commands

type CreateBusinessDomain struct {
	Name              string
	Description       string
	DomainArchitectID string
}

func (c CreateBusinessDomain) CommandName() string {
	return "CreateBusinessDomain"
}
