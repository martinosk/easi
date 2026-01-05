package commands

type ChangeViewVisibility struct {
	ViewID    string
	IsPrivate bool
}

func (c *ChangeViewVisibility) CommandName() string {
	return "ChangeViewVisibility"
}
