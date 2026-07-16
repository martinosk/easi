package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustFieldName(t *testing.T, v string) FieldName {
	t.Helper()
	name, err := NewFieldName(v)
	require.NoError(t, err)
	return name
}

func mustFieldType(t *testing.T, v string) FieldType {
	t.Helper()
	ft, err := NewFieldType(v)
	require.NoError(t, err)
	return ft
}

func mustOption(t *testing.T, label string) SelectionOption {
	t.Helper()
	optionLabel, err := NewOptionLabel(label)
	require.NoError(t, err)
	return NewSelectionOption(NewOptionID(), optionLabel)
}

func newFieldOfType(t *testing.T, name, typeName string) CustomField {
	t.Helper()
	field, err := NewCustomField(CustomFieldParams{
		ID:   NewFieldID(),
		Name: mustFieldName(t, name),
		Type: mustFieldType(t, typeName),
	})
	require.NoError(t, err)
	return field
}

func newTextField(t *testing.T, name string) CustomField {
	t.Helper()
	return newFieldOfType(t, name, "text")
}

func newSelectionField(t *testing.T, labels ...string) CustomField {
	t.Helper()
	options := make([]SelectionOption, len(labels))
	for i, l := range labels {
		options[i] = mustOption(t, l)
	}
	field, err := NewCustomField(CustomFieldParams{
		ID:      NewFieldID(),
		Name:    mustFieldName(t, "Hosting model"),
		Type:    mustFieldType(t, "selection"),
		Options: options,
	})
	require.NoError(t, err)
	return field
}

func TestNewCustomField_DefaultsToActiveOptional(t *testing.T) {
	field := newTextField(t, "Business summary")
	assert.True(t, field.IsActive())
	assert.False(t, field.IsRequired())
	assert.Equal(t, "Business summary", field.Name().Value())
	assert.Equal(t, "text", field.Type().Value())
}

func TestNewCustomField_SelectionRequiresAtLeastOneOption(t *testing.T) {
	_, err := NewCustomField(CustomFieldParams{
		ID:   NewFieldID(),
		Name: mustFieldName(t, "Hosting model"),
		Type: mustFieldType(t, "selection"),
	})
	assert.ErrorIs(t, err, ErrSelectionOptionRequired)
}

func TestNewCustomField_NonSelectionRejectsOptions(t *testing.T) {
	_, err := NewCustomField(CustomFieldParams{
		ID:      NewFieldID(),
		Name:    mustFieldName(t, "Annual cost"),
		Type:    mustFieldType(t, "number"),
		Options: []SelectionOption{mustOption(t, "One")},
	})
	assert.ErrorIs(t, err, ErrOptionsNotAllowed)
}

func TestNewCustomField_RejectsDuplicateOptionLabels(t *testing.T) {
	_, err := NewCustomField(CustomFieldParams{
		ID:      NewFieldID(),
		Name:    mustFieldName(t, "Hosting model"),
		Type:    mustFieldType(t, "selection"),
		Options: []SelectionOption{mustOption(t, "Cloud"), mustOption(t, "cloud")},
	})
	assert.ErrorIs(t, err, ErrDuplicateOptionLabel)
}

func TestCustomField_RenamedKeepsIdentityTypeAndRequirement(t *testing.T) {
	field := newTextField(t, "Contract")
	help, _ := NewHelpText("Link to the signed contract")

	renamed := field.Renamed(mustFieldName(t, "Contract link"), help)

	assert.Equal(t, field.ID().Value(), renamed.ID().Value())
	assert.Equal(t, field.Type(), renamed.Type())
	assert.Equal(t, field.IsRequired(), renamed.IsRequired())
	assert.Equal(t, "Contract link", renamed.Name().Value())
	assert.Equal(t, "Link to the signed contract", renamed.HelpText().Value())
}

func TestCustomField_WithRequirement(t *testing.T) {
	field := newTextField(t, "Product owner")
	required := field.WithRequirement(true)
	assert.True(t, required.IsRequired())
	assert.False(t, field.IsRequired())
}

func TestCustomField_RetiredAndReactivated(t *testing.T) {
	field := newSelectionField(t, "On-prem", "Cloud")
	retired := field.Retired()
	assert.False(t, retired.IsActive())

	reactivated := retired.Reactivated()
	assert.True(t, reactivated.IsActive())
	assert.Equal(t, field.ID().Value(), reactivated.ID().Value())
	assert.Len(t, reactivated.Options(), 2)
}

func TestCustomField_WithAddedOption(t *testing.T) {
	field := newSelectionField(t, "On-prem", "Cloud")

	updated, err := field.WithAddedOption(mustOption(t, "Hybrid"))
	require.NoError(t, err)
	assert.Len(t, updated.Options(), 3)
}

func TestCustomField_WithAddedOption_RejectsDuplicateActiveLabel(t *testing.T) {
	field := newSelectionField(t, "On-prem", "Cloud")
	_, err := field.WithAddedOption(mustOption(t, "cloud"))
	assert.ErrorIs(t, err, ErrDuplicateOptionLabel)
}

func TestCustomField_WithAddedOption_RejectsNonSelection(t *testing.T) {
	field := newTextField(t, "Business summary")
	_, err := field.WithAddedOption(mustOption(t, "Cloud"))
	assert.ErrorIs(t, err, ErrNotSelectionField)
}

