package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEnterpriseStrategicImportanceID(t *testing.T) {
	id := NewEnterpriseStrategicImportanceID()
	assert.NotEmpty(t, id.Value())
}

func TestNewEnterpriseStrategicImportanceIDFromString_Valid(t *testing.T) {
	id := NewEnterpriseStrategicImportanceID()
	parsed, err := NewEnterpriseStrategicImportanceIDFromString(id.Value())
	require.NoError(t, err)
	assert.Equal(t, id.Value(), parsed.Value())
}

func TestNewEnterpriseStrategicImportanceIDFromString_Empty(t *testing.T) {
	_, err := NewEnterpriseStrategicImportanceIDFromString("")
	assert.Error(t, err)
}

func TestNewEnterpriseStrategicImportanceIDFromString_Invalid(t *testing.T) {
	_, err := NewEnterpriseStrategicImportanceIDFromString("not-a-uuid")
	assert.Error(t, err)
}

func TestEnterpriseStrategicImportanceID_Equals(t *testing.T) {
	id := NewEnterpriseStrategicImportanceID()
	parsed, _ := NewEnterpriseStrategicImportanceIDFromString(id.Value())
	assert.True(t, id.Equals(parsed))
}

func TestNewEnterpriseStrategicImportanceIDFromComposite_IsDeterministic(t *testing.T) {
	capabilityID := NewEnterpriseCapabilityID()
	pillarID := NewPillarID()

	id1 := NewEnterpriseStrategicImportanceIDFromComposite(capabilityID, pillarID)
	id2 := NewEnterpriseStrategicImportanceIDFromComposite(capabilityID, pillarID)

	assert.Equal(t, id1.Value(), id2.Value())
	assert.True(t, id1.Equals(id2))
}

func TestNewEnterpriseStrategicImportanceIDFromComposite_DifferentInputsProduceDifferentIDs(t *testing.T) {
	capabilityID1 := NewEnterpriseCapabilityID()
	capabilityID2 := NewEnterpriseCapabilityID()
	pillarID := NewPillarID()

	id1 := NewEnterpriseStrategicImportanceIDFromComposite(capabilityID1, pillarID)
	id2 := NewEnterpriseStrategicImportanceIDFromComposite(capabilityID2, pillarID)

	assert.NotEqual(t, id1.Value(), id2.Value())
	assert.False(t, id1.Equals(id2))
}

func TestNewEnterpriseStrategicImportanceIDFromComposite_DifferentPillarsProduceDifferentIDs(t *testing.T) {
	capabilityID := NewEnterpriseCapabilityID()
	pillarID1 := NewPillarID()
	pillarID2 := NewPillarID()

	id1 := NewEnterpriseStrategicImportanceIDFromComposite(capabilityID, pillarID1)
	id2 := NewEnterpriseStrategicImportanceIDFromComposite(capabilityID, pillarID2)

	assert.NotEqual(t, id1.Value(), id2.Value())
	assert.False(t, id1.Equals(id2))
}
