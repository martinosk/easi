package projectors

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architecturemodeling/application/commands"
	"easi/backend/internal/architecturemodeling/domain/events"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTeamOwnedComponentFinder struct {
	componentIDs []string
	err          error
}

func (f *fakeTeamOwnedComponentFinder) FindComponentIDsByTeamOwner(_ context.Context, _ string) ([]string, error) {
	return f.componentIDs, f.err
}

type fakeCommandDispatcher struct {
	dispatched []cqrs.Command
	err        error
}

func (f *fakeCommandDispatcher) Dispatch(_ context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	if f.err != nil {
		return cqrs.EmptyResult(), f.err
	}
	f.dispatched = append(f.dispatched, cmd)
	return cqrs.EmptyResult(), nil
}

func TestTeamOwnershipDeletionReactor_ClearsEachOwnedComponent(t *testing.T) {
	finder := &fakeTeamOwnedComponentFinder{componentIDs: []string{"comp-1", "comp-2"}}
	dispatcher := &fakeCommandDispatcher{}
	reactor := NewTeamOwnershipDeletionReactor(finder, dispatcher)

	require.NoError(t, reactor.Handle(context.Background(), events.NewInternalTeamDeleted("team-1", "Platform Ops")))

	require.Len(t, dispatcher.dispatched, 2)
	first, ok := dispatcher.dispatched[0].(*commands.ClearApplicationComponentOwnership)
	require.True(t, ok)
	assert.Equal(t, "comp-1", first.ComponentID)
}

func TestTeamOwnershipDeletionReactor_NoOwnedComponents(t *testing.T) {
	reactor := NewTeamOwnershipDeletionReactor(&fakeTeamOwnedComponentFinder{}, &fakeCommandDispatcher{})

	require.NoError(t, reactor.Handle(context.Background(), events.NewInternalTeamDeleted("team-1", "Platform Ops")))
}

func TestTeamOwnershipDeletionReactor_FinderErrorFails(t *testing.T) {
	finder := &fakeTeamOwnedComponentFinder{err: errors.New("db down")}
	reactor := NewTeamOwnershipDeletionReactor(finder, &fakeCommandDispatcher{})

	err := reactor.Handle(context.Background(), events.NewInternalTeamDeleted("team-1", "Platform Ops"))

	assert.Error(t, err)
}

func TestTeamOwnershipDeletionReactor_DispatchErrorFails(t *testing.T) {
	finder := &fakeTeamOwnedComponentFinder{componentIDs: []string{"comp-1"}}
	reactor := NewTeamOwnershipDeletionReactor(finder, &fakeCommandDispatcher{err: errors.New("boom")})

	err := reactor.Handle(context.Background(), events.NewInternalTeamDeleted("team-1", "Platform Ops"))

	assert.Error(t, err)
}

func TestTeamOwnershipDeletionReactor_IgnoresOtherEvents(t *testing.T) {
	dispatcher := &fakeCommandDispatcher{}
	reactor := NewTeamOwnershipDeletionReactor(&fakeTeamOwnedComponentFinder{componentIDs: []string{"comp-1"}}, dispatcher)

	require.NoError(t, reactor.Handle(context.Background(), events.NewInternalTeamUpdated("team-1", "Platform Ops", "", "", "")))

	assert.Empty(t, dispatcher.dispatched)
}
