package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLayoutContextType_ValidValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected LayoutContextType
	}{
		{"ArchitectureCanvas", "architecture-canvas", ContextTypeArchitectureCanvas},
		{"BusinessDomainGrid", "business-domain-grid", ContextTypeBusinessDomainGrid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contextType, err := NewLayoutContextType(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, contextType)
		})
	}
}

func TestNewLayoutContextType_TrimSpace(t *testing.T) {
	contextType, err := NewLayoutContextType("  architecture-canvas  ")
	assert.NoError(t, err)
	assert.Equal(t, ContextTypeArchitectureCanvas, contextType)
}

func TestNewLayoutContextType_Empty(t *testing.T) {
	_, err := NewLayoutContextType("")
	assert.Error(t, err)
	assert.Equal(t, ErrEmptyContextType, err)
}

func TestNewLayoutContextType_InvalidValue(t *testing.T) {
	_, err := NewLayoutContextType("invalid-type")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidContextType, err)
}

func TestLayoutContextType_Value(t *testing.T) {
	contextType := ContextTypeArchitectureCanvas
	assert.Equal(t, "architecture-canvas", contextType.Value())
}

func TestLayoutContextType_String(t *testing.T) {
	contextType := ContextTypeBusinessDomainGrid
	assert.Equal(t, "business-domain-grid", contextType.String())
}

func TestLayoutContextType_Equals(t *testing.T) {
	type1 := ContextTypeArchitectureCanvas
	type2 := ContextTypeArchitectureCanvas
	type3 := ContextTypeBusinessDomainGrid

	assert.True(t, type1.Equals(type2))
	assert.False(t, type1.Equals(type3))
}
