package commands

type CreateApplicationComponent struct {
	Name        string
	Description string
	// ID will be populated by the handler after creation
	ID string
}

func (c CreateApplicationComponent) CommandName() string {
	return "CreateApplicationComponent"
}
