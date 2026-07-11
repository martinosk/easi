package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuiltInFieldRef(t *testing.T) {
	ref, err := NewBuiltInFieldRef("maturity")
	require.NoError(t, err)
	assert.Equal(t, FieldRefKindBuiltIn, ref.Kind())
	assert.Equal(t, "maturity", ref.RefID())
}

func TestNewCustomFieldRef(t *testing.T) {
	id := NewFieldID()
	ref := NewCustomFieldRef(id)
	assert.Equal(t, FieldRefKindCustom, ref.Kind())
	assert.Equal(t, id.Value(), ref.RefID())
}

func TestNewBuiltInFieldRef_RejectsEmptyID(t *testing.T) {
	_, err := NewBuiltInFieldRef("")
	assert.ErrorIs(t, err, ErrFieldRefIDEmpty)
}

func TestNewFieldRef_ValidatesKind(t *testing.T) {
	ref, err := NewFieldRef("builtIn", "name")
	require.NoError(t, err)
	assert.Equal(t, FieldRefKindBuiltIn, ref.Kind())

	_, err = NewFieldRef("magic", "name")
	assert.ErrorIs(t, err, ErrInvalidFieldRefKind)
}

func TestFieldRef_Equals(t *testing.T) {
	a, _ := NewFieldRef("builtIn", "name")
	b, _ := NewFieldRef("builtIn", "name")
	c, _ := NewFieldRef("custom", "name")
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
