package commands

type UpdateBusinessDomain struct {
	ID          string
	Name        string
	Description string
}

func (c UpdateBusinessDomain) CommandName() string {
	return "UpdateBusinessDomain"
}
