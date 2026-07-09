package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFieldType_AcceptsAllSixKinds(t *testing.T) {
	values := []string{"text", "number", "date", "link", "selection", "contact-person"}
	for _, v := range values {
		ft, err := NewFieldType(v)
		require.NoError(t, err, v)
		assert.Equal(t, v, ft.Value())
	}
}

func TestNewFieldType_RejectsUnknownKind(t *testing.T) {
	_, err := NewFieldType("checkbox")
	assert.ErrorIs(t, err, ErrInvalidFieldType)
}

func TestFieldType_IsSelection(t *testing.T) {
	selection, _ := NewFieldType("selection")
	text, _ := NewFieldType("text")
	assert.True(t, selection.IsSelection())
	assert.False(t, text.IsSelection())
}

func TestFieldType_Equals(t *testing.T) {
	a, _ := NewFieldType("link")
	b, _ := NewFieldType("link")
	c, _ := NewFieldType("date")
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
