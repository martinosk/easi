package ports

import "context"

type InvitationRequest struct {
	GranteeEmail string
	GrantorID    string
	GrantorEmail string
}

type InvitationRequester interface {
	RequestInvitation(ctx context.Context, request InvitationRequest) error
}
