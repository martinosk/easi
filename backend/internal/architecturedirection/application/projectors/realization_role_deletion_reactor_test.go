package projectors

import (
	"context"
	"testing"

	"easi/backend/internal/architecturedirection/application/commands"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRealizationRolePairFinder struct {
	byRealizationID map[string][2]string
}

func (f *fakeRealizationRolePairFinder) FindPairByRealizationID(_ context.Context, realizationID string) (string, string, bool, error) {
	pair, ok := f.byRealizationID[realizationID]
	if !ok {
		return "", "", false, nil
	}
	return pair[0], pair[1], true, nil
}

func TestRealizationRoleDeletionReactor_KnownRealization_DispatchesClear(t *testing.T) {
	realizationID := uuid.New().String()
	capID := uuid.New().String()
	compID := uuid.New().String()
	finder := &fakeRealizationRolePairFinder{byRealizationID: map[string][2]string{
		realizationID: {capID, compID},
	}}
	dispatcher := &fakeDispatcher{}
	reactor := NewRealizationRoleDeletionReactor(finder, dispatcher)

	err := reactor.ProjectEvent(context.Background(), "SystemRealizationDeleted", []byte(`{"id":"`+realizationID+`"}`))

	require.NoError(t, err)
	require.Len(t, dispatcher.dispatched, 1, "R6: deleting a realisation clears its role via a recorded reaction")
	cmd, ok := dispatcher.dispatched[0].(*commands.ClearRealizationRole)
	require.True(t, ok)
	assert.Equal(t, capID, cmd.CapabilityID)
	assert.Equal(t, compID, cmd.ComponentID)
	assert.Equal(t, "system:realization-deleted", cmd.ClearedBy)
}

func TestRealizationRoleDeletionReactor_UnknownRealization_NoOp(t *testing.T) {
	finder := &fakeRealizationRolePairFinder{byRealizationID: map[string][2]string{}}
	dispatcher := &fakeDispatcher{}
	reactor := NewRealizationRoleDeletionReactor(finder, dispatcher)

	err := reactor.ProjectEvent(context.Background(), "SystemRealizationDeleted", []byte(`{"id":"`+uuid.New().String()+`"}`))

	require.NoError(t, err)
	assert.Empty(t, dispatcher.dispatched)
}

func TestRealizationRoleDeletionReactor_OtherEventTypes_Ignored(t *testing.T) {
	dispatcher := &fakeDispatcher{}
	reactor := NewRealizationRoleDeletionReactor(&fakeRealizationRolePairFinder{}, dispatcher)

	err := reactor.ProjectEvent(context.Background(), "SomethingElse", []byte(`{}`))

	require.NoError(t, err)
	assert.Empty(t, dispatcher.dispatched)
}
