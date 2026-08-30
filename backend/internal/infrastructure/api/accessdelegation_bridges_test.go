package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	acdPorts "easi/backend/internal/accessdelegation/application/ports"
	authCommands "easi/backend/internal/auth/application/commands"
	"easi/backend/internal/shared/cqrs"
)

type recordingCommandBus struct {
	dispatched []cqrs.Command
}

func (b *recordingCommandBus) Dispatch(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	b.dispatched = append(b.dispatched, cmd)
	return cqrs.NewResult("invitation-1"), nil
}

func (b *recordingCommandBus) Register(commandName string, handler cqrs.CommandHandler) {}

func TestInvitationRequester_CreatesStakeholderInvitationForGrantee(t *testing.T) {
	bus := &recordingCommandBus{}
	requester := invitationRequester{commandBus: bus}

	err := requester.RequestInvitation(context.Background(), acdPorts.InvitationRequest{
		GranteeEmail: "newcomer@dfds.com",
		GrantorID:    "grantor-1",
		GrantorEmail: "grantor@dfds.com",
	})

	require.NoError(t, err)
	require.Len(t, bus.dispatched, 1)
	cmd, ok := bus.dispatched[0].(*authCommands.CreateInvitation)
	require.True(t, ok)
	assert.Equal(t, &authCommands.CreateInvitation{
		Email:        "newcomer@dfds.com",
		Role:         "stakeholder",
		InviterID:    "grantor-1",
		InviterEmail: "grantor@dfds.com",
	}, cmd)
}
