package valueobjects

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPillarName_ValidValue(t *testing.T) {
	name, err := NewPillarName("Digital Transformation")
	assert.NoError(t, err)
	assert.Equal(t, "Digital Transformation", name.Value())
}

func TestNewPillarName_Empty(t *testing.T) {
	_, err := NewPillarName("")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPillarNameEmpty)
}

func TestNewPillarName_WhitespaceOnly(t *testing.T) {
	_, err := NewPillarName("   ")
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPillarNameEmpty)
}

func TestNewPillarName_TrimsWhitespace(t *testing.T) {
	name, err := NewPillarName("  Customer Focus  ")
	assert.NoError(t, err)
	assert.Equal(t, "Customer Focus", name.Value())
}

func TestNewPillarName_MaxLength(t *testing.T) {
	text := strings.Repeat("a", 100)
	name, err := NewPillarName(text)
	assert.NoError(t, err)
	assert.Equal(t, 100, len(name.Value()))
}

func TestNewPillarName_ExceedsMaxLength(t *testing.T) {
	text := strings.Repeat("a", 101)
	_, err := NewPillarName(text)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrPillarNameTooLong)
}

func TestPillarName_Equals(t *testing.T) {
	n1, _ := NewPillarName("Digital")
	n2, _ := NewPillarName("Digital")
	n3, _ := NewPillarName("Innovation")

	assert.True(t, n1.Equals(n2))
	assert.False(t, n1.Equals(n3))
}

func TestPillarName_Equals_DifferentType(t *testing.T) {
	name, _ := NewPillarName("Digital")
	score, _ := NewFitScore(3)

	assert.False(t, name.Equals(score))
}

func TestPillarName_String(t *testing.T) {
	name, _ := NewPillarName("Digital Transformation")
	assert.Equal(t, "Digital Transformation", name.String())
}
