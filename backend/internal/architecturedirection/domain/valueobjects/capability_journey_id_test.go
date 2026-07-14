package valueobjects

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCapabilityJourneyID_GeneratesUniqueValue(t *testing.T) {
	a := NewCapabilityJourneyID()
	b := NewCapabilityJourneyID()
	assert.NotEmpty(t, a.Value())
	assert.NotEqual(t, a.Value(), b.Value())
}

func TestNewCapabilityJourneyIDFromString_Valid(t *testing.T) {
	id := uuid.New().String()
	journeyID, err := NewCapabilityJourneyIDFromString(id)
	require.NoError(t, err)
	assert.Equal(t, id, journeyID.Value())
}

func TestNewCapabilityJourneyIDFromString_Invalid(t *testing.T) {
	_, err := NewCapabilityJourneyIDFromString("not-a-uuid")
	assert.Error(t, err)
}

func TestCapabilityJourneyID_Equals(t *testing.T) {
	id := uuid.New().String()
	a, _ := NewCapabilityJourneyIDFromString(id)
	b, _ := NewCapabilityJourneyIDFromString(id)
	c := NewCapabilityJourneyID()
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
