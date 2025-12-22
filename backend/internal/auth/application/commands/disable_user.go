package commands

type DisableUser struct {
	UserID       string
	DisabledByID string
}

func (c DisableUser) CommandName() string {
	return "DisableUser"
}
