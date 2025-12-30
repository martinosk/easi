package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEnterpriseCapabilityLinkID(t *testing.T) {
	id := NewEnterpriseCapabilityLinkID()
	assert.NotEmpty(t, id.Value())
}

func TestNewEnterpriseCapabilityLinkIDFromString_Valid(t *testing.T) {
	id := NewEnterpriseCapabilityLinkID()
	parsed, err := NewEnterpriseCapabilityLinkIDFromString(id.Value())
	require.NoError(t, err)
	assert.Equal(t, id.Value(), parsed.Value())
}

func TestNewEnterpriseCapabilityLinkIDFromString_Empty(t *testing.T) {
	_, err := NewEnterpriseCapabilityLinkIDFromString("")
	assert.Error(t, err)
}

func TestNewEnterpriseCapabilityLinkIDFromString_Invalid(t *testing.T) {
	_, err := NewEnterpriseCapabilityLinkIDFromString("not-a-uuid")
	assert.Error(t, err)
}

func TestEnterpriseCapabilityLinkID_Equals(t *testing.T) {
	id := NewEnterpriseCapabilityLinkID()
	parsed, _ := NewEnterpriseCapabilityLinkIDFromString(id.Value())
	assert.True(t, id.Equals(parsed))
}
