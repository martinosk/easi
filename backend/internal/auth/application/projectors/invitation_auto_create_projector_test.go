package projectors

import (
	"context"
	"errors"
	"testing"
	"time"

	authCommands "easi/backend/internal/auth/application/commands"
	"easi/backend/internal/shared/cqrs"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type spyAutoInviteCommandBus struct {
	dispatched []cqrs.Command
	err        error
}

func (s *spyAutoInviteCommandBus) Dispatch(_ context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	s.dispatched = append(s.dispatched, cmd)
	return cqrs.EmptyResult(), s.err
}

func (s *spyAutoInviteCommandBus) Register(_ string, _ cqrs.CommandHandler) {}

func autoInviteEvent(data map[string]interface{}) domain.DomainEvent {
	return &autoInviteTestEvent{data: data}
}

type autoInviteTestEvent struct {
	data map[string]interface{}
}

func (e *autoInviteTestEvent) AggregateID() string               { return "" }
func (e *autoInviteTestEvent) EventType() string                 { return "EditGrantForNonUserCreated" }
func (e *autoInviteTestEvent) OccurredAt() time.Time             { return time.Now() }
func (e *autoInviteTestEvent) EventData() map[string]interface{} { return e.data }

func TestInvitationAutoCreateProjector_DispatchesCreateInvitationCommand(t *testing.T) {
	bus := &spyAutoInviteCommandBus{}
	projector := NewInvitationAutoCreateProjector(bus)

	err := projector.Handle(context.Background(), autoInviteEvent(map[string]interface{}{
		"granteeEmail": "new@example.com",
		"grantorId":    "grantor-123",
		"grantorEmail": "grantor@example.com",
	}))
	require.NoError(t, err)

	require.Len(t, bus.dispatched, 1)
	cmd, ok := bus.dispatched[0].(*authCommands.CreateInvitation)
	require.True(t, ok)
	assert.Equal(t, "new@example.com", cmd.Email)
	assert.Equal(t, "stakeholder", cmd.Role)
	assert.Equal(t, "grantor-123", cmd.InviterID)
	assert.Equal(t, "grantor@example.com", cmd.InviterEmail)
}

func TestInvitationAutoCreateProjector_CommandBusError_DoesNotReturnError(t *testing.T) {
	bus := &spyAutoInviteCommandBus{err: errors.New("invitation already exists")}
	projector := NewInvitationAutoCreateProjector(bus)

	err := projector.Handle(context.Background(), autoInviteEvent(map[string]interface{}{
		"granteeEmail": "fail@example.com",
		"grantorId":    "grantor-123",
		"grantorEmail": "grantor@example.com",
	}))
	assert.NoError(t, err)
}

func TestInvitationAutoCreateProjector_UnmarshalableEventData_ReturnsError(t *testing.T) {
	bus := &spyAutoInviteCommandBus{}
	projector := NewInvitationAutoCreateProjector(bus)

	err := projector.Handle(context.Background(), autoInviteEvent(map[string]interface{}{
		"unexpected": make(chan int),
	}))
	assert.Error(t, err)
	assert.Empty(t, bus.dispatched)
}

func TestInvitationAutoCreateProjector_EmptyEventData_DispatchesWithEmptyFields(t *testing.T) {
	bus := &spyAutoInviteCommandBus{}
	projector := NewInvitationAutoCreateProjector(bus)

	err := projector.Handle(context.Background(), autoInviteEvent(map[string]interface{}{}))
	require.NoError(t, err)

	require.Len(t, bus.dispatched, 1)
	cmd, ok := bus.dispatched[0].(*authCommands.CreateInvitation)
	require.True(t, ok)
	assert.Equal(t, "", cmd.Email)
	assert.Equal(t, "stakeholder", cmd.Role)
}

func TestInvitationAutoCreateProjector_AlwaysAssignsStakeholderRole(t *testing.T) {
	bus := &spyAutoInviteCommandBus{}
	projector := NewInvitationAutoCreateProjector(bus)

	_ = projector.Handle(context.Background(), autoInviteEvent(map[string]interface{}{
		"granteeEmail": "user@example.com",
		"grantorId":    "admin-1",
		"grantorEmail": "admin@example.com",
	}))

	require.Len(t, bus.dispatched, 1)
	cmd := bus.dispatched[0].(*authCommands.CreateInvitation)
	assert.Equal(t, "stakeholder", cmd.Role)
}
