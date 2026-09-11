package aggregates

import (
	"testing"

	"easi/backend/internal/metamodel/domain/events"
	"easi/backend/internal/metamodel/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSchema(t *testing.T) (*SubjectAttributeSchema, valueobjects.UserEmail) {
	t.Helper()
	tenantID, err := sharedvo.NewTenantID("tenant-123")
	require.NoError(t, err)
	subjectType, err := valueobjects.NewSubjectType("application")
	require.NoError(t, err)
	steward, err := valueobjects.NewUserEmail("steward@example.com")
	require.NoError(t, err)
	schema, err := NewSubjectAttributeSchema(tenantID, subjectType, steward)
	require.NoError(t, err)
	schema.MarkChangesAsCommitted()
	return schema, steward
}

func name(t *testing.T, v string) valueobjects.AttributeName {
	t.Helper()
	n, err := valueobjects.NewAttributeName(v)
	require.NoError(t, err)
	return n
}

func attrType(t *testing.T, v string) valueobjects.AttributeType {
	t.Helper()
	at, err := valueobjects.NewAttributeType(v)
	require.NoError(t, err)
	return at
}

func label(t *testing.T, v string) valueobjects.OptionLabel {
	t.Helper()
	l, err := valueobjects.NewOptionLabel(v)
	require.NoError(t, err)
	return l
}

func defineSelection(t *testing.T, schema *SubjectAttributeSchema, by valueobjects.UserEmail, labels ...string) valueobjects.AttributeID {
	t.Helper()
	optionLabels := make([]valueobjects.OptionLabel, len(labels))
	for i, l := range labels {
		optionLabels[i] = label(t, l)
	}
	id, err := schema.DefineAttribute(DefineAttributeParams{
		Name:         name(t, "Hosting Region"),
		Type:         attrType(t, "selection"),
		OptionLabels: optionLabels,
	}, by)
	require.NoError(t, err)
	return id
}

func eventTypes(schema *SubjectAttributeSchema) []string {
	changes := schema.GetUncommittedChanges()
	types := make([]string, len(changes))
	for i, e := range changes {
		types[i] = e.EventType()
	}
	return types
}

func TestNewSubjectAttributeSchema_RaisesCreatedEvent(t *testing.T) {
	tenantID, _ := sharedvo.NewTenantID("tenant-123")
	subjectType, _ := valueobjects.NewSubjectType("capability")
	steward, _ := valueobjects.NewUserEmail("steward@example.com")

	schema, err := NewSubjectAttributeSchema(tenantID, subjectType, steward)

	require.NoError(t, err)
	assert.Equal(t, 1, schema.Version())
	assert.Equal(t, "capability", schema.SubjectType().Value())
	assert.Empty(t, schema.Attributes())
	changes := schema.GetUncommittedChanges()
	require.Len(t, changes, 1)
	created, ok := changes[0].(events.SubjectAttributeSchemaCreated)
	require.True(t, ok)
	assert.Equal(t, "capability", created.SubjectType)
	assert.Equal(t, "tenant-123", created.TenantID)
}

func TestDefineAttribute_AddsActiveAttributeAndRaisesEvent(t *testing.T) {
	schema, steward := newSchema(t)

	id := defineSelection(t, schema, steward, "EU", "US")

	attribute, found := schema.AttributeByID(id)
	require.True(t, found)
	assert.True(t, attribute.IsActive())
	assert.Len(t, attribute.Options(), 2)
	require.Equal(t, []string{"SubjectAttributeDefined"}, eventTypes(schema))
	defined := schema.GetUncommittedChanges()[0].(events.SubjectAttributeDefined)
	assert.Equal(t, "application", defined.SubjectType)
	assert.Equal(t, id.Value(), defined.AttributeID)
	assert.Equal(t, "selection", defined.AttributeType)
	assert.Equal(t, 2, defined.Version)
}

func TestDefineAttribute_RejectsDuplicateActiveNameIgnoringCase(t *testing.T) {
	schema, steward := newSchema(t)
	defineSelection(t, schema, steward, "EU")

	_, err := schema.DefineAttribute(DefineAttributeParams{Name: name(t, "hosting region"), Type: attrType(t, "text")}, steward)

	assert.ErrorIs(t, err, ErrDuplicateAttributeName)
}

