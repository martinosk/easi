package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFieldName_TrimsWhitespace(t *testing.T) {
	name, err := NewFieldName("  Contract link  ")
	require.NoError(t, err)
	assert.Equal(t, "Contract link", name.Value())
}

func TestNewFieldName_RejectsEmpty(t *testing.T) {
	_, err := NewFieldName("   ")
	assert.ErrorIs(t, err, ErrFieldNameEmpty)
}

func TestNewFieldName_RejectsTooLong(t *testing.T) {
	_, err := NewFieldName(strings.Repeat("a", 101))
	assert.ErrorIs(t, err, ErrFieldNameTooLong)
}

func TestFieldName_EqualsIgnoreCase(t *testing.T) {
	a, _ := NewFieldName("Contract Link")
	b, _ := NewFieldName("contract link")
	c, _ := NewFieldName("Contract")
	assert.True(t, a.EqualsIgnoreCase(b))
	assert.False(t, a.EqualsIgnoreCase(c))
}

func TestNewHelpText_AllowsEmpty(t *testing.T) {
	help, err := NewHelpText("")
	require.NoError(t, err)
	assert.Equal(t, "", help.Value())
}

func TestNewHelpText_TrimsAndCapsLength(t *testing.T) {
	help, err := NewHelpText("  guidance  ")
	require.NoError(t, err)
	assert.Equal(t, "guidance", help.Value())

	_, err = NewHelpText(strings.Repeat("a", 501))
	assert.ErrorIs(t, err, ErrHelpTextTooLong)
}

func TestNewOptionLabel_RejectsEmptyAndTooLong(t *testing.T) {
	_, err := NewOptionLabel(" ")
	assert.ErrorIs(t, err, ErrOptionLabelEmpty)

	_, err = NewOptionLabel(strings.Repeat("a", 101))
	assert.ErrorIs(t, err, ErrOptionLabelTooLong)

	label, err := NewOptionLabel(" On-prem ")
	require.NoError(t, err)
	assert.Equal(t, "On-prem", label.Value())
}
