package aggregates

import (
	"testing"

	"easi/backend/internal/onepagers/domain/events"
	"easi/backend/internal/onepagers/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func adminEmail(t *testing.T) valueobjects.UserEmail {
	t.Helper()
	email, err := valueobjects.NewUserEmail("admin@example.com")
	require.NoError(t, err)
	return email
}

func subjectType(t *testing.T, value string) valueobjects.SubjectType {
	t.Helper()
	st, err := valueobjects.NewSubjectType(value)
	require.NoError(t, err)
	return st
}

func newCommittedApplicationConfig(t *testing.T) *OnePagerConfiguration {
	t.Helper()
	tenantID, err := sharedvo.NewTenantID("tenant-123")
	require.NoError(t, err)
	config, err := NewOnePagerConfiguration(tenantID, subjectType(t, "application"), adminEmail(t))
	require.NoError(t, err)
	config.MarkChangesAsCommitted()
	return config
}

func includeField(t *testing.T, config *OnePagerConfiguration) valueobjects.FieldID {
	t.Helper()
	fieldID := valueobjects.NewFieldID()
	require.NoError(t, config.IncludeCustomField(fieldID, adminEmail(t)))
	return fieldID
}

func orderRefIDs(config *OnePagerConfiguration) []string {
	order := config.DisplayOrder()
	ids := make([]string, len(order))
	for i, ref := range order {
		ids[i] = string(ref.Kind()) + ":" + ref.RefID()
	}
	return ids
}

func lastEvent(t *testing.T, config *OnePagerConfiguration) domain.DomainEvent {
	t.Helper()
	changes := config.GetUncommittedChanges()
	require.NotEmpty(t, changes)
	return changes[len(changes)-1]
}

func modifyParams(config *OnePagerConfiguration) events.ModifyConfigurationParams {
	return events.ModifyConfigurationParams{ConfigID: config.ID(), TenantID: "tenant-123", Version: config.Version() + 1, ModifiedBy: "admin@example.com"}
}

func TestNewOnePagerConfiguration_DefaultsToFullCatalogInOrder(t *testing.T) {
	tenantID, err := sharedvo.NewTenantID("tenant-123")
	require.NoError(t, err)

	config, err := NewOnePagerConfiguration(tenantID, subjectType(t, "application"), adminEmail(t))

	require.NoError(t, err)
	assert.Equal(t, 1, config.Version())
	assert.Equal(t, "application", config.SubjectType().Value())
	assert.Equal(t, []string{"builtIn:name", "builtIn:description", "builtIn:experts"}, orderRefIDs(config))
	changes := config.GetUncommittedChanges()
	require.Len(t, changes, 1)
	created, ok := changes[0].(events.OnePagerConfigurationCreated)
	require.True(t, ok)
	assert.Equal(t, []string{"name", "description", "experts"}, created.BuiltIns)
}

func TestIncludeCustomField_AppendsToOrderAndRaisesEvent(t *testing.T) {
	config := newCommittedApplicationConfig(t)

	fieldID := includeField(t, config)

	assert.Equal(t, []string{"builtIn:name", "builtIn:description", "builtIn:experts", "custom:" + fieldID.Value()}, orderRefIDs(config))
	assert.True(t, config.IsCustomFieldIncluded(fieldID))
	assert.False(t, config.IsCustomFieldRequired(fieldID))
	included, ok := lastEvent(t, config).(events.CustomFieldIncluded)
	require.True(t, ok)
	assert.Equal(t, fieldID.Value(), included.FieldID)
	assert.Equal(t, 2, included.Version)
}

func TestIncludeCustomField_RejectsAlreadyIncluded(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := includeField(t, config)

	assert.ErrorIs(t, config.IncludeCustomField(fieldID, adminEmail(t)), ErrCustomFieldAlreadyIncluded)
}

func TestExcludeCustomField_LeavesDisplayOrderAndKeepsRequirementDormant(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := includeField(t, config)
	require.NoError(t, config.ChangeCustomFieldRequirement(fieldID, true, adminEmail(t)))

	require.NoError(t, config.ExcludeCustomField(fieldID, adminEmail(t)))

	assert.Equal(t, []string{"builtIn:name", "builtIn:description", "builtIn:experts"}, orderRefIDs(config))
	assert.False(t, config.IsCustomFieldIncluded(fieldID))
	_, ok := lastEvent(t, config).(events.CustomFieldExcluded)
	assert.True(t, ok)
	assert.ErrorIs(t, config.ExcludeCustomField(fieldID, adminEmail(t)), ErrCustomFieldNotIncluded)

	require.NoError(t, config.IncludeCustomField(fieldID, adminEmail(t)))
	assert.True(t, config.IsCustomFieldRequired(fieldID), "requirement survives exclude and re-include")
}