func TestDefineAttribute_NumberWithBounds(t *testing.T) {
	schema, steward := newSchema(t)
	min, max := 0.0, 5.0

	id, err := schema.DefineAttribute(DefineAttributeParams{Name: name(t, "Score"), Type: attrType(t, "number"), Min: &min, Max: &max}, steward)

	require.NoError(t, err)
	attribute, _ := schema.AttributeByID(id)
	assert.Equal(t, 5.0, *attribute.Max())
	defined := schema.GetUncommittedChanges()[0].(events.SubjectAttributeDefined)
	assert.Equal(t, 0.0, *defined.Min)
}

func TestRenameAttribute_KeepsTypeAndRejectsTypeChange(t *testing.T) {
	schema, steward := newSchema(t)
	id := defineSelection(t, schema, steward, "EU")
	help, _ := valueobjects.NewHelpText("Where it runs")

	err := schema.RenameAttribute(RenameAttributeParams{AttributeID: id, Name: name(t, "Region"), HelpText: help, RequestedType: "text"}, steward)
	assert.ErrorIs(t, err, ErrAttributeTypeImmutable)

	require.NoError(t, schema.RenameAttribute(RenameAttributeParams{AttributeID: id, Name: name(t, "Region"), HelpText: help}, steward))
	attribute, _ := schema.AttributeByID(id)
	assert.Equal(t, "Region", attribute.Name().Value())
	assert.Equal(t, "Where it runs", attribute.HelpText().Value())
	assert.Contains(t, eventTypes(schema), "SubjectAttributeRenamed")
}

func TestRetireAndReactivateAttribute(t *testing.T) {
	schema, steward := newSchema(t)
	id := defineSelection(t, schema, steward, "EU")

	require.NoError(t, schema.RetireAttribute(id, steward))
	assert.ErrorIs(t, schema.RetireAttribute(id, steward), ErrAttributeAlreadyRetired)
	attribute, _ := schema.AttributeByID(id)
	assert.False(t, attribute.IsActive())

	_, err := schema.AddOption(id, label(t, "US"), steward)
	assert.ErrorIs(t, err, ErrAttributeRetired)

	require.NoError(t, schema.ReactivateAttribute(id, steward))
	assert.ErrorIs(t, schema.ReactivateAttribute(id, steward), ErrAttributeAlreadyActive)
	assert.Equal(t, []string{"SubjectAttributeDefined", "SubjectAttributeRetired", "SubjectAttributeReactivated"}, eventTypes(schema))
}

func TestReactivateAttribute_RejectsWhenNameNowTaken(t *testing.T) {
	schema, steward := newSchema(t)
	id := defineSelection(t, schema, steward, "EU")
	require.NoError(t, schema.RetireAttribute(id, steward))
	defineSelection(t, schema, steward, "EU")

	assert.ErrorIs(t, schema.ReactivateAttribute(id, steward), ErrDuplicateAttributeName)
}

func TestOptionsAndBounds(t *testing.T) {
	schema, steward := newSchema(t)
	selection := defineSelection(t, schema, steward, "EU")
	number, err := schema.DefineAttribute(DefineAttributeParams{Name: name(t, "Score"), Type: attrType(t, "number")}, steward)
	require.NoError(t, err)

	optionID, err := schema.AddOption(selection, label(t, "US"), steward)
	require.NoError(t, err)
	_, err = schema.AddOption(number, label(t, "US"), steward)
	assert.ErrorIs(t, err, valueobjects.ErrNotSelectionAttribute)
	require.NoError(t, schema.RetireOption(selection, optionID, steward))
	assert.ErrorIs(t, schema.RetireOption(selection, optionID, steward), valueobjects.ErrOptionAlreadyRetired)

	min := 1.0
	require.NoError(t, schema.SetBounds(number, &min, nil, steward))
	assert.ErrorIs(t, schema.SetBounds(selection, &min, nil, steward), valueobjects.ErrBoundsNotAllowed)
	assert.ErrorIs(t, schema.SetBounds(valueobjects.NewAttributeID(), &min, nil, steward), ErrAttributeNotFound)

	assert.Equal(t, []string{
		"SubjectAttributeDefined", "SubjectAttributeDefined",
		"SubjectAttributeOptionAdded", "SubjectAttributeOptionRetired", "SubjectAttributeBoundsChanged",
	}, eventTypes(schema))
}

