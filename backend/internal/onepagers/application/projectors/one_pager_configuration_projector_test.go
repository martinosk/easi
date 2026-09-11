package projectors

import (
	"context"
	"testing"

	"easi/backend/internal/onepagers/application/readmodels"
	"easi/backend/internal/onepagers/domain/aggregates"
	"easi/backend/internal/onepagers/domain/events"
	"easi/backend/internal/onepagers/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
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

func TestProjector_ProjectsCustomFieldInclusionAndRequirement(t *testing.T) {
	store := newInMemoryStore()
	config := newAggregate(t)
	email := adminEmail(t)
	fieldID := valueobjects.NewFieldID()

	require.NoError(t, config.IncludeCustomField(fieldID, email))
	project(t, store, config)

	record, err := store.GetByID(context.Background(), config.ID())
	require.NoError(t, err)
	assert.Equal(t, []readmodels.FieldRequirementRecord{{ID: fieldID.Value(), Required: false}}, record.Document.CustomFields)
	assert.Equal(t, readmodels.FieldRefRecord{Kind: "custom", ID: fieldID.Value()}, record.Document.DisplayOrder[3])

	require.NoError(t, config.ChangeCustomFieldRequirement(fieldID, true, email))
	require.NoError(t, config.ExcludeCustomField(fieldID, email))
	project(t, store, config)

	record, err = store.GetByID(context.Background(), config.ID())
	require.NoError(t, err)
	assert.True(t, record.Document.CustomFieldRequired(fieldID.Value()))
	assert.Len(t, record.Document.DisplayOrder, 3)

	require.NoError(t, config.IncludeCustomField(fieldID, email))
	project(t, store, config)
	record, err = store.GetByID(context.Background(), config.ID())
	require.NoError(t, err)
	assert.Equal(t, []string{fieldID.Value()}, record.Document.RequiredCustomFieldIDs())
	assert.Len(t, record.Document.DisplayOrder, 4)
	assert.Equal(t, config.Version(), record.Version)
}

func legacyParams(configID string, version int) events.ModifyConfigurationParams {
	return events.ModifyConfigurationParams{ConfigID: configID, TenantID: "tenant-123", Version: version, ModifiedBy: "admin@example.com"}
}

func TestProjector_LegacySchemaEventsStillProjectInclusionAndRequirement(t *testing.T) {
	store := newInMemoryStore()
	config := newAggregate(t)
	project(t, store, config)
	projector := NewOnePagerConfigurationProjector(store)
	fieldID := valueobjects.NewFieldID().Value()
	legacy := []domain.DomainEvent{
		events.NewCustomFieldDefined(legacyParams(config.ID(), 2), events.CustomFieldData{FieldID: fieldID, Name: "Hosting", FieldType: "selection", Required: true, Options: []events.SelectionOptionData{{ID: valueobjects.NewOptionID().Value(), Label: "Cloud", Active: true}}}),
		events.NewCustomFieldRenamed(legacyParams(config.ID(), 3), events.FieldRenameData{FieldID: fieldID, NewName: "Deployment"}),
		events.NewSelectionOptionAdded(legacyParams(config.ID(), 4), fieldID, valueobjects.NewOptionID().Value(), "Hybrid"),
		events.NewNumberFieldBoundsChanged(legacyParams(config.ID(), 5), fieldID, nil, nil),
		events.NewCustomFieldRetired(legacyParams(config.ID(), 6), fieldID),
	}
	for _, event := range legacy {
		require.NoError(t, projector.Handle(context.Background(), event))
	}

	record, err := store.GetByID(context.Background(), config.ID())
	require.NoError(t, err)
	assert.Equal(t, 6, record.Version)
	assert.True(t, record.Document.CustomFieldRequired(fieldID))
	assert.Len(t, record.Document.DisplayOrder, 3, "retired legacy field leaves the display order")

	require.NoError(t, projector.Handle(context.Background(), events.NewCustomFieldReactivated(legacyParams(config.ID(), 7), fieldID)))
	record, err = store.GetByID(context.Background(), config.ID())
	require.NoError(t, err)
	assert.Equal(t, []string{fieldID}, record.Document.RequiredCustomFieldIDs())
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
	require.NoError(t, config.IncludeCustomField(valueobjects.NewFieldID(), adminEmail(t)))

	projector := NewOnePagerConfigurationProjector(store)
	var projectErr error
	for _, event := range config.GetUncommittedChanges() {
		projectErr = projector.Handle(context.Background(), event)
	}

	assert.NoError(t, projectErr)
}