func TestCustomField_WithRetiredOption_KeepsOptionOnDefinition(t *testing.T) {
	field := newSelectionField(t, "On-prem", "Cloud")
	target := field.Options()[0]

	updated, err := field.WithRetiredOption(target.ID())
	require.NoError(t, err)

	require.Len(t, updated.Options(), 2)
	assert.False(t, updated.Options()[0].IsActive())
	assert.True(t, updated.Options()[1].IsActive())
}

func TestCustomField_WithRetiredOption_RejectsUnknownOption(t *testing.T) {
	field := newSelectionField(t, "On-prem", "Cloud")
	_, err := field.WithRetiredOption(NewOptionID())
	assert.ErrorIs(t, err, ErrOptionNotFound)
}

func TestCustomField_WithRetiredOption_RejectsAlreadyRetired(t *testing.T) {
	field := newSelectionField(t, "On-prem", "Cloud")
	target := field.Options()[0]
	updated, err := field.WithRetiredOption(target.ID())
	require.NoError(t, err)

	_, err = updated.WithRetiredOption(target.ID())
	assert.ErrorIs(t, err, ErrOptionAlreadyRetired)
}

func TestCustomField_WithRetiredOption_RejectsLastActiveOption(t *testing.T) {
	field := newSelectionField(t, "On-prem")
	_, err := field.WithRetiredOption(field.Options()[0].ID())
	assert.ErrorIs(t, err, ErrLastActiveOption)
}

func TestCustomField_HasActiveOption(t *testing.T) {
	field := newSelectionField(t, "On-prem", "Cloud")
	target := field.Options()[0]
	updated, err := field.WithRetiredOption(target.ID())
	require.NoError(t, err)

	assert.False(t, updated.HasActiveOption(target.ID()))
	assert.True(t, updated.HasActiveOption(field.Options()[1].ID()))
}

func floatPtr(v float64) *float64 {
	return &v
}

func newNumberField(t *testing.T, name string) CustomField {
	t.Helper()
	return newFieldOfType(t, name, "number")
}

func TestNewCustomField_NumberAcceptsEitherBothOrNeitherBound(t *testing.T) {
	cases := []struct {
		name string
		min  *float64
		max  *float64
	}{
		{"both bounds", floatPtr(0), floatPtr(5)},
		{"minimum only", floatPtr(0), nil},
		{"maximum only", nil, floatPtr(100)},
		{"no bounds", nil, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			field, err := NewCustomField(CustomFieldParams{
				ID:   NewFieldID(),
				Name: mustFieldName(t, "Maturity score"),
				Type: mustFieldType(t, "number"),
				Min:  tc.min,
				Max:  tc.max,
			})
			require.NoError(t, err)
			assert.Equal(t, tc.min, field.Min())
			assert.Equal(t, tc.max, field.Max())
		})
	}
}

func TestNewCustomField_RejectsInvalidBounds(t *testing.T) {
	cases := []struct {
		name     string
		typeName string
		min      *float64
		max      *float64
		wantErr  error
	}{
		{"minimum exceeds maximum", "number", floatPtr(10), floatPtr(5), ErrMinExceedsMax},
		{"bounds on non-number field", "text", floatPtr(0), nil, ErrBoundsNotAllowed},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewCustomField(CustomFieldParams{
				ID:   NewFieldID(),
				Name: mustFieldName(t, "Maturity score"),
				Type: mustFieldType(t, tc.typeName),
				Min:  tc.min,
				Max:  tc.max,
			})
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestCustomField_WithBounds_UpdatesBoundsAndKeepsIdentity(t *testing.T) {
	field := newNumberField(t, "Maturity score")

	updated, err := field.WithBounds(floatPtr(0), floatPtr(5))

	require.NoError(t, err)
	assert.Equal(t, floatPtr(0), updated.Min())
	assert.Equal(t, floatPtr(5), updated.Max())
	assert.Equal(t, field.ID().Value(), updated.ID().Value())
	assert.Equal(t, field.Type(), updated.Type())
	assert.Equal(t, field.IsRequired(), updated.IsRequired())
	assert.Nil(t, field.Min())
}

func TestCustomField_WithBounds_CanClearBounds(t *testing.T) {
	field := newNumberField(t, "Headcount")
	withBounds, err := field.WithBounds(floatPtr(0), floatPtr(500))
	require.NoError(t, err)

	cleared, err := withBounds.WithBounds(floatPtr(0), nil)

	require.NoError(t, err)
	assert.Equal(t, floatPtr(0), cleared.Min())
	assert.Nil(t, cleared.Max())
}

func TestCustomField_WithBounds_RejectsMinimumGreaterThanMaximum(t *testing.T) {
	field := newNumberField(t, "Maturity score")

	_, err := field.WithBounds(floatPtr(10), floatPtr(5))

	assert.ErrorIs(t, err, ErrMinExceedsMax)
}

func TestCustomField_WithBounds_RejectsNonNumberField(t *testing.T) {
	field := newTextField(t, "Business summary")

	_, err := field.WithBounds(floatPtr(0), floatPtr(5))

	assert.ErrorIs(t, err, ErrBoundsNotAllowed)
}

func TestSelectionOption_Retired(t *testing.T) {
	option := mustOption(t, "On-prem")
	retired := option.Retired()
	assert.False(t, retired.IsActive())
	assert.True(t, option.IsActive())
	assert.Equal(t, option.ID().Value(), retired.ID().Value())
	assert.Equal(t, "On-prem", retired.Label().Value())
}
