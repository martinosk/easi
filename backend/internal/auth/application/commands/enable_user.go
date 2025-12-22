package commands

type EnableUser struct {
	UserID      string
	EnabledByID string
}

func (c EnableUser) CommandName() string {
	return "EnableUser"
}
