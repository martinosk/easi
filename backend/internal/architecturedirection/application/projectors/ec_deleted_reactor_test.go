package projectors

import (
	"context"
	"testing"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/shared/cqrs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeActiveDirectionFinder struct {
	byEC map[string]*readmodels.DirectionDTO
}

func (f *fakeActiveDirectionFinder) GetActiveByEnterpriseCapabilityID(_ context.Context, ecID string) (*readmodels.DirectionDTO, error) {
	return f.byEC[ecID], nil
}

type fakeDispatcher struct {
	dispatched []cqrs.Command
}

func (f *fakeDispatcher) Dispatch(_ context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	f.dispatched = append(f.dispatched, cmd)
	return cqrs.EmptyResult(), nil
}

func TestECDeletedReactor_RejectsActiveDirection(t *testing.T) {
	ecID := uuid.New().String()
	directionID := uuid.New().String()
	finder := &fakeActiveDirectionFinder{byEC: map[string]*readmodels.DirectionDTO{
		ecID: {ID: directionID, EnterpriseCapabilityID: ecID, Status: "agreed"},
	}}
	dispatcher := &fakeDispatcher{}
	reactor := NewEnterpriseCapabilityDeletedReactor(finder, dispatcher)

	err := reactor.ProjectEvent(context.Background(), "EnterpriseCapabilityDeleted", []byte(`{"id":"`+ecID+`"}`))

	require.NoError(t, err)
	require.Len(t, dispatcher.dispatched, 1, "deleting an EC rejects its active direction (R6)")
	cmd, ok := dispatcher.dispatched[0].(*commands.RejectDirection)
	require.True(t, ok)
	assert.Equal(t, directionID, cmd.DirectionID)
}

func TestECDeletedReactor_NoActiveDirection_NoCommand(t *testing.T) {
	finder := &fakeActiveDirectionFinder{byEC: map[string]*readmodels.DirectionDTO{}}
	dispatcher := &fakeDispatcher{}
	reactor := NewEnterpriseCapabilityDeletedReactor(finder, dispatcher)

	err := reactor.ProjectEvent(context.Background(), "EnterpriseCapabilityDeleted", []byte(`{"id":"`+uuid.New().String()+`"}`))

	require.NoError(t, err)
	assert.Empty(t, dispatcher.dispatched)
}

func TestECDeletedReactor_OtherEventTypes_Ignored(t *testing.T) {
	dispatcher := &fakeDispatcher{}
	reactor := NewEnterpriseCapabilityDeletedReactor(&fakeActiveDirectionFinder{}, dispatcher)

	err := reactor.ProjectEvent(context.Background(), "SomethingElse", []byte(`{}`))

	require.NoError(t, err)
	assert.Empty(t, dispatcher.dispatched)
}
