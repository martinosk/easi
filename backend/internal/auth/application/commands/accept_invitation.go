package commands

type AcceptInvitation struct {
	Email string
}

func (c AcceptInvitation) CommandName() string {
	return "AcceptInvitation"
}
