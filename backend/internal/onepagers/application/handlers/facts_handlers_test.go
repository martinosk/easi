package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"easi/backend/internal/onepagers/application/commands"
	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/aggregates"
	"easi/backend/internal/onepagers/domain/valueobjects"
	"easi/backend/internal/onepagers/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeConfigReader struct {
	record *readmodels.ConfigurationRecord
	err    error
}

func (f *fakeConfigReader) GetBySubjectType(_ context.Context, _ string) (*readmodels.ConfigurationRecord, error) {
	return f.record, f.err
}

type fakeFactsLookup struct {
	ids map[string]string
	err error
}

func newFakeFactsLookup() *fakeFactsLookup {
	return &fakeFactsLookup{ids: make(map[string]string)}
}

func (f *fakeFactsLookup) FactsIDForSubject(_ context.Context, subject readmodels.SubjectKey) (string, error) {
	return f.ids[subject.SubjectType+"/"+subject.SubjectID], f.err
}

type fakeSubjects struct {
	exists bool
	err    error
	calls  int
}

func (f *fakeSubjects) SubjectExists(_ context.Context, _, _ string) (bool, error) {
	f.calls++
	return f.exists, f.err
}

func configRecordWith(fields ...readmodels.CustomFieldRecord) *readmodels.ConfigurationRecord {
	return &readmodels.ConfigurationRecord{
		ID:          uuid.New().String(),
		TenantID:    "tenant-123",
		SubjectType: "application",
		Document:    readmodels.ConfigurationDocument{CustomFields: fields},
	}
}

func activeField(fieldType string) readmodels.CustomFieldRecord {
	return readmodels.CustomFieldRecord{
		ID:     uuid.New().String(),
		Name:   "Field under test",
		Type:   fieldType,
		Active: true,
	}
}

func envelope(t *testing.T, valueType, rawValue string) valueobjects.ValueEnvelope {
	t.Helper()
	return valueobjects.ValueEnvelope{Type: valueType, Version: 1, Value: json.RawMessage(rawValue)}
}

type factsTestEnv struct {
	repo     *repositories.OnePagerFactsRepository
	configs  *fakeConfigReader
	lookup   *fakeFactsLookup
	subjects *fakeSubjects
}

func newFactsTestEnv(config *readmodels.ConfigurationRecord) *factsTestEnv {
	return &factsTestEnv{
		repo:     repositories.NewOnePagerFactsRepository(newInMemoryEventStore()),
		configs:  &fakeConfigReader{record: config},
		lookup:   newFakeFactsLookup(),
		subjects: &fakeSubjects{exists: true},
	}
}

func (env *factsTestEnv) recordHandler() cqrs.CommandHandler {
	return NewRecordFieldValueHandler(env.repo, env.configs, env.lookup, env.subjects)
}

func (env *factsTestEnv) clearHandler() cqrs.CommandHandler {
	return NewClearFieldValueHandler(env.repo, env.configs, env.lookup)
}

func subjectField(fieldID string) commands.FactsSubjectField {
	return commands.FactsSubjectField{
		TenantID:    "tenant-123",
		SubjectType: "application",
		SubjectID:   "app-1",
		FieldID:     fieldID,
		ModifiedBy:  "architect@example.com",
	}
}

func recordCommand(fieldID string, value valueobjects.ValueEnvelope) *commands.RecordFieldValue {
	return &commands.RecordFieldValue{FactsSubjectField: subjectField(fieldID), Value: value}
}

func recordInitialValue(t *testing.T, env *factsTestEnv, fieldID, payload string) string {
	t.Helper()
	first, err := env.recordHandler().Handle(context.Background(), recordCommand(fieldID, envelope(t, "text", payload)))
	require.NoError(t, err)
	env.lookup.ids["application/app-1"] = first.CreatedID
	return first.CreatedID
}

func TestRecordFieldValueHandler_CreatesFactsOnFirstWrite(t *testing.T) {
	field := activeField("text")
	env := newFactsTestEnv(configRecordWith(field))

	result, err := env.recordHandler().Handle(context.Background(), recordCommand(field.ID, envelope(t, "text", `"Runs on shared Kubernetes cluster"`)))

	require.NoError(t, err)
	require.NotEmpty(t, result.CreatedID)
	assert.Equal(t, 1, env.subjects.calls)

	facts, err := env.repo.GetByID(context.Background(), result.CreatedID)
	require.NoError(t, err)
	assert.Equal(t, "application", facts.SubjectRef().SubjectType().Value())
	assert.Equal(t, "app-1", facts.SubjectRef().SubjectID())
	require.Len(t, facts.Values(), 1)
}

