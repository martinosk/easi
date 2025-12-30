package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEnterpriseCapabilityID(t *testing.T) {
	id := NewEnterpriseCapabilityID()
	assert.NotEmpty(t, id.Value())
}

func TestNewEnterpriseCapabilityIDFromString_Valid(t *testing.T) {
	id := NewEnterpriseCapabilityID()
	parsed, err := NewEnterpriseCapabilityIDFromString(id.Value())
	require.NoError(t, err)
	assert.Equal(t, id.Value(), parsed.Value())
}

func TestNewEnterpriseCapabilityIDFromString_Empty(t *testing.T) {
	_, err := NewEnterpriseCapabilityIDFromString("")
	assert.Error(t, err)
}

func TestNewEnterpriseCapabilityIDFromString_Invalid(t *testing.T) {
	_, err := NewEnterpriseCapabilityIDFromString("not-a-uuid")
	assert.Error(t, err)
}

func TestEnterpriseCapabilityID_Equals(t *testing.T) {
	id := NewEnterpriseCapabilityID()
	parsed, _ := NewEnterpriseCapabilityIDFromString(id.Value())
	assert.True(t, id.Equals(parsed))
}

func TestEnterpriseCapabilityID_NotEquals(t *testing.T) {
	id1 := NewEnterpriseCapabilityID()
	id2 := NewEnterpriseCapabilityID()
	assert.False(t, id1.Equals(id2))
}
