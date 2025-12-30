package commands

type CreateApplicationComponent struct {
	Name        string
	Description string
}

func (c CreateApplicationComponent) CommandName() string {
	return "CreateApplicationComponent"
}
