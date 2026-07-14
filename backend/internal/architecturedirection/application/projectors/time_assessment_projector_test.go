package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/events"
	"easi/backend/internal/architecturedirection/domain/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockTimeAssessmentStore struct {
	upserts   []readmodels.UpsertTimeAssessmentParams
	deletedID []string
	upsertErr error
	deleteErr error
}

func (m *mockTimeAssessmentStore) UpsertCurrent(_ context.Context, p readmodels.UpsertTimeAssessmentParams) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.upserts = append(m.upserts, p)
	return nil
}

func (m *mockTimeAssessmentStore) Delete(_ context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedID = append(m.deletedID, id)
	return nil
}

func projectTimeAssessmentEvent(t *testing.T, projector *TimeAssessmentProjector, eventType string, payload map[string]interface{}) error {
	t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	return projector.ProjectEvent(context.Background(), eventType, data)
}

func TestTimeAssessmentProjector_Recorded_UpsertsRow(t *testing.T) {
	store := &mockTimeAssessmentStore{}
	projector := NewTimeAssessmentProjector(store)

	id := uuid.New().String()
	capID := uuid.New().String()
	compID := uuid.New().String()
	realizationID := uuid.New().String()
	evt := events.NewTimeAssessmentRecorded(events.TimeAssessmentRecordedFields{
		ID:            id,
		CapabilityID:  capID,
		ComponentID:   compID,
		RealizationID: realizationID,
		Grade:         valueobjects.TimeGradeMigrate,
		Rationale:     "carve-out candidate",
		AssessedBy:    "a@example.com",
	})

	require.NoError(t, projectTimeAssessmentEvent(t, projector, evt.EventType(), evt.EventData()))

	require.Len(t, store.upserts, 1)
	assert.Equal(t, id, store.upserts[0].ID)
	assert.Equal(t, capID, store.upserts[0].CapabilityID)
	assert.Equal(t, compID, store.upserts[0].ComponentID)
	assert.Equal(t, realizationID, store.upserts[0].RealizationID)
	assert.Equal(t, valueobjects.TimeGradeMigrate, store.upserts[0].Grade)
	assert.Equal(t, "carve-out candidate", store.upserts[0].Rationale)
	assert.Equal(t, "a@example.com", store.upserts[0].AssessedBy)
	assert.False(t, store.upserts[0].AssessedAt.IsZero())
}

func TestTimeAssessmentProjector_Removed_DeletesRow(t *testing.T) {
	store := &mockTimeAssessmentStore{}
	projector := NewTimeAssessmentProjector(store)

	id := uuid.New().String()
	evt := events.NewTimeAssessmentRemoved(events.TimeAssessmentRemovedFields{
		ID:           id,
		CapabilityID: uuid.New().String(),
		ComponentID:  uuid.New().String(),
		RemovedBy:    "a@example.com",
	})

	require.NoError(t, projectTimeAssessmentEvent(t, projector, evt.EventType(), evt.EventData()))

	assert.Equal(t, []string{id}, store.deletedID)
}

func TestTimeAssessmentProjector_UnknownEvent_NoOp(t *testing.T) {
	store := &mockTimeAssessmentStore{}
	projector := NewTimeAssessmentProjector(store)

	err := projector.ProjectEvent(context.Background(), "SomeUnrelatedEvent", []byte(`{}`))

	assert.NoError(t, err)
	assert.Empty(t, store.upserts)
	assert.Empty(t, store.deletedID)
}

func TestTimeAssessmentProjector_StoreErrors_Propagate(t *testing.T) {
	recorded := events.NewTimeAssessmentRecorded(events.TimeAssessmentRecordedFields{
		ID:           uuid.New().String(),
		CapabilityID: uuid.New().String(),
		ComponentID:  uuid.New().String(),
		Grade:        valueobjects.TimeGradeInvest,
		AssessedBy:   "a@example.com",
	})
	removed := events.NewTimeAssessmentRemoved(events.TimeAssessmentRemovedFields{
		ID:           uuid.New().String(),
		CapabilityID: uuid.New().String(),
		ComponentID:  uuid.New().String(),
		RemovedBy:    "a@example.com",
	})

	cases := []struct {
		name      string
		store     *mockTimeAssessmentStore
		eventType string
		payload   map[string]interface{}
	}{
		{"upsert error on recorded", &mockTimeAssessmentStore{upsertErr: errors.New("db")}, recorded.EventType(), recorded.EventData()},
		{"delete error on removed", &mockTimeAssessmentStore{deleteErr: errors.New("db")}, removed.EventType(), removed.EventData()},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			projector := NewTimeAssessmentProjector(tc.store)
			err := projectTimeAssessmentEvent(t, projector, tc.eventType, tc.payload)
			assert.Error(t, err)
		})
	}
}
