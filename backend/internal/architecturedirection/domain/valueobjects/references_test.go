package valueobjects

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEnterpriseCapabilityRef_Valid(t *testing.T) {
	id := uuid.New().String()
	ref, err := NewEnterpriseCapabilityRef(id)
	require.NoError(t, err)
	assert.Equal(t, id, ref.Value())
}

func TestNewEnterpriseCapabilityRef_Invalid(t *testing.T) {
	_, err := NewEnterpriseCapabilityRef("xyz")
	assert.Error(t, err)
}

func TestNewPhysicalCapabilityRef_Valid(t *testing.T) {
	id := uuid.New().String()
	ref, err := NewPhysicalCapabilityRef(id)
	require.NoError(t, err)
	assert.Equal(t, id, ref.Value())
}

func TestNewPhysicalCapabilityRef_Invalid(t *testing.T) {
	_, err := NewPhysicalCapabilityRef("not-uuid")
	assert.Error(t, err)
}

func TestPhysicalCapabilityRef_Equals(t *testing.T) {
	id := uuid.New().String()
	a, _ := NewPhysicalCapabilityRef(id)
	b, _ := NewPhysicalCapabilityRef(id)
	c, _ := NewPhysicalCapabilityRef(uuid.New().String())
	assert.True(t, a.Equals(b))
	assert.False(t, a.Equals(c))
}
