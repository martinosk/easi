package projectors

import (
	"context"
	"testing"

	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/events"
	"easi/backend/internal/onepagers/domain/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeFactsStore struct {
	upserts  []readmodels.FactRecord
	clears   []readmodels.ClearFactParams
	deletes  [][2]string
	failWith error
}

func (s *fakeFactsStore) Upsert(_ context.Context, record readmodels.FactRecord) error {
	s.upserts = append(s.upserts, record)
	return s.failWith
}

func (s *fakeFactsStore) Clear(_ context.Context, params readmodels.ClearFactParams) error {
	s.clears = append(s.clears, params)
	return s.failWith
}

func (s *fakeFactsStore) DeleteForSubject(_ context.Context, subject readmodels.SubjectKey) error {
	s.deletes = append(s.deletes, [2]string{subject.SubjectType, subject.SubjectID})
	return s.failWith
}

func factsParams(factsID string) events.ModifyFactsParams {
	return events.ModifyFactsParams{
		FactsID:     factsID,
		TenantID:    "tenant-123",
		SubjectType: "application",
		SubjectID:   "app-1",
		Version:     1,
		ModifiedBy:  "architect@example.com",
	}
}

func TestFactsProjector_UpsertsOnFieldValueRecorded(t *testing.T) {
	store := &fakeFactsStore{}
	projector := NewOnePagerFactsProjector(store)
	factsID := uuid.New().String()
	fieldID := uuid.New().String()
	value, err := valueobjects.NewTextValue("Runs on shared Kubernetes cluster")
	require.NoError(t, err)
	envelope, err := valueobjects.NewValueEnvelope(value)
	require.NoError(t, err)
	event := events.NewFieldValueRecorded(factsParams(factsID), fieldID, envelope)

	require.NoError(t, projector.Handle(context.Background(), event))

	require.Len(t, store.upserts, 1)
	record := store.upserts[0]
	assert.Equal(t, factsID, record.FactsID)
	assert.Equal(t, "tenant-123", record.TenantID)
	assert.Equal(t, "application", record.SubjectType)
	assert.Equal(t, "app-1", record.SubjectID)
	assert.Equal(t, fieldID, record.FieldID)
	assert.Equal(t, "text", record.ValueType)
	assert.Equal(t, "Runs on shared Kubernetes cluster", record.DisplayText)
	assert.Equal(t, "architect@example.com", record.ModifiedBy)
	require.NotNil(t, record.Value)
	decoded, err := valueobjects.FieldValueFromEnvelope(*record.Value)
	require.NoError(t, err)
	assert.True(t, decoded.Equals(value))
}

func TestFactsProjector_ClearsOnFieldValueCleared(t *testing.T) {
	store := &fakeFactsStore{}
	projector := NewOnePagerFactsProjector(store)
	fieldID := uuid.New().String()
	event := events.NewFieldValueCleared(factsParams(uuid.New().String()), fieldID)

	require.NoError(t, projector.Handle(context.Background(), event))

	require.Len(t, store.clears, 1)
	assert.Equal(t, "application", store.clears[0].SubjectType)
	assert.Equal(t, "app-1", store.clears[0].SubjectID)
	assert.Equal(t, fieldID, store.clears[0].FieldID)
}

func TestFactsProjector_DeletesRowsOnArchived(t *testing.T) {
	store := &fakeFactsStore{}
	projector := NewOnePagerFactsProjector(store)
	event := events.NewOnePagerFactsArchived(factsParams(uuid.New().String()), "subject deleted")

	require.NoError(t, projector.Handle(context.Background(), event))

	require.Len(t, store.deletes, 1)
	assert.Equal(t, [2]string{"application", "app-1"}, store.deletes[0])
}

func TestFactsProjector_IgnoresUnknownEventTypes(t *testing.T) {
	store := &fakeFactsStore{}
	projector := NewOnePagerFactsProjector(store)

	require.NoError(t, projector.ProjectEvent(context.Background(), "SomethingElse", []byte(`{}`)))

	assert.Empty(t, store.upserts)
	assert.Empty(t, store.clears)
	assert.Empty(t, store.deletes)
}
