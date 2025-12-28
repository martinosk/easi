package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestPillar(name, description string) (StrategyPillar, error) {
	id := NewStrategyPillarID()
	pillarName, err := NewPillarName(name)
	if err != nil {
		return StrategyPillar{}, err
	}
	pillarDesc, err := NewPillarDescription(description)
	if err != nil {
		return StrategyPillar{}, err
	}
	return NewStrategyPillar(id, pillarName, pillarDesc)
}

func TestNewStrategyPillar_Valid(t *testing.T) {
	id := NewStrategyPillarID()
	name, _ := NewPillarName("Always On")
	desc, _ := NewPillarDescription("Core capabilities")

	pillar, err := NewStrategyPillar(id, name, desc)

	require.NoError(t, err)
	assert.Equal(t, id.Value(), pillar.ID().Value())
	assert.Equal(t, "Always On", pillar.Name().Value())
	assert.Equal(t, "Core capabilities", pillar.Description().Value())
	assert.True(t, pillar.IsActive())
}

func TestNewStrategyPillar_EmptyDescription(t *testing.T) {
	id := NewStrategyPillarID()
	name, _ := NewPillarName("Always On")
	desc, _ := NewPillarDescription("")

	pillar, err := NewStrategyPillar(id, name, desc)

	require.NoError(t, err)
	assert.True(t, pillar.Description().IsEmpty())
	assert.True(t, pillar.IsActive())
}

func TestStrategyPillar_WithUpdatedDetails(t *testing.T) {
	pillar, _ := createTestPillar("Original", "Original description")
	newName, _ := NewPillarName("Updated")
	newDesc, _ := NewPillarDescription("Updated description")

	updated, err := pillar.WithUpdatedDetails(newName, newDesc)

	require.NoError(t, err)
	assert.Equal(t, pillar.ID().Value(), updated.ID().Value())
	assert.Equal(t, "Updated", updated.Name().Value())
	assert.Equal(t, "Updated description", updated.Description().Value())
	assert.True(t, updated.IsActive())
}

func TestStrategyPillar_Deactivate(t *testing.T) {
	pillar, _ := createTestPillar("Always On", "Core capabilities")
	assert.True(t, pillar.IsActive())

	deactivated := pillar.Deactivate()

	assert.Equal(t, pillar.ID().Value(), deactivated.ID().Value())
	assert.Equal(t, pillar.Name().Value(), deactivated.Name().Value())
	assert.False(t, deactivated.IsActive())
	assert.True(t, pillar.IsActive())
}