func TestChangeCustomFieldRequirement_RecordsFlagOnIncludedField(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := includeField(t, config)

	require.NoError(t, config.ChangeCustomFieldRequirement(fieldID, true, adminEmail(t)))

	assert.True(t, config.IsCustomFieldRequired(fieldID))
	changed, ok := lastEvent(t, config).(events.CustomFieldRequirementChanged)
	require.True(t, ok)
	assert.Equal(t, fieldID.Value(), changed.FieldID)
	assert.True(t, changed.Required)
	require.NoError(t, config.ChangeCustomFieldRequirement(fieldID, false, adminEmail(t)))
	assert.False(t, config.IsCustomFieldRequired(fieldID))
}

func TestChangeCustomFieldRequirement_RejectsFieldNotIncluded(t *testing.T) {
	config := newCommittedApplicationConfig(t)

	err := config.ChangeCustomFieldRequirement(valueobjects.NewFieldID(), true, adminEmail(t))

	assert.ErrorIs(t, err, ErrCustomFieldNotIncluded)
}

func TestExcludeBuiltInField_LeavesDisplayOrder(t *testing.T) {
	config := newCommittedApplicationConfig(t)

	require.NoError(t, config.ExcludeBuiltInField("experts", adminEmail(t)))

	assert.Equal(t, []string{"builtIn:name", "builtIn:description"}, orderRefIDs(config))
	assert.ErrorIs(t, config.ExcludeBuiltInField("experts", adminEmail(t)), ErrBuiltInFieldNotIncluded)
}

func TestIncludeBuiltInField_ReappearsAtEndOfOrder(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	require.NoError(t, config.ExcludeBuiltInField("name", adminEmail(t)))

	require.NoError(t, config.IncludeBuiltInField("name", adminEmail(t)))

	assert.Equal(t, []string{"builtIn:description", "builtIn:experts", "builtIn:name"}, orderRefIDs(config))
	assert.ErrorIs(t, config.IncludeBuiltInField("name", adminEmail(t)), ErrBuiltInFieldAlreadyIncluded)
	assert.ErrorIs(t, config.IncludeBuiltInField("no-such-entry", adminEmail(t)), ErrUnknownBuiltInField)
}

func TestChangeBuiltInFieldRequirement_RecordsFlagOnIncludedBuiltIn(t *testing.T) {
	config := newCommittedApplicationConfig(t)

	require.NoError(t, config.ChangeBuiltInFieldRequirement("experts", true, adminEmail(t)))

	assert.True(t, config.IsBuiltInRequired("experts"))
	changed, ok := lastEvent(t, config).(events.BuiltInFieldRequirementChanged)
	require.True(t, ok)
	assert.Equal(t, "experts", changed.EntryID)
	assert.True(t, changed.Required)
}

func TestChangeBuiltInFieldRequirement_RejectsExcludedOrUnknownBuiltIn(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	require.NoError(t, config.ExcludeBuiltInField("experts", adminEmail(t)))

	assert.ErrorIs(t, config.ChangeBuiltInFieldRequirement("experts", true, adminEmail(t)), ErrBuiltInFieldNotIncluded)
	assert.ErrorIs(t, config.ChangeBuiltInFieldRequirement("no-such-entry", true, adminEmail(t)), ErrUnknownBuiltInField)
}

func TestChangeBuiltInFieldRequirement_ExcludeRetainsFlagDormantAndReincludeRestores(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	require.NoError(t, config.ChangeBuiltInFieldRequirement("experts", true, adminEmail(t)))
	require.NoError(t, config.ExcludeBuiltInField("experts", adminEmail(t)))
	require.NoError(t, config.IncludeBuiltInField("experts", adminEmail(t)))

	assert.True(t, config.IsBuiltInRequired("experts"))
}

func TestReorderFields_InterleavesBuiltInAndCustomFields(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := includeField(t, config)
	nameRef, _ := valueobjects.NewBuiltInFieldRef("name")
	descriptionRef, _ := valueobjects.NewBuiltInFieldRef("description")
	expertsRef, _ := valueobjects.NewBuiltInFieldRef("experts")

	require.NoError(t, config.ReorderFields([]valueobjects.FieldRef{
		valueobjects.NewCustomFieldRef(fieldID), nameRef, expertsRef, descriptionRef,
	}, adminEmail(t)))

	assert.Equal(t, []string{"custom:" + fieldID.Value(), "builtIn:name", "builtIn:experts", "builtIn:description"}, orderRefIDs(config))
}

func TestReorderFields_RejectsNonPermutation(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	nameRef, _ := valueobjects.NewBuiltInFieldRef("name")
	descriptionRef, _ := valueobjects.NewBuiltInFieldRef("description")

	assert.ErrorIs(t, config.ReorderFields([]valueobjects.FieldRef{nameRef, descriptionRef}, adminEmail(t)), ErrInvalidDisplayOrder)
	assert.ErrorIs(t, config.ReorderFields([]valueobjects.FieldRef{nameRef, nameRef, descriptionRef}, adminEmail(t)), ErrInvalidDisplayOrder)
	assert.ErrorIs(t, config.ReorderFields([]valueobjects.FieldRef{nameRef, descriptionRef, valueobjects.NewCustomFieldRef(valueobjects.NewFieldID())}, adminEmail(t)), ErrInvalidDisplayOrder)
}

