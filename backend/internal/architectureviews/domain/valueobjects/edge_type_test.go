package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEdgeType_Default(t *testing.T) {
	edgeType, err := NewEdgeType("default")
	assert.NoError(t, err)
	assert.Equal(t, "default", edgeType.Value())
}

func TestNewEdgeType_Step(t *testing.T) {
	edgeType, err := NewEdgeType("step")
	assert.NoError(t, err)
	assert.Equal(t, "step", edgeType.Value())
}

func TestNewEdgeType_SmoothStep(t *testing.T) {
	edgeType, err := NewEdgeType("smoothstep")
	assert.NoError(t, err)
	assert.Equal(t, "smoothstep", edgeType.Value())
}

func TestNewEdgeType_Straight(t *testing.T) {
	edgeType, err := NewEdgeType("straight")
	assert.NoError(t, err)
	assert.Equal(t, "straight", edgeType.Value())
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

func TestNewEdgeType_CaseSensitive(t *testing.T) {
	_, err := NewEdgeType("DEFAULT")
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

func TestEdgeType_String(t *testing.T) {
	edgeType, _ := NewEdgeType("smoothstep")
	assert.Equal(t, "smoothstep", edgeType.String())
}
