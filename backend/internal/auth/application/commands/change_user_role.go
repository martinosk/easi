package commands

type ChangeUserRole struct {
	UserID      string
	NewRole     string
	ChangedByID string
}

func (c ChangeUserRole) CommandName() string {
	return "ChangeUserRole"
}
