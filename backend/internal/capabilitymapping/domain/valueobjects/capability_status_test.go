package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCapabilityStatus_Valid(t *testing.T) {
	status, err := NewCapabilityStatus("Active")
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

func TestCapabilityStatus_Equals(t *testing.T) {
	status1 := StatusActive
	status2 := StatusActive
	status3 := StatusPlanned

	assert.True(t, status1.Equals(status2))
	assert.False(t, status1.Equals(status3))
}
