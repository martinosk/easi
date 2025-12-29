package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRationale_ValidValue(t *testing.T) {
	text := "This capability is critical for our digital transformation strategy."
	rationale, err := NewRationale(text)
	assert.NoError(t, err)
	assert.Equal(t, text, rationale.Value())
}

func TestNewRationale_Empty(t *testing.T) {
	rationale, err := NewRationale("")
	assert.NoError(t, err)
	assert.Equal(t, "", rationale.Value())
	assert.True(t, rationale.IsEmpty())
}

func TestNewRationale_WhitespaceOnly(t *testing.T) {
	rationale, err := NewRationale("   ")
	assert.NoError(t, err)
	assert.Equal(t, "", rationale.Value())
	assert.True(t, rationale.IsEmpty())
}

func TestNewRationale_TrimsWhitespace(t *testing.T) {
	rationale, err := NewRationale("  Important for growth  ")
	assert.NoError(t, err)
	assert.Equal(t, "Important for growth", rationale.Value())
}

func TestNewRationale_MaxLength(t *testing.T) {
	text := strings.Repeat("a", 500)
	rationale, err := NewRationale(text)
	assert.NoError(t, err)
	assert.Equal(t, 500, len(rationale.Value()))
}

func TestNewRationale_ExceedsMaxLength(t *testing.T) {
	text := strings.Repeat("a", 501)
	_, err := NewRationale(text)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrRationaleTooLong)
}

func TestRationale_Equals(t *testing.T) {
	r1, _ := NewRationale("Important")
	r2, _ := NewRationale("Important")
	r3, _ := NewRationale("Critical")

	assert.True(t, r1.Equals(r2))
	assert.False(t, r1.Equals(r3))
}

func TestEmptyRationale(t *testing.T) {
	rationale := EmptyRationale()
	assert.True(t, rationale.IsEmpty())
	assert.Equal(t, "", rationale.Value())
}
