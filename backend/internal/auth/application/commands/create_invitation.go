package commands

type CreateInvitation struct {
	Email        string
	Role         string
	InviterID    string
	InviterEmail string
	ID           string
}

func (c CreateInvitation) CommandName() string {
	return "CreateInvitation"
}
