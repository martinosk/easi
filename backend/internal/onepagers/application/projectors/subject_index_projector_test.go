package projectors_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/onepagers/application/ports"
	"easi/backend/internal/onepagers/application/projectors"
	"easi/backend/internal/onepagers/application/readmodels"
	opevents "easi/backend/internal/onepagers/domain/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type appliedCompleteness struct {
	subjectType string
	required    int
	filled      map[string]int
}

type fakeIndexStore struct {
	upserts    []readmodels.SubjectIndexRecord
	deletes    []readmodels.SubjectKey
	changes    []readmodels.SubjectChange
	recomputes []appliedCompleteness
	idsByType  map[string][]string
}

func (f *fakeIndexStore) Upsert(_ context.Context, record readmodels.SubjectIndexRecord) error {
	f.upserts = append(f.upserts, record)
	return nil
}

func (f *fakeIndexStore) Delete(_ context.Context, subject readmodels.SubjectKey) error {
	f.deletes = append(f.deletes, subject)
	return nil
}

func (f *fakeIndexStore) ApplySubjectChange(_ context.Context, change readmodels.SubjectChange) error {
	f.changes = append(f.changes, change)
	return nil
}

func (f *fakeIndexStore) ApplyCompleteness(_ context.Context, subjectType string, required int, filledBySubject map[string]int) error {
	f.recomputes = append(f.recomputes, appliedCompleteness{subjectType: subjectType, required: required, filled: filledBySubject})
	return nil
}

func (f *fakeIndexStore) SubjectIDs(_ context.Context, subjectType string) ([]string, error) {
	return f.idsByType[subjectType], nil
}

type fakeCounter struct {
	required int
	filled   map[string]int
}

func (f fakeCounter) CountsForSubjects(_ context.Context, _ string, subjectIDs []string) (int, map[string]int, error) {
	filled := make(map[string]int, len(subjectIDs))
	for _, id := range subjectIDs {
		filled[id] = f.filled[id]
	}
	return f.required, filled, nil
}

type fakeAuditReader struct {
	audit ports.SubjectAudit
}

func (f fakeAuditReader) Created(_ context.Context, _ string) (ports.SubjectAudit, error) {
	return f.audit, nil
}

type fakeConfigLookup struct {
	bySubjectType map[string]string
}

func (f fakeConfigLookup) GetByID(_ context.Context, id string) (*readmodels.ConfigurationRecord, error) {
	subjectType, ok := f.bySubjectType[id]
	if !ok {
		return nil, nil
	}
	return &readmodels.ConfigurationRecord{ID: id, SubjectType: subjectType}, nil
}

type projectorFakes struct {
	store   *fakeIndexStore
	counter fakeCounter
	audit   fakeAuditReader
	lookup  fakeConfigLookup
}

type projectorHarness struct {
	t         *testing.T
	store     *fakeIndexStore
	projector *projectors.SubjectIndexProjector
}

func newHarness(t *testing.T, fakes projectorFakes) *projectorHarness {
	return &projectorHarness{
		t:         t,
		store:     fakes.store,
		projector: projectors.NewSubjectIndexProjector(fakes.store, fakes.counter, fakes.audit, fakes.lookup),
	}
}

func (h *projectorHarness) project(eventType string, occurredAt time.Time, payload map[string]any) {
	h.t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(h.t, err)
	require.NoError(h.t, h.projector.ProjectEvent(context.Background(), eventType, occurredAt, data))
}

func subjectKey(subjectType, subjectID string) readmodels.SubjectKey {
	return readmodels.SubjectKey{SubjectType: subjectType, SubjectID: subjectID}
}

func TestSubjectIndexProjector_Created_InsertsRow(t *testing.T) {
	store := &fakeIndexStore{}
	created := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	audit := fakeAuditReader{audit: ports.SubjectAudit{ActorID: "user-1", ActorEmail: "a@dfds.com", CreatedAt: created, Found: true}}
	h := newHarness(t, projectorFakes{store: store, counter: fakeCounter{required: 2, filled: map[string]int{"cap-1": 1}}, audit: audit})

	h.project(capPL.CapabilityCreated, created, map[string]any{"id": "cap-1", "name": "Billing"})

	require.Len(t, store.upserts, 1)
	assert.Equal(t, readmodels.SubjectIndexRecord{
		SubjectType: "capability", SubjectID: "cap-1", Name: "Billing",
		CreatorActorID: "user-1", CreatorEmail: "a@dfds.com",
		CreatedAt: created, LastUpdatedAt: created, RequiredCount: 2, FilledCount: 1,
	}, store.upserts[0])
}

func TestSubjectIndexProjector_Deleted_RemovesRow(t *testing.T) {
	store := &fakeIndexStore{}
	h := newHarness(t, projectorFakes{store: store})

	h.project(capPL.CapabilityDeleted, time.Now(), map[string]any{"id": "cap-1"})

	assert.Equal(t, []readmodels.SubjectKey{subjectKey("capability", "cap-1")}, store.deletes)
}

