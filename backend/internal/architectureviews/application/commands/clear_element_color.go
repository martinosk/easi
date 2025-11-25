package commands

type ClearElementColor struct {
	ViewID      string
	ElementID   string
	ElementType string
}

func (c *ClearElementColor) CommandName() string {
	return "ClearElementColor"
}
