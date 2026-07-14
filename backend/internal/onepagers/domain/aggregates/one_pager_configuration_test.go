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

func fieldName(t *testing.T, value string) valueobjects.FieldName {
	t.Helper()
	name, err := valueobjects.NewFieldName(value)
	require.NoError(t, err)
	return name
}

func fieldType(t *testing.T, value string) valueobjects.FieldType {
	t.Helper()
	ft, err := valueobjects.NewFieldType(value)
	require.NoError(t, err)
	return ft
}

func helpText(t *testing.T, value string) valueobjects.HelpText {
	t.Helper()
	help, err := valueobjects.NewHelpText(value)
	require.NoError(t, err)
	return help
}

func optionLabels(t *testing.T, labels ...string) []valueobjects.OptionLabel {
	t.Helper()
	result := make([]valueobjects.OptionLabel, len(labels))
	for i, l := range labels {
		label, err := valueobjects.NewOptionLabel(l)
		require.NoError(t, err)
		result[i] = label
	}
	return result
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

func mustDefineField(t *testing.T, config *OnePagerConfiguration, params DefineCustomFieldParams) valueobjects.FieldID {
	t.Helper()
	fieldID, err := config.DefineCustomField(params, adminEmail(t))
	require.NoError(t, err)
	return fieldID
}

func defineField(t *testing.T, config *OnePagerConfiguration, name, typeName string) valueobjects.FieldID {
	t.Helper()
	return mustDefineField(t, config, DefineCustomFieldParams{
		Name: fieldName(t, name),
		Type: fieldType(t, typeName),
	})
}

func defineSelectionField(t *testing.T, config *OnePagerConfiguration, labels ...string) valueobjects.FieldID {
	t.Helper()
	return mustDefineField(t, config, DefineCustomFieldParams{
		Name:         fieldName(t, "Hosting model"),
		Type:         fieldType(t, "selection"),
		OptionLabels: optionLabels(t, labels...),
	})
}

func customFieldByID(t *testing.T, config *OnePagerConfiguration, fieldID valueobjects.FieldID) valueobjects.CustomField {
	t.Helper()
	field, found := config.CustomFieldByID(fieldID)
	require.True(t, found)
	return field
}

func orderRefIDs(config *OnePagerConfiguration) []string {
	order := config.DisplayOrder()
	ids := make([]string, len(order))
	for i, ref := range order {
		ids[i] = string(ref.Kind()) + ":" + ref.RefID()
	}
	return ids
}

func floatPtr(v float64) *float64 {
	return &v
}

func lastEvent(t *testing.T, config *OnePagerConfiguration) domain.DomainEvent {
	t.Helper()
	changes := config.GetUncommittedChanges()
	require.NotEmpty(t, changes)
	return changes[len(changes)-1]
}

func TestNewOnePagerConfiguration_DefaultsToFullCatalogInOrder(t *testing.T) {
	tenantID, err := sharedvo.NewTenantID("tenant-123")
	require.NoError(t, err)

	config, err := NewOnePagerConfiguration(tenantID, subjectType(t, "vendor"), adminEmail(t))

	require.NoError(t, err)
	assert.NotEmpty(t, config.ID())
	assert.Equal(t, "vendor", config.SubjectType().Value())
	assert.Equal(t, []string{"builtIn:name", "builtIn:implementation-partner", "builtIn:notes"}, orderRefIDs(config))
	assert.Empty(t, config.CustomFields())
	assert.Equal(t, 1, config.Version())

	changes := config.GetUncommittedChanges()
	require.Len(t, changes, 1)
	created, ok := changes[0].(events.OnePagerConfigurationCreated)
	require.True(t, ok)
	assert.Equal(t, config.ID(), created.ID)
	assert.Equal(t, "tenant-123", created.TenantID)
	assert.Equal(t, "vendor", created.SubjectType)
	assert.Equal(t, []string{"name", "implementation-partner", "notes"}, created.BuiltIns)
	assert.Equal(t, "admin@example.com", created.CreatedBy)
}

func TestDefineCustomField_EachFieldType(t *testing.T) {
	cases := []struct {
		name      string
		fieldType string
		labels    []string
	}{
		{"Business summary", "text", nil},
		{"Annual cost", "number", nil},
		{"Contract renewal", "date", nil},
		{"Contract", "link", nil},
		{"Hosting model", "selection", []string{"On-prem", "Cloud"}},
		{"Product owner", "contact-person", nil},
	}

	for _, tc := range cases {
		t.Run(tc.fieldType, func(t *testing.T) {
			config := newCommittedApplicationConfig(t)

			fieldID := mustDefineField(t, config, DefineCustomFieldParams{
				Name:         fieldName(t, tc.name),
				Type:         fieldType(t, tc.fieldType),
				OptionLabels: optionLabels(t, tc.labels...),
			})

			field := customFieldByID(t, config, fieldID)
			assert.True(t, field.IsActive())
			assert.Equal(t, tc.name, field.Name().Value())
			assert.Equal(t, tc.fieldType, field.Type().Value())
			assert.NotEmpty(t, field.ID().Value())

			order := config.DisplayOrder()
			last := order[len(order)-1]
			assert.Equal(t, valueobjects.FieldRefKindCustom, last.Kind())
			assert.Equal(t, fieldID.Value(), last.RefID())

			defined, ok := lastEvent(t, config).(events.CustomFieldDefined)
			require.True(t, ok)
			assert.Equal(t, fieldID.Value(), defined.FieldID)
			assert.Equal(t, tc.fieldType, defined.FieldType)
			assert.Equal(t, len(tc.labels), len(defined.Options))
		})
	}
}

func TestDefineCustomField_RejectsDuplicateActiveNameCaseInsensitive(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	defineField(t, config, "Contract link", "link")

	_, err := config.DefineCustomField(DefineCustomFieldParams{
		Name: fieldName(t, "contract LINK"),
		Type: fieldType(t, "text"),
	}, adminEmail(t))

	assert.ErrorIs(t, err, ErrDuplicateFieldName)
}

func TestDefineCustomField_RejectsNameCollidingWithIncludedBuiltInLabel(t *testing.T) {
	config := newCommittedApplicationConfig(t)

	_, err := config.DefineCustomField(DefineCustomFieldParams{
		Name: fieldName(t, "experts"),
		Type: fieldType(t, "text"),
	}, adminEmail(t))

	assert.ErrorIs(t, err, ErrDuplicateFieldName)
}

func TestDefineCustomField_AllowsNameOfRetiredField(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Contract link", "link")
	require.NoError(t, config.RetireCustomField(fieldID, adminEmail(t)))

	_, err := config.DefineCustomField(DefineCustomFieldParams{
		Name: fieldName(t, "Contract link"),
		Type: fieldType(t, "text"),
	}, adminEmail(t))

	assert.NoError(t, err)
}

func TestDefineCustomField_SelectionWithoutOptionsRejected(t *testing.T) {
	config := newCommittedApplicationConfig(t)

	_, err := config.DefineCustomField(DefineCustomFieldParams{
		Name: fieldName(t, "Hosting model"),
		Type: fieldType(t, "selection"),
	}, adminEmail(t))

	assert.ErrorIs(t, err, valueobjects.ErrSelectionOptionRequired)
}

func TestRenameCustomField_PreservesIdentityTypeRequirementAndPosition(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Contract", "link")
	defineField(t, config, "Product owner", "contact-person")
	originalOrder := orderRefIDs(config)
	config.MarkChangesAsCommitted()

	err := config.RenameCustomField(RenameCustomFieldParams{
		FieldID:  fieldID,
		Name:     fieldName(t, "Contract link"),
		HelpText: helpText(t, "Link to the contract"),
	}, adminEmail(t))

	require.NoError(t, err)
	field := customFieldByID(t, config, fieldID)
	assert.Equal(t, "Contract link", field.Name().Value())
	assert.Equal(t, "Link to the contract", field.HelpText().Value())
	assert.Equal(t, "link", field.Type().Value())
	assert.False(t, field.IsRequired())
	assert.Equal(t, originalOrder, orderRefIDs(config))

	renamed, ok := lastEvent(t, config).(events.CustomFieldRenamed)
	require.True(t, ok)
	assert.Equal(t, fieldID.Value(), renamed.FieldID)
	assert.Equal(t, "Contract link", renamed.NewName)
}

func TestRenameCustomField_RejectsDuplicateName(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	defineField(t, config, "Contract link", "link")
	other := defineField(t, config, "Product owner", "contact-person")

	err := config.RenameCustomField(RenameCustomFieldParams{
		FieldID: other,
		Name:    fieldName(t, "contract link"),
	}, adminEmail(t))

	assert.ErrorIs(t, err, ErrDuplicateFieldName)
}

func TestRenameCustomField_AllowsKeepingOwnName(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Contract link", "link")

	err := config.RenameCustomField(RenameCustomFieldParams{
		FieldID:  fieldID,
		Name:     fieldName(t, "Contract Link"),
		HelpText: helpText(t, "updated"),
	}, adminEmail(t))

	assert.NoError(t, err)
}

func TestRenameCustomField_RejectsFieldTypeChange(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Contract link", "link")
	config.MarkChangesAsCommitted()

	err := config.RenameCustomField(RenameCustomFieldParams{
		FieldID:       fieldID,
		Name:          fieldName(t, "Contract link"),
		RequestedType: "text",
	}, adminEmail(t))

	assert.ErrorIs(t, err, ErrFieldTypeImmutable)
	assert.Equal(t, "link", customFieldByID(t, config, fieldID).Type().Value())
	assert.Empty(t, config.GetUncommittedChanges())
}

func TestRenameCustomField_AllowsRestatingSameType(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Contract link", "link")

	err := config.RenameCustomField(RenameCustomFieldParams{
		FieldID:       fieldID,
		Name:          fieldName(t, "Contract URL"),
		RequestedType: "link",
	}, adminEmail(t))

	assert.NoError(t, err)
}

func TestRetireCustomField_LeavesDisplayOrderAndStaysListed(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Contract link", "link")
	config.MarkChangesAsCommitted()

	err := config.RetireCustomField(fieldID, adminEmail(t))

	require.NoError(t, err)
	assert.NotContains(t, orderRefIDs(config), "custom:"+fieldID.Value())
	field := customFieldByID(t, config, fieldID)
	assert.False(t, field.IsActive())

	retired, ok := lastEvent(t, config).(events.CustomFieldRetired)
	require.True(t, ok)
	assert.Equal(t, fieldID.Value(), retired.FieldID)
}

func TestRetireCustomField_RejectsAlreadyRetired(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Contract link", "link")
	require.NoError(t, config.RetireCustomField(fieldID, adminEmail(t)))

	err := config.RetireCustomField(fieldID, adminEmail(t))

	assert.ErrorIs(t, err, ErrFieldAlreadyRetired)
}

func TestReactivateCustomField_RestoresDefinitionAndAppendsToOrder(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineSelectionField(t, config, "On-prem", "Cloud")
	require.NoError(t, config.RetireCustomField(fieldID, adminEmail(t)))
	config.MarkChangesAsCommitted()

	err := config.ReactivateCustomField(fieldID, adminEmail(t))

	require.NoError(t, err)
	field := customFieldByID(t, config, fieldID)
	assert.True(t, field.IsActive())
	assert.Equal(t, "selection", field.Type().Value())
	assert.Len(t, field.Options(), 2)

	order := orderRefIDs(config)
	assert.Equal(t, "custom:"+fieldID.Value(), order[len(order)-1])

	reactivated, ok := lastEvent(t, config).(events.CustomFieldReactivated)
	require.True(t, ok)
	assert.Equal(t, fieldID.Value(), reactivated.FieldID)
}

func TestReactivateCustomField_RejectsActiveField(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Contract link", "link")

	err := config.ReactivateCustomField(fieldID, adminEmail(t))

	assert.ErrorIs(t, err, ErrFieldAlreadyActive)
}

func TestReactivateCustomField_RejectsNameCollisionWithActiveField(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Contract link", "link")
	require.NoError(t, config.RetireCustomField(fieldID, adminEmail(t)))
	defineField(t, config, "Contract link", "text")

	err := config.ReactivateCustomField(fieldID, adminEmail(t))

	assert.ErrorIs(t, err, ErrDuplicateFieldName)
}

func TestChangeCustomFieldRequirement_RecordsFlag(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Product owner", "contact-person")
	config.MarkChangesAsCommitted()

	err := config.ChangeCustomFieldRequirement(fieldID, true, adminEmail(t))

	require.NoError(t, err)
	assert.True(t, customFieldByID(t, config, fieldID).IsRequired())

	changed, ok := lastEvent(t, config).(events.CustomFieldRequirementChanged)
	require.True(t, ok)
	assert.Equal(t, fieldID.Value(), changed.FieldID)
	assert.True(t, changed.Required)
}

func TestChangeCustomFieldRequirement_RejectsRetiredField(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Product owner", "contact-person")
	require.NoError(t, config.RetireCustomField(fieldID, adminEmail(t)))

	err := config.ChangeCustomFieldRequirement(fieldID, true, adminEmail(t))

	assert.ErrorIs(t, err, ErrFieldRetired)
}

func TestExcludeBuiltInField_LeavesDisplayOrder(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	config.MarkChangesAsCommitted()

	err := config.ExcludeBuiltInField("experts", adminEmail(t))

	require.NoError(t, err)
	assert.NotContains(t, orderRefIDs(config), "builtIn:experts")

	excluded, ok := lastEvent(t, config).(events.BuiltInFieldExcluded)
	require.True(t, ok)
	assert.Equal(t, "experts", excluded.EntryID)
}

func TestIncludeBuiltInField_ReappearsAtEndOfOrder(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	require.NoError(t, config.ExcludeBuiltInField("experts", adminEmail(t)))
	config.MarkChangesAsCommitted()

	err := config.IncludeBuiltInField("experts", adminEmail(t))

	require.NoError(t, err)
	order := orderRefIDs(config)
	assert.Equal(t, "builtIn:experts", order[len(order)-1])

	included, ok := lastEvent(t, config).(events.BuiltInFieldIncluded)
	require.True(t, ok)
	assert.Equal(t, "experts", included.EntryID)
}

func TestIncludeBuiltInField_RejectsUnknownCatalogEntry(t *testing.T) {
	config := newCommittedApplicationConfig(t)

	err := config.IncludeBuiltInField("maturity", adminEmail(t))

	assert.ErrorIs(t, err, ErrUnknownBuiltInField)
}

func TestIncludeBuiltInField_RejectsAlreadyIncluded(t *testing.T) {
	config := newCommittedApplicationConfig(t)

	err := config.IncludeBuiltInField("experts", adminEmail(t))

	assert.ErrorIs(t, err, ErrBuiltInFieldAlreadyIncluded)
}

func TestExcludeBuiltInField_RejectsNotIncluded(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	require.NoError(t, config.ExcludeBuiltInField("experts", adminEmail(t)))

	err := config.ExcludeBuiltInField("experts", adminEmail(t))

	assert.ErrorIs(t, err, ErrBuiltInFieldNotIncluded)
}

func TestIncludeBuiltInField_RejectsLabelCollisionWithActiveCustomField(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	require.NoError(t, config.ExcludeBuiltInField("experts", adminEmail(t)))
	defineField(t, config, "Experts", "text")

	err := config.IncludeBuiltInField("experts", adminEmail(t))

	assert.ErrorIs(t, err, ErrDuplicateFieldName)
}

func TestChangeBuiltInFieldRequirement_RecordsFlagOnIncludedBuiltIn(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	config.MarkChangesAsCommitted()

	err := config.ChangeBuiltInFieldRequirement("experts", true, adminEmail(t))

	require.NoError(t, err)
	assert.True(t, config.IsBuiltInRequired("experts"))

	changed, ok := lastEvent(t, config).(events.BuiltInFieldRequirementChanged)
	require.True(t, ok)
	assert.Equal(t, "experts", changed.EntryID)
	assert.True(t, changed.Required)
}

func TestChangeBuiltInFieldRequirement_RejectsExcludedBuiltIn(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	require.NoError(t, config.ExcludeBuiltInField("experts", adminEmail(t)))

	err := config.ChangeBuiltInFieldRequirement("experts", true, adminEmail(t))

	assert.ErrorIs(t, err, ErrBuiltInFieldNotIncluded)
}

func TestChangeBuiltInFieldRequirement_RejectsUnknownCatalogEntry(t *testing.T) {
	config := newCommittedApplicationConfig(t)

	err := config.ChangeBuiltInFieldRequirement("maturity", true, adminEmail(t))

	assert.ErrorIs(t, err, ErrUnknownBuiltInField)
}

func TestChangeBuiltInFieldRequirement_ReplaysFromHistory(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	require.NoError(t, config.ChangeBuiltInFieldRequirement("experts", true, adminEmail(t)))

	created := events.NewOnePagerConfigurationCreated(events.CreateConfigurationParams{
		ID:          config.ID(),
		TenantID:    "tenant-123",
		SubjectType: "application",
		BuiltIns:    []string{"name", "description", "experts"},
		CreatedBy:   "admin@example.com",
	})
	history := append([]domain.DomainEvent{created}, config.GetUncommittedChanges()...)

	replayed, err := LoadOnePagerConfigurationFromHistory(history)

	require.NoError(t, err)
	assert.True(t, replayed.IsBuiltInRequired("experts"))
	assert.False(t, replayed.IsBuiltInRequired("description"))
}

func TestChangeBuiltInFieldRequirement_ExcludeRetainsFlagDormantAndReincludeRestores(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	require.NoError(t, config.ChangeBuiltInFieldRequirement("experts", true, adminEmail(t)))

	require.NoError(t, config.ExcludeBuiltInField("experts", adminEmail(t)))
	assert.True(t, config.IsBuiltInRequired("experts"), "excluding retains the required flag dormant")

	require.NoError(t, config.IncludeBuiltInField("experts", adminEmail(t)))
	assert.True(t, config.IsBuiltInRequired("experts"), "re-including restores the prior required flag")
}

func TestReorderFields_InterleavesBuiltInAndCustomFields(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Contract link", "link")
	config.MarkChangesAsCommitted()

	nameRef, _ := valueobjects.NewBuiltInFieldRef("name")
	descriptionRef, _ := valueobjects.NewBuiltInFieldRef("description")
	expertsRef, _ := valueobjects.NewBuiltInFieldRef("experts")
	customRef := valueobjects.NewCustomFieldRef(fieldID)

	err := config.ReorderFields([]valueobjects.FieldRef{nameRef, descriptionRef, customRef, expertsRef}, adminEmail(t))

	require.NoError(t, err)
	assert.Equal(t, []string{"builtIn:name", "builtIn:description", "custom:" + fieldID.Value(), "builtIn:experts"}, orderRefIDs(config))

	reordered, ok := lastEvent(t, config).(events.OnePagerFieldsReordered)
	require.True(t, ok)
	assert.Len(t, reordered.Order, 4)
}

func TestReorderFields_RejectsNonPermutation(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	nameRef, _ := valueobjects.NewBuiltInFieldRef("name")
	descriptionRef, _ := valueobjects.NewBuiltInFieldRef("description")

	err := config.ReorderFields([]valueobjects.FieldRef{nameRef, descriptionRef}, adminEmail(t))

	assert.ErrorIs(t, err, ErrInvalidDisplayOrder)
}

func TestReorderFields_RejectsDuplicateEntries(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	nameRef, _ := valueobjects.NewBuiltInFieldRef("name")
	descriptionRef, _ := valueobjects.NewBuiltInFieldRef("description")
	expertsRef, _ := valueobjects.NewBuiltInFieldRef("experts")

	err := config.ReorderFields([]valueobjects.FieldRef{nameRef, descriptionRef, expertsRef, nameRef}, adminEmail(t))

	assert.ErrorIs(t, err, ErrInvalidDisplayOrder)
}

func TestReorderFields_RejectsRetiredFieldInOrder(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Contract link", "link")
	require.NoError(t, config.RetireCustomField(fieldID, adminEmail(t)))

	nameRef, _ := valueobjects.NewBuiltInFieldRef("name")
	descriptionRef, _ := valueobjects.NewBuiltInFieldRef("description")
	expertsRef, _ := valueobjects.NewBuiltInFieldRef("experts")
	customRef := valueobjects.NewCustomFieldRef(fieldID)

	err := config.ReorderFields([]valueobjects.FieldRef{nameRef, descriptionRef, expertsRef, customRef}, adminEmail(t))

	assert.ErrorIs(t, err, ErrInvalidDisplayOrder)
}

func TestSelectionOptionLifecycle_AddAndRetire(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineSelectionField(t, config, "On-prem", "Cloud")
	config.MarkChangesAsCommitted()

	optionID, err := config.AddSelectionOption(fieldID, optionLabels(t, "Hybrid")[0], adminEmail(t))
	require.NoError(t, err)
	assert.NotEmpty(t, optionID.Value())

	added, ok := lastEvent(t, config).(events.SelectionOptionAdded)
	require.True(t, ok)
	assert.Equal(t, "Hybrid", added.Label)

	field := customFieldByID(t, config, fieldID)
	onPremID := field.Options()[0].ID()
	require.NoError(t, config.RetireSelectionOption(fieldID, onPremID, adminEmail(t)))

	field = customFieldByID(t, config, fieldID)
	require.Len(t, field.Options(), 3)
	assert.False(t, field.Options()[0].IsActive())
	assert.True(t, field.Options()[1].IsActive())
	assert.True(t, field.Options()[2].IsActive())

	retired, ok := lastEvent(t, config).(events.SelectionOptionRetired)
	require.True(t, ok)
	assert.Equal(t, onPremID.Value(), retired.OptionID)
}

func TestAddSelectionOption_RejectsNonSelectionField(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Business summary", "text")

	_, err := config.AddSelectionOption(fieldID, optionLabels(t, "Cloud")[0], adminEmail(t))

	assert.ErrorIs(t, err, valueobjects.ErrNotSelectionField)
}

func TestAddSelectionOption_RejectsRetiredField(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineSelectionField(t, config, "On-prem")
	require.NoError(t, config.RetireCustomField(fieldID, adminEmail(t)))

	_, err := config.AddSelectionOption(fieldID, optionLabels(t, "Cloud")[0], adminEmail(t))

	assert.ErrorIs(t, err, ErrFieldRetired)
}

func TestRetireSelectionOption_RejectsLastActiveOption(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineSelectionField(t, config, "On-prem")
	optionID := customFieldByID(t, config, fieldID).Options()[0].ID()

	err := config.RetireSelectionOption(fieldID, optionID, adminEmail(t))

	assert.ErrorIs(t, err, valueobjects.ErrLastActiveOption)
}

func TestSetNumberFieldBounds_UpdatesBoundsAndPreservesIdentity(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Maturity score", "number")
	config.MarkChangesAsCommitted()

	err := config.SetNumberFieldBounds(fieldID, floatPtr(0), floatPtr(5), adminEmail(t))

	require.NoError(t, err)
	field := customFieldByID(t, config, fieldID)
	assert.Equal(t, floatPtr(0), field.Min())
	assert.Equal(t, floatPtr(5), field.Max())
	assert.Equal(t, "number", field.Type().Value())
	assert.False(t, field.IsRequired())

	changed, ok := lastEvent(t, config).(events.NumberFieldBoundsChanged)
	require.True(t, ok)
	assert.Equal(t, fieldID.Value(), changed.FieldID)
	assert.Equal(t, floatPtr(0), changed.Min)
	assert.Equal(t, floatPtr(5), changed.Max)
}

func TestSetNumberFieldBounds_CanClearABound(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Headcount", "number")
	require.NoError(t, config.SetNumberFieldBounds(fieldID, floatPtr(0), floatPtr(500), adminEmail(t)))

	err := config.SetNumberFieldBounds(fieldID, floatPtr(0), nil, adminEmail(t))

	require.NoError(t, err)
	field := customFieldByID(t, config, fieldID)
	assert.Equal(t, floatPtr(0), field.Min())
	assert.Nil(t, field.Max())
}

func TestSetNumberFieldBounds_DoesNotAppendFactsEvent(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Maturity score", "number")
	config.MarkChangesAsCommitted()

	require.NoError(t, config.SetNumberFieldBounds(fieldID, floatPtr(0), floatPtr(3), adminEmail(t)))

	changes := config.GetUncommittedChanges()
	require.Len(t, changes, 1)
	assert.Equal(t, "NumberFieldBoundsChanged", changes[0].EventType())
}

func TestSetNumberFieldBounds_RejectsMinimumGreaterThanMaximum(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Maturity score", "number")
	config.MarkChangesAsCommitted()

	err := config.SetNumberFieldBounds(fieldID, floatPtr(10), floatPtr(5), adminEmail(t))

	assert.ErrorIs(t, err, valueobjects.ErrMinExceedsMax)
	assert.Empty(t, config.GetUncommittedChanges())
	assert.Nil(t, customFieldByID(t, config, fieldID).Min())
}

func TestSetNumberFieldBounds_RejectsNonNumberField(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Business summary", "text")

	err := config.SetNumberFieldBounds(fieldID, floatPtr(0), floatPtr(5), adminEmail(t))

	assert.ErrorIs(t, err, valueobjects.ErrBoundsNotAllowed)
}

func TestSetNumberFieldBounds_RejectsRetiredField(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	fieldID := defineField(t, config, "Maturity score", "number")
	require.NoError(t, config.RetireCustomField(fieldID, adminEmail(t)))

	err := config.SetNumberFieldBounds(fieldID, floatPtr(0), floatPtr(5), adminEmail(t))

	assert.ErrorIs(t, err, ErrFieldRetired)
}

func TestSetNumberFieldBounds_RejectsUnknownField(t *testing.T) {
	config := newCommittedApplicationConfig(t)

	err := config.SetNumberFieldBounds(valueobjects.NewFieldID(), floatPtr(0), floatPtr(5), adminEmail(t))

	assert.ErrorIs(t, err, ErrFieldNotFound)
}

func TestCommandsOnUnknownField_ReturnFieldNotFound(t *testing.T) {
	config := newCommittedApplicationConfig(t)
	unknown := valueobjects.NewFieldID()

	assert.ErrorIs(t, config.RetireCustomField(unknown, adminEmail(t)), ErrFieldNotFound)
	assert.ErrorIs(t, config.ReactivateCustomField(unknown, adminEmail(t)), ErrFieldNotFound)
	assert.ErrorIs(t, config.ChangeCustomFieldRequirement(unknown, true, adminEmail(t)), ErrFieldNotFound)
	assert.ErrorIs(t, config.RenameCustomField(RenameCustomFieldParams{
		FieldID: unknown,
		Name:    fieldName(t, "Anything"),
	}, adminEmail(t)), ErrFieldNotFound)
	_, err := config.AddSelectionOption(unknown, optionLabels(t, "Cloud")[0], adminEmail(t))
	assert.ErrorIs(t, err, ErrFieldNotFound)
	assert.ErrorIs(t, config.RetireSelectionOption(unknown, valueobjects.NewOptionID(), adminEmail(t)), ErrFieldNotFound)
}

func TestReplay_ReconstructsFullConfiguration(t *testing.T) {
	config := newCommittedApplicationConfig(t)

	contractID := defineField(t, config, "Contract", "link")
	hostingID := defineSelectionField(t, config, "On-prem", "Cloud")
	require.NoError(t, config.RenameCustomField(RenameCustomFieldParams{
		FieldID:  contractID,
		Name:     fieldName(t, "Contract link"),
		HelpText: helpText(t, "Signed contract"),
	}, adminEmail(t)))
	require.NoError(t, config.ChangeCustomFieldRequirement(contractID, true, adminEmail(t)))
	maturityID := defineField(t, config, "Maturity score", "number")
	require.NoError(t, config.SetNumberFieldBounds(maturityID, floatPtr(0), floatPtr(5), adminEmail(t)))
	require.NoError(t, config.ExcludeBuiltInField("experts", adminEmail(t)))
	require.NoError(t, config.IncludeBuiltInField("experts", adminEmail(t)))
	_, err := config.AddSelectionOption(hostingID, optionLabels(t, "Hybrid")[0], adminEmail(t))
	require.NoError(t, err)
	onPremID := customFieldByID(t, config, hostingID).Options()[0].ID()
	require.NoError(t, config.RetireSelectionOption(hostingID, onPremID, adminEmail(t)))
	require.NoError(t, config.RetireCustomField(hostingID, adminEmail(t)))
	require.NoError(t, config.ReactivateCustomField(hostingID, adminEmail(t)))

	nameRef, _ := valueobjects.NewBuiltInFieldRef("name")
	descriptionRef, _ := valueobjects.NewBuiltInFieldRef("description")
	expertsRef, _ := valueobjects.NewBuiltInFieldRef("experts")
	require.NoError(t, config.ReorderFields([]valueobjects.FieldRef{
		valueobjects.NewCustomFieldRef(hostingID),
		nameRef,
		valueobjects.NewCustomFieldRef(contractID),
		descriptionRef,
		valueobjects.NewCustomFieldRef(maturityID),
		expertsRef,
	}, adminEmail(t)))

	created := events.NewOnePagerConfigurationCreated(events.CreateConfigurationParams{
		ID:          config.ID(),
		TenantID:    "tenant-123",
		SubjectType: "application",
		BuiltIns:    []string{"name", "description", "experts"},
		CreatedBy:   "admin@example.com",
	})
	history := append([]domain.DomainEvent{created}, config.GetUncommittedChanges()...)

	replayed, err := LoadOnePagerConfigurationFromHistory(history)
	require.NoError(t, err)

	assert.Equal(t, config.ID(), replayed.ID())
	assert.Equal(t, config.Version(), replayed.Version())
	assert.Equal(t, config.SubjectType().Value(), replayed.SubjectType().Value())
	assert.Equal(t, orderRefIDs(config), orderRefIDs(replayed))

	originalFields := config.CustomFields()
	replayedFields := replayed.CustomFields()
	require.Len(t, replayedFields, len(originalFields))
	for i := range originalFields {
		assert.Equal(t, originalFields[i].ID().Value(), replayedFields[i].ID().Value())
		assert.Equal(t, originalFields[i].Name().Value(), replayedFields[i].Name().Value())
		assert.Equal(t, originalFields[i].Type().Value(), replayedFields[i].Type().Value())
		assert.Equal(t, originalFields[i].IsRequired(), replayedFields[i].IsRequired())
		assert.Equal(t, originalFields[i].IsActive(), replayedFields[i].IsActive())
		assert.Equal(t, originalFields[i].HelpText().Value(), replayedFields[i].HelpText().Value())
		assert.Equal(t, originalFields[i].Options(), replayedFields[i].Options())
		assert.Equal(t, originalFields[i].Min(), replayedFields[i].Min())
		assert.Equal(t, originalFields[i].Max(), replayedFields[i].Max())
	}
}
