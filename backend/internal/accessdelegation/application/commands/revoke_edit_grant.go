package commands

type RevokeEditGrant struct {
	ID        string
	RevokedBy string
}

func (c RevokeEditGrant) CommandName() string {
	return "RevokeEditGrant"
}
