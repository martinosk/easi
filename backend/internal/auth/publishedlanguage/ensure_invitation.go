package publishedlanguage

type EnsureInvitation struct {
	Email        string
	Role         string
	InviterID    string
	InviterEmail string
}

func (c EnsureInvitation) CommandName() string {
	return "EnsureInvitation"
}
