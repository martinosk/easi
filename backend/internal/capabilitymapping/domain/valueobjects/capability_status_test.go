package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCapabilityStatus_ValidValues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected CapabilityStatus
	}{
		{"Active", "Active", StatusActive},
		{"Planned", "Planned", StatusPlanned},
		{"Deprecated", "Deprecated", StatusDeprecated},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := NewCapabilityStatus(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, status)
		})
	}
}

func TestNewCapabilityStatus_TrimSpace(t *testing.T) {
	status, err := NewCapabilityStatus("  Active  ")
	assert.NoError(t, err)
	assert.Equal(t, StatusActive, status)
}

func TestNewCapabilityStatus_Empty(t *testing.T) {
	status, err := NewCapabilityStatus("")
	assert.NoError(t, err)
	assert.Equal(t, StatusActive, status)
}

func TestNewCapabilityStatus_InvalidValue(t *testing.T) {
	_, err := NewCapabilityStatus("InvalidStatus")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCapabilityStatus, err)
}

func TestCapabilityStatus_Value(t *testing.T) {
	status := StatusPlanned
	assert.Equal(t, "Planned", status.Value())
}

func TestCapabilityStatus_String(t *testing.T) {
	status := StatusDeprecated
	assert.Equal(t, "Deprecated", status.String())
}

func TestCapabilityStatus_Equals(t *testing.T) {
	status1 := StatusActive
	status2 := StatusActive
	status3 := StatusPlanned

	assert.True(t, status1.Equals(status2))
	assert.False(t, status1.Equals(status3))
}
