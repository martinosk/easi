package commands

type CreateBusinessDomain struct {
	Name        string
	Description string
}

func (c CreateBusinessDomain) CommandName() string {
	return "CreateBusinessDomain"
}