func TestRecordFieldValueHandler_AcceptsEveryFieldType(t *testing.T) {
	optionID := uuid.New().String()
	cases := []struct {
		fieldType string
		options   []readmodels.OptionRecord
		payload   string
	}{
		{"text", nil, `"Runs on shared Kubernetes cluster"`},
		{"number", nil, `42.5`},
		{"date", nil, `"2026-03-01"`},
		{"link", nil, `{"label":"MSA","url":"https://contracts.example.com"}`},
		{"selection", []readmodels.OptionRecord{{ID: optionID, Label: "Tier 1", Active: true}}, fmt.Sprintf(`{"optionId":%q}`, optionID)},
		{"contact-person", nil, `{"name":"A. Larsen","email":"al@ext.example","company":"Ext ApS"}`},
	}

	for _, tc := range cases {
		t.Run(tc.fieldType, func(t *testing.T) {
			field := activeField(tc.fieldType)
			field.Options = tc.options
			env := newFactsTestEnv(configRecordWith(field))

			result, err := env.recordHandler().Handle(context.Background(), recordCommand(field.ID, envelope(t, tc.fieldType, tc.payload)))

			require.NoError(t, err)
			assert.NotEmpty(t, result.CreatedID)
		})
	}
}

func TestRecordFieldValueHandler_ReusesExistingAggregate(t *testing.T) {
	fieldA := activeField("text")
	fieldB := activeField("number")
	env := newFactsTestEnv(configRecordWith(fieldA, fieldB))
	ctx := context.Background()

	first, err := env.recordHandler().Handle(ctx, recordCommand(fieldA.ID, envelope(t, "text", `"v"`)))
	require.NoError(t, err)
	env.lookup.ids["application/app-1"] = first.CreatedID
	env.subjects.exists = false

	second, err := env.recordHandler().Handle(ctx, recordCommand(fieldB.ID, envelope(t, "number", `42.5`)))

	require.NoError(t, err)
	assert.Equal(t, first.CreatedID, second.CreatedID)
	assert.Equal(t, 1, env.subjects.calls)

	facts, err := env.repo.GetByID(ctx, first.CreatedID)
	require.NoError(t, err)
	assert.Len(t, facts.Values(), 2)
}

func TestRecordFieldValueHandler_SuppressesNoOpWrite(t *testing.T) {
	field := activeField("text")
	env := newFactsTestEnv(configRecordWith(field))
	ctx := context.Background()

	factsID := recordInitialValue(t, env, field.ID, `"same"`)

	_, err := env.recordHandler().Handle(ctx, recordCommand(field.ID, envelope(t, "text", `"same"`)))
	require.NoError(t, err)

	facts, err := env.repo.GetByID(ctx, factsID)
	require.NoError(t, err)
	assert.Equal(t, 1, facts.Version())
}

