package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEdgeType_Valid(t *testing.T) {
	edgeType, err := NewEdgeType("default")
	assert.NoError(t, err)
	assert.Equal(t, "default", edgeType.Value())
}

func TestNewEdgeType_Invalid(t *testing.T) {
	_, err := NewEdgeType("curved")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidEdgeType, err)
}

func TestNewEdgeType_EmptyString(t *testing.T) {
	_, err := NewEdgeType("")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidEdgeType, err)
}

func TestDefaultEdgeType(t *testing.T) {
	edgeType := DefaultEdgeType()
	assert.Equal(t, "default", edgeType.Value())
}

func TestEdgeType_Equals(t *testing.T) {
	edgeType1, _ := NewEdgeType("default")
	edgeType2, _ := NewEdgeType("default")
	edgeType3, _ := NewEdgeType("step")

	assert.True(t, edgeType1.Equals(edgeType2))
	assert.False(t, edgeType1.Equals(edgeType3))
}
