package aggregates

import (
	"testing"

	"easi/backend/internal/onepagers/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestFacts(t *testing.T) *OnePagerFacts {
	t.Helper()
	tenantID, err := sharedvo.NewTenantID("tenant-123")
	require.NoError(t, err)
	subjectRef, err := valueobjects.NewSubjectRef("application", uuid.New().String())
	require.NoError(t, err)
	return NewOnePagerFacts(tenantID, subjectRef)
}

func testEmail(t *testing.T) valueobjects.UserEmail {
	t.Helper()
	email, err := valueobjects.NewUserEmail("architect@example.com")
	require.NoError(t, err)
	return email
}

func textValue(t *testing.T, raw string) valueobjects.FieldValue {
	t.Helper()
	value, err := valueobjects.NewTextValue(raw)
	require.NoError(t, err)
	return value
}

func TestNewOnePagerFacts_HasIntrinsicIdentity(t *testing.T) {
	facts := newTestFacts(t)

	_, err := uuid.Parse(facts.ID())
	require.NoError(t, err)
	assert.NotEqual(t, facts.SubjectRef().SubjectID(), facts.ID())
	assert.False(t, facts.IsArchived())
	assert.Empty(t, facts.Values())
}

func TestRecordFieldValue_RecordsValueAndRaisesEvent(t *testing.T) {
	facts := newTestFacts(t)
	fieldID := valueobjects.NewFieldID()

	err := facts.RecordFieldValue(fieldID, textValue(t, "Runs on shared Kubernetes cluster"), testEmail(t))

	require.NoError(t, err)
	value, found := facts.Value(fieldID)
	require.True(t, found)
	assert.True(t, value.Equals(textValue(t, "Runs on shared Kubernetes cluster")))
	changes := facts.GetUncommittedChanges()
	require.Len(t, changes, 1)
	assert.Equal(t, "FieldValueRecorded", changes[0].EventType())
	assert.Equal(t, facts.ID(), changes[0].AggregateID())
}

func TestRecordFieldValue_ReplacesExistingValue(t *testing.T) {
	facts := newTestFacts(t)
	fieldID := valueobjects.NewFieldID()

	require.NoError(t, facts.RecordFieldValue(fieldID, textValue(t, "first"), testEmail(t)))
	require.NoError(t, facts.RecordFieldValue(fieldID, textValue(t, "second"), testEmail(t)))

	require.Len(t, facts.Values(), 1)
	value, found := facts.Value(fieldID)
	require.True(t, found)
	assert.True(t, value.Equals(textValue(t, "second")))
	assert.Len(t, facts.GetUncommittedChanges(), 2)
}

func TestRecordFieldValue_SuppressesNoOpWrites(t *testing.T) {
	facts := newTestFacts(t)
	fieldID := valueobjects.NewFieldID()

	require.NoError(t, facts.RecordFieldValue(fieldID, textValue(t, "same"), testEmail(t)))
	require.NoError(t, facts.RecordFieldValue(fieldID, textValue(t, "same"), testEmail(t)))

	assert.Len(t, facts.GetUncommittedChanges(), 1)
}

func TestRecordFieldValue_RejectsMissingFieldIDOrValue(t *testing.T) {
	facts := newTestFacts(t)

	err := facts.RecordFieldValue(valueobjects.FieldID{}, textValue(t, "x"), testEmail(t))
	assert.ErrorIs(t, err, ErrFieldIDRequired)

	err = facts.RecordFieldValue(valueobjects.NewFieldID(), nil, testEmail(t))
	assert.ErrorIs(t, err, ErrFieldValueRequired)
}

func TestClearFieldValue_RemovesValue(t *testing.T) {
	facts := newTestFacts(t)
	fieldID := valueobjects.NewFieldID()
	require.NoError(t, facts.RecordFieldValue(fieldID, textValue(t, "to be cleared"), testEmail(t)))

	require.NoError(t, facts.ClearFieldValue(fieldID, testEmail(t)))

	_, found := facts.Value(fieldID)
	assert.False(t, found)
	changes := facts.GetUncommittedChanges()
	require.Len(t, changes, 2)
	assert.Equal(t, "FieldValueCleared", changes[1].EventType())
}

func TestClearFieldValue_SuppressesClearingAbsentValue(t *testing.T) {
	facts := newTestFacts(t)

	require.NoError(t, facts.ClearFieldValue(valueobjects.NewFieldID(), testEmail(t)))

	assert.Empty(t, facts.GetUncommittedChanges())
}

func TestClearFieldValue_RejectsMissingFieldID(t *testing.T) {
	facts := newTestFacts(t)

	err := facts.ClearFieldValue(valueobjects.FieldID{}, testEmail(t))
	assert.ErrorIs(t, err, ErrFieldIDRequired)
}

func TestArchive_BlocksAllSubsequentWrites(t *testing.T) {
	facts := newTestFacts(t)
	fieldID := valueobjects.NewFieldID()
	require.NoError(t, facts.RecordFieldValue(fieldID, textValue(t, "kept"), testEmail(t)))

	require.NoError(t, facts.Archive("subject deleted"))

	assert.True(t, facts.IsArchived())
	assert.ErrorIs(t, facts.RecordFieldValue(fieldID, textValue(t, "new"), testEmail(t)), ErrFactsArchived)
	assert.ErrorIs(t, facts.ClearFieldValue(fieldID, testEmail(t)), ErrFactsArchived)
	changes := facts.GetUncommittedChanges()
	require.Len(t, changes, 2)
	assert.Equal(t, "OnePagerFactsArchived", changes[1].EventType())
}

func TestArchive_IsIdempotent(t *testing.T) {
	facts := newTestFacts(t)

	require.NoError(t, facts.Archive("subject deleted"))
	require.NoError(t, facts.Archive("subject deleted"))

	assert.Len(t, facts.GetUncommittedChanges(), 1)
}

func TestLoadOnePagerFactsFromHistory_RebuildsState(t *testing.T) {
	facts := newTestFacts(t)
	keptField := valueobjects.NewFieldID()
	clearedField := valueobjects.NewFieldID()
	require.NoError(t, facts.RecordFieldValue(keptField, textValue(t, "kept"), testEmail(t)))
	require.NoError(t, facts.RecordFieldValue(clearedField, textValue(t, "gone"), testEmail(t)))
	require.NoError(t, facts.ClearFieldValue(clearedField, testEmail(t)))

	history := make([]domain.DomainEvent, len(facts.GetUncommittedChanges()))
	copy(history, facts.GetUncommittedChanges())

	loaded, err := LoadOnePagerFactsFromHistory(history)
	require.NoError(t, err)
	assert.Equal(t, facts.ID(), loaded.ID())
	assert.Equal(t, facts.Version(), loaded.Version())
	assert.Equal(t, "tenant-123", loaded.TenantID().Value())
	assert.True(t, loaded.SubjectRef().Equals(facts.SubjectRef()))
	require.Len(t, loaded.Values(), 1)
	value, found := loaded.Value(keptField)
	require.True(t, found)
	assert.True(t, value.Equals(textValue(t, "kept")))
	assert.False(t, loaded.IsArchived())
}

func TestLoadOnePagerFactsFromHistory_RestoresArchivedFlag(t *testing.T) {
	facts := newTestFacts(t)
	require.NoError(t, facts.RecordFieldValue(valueobjects.NewFieldID(), textValue(t, "v"), testEmail(t)))
	require.NoError(t, facts.Archive("subject deleted"))

	loaded, err := LoadOnePagerFactsFromHistory(facts.GetUncommittedChanges())
	require.NoError(t, err)
	assert.True(t, loaded.IsArchived())
}