func TestReplay_LegacySchemaEventsStillShapeInclusionAndRequirement(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	legacyField := valueobjects.NewFieldID().Value()
	retiredField := valueobjects.NewFieldID().Value()
	created := events.NewOnePagerConfigurationCreated(events.CreateConfigurationParams{
		ID: config.ID(), TenantID: "tenant-123", SubjectType: "application", BuiltIns: []string{"name", "description", "experts"}, CreatedBy: "admin@example.com",
	})
	history := []domain.DomainEvent{
		created,
		events.NewCustomFieldDefined(modifyParams(config), events.CustomFieldData{FieldID: legacyField, Name: "Contract", FieldType: "link", Required: true}),
		events.NewCustomFieldDefined(modifyParams(config), events.CustomFieldData{FieldID: retiredField, Name: "Hosting", FieldType: "selection", Options: []events.SelectionOptionData{{ID: valueobjects.NewOptionID().Value(), Label: "Cloud", Active: true}}}),
		events.NewCustomFieldRenamed(modifyParams(config), events.FieldRenameData{FieldID: legacyField, NewName: "Contract link"}),
		events.NewSelectionOptionAdded(modifyParams(config), retiredField, valueobjects.NewOptionID().Value(), "On-prem"),
		events.NewNumberFieldBoundsChanged(modifyParams(config), legacyField, nil, nil),
		events.NewCustomFieldRequirementChanged(modifyParams(config), legacyField, false),
		events.NewCustomFieldRetired(modifyParams(config), retiredField),
		events.NewCustomFieldReactivated(modifyParams(config), retiredField),
		events.NewCustomFieldRetired(modifyParams(config), retiredField),
	}

	replayed, err := LoadOnePagerConfigurationFromHistory(history)

	require.NoError(t, err)
	assert.Equal(t, len(history), replayed.Version())
	assert.Equal(t, []string{"builtIn:name", "builtIn:description", "builtIn:experts", "custom:" + legacyField}, orderRefIDs(replayed))
	legacyID, _ := valueobjects.NewFieldIDFromString(legacyField)
	retiredID, _ := valueobjects.NewFieldIDFromString(retiredField)
	assert.True(t, replayed.IsCustomFieldIncluded(legacyID))
	assert.False(t, replayed.IsCustomFieldRequired(legacyID))
	assert.False(t, replayed.IsCustomFieldIncluded(retiredID))
	assert.Equal(t, "admin@example.com", replayed.ModifiedBy().Value())
}

func TestReplay_ReconstructsPresentationState(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	contractID := includeField(t, config)
	hostingID := includeField(t, config)
	require.NoError(t, config.ChangeCustomFieldRequirement(contractID, true, adminEmail(t)))
	require.NoError(t, config.ExcludeBuiltInField("experts", adminEmail(t)))
	require.NoError(t, config.IncludeBuiltInField("experts", adminEmail(t)))
	require.NoError(t, config.ChangeBuiltInFieldRequirement("experts", true, adminEmail(t)))
	require.NoError(t, config.ExcludeCustomField(hostingID, adminEmail(t)))
	require.NoError(t, config.IncludeCustomField(hostingID, adminEmail(t)))
	nameRef, _ := valueobjects.NewBuiltInFieldRef("name")
	descriptionRef, _ := valueobjects.NewBuiltInFieldRef("description")
	expertsRef, _ := valueobjects.NewBuiltInFieldRef("experts")
	require.NoError(t, config.ReorderFields([]valueobjects.FieldRef{
		valueobjects.NewCustomFieldRef(hostingID), nameRef, valueobjects.NewCustomFieldRef(contractID), descriptionRef, expertsRef,
	}, adminEmail(t)))

	created := events.NewOnePagerConfigurationCreated(events.CreateConfigurationParams{
		ID: config.ID(), TenantID: "tenant-123", SubjectType: "application", BuiltIns: []string{"name", "description", "experts"}, CreatedBy: "admin@example.com",
	})
	history := append([]domain.DomainEvent{created}, config.GetUncommittedChanges()...)
	replayed, err := LoadOnePagerConfigurationFromHistory(history)

	require.NoError(t, err)
	assert.Equal(t, config.ID(), replayed.ID())
	assert.Equal(t, config.Version(), replayed.Version())
	assert.Equal(t, orderRefIDs(config), orderRefIDs(replayed))
	assert.True(t, replayed.IsCustomFieldRequired(contractID))
	assert.False(t, replayed.IsCustomFieldRequired(hostingID))
	assert.True(t, replayed.IsBuiltInRequired("experts"))
}
