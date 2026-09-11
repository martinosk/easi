package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mustAttributeName(t *testing.T, v string) AttributeName {
	t.Helper()
	name, err := NewAttributeName(v)
	require.NoError(t, err)
	return name
}

func mustAttributeType(t *testing.T, v string) AttributeType {
	t.Helper()
	at, err := NewAttributeType(v)
	require.NoError(t, err)
	return at
}

func mustOptionLabel(t *testing.T, v string) OptionLabel {
	t.Helper()
	label, err := NewOptionLabel(v)
	require.NoError(t, err)
	return label
}

func TestNewSubjectType_AcceptsEveryOnePagerSubjectType(t *testing.T) {
	for _, value := range []string{"capability", "application", "acquired-entity", "vendor", "internal-team"} {
		st, err := NewSubjectType(value)
		require.NoError(t, err)
		assert.Equal(t, value, st.Value())
	}
	_, err := NewSubjectType("spaceship")
	assert.ErrorIs(t, err, ErrInvalidSubjectType)
}

func TestNewAttributeType_AcceptsSchemaTypesOnly(t *testing.T) {
	for _, value := range []string{"text", "number", "date", "link", "selection", "contact-person"} {
		_, err := NewAttributeType(value)
		require.NoError(t, err)
	}
	_, err := NewAttributeType("boolean")
	assert.ErrorIs(t, err, ErrInvalidAttributeType)
	assert.True(t, mustAttributeType(t, "selection").IsSelection())
	assert.True(t, mustAttributeType(t, "number").IsNumber())
}

func TestNewAttributeName_TrimsAndValidates(t *testing.T) {
	name, err := NewAttributeName("  Hosting Region ")
	require.NoError(t, err)
	assert.Equal(t, "Hosting Region", name.Value())
	assert.True(t, name.EqualsIgnoreCase(mustAttributeName(t, "hosting region")))

	_, err = NewAttributeName("   ")
	assert.ErrorIs(t, err, ErrAttributeNameEmpty)
}

func TestNewSubjectAttribute_RejectsShapeMismatchingType(t *testing.T) {
	min := 1.0
	cases := []struct {
		name    string
		params  SubjectAttributeParams
		wantErr error
	}{
		{"selection without options", SubjectAttributeParams{Name: mustAttributeName(t, "Region"), Type: mustAttributeType(t, "selection")}, ErrSelectionOptionRequired},
		{"options on a text attribute", SubjectAttributeParams{Name: mustAttributeName(t, "Summary"), Type: mustAttributeType(t, "text"), Options: []AttributeOption{NewAttributeOption(NewOptionID(), mustOptionLabel(t, "A"))}}, ErrOptionsNotAllowed},
		{"bounds on a text attribute", SubjectAttributeParams{Name: mustAttributeName(t, "Summary"), Type: mustAttributeType(t, "text"), Min: &min}, ErrBoundsNotAllowed},
		{"duplicate option labels", SubjectAttributeParams{Name: mustAttributeName(t, "Region"), Type: mustAttributeType(t, "selection"), Options: []AttributeOption{NewAttributeOption(NewOptionID(), mustOptionLabel(t, "EU")), NewAttributeOption(NewOptionID(), mustOptionLabel(t, "eu"))}}, ErrDuplicateOptionLabel},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.params.ID = NewAttributeID()

			_, err := NewSubjectAttribute(tc.params)

			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestSubjectAttribute_OptionLifecycle(t *testing.T) {
	first := NewAttributeOption(NewOptionID(), mustOptionLabel(t, "EU"))
	attribute, err := NewSubjectAttribute(SubjectAttributeParams{
		ID:      NewAttributeID(),
		Name:    mustAttributeName(t, "Region"),
		Type:    mustAttributeType(t, "selection"),
		Options: []AttributeOption{first},
	})
	require.NoError(t, err)

	_, err = attribute.WithAddedOption(NewAttributeOption(NewOptionID(), mustOptionLabel(t, "eu")))
	assert.ErrorIs(t, err, ErrDuplicateOptionLabel)

	_, err = attribute.WithRetiredOption(first.ID())
	assert.ErrorIs(t, err, ErrLastActiveOption)

	second := NewAttributeOption(NewOptionID(), mustOptionLabel(t, "US"))
	attribute, err = attribute.WithAddedOption(second)
	require.NoError(t, err)
	attribute, err = attribute.WithRetiredOption(first.ID())
	require.NoError(t, err)
	assert.False(t, attribute.Options()[0].IsActive())
	assert.True(t, attribute.HasActiveOption(second.ID()))
	assert.False(t, attribute.HasActiveOption(first.ID()))
}

func TestSubjectAttribute_BoundsMustBeOrdered(t *testing.T) {
	attribute, err := NewSubjectAttribute(SubjectAttributeParams{
		ID:   NewAttributeID(),
		Name: mustAttributeName(t, "Score"),
		Type: mustAttributeType(t, "number"),
	})
	require.NoError(t, err)
	min, max := 10.0, 5.0
	_, err = attribute.WithBounds(&min, &max)
	assert.ErrorIs(t, err, ErrMinExceedsMax)

	max = 20.0
	bounded, err := attribute.WithBounds(&min, &max)
	require.NoError(t, err)
	assert.Equal(t, 10.0, *bounded.Min())
	assert.Equal(t, 20.0, *bounded.Max())
	assert.Nil(t, attribute.Min())
}

func TestSubjectAttribute_RetireAndReactivate(t *testing.T) {
	attribute, err := NewSubjectAttribute(SubjectAttributeParams{
		ID:   NewAttributeID(),
		Name: mustAttributeName(t, "Score"),
		Type: mustAttributeType(t, "number"),
	})
	require.NoError(t, err)
	assert.True(t, attribute.IsActive())
	assert.False(t, attribute.Retired().IsActive())
	assert.True(t, attribute.Retired().Reactivated().IsActive())
}
