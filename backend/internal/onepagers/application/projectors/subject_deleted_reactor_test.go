package projectors

import (
	"context"
	"fmt"
	"testing"

	"easi/backend/internal/onepagers/application/commands"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/shared/cqrs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDispatcher struct {
	dispatched []cqrs.Command
	err        error
}

func (d *fakeDispatcher) Dispatch(_ context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	d.dispatched = append(d.dispatched, cmd)
	return cqrs.EmptyResult(), d.err
}

type fakeFactsFinder struct {
	ids map[string]string
	err error
}

func (f *fakeFactsFinder) FactsIDForSubject(_ context.Context, subject readmodels.SubjectKey) (string, error) {
	return f.ids[subject.SubjectType+"/"+subject.SubjectID], f.err
}

func TestSubjectDeletedReactor_ArchivesFactsForEachSubjectType(t *testing.T) {
	cases := []struct {
		eventType   string
		subjectType string
	}{
		{"CapabilityDeleted", "capability"},
		{"ApplicationComponentDeleted", "application"},
		{"AcquiredEntityDeleted", "acquired-entity"},
		{"VendorDeleted", "vendor"},
		{"InternalTeamDeleted", "internal-team"},
	}

	for _, tc := range cases {
		t.Run(tc.eventType, func(t *testing.T) {
			subjectID := uuid.New().String()
			factsID := uuid.New().String()
			finder := &fakeFactsFinder{ids: map[string]string{tc.subjectType + "/" + subjectID: factsID}}
			dispatcher := &fakeDispatcher{}
			reactor := NewSubjectDeletedReactor(finder, dispatcher)

			err := reactor.ProjectEvent(context.Background(), tc.eventType, []byte(fmt.Sprintf(`{"id":%q}`, subjectID)))

			require.NoError(t, err)
			require.Len(t, dispatcher.dispatched, 1)
			archive, ok := dispatcher.dispatched[0].(*commands.ArchiveOnePagerFacts)
			require.True(t, ok)
			assert.Equal(t, factsID, archive.FactsID)
			assert.Equal(t, "subject deleted", archive.Reason)
		})
	}
}

func TestSubjectDeletedReactor_NoFactsMeansNoDispatch(t *testing.T) {
	dispatcher := &fakeDispatcher{}
	reactor := NewSubjectDeletedReactor(&fakeFactsFinder{ids: map[string]string{}}, dispatcher)

	err := reactor.ProjectEvent(context.Background(), "CapabilityDeleted", []byte(`{"id":"cap-1"}`))

	require.NoError(t, err)
	assert.Empty(t, dispatcher.dispatched)
}

func TestSubjectDeletedReactor_IgnoresUnrelatedEvents(t *testing.T) {
	dispatcher := &fakeDispatcher{}
	reactor := NewSubjectDeletedReactor(&fakeFactsFinder{ids: map[string]string{}}, dispatcher)

	err := reactor.ProjectEvent(context.Background(), "CapabilityCreated", []byte(`{"id":"cap-1"}`))

	require.NoError(t, err)
	assert.Empty(t, dispatcher.dispatched)
}

func TestSubjectDeletedReactor_SubscribedEventTypesCoverAllFiveSubjects(t *testing.T) {
	assert.ElementsMatch(t, []string{
		"CapabilityDeleted",
		"ApplicationComponentDeleted",
		"AcquiredEntityDeleted",
		"VendorDeleted",
		"InternalTeamDeleted",
	}, SubjectDeletionEventTypes())
}