func TestImportAttribute_PreservesIdentityOptionsBoundsAndRetirement(t *testing.T) {
	schema, steward := newSchema(t)
	attributeID := valueobjects.NewAttributeID()
	activeOption := valueobjects.NewAttributeOption(valueobjects.NewOptionID(), label(t, "EU"))
	retiredOption := valueobjects.NewRetiredAttributeOption(valueobjects.NewOptionID(), label(t, "Mars"))
	imported, err := valueobjects.NewSubjectAttribute(valueobjects.SubjectAttributeParams{
		ID:      attributeID,
		Name:    name(t, "Hosting Region"),
		Type:    attrType(t, "selection"),
		Options: []valueobjects.AttributeOption{activeOption, retiredOption},
	})
	require.NoError(t, err)
	imported = imported.Retired()

	require.NoError(t, schema.ImportAttribute(imported, steward))

	stored, found := schema.AttributeByID(attributeID)
	require.True(t, found)
	assert.False(t, stored.IsActive())
	assert.Equal(t, imported.Options(), stored.Options())
	assert.Equal(t, []string{"SubjectAttributeDefined", "SubjectAttributeRetired"}, eventTypes(schema))
	defined := schema.GetUncommittedChanges()[0].(events.SubjectAttributeDefined)
	assert.Equal(t, attributeID.Value(), defined.AttributeID)
	assert.Equal(t, retiredOption.ID().Value(), defined.Options[1].ID)
	assert.False(t, defined.Options[1].Active)

	assert.ErrorIs(t, schema.ImportAttribute(imported, steward), ErrAttributeAlreadyDefined)
}

func TestImportAttribute_WithBoundsRaisesSingleDefinedEvent(t *testing.T) {
	schema, steward := newSchema(t)
	min, max := 0.0, 10.0
	imported, err := valueobjects.NewSubjectAttribute(valueobjects.SubjectAttributeParams{
		ID: valueobjects.NewAttributeID(), Name: name(t, "Score"), Type: attrType(t, "number"), Min: &min, Max: &max,
	})
	require.NoError(t, err)

	require.NoError(t, schema.ImportAttribute(imported, steward))

	assert.Equal(t, []string{"SubjectAttributeDefined"}, eventTypes(schema))
	defined := schema.GetUncommittedChanges()[0].(events.SubjectAttributeDefined)
	assert.Equal(t, 10.0, *defined.Max)
}

func TestLoadSubjectAttributeSchemaFromHistory_RebuildsState(t *testing.T) {
	schema, steward := newSchema(t)
	selection := defineSelection(t, schema, steward, "EU")
	optionID, err := schema.AddOption(selection, label(t, "US"), steward)
	require.NoError(t, err)
	number, err := schema.DefineAttribute(DefineAttributeParams{Name: name(t, "Score"), Type: attrType(t, "number")}, steward)
	require.NoError(t, err)
	min := 1.0
	require.NoError(t, schema.SetBounds(number, &min, nil, steward))
	help, _ := valueobjects.NewHelpText("help")
	require.NoError(t, schema.RenameAttribute(RenameAttributeParams{AttributeID: number, Name: name(t, "Rating"), HelpText: help}, steward))
	require.NoError(t, schema.RetireOption(selection, optionID, steward))
	require.NoError(t, schema.RetireAttribute(number, steward))

	history := append([]domain.DomainEvent{createdEventOf(t, schema)}, schema.GetUncommittedChanges()...)
	loaded, err := LoadSubjectAttributeSchemaFromHistory(history)

	require.NoError(t, err)
	assert.Equal(t, schema.ID(), loaded.ID())
	assert.Equal(t, schema.Version(), loaded.Version())
	assert.Equal(t, schema.Attributes(), loaded.Attributes())
	assert.Equal(t, schema.ModifiedBy(), loaded.ModifiedBy())
}

func createdEventOf(t *testing.T, schema *SubjectAttributeSchema) domain.DomainEvent {
	t.Helper()
	return events.NewSubjectAttributeSchemaCreated(events.CreateSchemaParams{
		ID:          schema.ID(),
		TenantID:    schema.TenantID().Value(),
		SubjectType: schema.SubjectType().Value(),
		CreatedBy:   "steward@example.com",
	})
}
