package commands

type UpdateElementColor struct {
	ViewID      string
	ElementID   string
	ElementType string
	Color       string
}

func (c *UpdateElementColor) CommandName() string {
	return "UpdateElementColor"
}
