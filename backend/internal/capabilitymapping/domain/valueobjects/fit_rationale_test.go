package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFitRationale_ValidValue(t *testing.T) {
	text := "This application has excellent alignment with our strategic goals."
	rationale, err := NewFitRationale(text)
	assert.NoError(t, err)
	assert.Equal(t, text, rationale.Value())
}

func TestNewFitRationale_Empty(t *testing.T) {
	rationale, err := NewFitRationale("")
	assert.NoError(t, err)
	assert.Equal(t, "", rationale.Value())
	assert.True(t, rationale.IsEmpty())
}

func TestNewFitRationale_WhitespaceOnly(t *testing.T) {
	rationale, err := NewFitRationale("   ")
	assert.NoError(t, err)
	assert.Equal(t, "", rationale.Value())
	assert.True(t, rationale.IsEmpty())
}

func TestNewFitRationale_TrimsWhitespace(t *testing.T) {
	rationale, err := NewFitRationale("  Strong alignment  ")
	assert.NoError(t, err)
	assert.Equal(t, "Strong alignment", rationale.Value())
}

func TestNewFitRationale_MaxLength(t *testing.T) {
	text := strings.Repeat("a", 500)
	rationale, err := NewFitRationale(text)
	assert.NoError(t, err)
	assert.Equal(t, 500, len(rationale.Value()))
}

func TestNewFitRationale_ExceedsMaxLength(t *testing.T) {
	text := strings.Repeat("a", 501)
	_, err := NewFitRationale(text)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFitRationaleTooLong)
}

func TestFitRationale_Equals(t *testing.T) {
	r1, _ := NewFitRationale("Strong alignment")
	r2, _ := NewFitRationale("Strong alignment")
	r3, _ := NewFitRationale("Weak alignment")

	assert.True(t, r1.Equals(r2))
	assert.False(t, r1.Equals(r3))
}

func TestFitRationale_Equals_DifferentType(t *testing.T) {
	rationale, _ := NewFitRationale("test")
	score, _ := NewFitScore(3)

	assert.False(t, rationale.Equals(score))
}

func TestFitRationale_String(t *testing.T) {
	rationale, _ := NewFitRationale("Strong alignment")
	assert.Equal(t, "Strong alignment", rationale.String())
}
