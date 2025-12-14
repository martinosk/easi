package commands

type MarkInvitationExpired struct {
	ID string
}

func (c MarkInvitationExpired) CommandName() string {
	return "MarkInvitationExpired"
}
