package projectors

import (
	"context"
	"testing"

	"easi/backend/internal/architecturedirection/application/commands"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTimeAssessmentPairFinder struct {
	byRealizationID map[string][2]string
}

func (f *fakeTimeAssessmentPairFinder) FindPairByRealizationID(_ context.Context, realizationID string) (string, string, bool, error) {
	pair, ok := f.byRealizationID[realizationID]
	if !ok {
		return "", "", false, nil
	}
	return pair[0], pair[1], true, nil
}

func TestTimeAssessmentDeletionReactor_KnownRealization_DispatchesRemove(t *testing.T) {
	realizationID := uuid.New().String()
	capID := uuid.New().String()
	compID := uuid.New().String()
	finder := &fakeTimeAssessmentPairFinder{byRealizationID: map[string][2]string{
		realizationID: {capID, compID},
	}}
	dispatcher := &fakeDispatcher{}
	reactor := NewTimeAssessmentDeletionReactor(finder, dispatcher)

	err := reactor.ProjectEvent(context.Background(), "SystemRealizationDeleted", []byte(`{"id":"`+realizationID+`"}`))

	require.NoError(t, err)
	require.Len(t, dispatcher.dispatched, 1, "deleting a realisation removes its TIME assessment via a recorded reaction")
	cmd, ok := dispatcher.dispatched[0].(*commands.RemoveTimeAssessment)
	require.True(t, ok)
	assert.Equal(t, capID, cmd.CapabilityID)
	assert.Equal(t, compID, cmd.ComponentID)
	assert.Equal(t, "system:realization-deleted", cmd.RemovedBy)
}

func TestTimeAssessmentDeletionReactor_UnknownRealization_NoOp(t *testing.T) {
	finder := &fakeTimeAssessmentPairFinder{byRealizationID: map[string][2]string{}}
	dispatcher := &fakeDispatcher{}
	reactor := NewTimeAssessmentDeletionReactor(finder, dispatcher)

	err := reactor.ProjectEvent(context.Background(), "SystemRealizationDeleted", []byte(`{"id":"`+uuid.New().String()+`"}`))

	require.NoError(t, err)
	assert.Empty(t, dispatcher.dispatched)
}

func TestTimeAssessmentDeletionReactor_OtherEventTypes_Ignored(t *testing.T) {
	dispatcher := &fakeDispatcher{}
	reactor := NewTimeAssessmentDeletionReactor(&fakeTimeAssessmentPairFinder{}, dispatcher)

	err := reactor.ProjectEvent(context.Background(), "SomethingElse", []byte(`{}`))

	require.NoError(t, err)
	assert.Empty(t, dispatcher.dispatched)
}
