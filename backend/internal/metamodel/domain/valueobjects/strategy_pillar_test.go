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

func TestStrategyPillar_FitType_DefaultsToEmpty(t *testing.T) {
	pillar, _ := createTestPillar("Always On", "Core capabilities")

	assert.True(t, pillar.FitType().IsEmpty())
}

func TestStrategyPillar_WithFitType(t *testing.T) {
	pillar, _ := createTestPillar("Transformation", "Digital transformation")
	technicalFit, _ := NewFitType("TECHNICAL")

	updated := pillar.WithFitType(technicalFit)

	assert.Equal(t, "TECHNICAL", updated.FitType().Value())
	assert.True(t, pillar.FitType().IsEmpty())
}

func TestStrategyPillar_WithFitConfiguration_IncludesFitType(t *testing.T) {
	pillar, _ := createTestPillar("Grow", "Business growth")
	criteria, _ := NewFitCriteria("Scalability, market reach")
	functionalFit, _ := NewFitType("FUNCTIONAL")

	updated := pillar.WithFitConfiguration(true, criteria, functionalFit)

	assert.True(t, updated.FitScoringEnabled())
	assert.Equal(t, "Scalability, market reach", updated.FitCriteria().Value())
	assert.Equal(t, "FUNCTIONAL", updated.FitType().Value())
}

func TestStrategyPillar_WithFitConfiguration_PreservesOtherFields(t *testing.T) {
	pillar, _ := createTestPillar("Always On", "Core capabilities")
	criteria, _ := NewFitCriteria("Reliability criteria")
	technicalFit, _ := NewFitType("TECHNICAL")

	updated := pillar.WithFitConfiguration(true, criteria, technicalFit)

	assert.Equal(t, pillar.ID().Value(), updated.ID().Value())
	assert.Equal(t, pillar.Name().Value(), updated.Name().Value())
	assert.Equal(t, pillar.Description().Value(), updated.Description().Value())
	assert.Equal(t, pillar.IsActive(), updated.IsActive())
}

func TestStrategyPillar_Equals_IncludesFitType(t *testing.T) {
	pillar1, _ := createTestPillar("Always On", "Core capabilities")
	technicalFit, _ := NewFitType("TECHNICAL")
	functionalFit, _ := NewFitType("FUNCTIONAL")

	pillar1WithTech := pillar1.WithFitType(technicalFit)
	pillar1WithFunc := pillar1.WithFitType(functionalFit)

	assert.False(t, pillar1WithTech.Equals(pillar1WithFunc))
	assert.True(t, pillar1WithTech.Equals(pillar1WithTech))
}
