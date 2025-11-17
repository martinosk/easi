package commands

type UpdateApplicationComponent struct {
	ID          string
	Name        string
	Description string
}

func (c UpdateApplicationComponent) CommandName() string {
	return "UpdateApplicationComponent"
}