func TestSubjectIndexProjector_Updated_RefreshesNameAndCompleteness(t *testing.T) {
	store := &fakeIndexStore{}
	at := time.Date(2026, 3, 3, 9, 0, 0, 0, time.UTC)
	h := newHarness(t, projectorFakes{store: store, counter: fakeCounter{required: 3, filled: map[string]int{"cap-1": 3}}})

	h.project(capPL.CapabilityUpdated, at, map[string]any{"id": "cap-1", "name": "Renamed"})

	require.Len(t, store.changes, 1)
	assert.Equal(t, readmodels.SubjectChange{
		Subject: subjectKey("capability", "cap-1"), Name: "Renamed",
		Counts: readmodels.CompletenessCounts{Required: 3, Filled: 3}, OccurredAt: at,
	}, store.changes[0])
}

func TestSubjectIndexProjector_ExpertEvents_ResolveSubjectFromContextIdKey(t *testing.T) {
	cases := []struct {
		name        string
		eventType   string
		payload     map[string]any
		wantSubject readmodels.SubjectKey
	}{
		{
			name:        "capability expert added",
			eventType:   capPL.CapabilityExpertAdded,
			payload:     map[string]any{"capabilityId": "cap-1", "expertName": "Jane"},
			wantSubject: subjectKey("capability", "cap-1"),
		},
		{
			name:        "capability expert removed",
			eventType:   capPL.CapabilityExpertRemoved,
			payload:     map[string]any{"capabilityId": "cap-1", "expertName": "Jane"},
			wantSubject: subjectKey("capability", "cap-1"),
		},
		{
			name:        "application expert added",
			eventType:   amPL.ApplicationComponentExpertAdded,
			payload:     map[string]any{"componentId": "app-9", "expertName": "Jane"},
			wantSubject: subjectKey("application", "app-9"),
		},
		{
			name:        "application expert removed",
			eventType:   amPL.ApplicationComponentExpertRemoved,
			payload:     map[string]any{"componentId": "app-9", "expertName": "Jane"},
			wantSubject: subjectKey("application", "app-9"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			store := &fakeIndexStore{}
			counter := fakeCounter{required: 1, filled: map[string]int{tc.wantSubject.SubjectID: 1}}
			h := newHarness(t, projectorFakes{store: store, counter: counter})

			h.project(tc.eventType, time.Now().UTC(), tc.payload)

			require.Len(t, store.changes, 1)
			assert.Equal(t, tc.wantSubject, store.changes[0].Subject)
			assert.Equal(t, "", store.changes[0].Name)
			assert.Equal(t, 1, store.changes[0].Counts.Filled)
		})
	}
}

func TestSubjectIndexProjector_SubjectUpdateWithoutSubjectId_Errors(t *testing.T) {
	store := &fakeIndexStore{}
	projector := projectors.NewSubjectIndexProjector(store, fakeCounter{}, fakeAuditReader{}, fakeConfigLookup{})

	err := projector.ProjectEvent(context.Background(), capPL.CapabilityExpertAdded, time.Now(), []byte(`{"expertName":"Jane"}`))

	require.Error(t, err)
	assert.Empty(t, store.changes)
}

func TestSubjectIndexProjector_FactsRecorded_RecomputesSubjectOnly(t *testing.T) {
	store := &fakeIndexStore{}
	h := newHarness(t, projectorFakes{store: store, counter: fakeCounter{required: 2, filled: map[string]int{"app-9": 2}}})

	h.project(opevents.TypeFieldValueRecorded, time.Now(), map[string]any{
		"subjectType": "application", "subjectId": "app-9", "fieldId": "f1",
	})

	require.Len(t, store.recomputes, 1)
	assert.Equal(t, appliedCompleteness{subjectType: "application", required: 2, filled: map[string]int{"app-9": 2}}, store.recomputes[0])
	assert.Empty(t, store.changes)
}

func TestSubjectIndexProjector_ConfigChange_RecomputesAllSubjectsOfType(t *testing.T) {
	store := &fakeIndexStore{idsByType: map[string][]string{"application": {"app-1", "app-2"}}}
	counter := fakeCounter{required: 1, filled: map[string]int{"app-1": 1, "app-2": 0}}
	lookup := fakeConfigLookup{bySubjectType: map[string]string{"cfg-app": "application"}}
	h := newHarness(t, projectorFakes{store: store, counter: counter, lookup: lookup})

	h.project(opevents.TypeBuiltInFieldRequirementChanged, time.Now(), map[string]any{"id": "cfg-app"})

	require.Len(t, store.recomputes, 1, "all subjects of the type recompute in a single batched store call")
	assert.Equal(t, appliedCompleteness{subjectType: "application", required: 1, filled: map[string]int{"app-1": 1, "app-2": 0}}, store.recomputes[0])
}

func TestSubjectIndexProjector_UnknownEvent_NoOp(t *testing.T) {
	store := &fakeIndexStore{}
	h := newHarness(t, projectorFakes{store: store})

	h.project("SomethingElse", time.Now(), map[string]any{"id": "x"})

	assert.Empty(t, store.upserts)
	assert.Empty(t, store.deletes)
	assert.Empty(t, store.changes)
	assert.Empty(t, store.recomputes)
}
