package commands

type DeleteBusinessDomain struct {
	ID string
}

func (c DeleteBusinessDomain) CommandName() string {
	return "DeleteBusinessDomain"
}
