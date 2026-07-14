package projectors

import (
	"context"
	"testing"

	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/aggregates"
	"easi/backend/internal/onepagers/domain/valueobjects"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type inMemoryStore struct {
	records map[string]readmodels.ConfigurationRecord
}

func newInMemoryStore() *inMemoryStore {
	return &inMemoryStore{records: make(map[string]readmodels.ConfigurationRecord)}
}

func (s *inMemoryStore) Insert(_ context.Context, record readmodels.ConfigurationRecord) error {
	s.records[record.ID] = record
	return nil
}

func (s *inMemoryStore) GetByID(_ context.Context, id string) (*readmodels.ConfigurationRecord, error) {
	record, ok := s.records[id]
	if !ok {
		return nil, nil
	}
	return &record, nil
}

func (s *inMemoryStore) Update(_ context.Context, params readmodels.UpdateParams) error {
	record, ok := s.records[params.ID]
	if !ok {
		return nil
	}
	record.Document = params.Document
	record.Version = params.Version
	record.ModifiedAt = params.ModifiedAt
	record.ModifiedBy = params.ModifiedBy
	s.records[params.ID] = record
	return nil
}

func adminEmail(t *testing.T) valueobjects.UserEmail {
	t.Helper()
	email, err := valueobjects.NewUserEmail("admin@example.com")
	require.NoError(t, err)
	return email
}

func newAggregate(t *testing.T) *aggregates.OnePagerConfiguration {
	t.Helper()
	tenantID, err := sharedvo.NewTenantID("tenant-123")
	require.NoError(t, err)
	subjectType, err := valueobjects.NewSubjectType("application")
	require.NoError(t, err)
	config, err := aggregates.NewOnePagerConfiguration(tenantID, subjectType, adminEmail(t))
	require.NoError(t, err)
	return config
}

func project(t *testing.T, store *inMemoryStore, config *aggregates.OnePagerConfiguration) {
	t.Helper()
	projector := NewOnePagerConfigurationProjector(store)
	for _, event := range config.GetUncommittedChanges() {
		require.NoError(t, projector.Handle(context.Background(), event))
	}
	config.MarkChangesAsCommitted()
}

func mustName(t *testing.T, v string) valueobjects.FieldName {
	t.Helper()
	name, err := valueobjects.NewFieldName(v)
	require.NoError(t, err)
	return name
}

func mustType(t *testing.T, v string) valueobjects.FieldType {
	t.Helper()
	ft, err := valueobjects.NewFieldType(v)
	require.NoError(t, err)
	return ft
}

func mustLabel(t *testing.T, v string) valueobjects.OptionLabel {
	t.Helper()
	label, err := valueobjects.NewOptionLabel(v)
	require.NoError(t, err)
	return label
}

func TestProjector_CreatedInsertsDefaultDocument(t *testing.T) {
	store := newInMemoryStore()
	config := newAggregate(t)

	project(t, store, config)

	record, err := store.GetByID(context.Background(), config.ID())
	require.NoError(t, err)
	require.NotNil(t, record)
	assert.Equal(t, "tenant-123", record.TenantID)
	assert.Equal(t, "application", record.SubjectType)
	assert.Equal(t, 1, record.Version)
	assert.Equal(t, "admin@example.com", record.ModifiedBy)
	assert.Empty(t, record.Document.CustomFields)
	assert.Equal(t, []readmodels.FieldRefRecord{
		{Kind: "builtIn", ID: "name"},
		{Kind: "builtIn", ID: "description"},
		{Kind: "builtIn", ID: "experts"},
	}, record.Document.DisplayOrder)
}

func TestProjector_ProjectsFullCustomFieldLifecycle(t *testing.T) {
	store := newInMemoryStore()
	config := newAggregate(t)
	email := adminEmail(t)

	fieldID, err := config.DefineCustomField(aggregates.DefineCustomFieldParams{
		Name:         mustName(t, "Hosting model"),
		Type:         mustType(t, "selection"),
		OptionLabels: []valueobjects.OptionLabel{mustLabel(t, "On-prem"), mustLabel(t, "Cloud")},
	}, email)
	require.NoError(t, err)
	project(t, store, config)

	record, err := store.GetByID(context.Background(), config.ID())
	require.NoError(t, err)
	require.Len(t, record.Document.CustomFields, 1)
	field := record.Document.CustomFields[0]
	assert.Equal(t, fieldID.Value(), field.ID)
	assert.Equal(t, "Hosting model", field.Name)
	assert.Equal(t, "selection", field.Type)
	assert.True(t, field.Active)
	require.Len(t, field.Options, 2)
	assert.Equal(t, readmodels.FieldRefRecord{Kind: "custom", ID: fieldID.Value()}, record.Document.DisplayOrder[3])

	help, err := valueobjects.NewHelpText("Where it runs")
	require.NoError(t, err)
	require.NoError(t, config.RenameCustomField(aggregates.RenameCustomFieldParams{
		FieldID:  fieldID,
		Name:     mustName(t, "Deployment model"),
		HelpText: help,
	}, email))
	require.NoError(t, config.ChangeCustomFieldRequirement(fieldID, true, email))
	optionID, err := config.AddSelectionOption(fieldID, mustLabel(t, "Hybrid"), email)
	require.NoError(t, err)
	onPrem, found := config.CustomFieldByID(fieldID)
	require.True(t, found)
	require.NoError(t, config.RetireSelectionOption(fieldID, onPrem.Options()[0].ID(), email))
	project(t, store, config)

	record, err = store.GetByID(context.Background(), config.ID())
	require.NoError(t, err)
	field = record.Document.CustomFields[0]
	assert.Equal(t, "Deployment model", field.Name)
	assert.Equal(t, "Where it runs", field.HelpText)
	assert.True(t, field.Required)
	require.Len(t, field.Options, 3)
	assert.False(t, field.Options[0].Active)
	assert.Equal(t, optionID.Value(), field.Options[2].ID)
	assert.Equal(t, "Hybrid", field.Options[2].Label)

	require.NoError(t, config.RetireCustomField(fieldID, email))
	project(t, store, config)
	record, err = store.GetByID(context.Background(), config.ID())
	require.NoError(t, err)
	assert.False(t, record.Document.CustomFields[0].Active)
	assert.Len(t, record.Document.DisplayOrder, 3)

	require.NoError(t, config.ReactivateCustomField(fieldID, email))
	project(t, store, config)
	record, err = store.GetByID(context.Background(), config.ID())
	require.NoError(t, err)
	assert.True(t, record.Document.CustomFields[0].Active)
	assert.Len(t, record.Document.DisplayOrder, 4)
	assert.Equal(t, config.Version(), record.Version)
}

func floatPtr(v float64) *float64 {
	return &v
}

func TestProjector_ProjectsNumberFieldBoundsChanged(t *testing.T) {
	store := newInMemoryStore()
	config := newAggregate(t)
	email := adminEmail(t)

	fieldID, err := config.DefineCustomField(aggregates.DefineCustomFieldParams{
		Name: mustName(t, "Maturity score"),
		Type: mustType(t, "number"),
	}, email)
	require.NoError(t, err)
	project(t, store, config)

	require.NoError(t, config.SetNumberFieldBounds(fieldID, floatPtr(0), floatPtr(5), email))
	project(t, store, config)

	record, err := store.GetByID(context.Background(), config.ID())
	require.NoError(t, err)
	field := record.Document.CustomFields[0]
	require.NotNil(t, field.Min)
	require.NotNil(t, field.Max)
	assert.Equal(t, 0.0, *field.Min)
	assert.Equal(t, 5.0, *field.Max)

	require.NoError(t, config.SetNumberFieldBounds(fieldID, floatPtr(0), nil, email))
	project(t, store, config)

	record, err = store.GetByID(context.Background(), config.ID())
	require.NoError(t, err)
	field = record.Document.CustomFields[0]
	require.NotNil(t, field.Min)
	assert.Equal(t, 0.0, *field.Min)
	assert.Nil(t, field.Max)
}

func TestProjector_ProjectsBuiltInInclusionAndReorder(t *testing.T) {
	store := newInMemoryStore()
	config := newAggregate(t)
	email := adminEmail(t)

	require.NoError(t, config.ExcludeBuiltInField("experts", email))
	project(t, store, config)
	record, err := store.GetByID(context.Background(), config.ID())
	require.NoError(t, err)
	assert.Len(t, record.Document.DisplayOrder, 2)

	require.NoError(t, config.IncludeBuiltInField("experts", email))
	project(t, store, config)
	record, err = store.GetByID(context.Background(), config.ID())
	require.NoError(t, err)
	assert.Equal(t, readmodels.FieldRefRecord{Kind: "builtIn", ID: "experts"}, record.Document.DisplayOrder[2])

	order := config.DisplayOrder()
	reversed := make([]valueobjects.FieldRef, len(order))
	for i, ref := range order {
		reversed[len(order)-1-i] = ref
	}
	require.NoError(t, config.ReorderFields(reversed, email))
	project(t, store, config)
	record, err = store.GetByID(context.Background(), config.ID())
	require.NoError(t, err)
	assert.Equal(t, readmodels.FieldRefRecord{Kind: "builtIn", ID: "experts"}, record.Document.DisplayOrder[0])
	assert.Equal(t, config.Version(), record.Version)
}

func requiredBuiltIn(record *readmodels.ConfigurationRecord, entryID string) bool {
	for _, builtIn := range record.Document.BuiltInFields {
		if builtIn.ID == entryID {
			return builtIn.Required
		}
	}
	return false
}

func TestProjector_ProjectsBuiltInFieldRequirementChanged(t *testing.T) {
	store := newInMemoryStore()
	config := newAggregate(t)
	email := adminEmail(t)

	require.NoError(t, config.ChangeBuiltInFieldRequirement("experts", true, email))
	project(t, store, config)
	record, err := store.GetByID(context.Background(), config.ID())
	require.NoError(t, err)
	assert.True(t, requiredBuiltIn(record, "experts"))

	require.NoError(t, config.ChangeBuiltInFieldRequirement("experts", false, email))
	project(t, store, config)
	record, err = store.GetByID(context.Background(), config.ID())
	require.NoError(t, err)
	assert.False(t, requiredBuiltIn(record, "experts"))
}

func TestProjector_BuiltInRequirementSurvivesExcludeAndReinclude(t *testing.T) {
	store := newInMemoryStore()
	config := newAggregate(t)
	email := adminEmail(t)

	require.NoError(t, config.ChangeBuiltInFieldRequirement("experts", true, email))
	require.NoError(t, config.ExcludeBuiltInField("experts", email))
	require.NoError(t, config.IncludeBuiltInField("experts", email))
	project(t, store, config)

	record, err := store.GetByID(context.Background(), config.ID())
	require.NoError(t, err)
	assert.True(t, requiredBuiltIn(record, "experts"), "required flag survives exclude and re-include in the projection")
}

func TestProjector_UnknownEventTypeIsIgnored(t *testing.T) {
	store := newInMemoryStore()
	projector := NewOnePagerConfigurationProjector(store)

	err := projector.ProjectEvent(context.Background(), "SomethingElseHappened", []byte(`{}`))

	assert.NoError(t, err)
}

func TestProjector_MissingRecordIsSkipped(t *testing.T) {
	store := newInMemoryStore()
	config := newAggregate(t)
	config.MarkChangesAsCommitted()
	_, err := config.DefineCustomField(aggregates.DefineCustomFieldParams{
		Name: mustName(t, "Orphan"),
		Type: mustType(t, "text"),
	}, adminEmail(t))
	require.NoError(t, err)

	projector := NewOnePagerConfigurationProjector(store)
	var projectErr error
	for _, event := range config.GetUncommittedChanges() {
		projectErr = projector.Handle(context.Background(), event)
	}

	assert.NoError(t, projectErr)
}