func TestRecordFieldValueHandler_DefinitionValidation(t *testing.T) {
	optionID := uuid.New().String()
	retiredField := activeField("text")
	retiredField.Active = false
	dateField := activeField("date")
	selectionField := activeField("selection")
	selectionField.Options = []readmodels.OptionRecord{
		{ID: optionID, Label: "Tier 1", Active: false},
	}

	cases := []struct {
		name    string
		config  *readmodels.ConfigurationRecord
		fieldID string
		value   valueobjects.ValueEnvelope
		wantErr error
	}{
		{"no configuration", nil, uuid.New().String(), envelope(t, "text", `"v"`), ErrFieldNotDefined},
		{"unknown field", configRecordWith(dateField), uuid.New().String(), envelope(t, "text", `"v"`), ErrFieldNotDefined},
		{"retired field", configRecordWith(retiredField), retiredField.ID, envelope(t, "text", `"v"`), ErrFieldDefinitionRetired},
		{"type mismatch", configRecordWith(dateField), dateField.ID, envelope(t, "text", `"v"`), ErrValueTypeMismatch},
		{"option not defined", configRecordWith(selectionField), selectionField.ID, envelope(t, "selection", fmt.Sprintf(`{"optionId":%q}`, uuid.New().String())), ErrOptionNotDefined},
		{"retired option", configRecordWith(selectionField), selectionField.ID, envelope(t, "selection", fmt.Sprintf(`{"optionId":%q}`, optionID)), ErrOptionRetired},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := newFactsTestEnv(tc.config)

			_, err := env.recordHandler().Handle(context.Background(), recordCommand(tc.fieldID, tc.value))

			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestRecordFieldValueHandler_RejectsInvalidValuePayload(t *testing.T) {
	field := activeField("text")
	env := newFactsTestEnv(configRecordWith(field))

	_, err := env.recordHandler().Handle(context.Background(), recordCommand(field.ID, envelope(t, "text", `"   "`)))

	assert.ErrorIs(t, err, valueobjects.ErrTextValueEmpty)
}

func TestRecordFieldValueHandler_RejectsMissingSubject(t *testing.T) {
	field := activeField("text")
	env := newFactsTestEnv(configRecordWith(field))
	env.subjects.exists = false

	_, err := env.recordHandler().Handle(context.Background(), recordCommand(field.ID, envelope(t, "text", `"v"`)))

	assert.ErrorIs(t, err, ErrSubjectNotFound)
}

func TestRecordFieldValueHandler_RejectsArchivedFacts(t *testing.T) {
	field := activeField("text")
	env := newFactsTestEnv(configRecordWith(field))
	ctx := context.Background()

	factsID := recordInitialValue(t, env, field.ID, `"v"`)

	_, err := NewArchiveOnePagerFactsHandler(env.repo).Handle(ctx, &commands.ArchiveOnePagerFacts{
		FactsID: factsID,
		Reason:  "subject deleted",
	})
	require.NoError(t, err)

	_, err = env.recordHandler().Handle(ctx, recordCommand(field.ID, envelope(t, "text", `"new"`)))
	assert.ErrorIs(t, err, aggregates.ErrFactsArchived)
}

func TestRecordFieldValueHandler_RejectsWrongCommandType(t *testing.T) {
	env := newFactsTestEnv(nil)
	_, err := env.recordHandler().Handle(context.Background(), &commands.CreateOnePagerConfiguration{})
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func clearCommand(fieldID string) *commands.ClearFieldValue {
	return &commands.ClearFieldValue{FactsSubjectField: subjectField(fieldID)}
}

func TestClearFieldValueHandler_ClearsRecordedValue(t *testing.T) {
	field := activeField("text")
	env := newFactsTestEnv(configRecordWith(field))
	ctx := context.Background()

	factsID := recordInitialValue(t, env, field.ID, `"v"`)

	_, err := env.clearHandler().Handle(ctx, clearCommand(field.ID))
	require.NoError(t, err)

	facts, err := env.repo.GetByID(ctx, factsID)
	require.NoError(t, err)
	assert.Empty(t, facts.Values())
}

func TestClearFieldValueHandler_NoFactsIsNoOp(t *testing.T) {
	field := activeField("text")
	env := newFactsTestEnv(configRecordWith(field))

	_, err := env.clearHandler().Handle(context.Background(), clearCommand(field.ID))

	assert.NoError(t, err)
}

func TestClearFieldValueHandler_RejectsUnknownField(t *testing.T) {
	env := newFactsTestEnv(configRecordWith(activeField("text")))

	_, err := env.clearHandler().Handle(context.Background(), clearCommand(uuid.New().String()))

	assert.ErrorIs(t, err, ErrFieldNotDefined)
}

func TestClearFieldValueHandler_RejectsWrongCommandType(t *testing.T) {
	env := newFactsTestEnv(nil)
	_, err := env.clearHandler().Handle(context.Background(), &commands.CreateOnePagerConfiguration{})
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestArchiveOnePagerFactsHandler_RejectsWrongCommandType(t *testing.T) {
	env := newFactsTestEnv(nil)
	_, err := NewArchiveOnePagerFactsHandler(env.repo).Handle(context.Background(), &commands.CreateOnePagerConfiguration{})
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestArchiveOnePagerFactsHandler_IsIdempotent(t *testing.T) {
	field := activeField("text")
	env := newFactsTestEnv(configRecordWith(field))
	ctx := context.Background()

	first, err := env.recordHandler().Handle(ctx, recordCommand(field.ID, envelope(t, "text", `"v"`)))
	require.NoError(t, err)

	archive := NewArchiveOnePagerFactsHandler(env.repo)
	_, err = archive.Handle(ctx, &commands.ArchiveOnePagerFacts{FactsID: first.CreatedID, Reason: "subject deleted"})
	require.NoError(t, err)
	_, err = archive.Handle(ctx, &commands.ArchiveOnePagerFacts{FactsID: first.CreatedID, Reason: "subject deleted"})
	require.NoError(t, err)

	facts, err := env.repo.GetByID(ctx, first.CreatedID)
	require.NoError(t, err)
	assert.True(t, facts.IsArchived())
	assert.Equal(t, 2, facts.Version())
}
