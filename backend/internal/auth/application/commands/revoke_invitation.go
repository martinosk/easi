package commands

type RevokeInvitation struct {
	ID string
}

func (c RevokeInvitation) CommandName() string {
	return "RevokeInvitation"
}
