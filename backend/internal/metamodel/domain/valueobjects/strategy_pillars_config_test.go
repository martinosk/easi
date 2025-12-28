package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createPillarWithName(name string) StrategyPillar {
	id := NewStrategyPillarID()
	pillarName, _ := NewPillarName(name)
	desc, _ := NewPillarDescription("")
	pillar, _ := NewStrategyPillar(id, pillarName, desc)
	return pillar
}

func TestNewStrategyPillarsConfig_Valid(t *testing.T) {
	pillars := []StrategyPillar{
		createPillarWithName("Always On"),
		createPillarWithName("Grow"),
		createPillarWithName("Transform"),
	}

	config, err := NewStrategyPillarsConfig(pillars)

	require.NoError(t, err)
	assert.Equal(t, 3, len(config.Pillars()))
	assert.Equal(t, 3, config.CountActive())
}

func TestNewStrategyPillarsConfig_EmptyIsValid(t *testing.T) {
	config, err := NewStrategyPillarsConfig([]StrategyPillar{})

	require.NoError(t, err)
	assert.Equal(t, 0, len(config.Pillars()))
}

func TestNewStrategyPillarsConfig_TooManyPillars(t *testing.T) {
	pillars := make([]StrategyPillar, 21)
	for i := 0; i < 21; i++ {
		pillars[i] = createPillarWithName("Pillar" + string(rune('A'+i)))
	}

	_, err := NewStrategyPillarsConfig(pillars)

	assert.Error(t, err)
	assert.Equal(t, ErrTooManyPillars, err)
}

func TestNewStrategyPillarsConfig_MaxPillarsAllowed(t *testing.T) {
	pillars := make([]StrategyPillar, 20)
	for i := 0; i < 20; i++ {
		pillars[i] = createPillarWithName("Pillar" + string(rune('A'+i)))
	}

	config, err := NewStrategyPillarsConfig(pillars)

	require.NoError(t, err)
	assert.Equal(t, 20, len(config.Pillars()))
}

func TestNewStrategyPillarsConfig_DuplicateNames(t *testing.T) {
	pillars := []StrategyPillar{
		createPillarWithName("Always On"),
		createPillarWithName("always on"),
	}

	_, err := NewStrategyPillarsConfig(pillars)

	assert.Error(t, err)
	assert.Equal(t, ErrPillarNameDuplicate, err)
}

func TestDefaultStrategyPillarsConfig(t *testing.T) {
	config := DefaultStrategyPillarsConfig()

	pillars := config.Pillars()
	assert.Equal(t, 3, len(pillars))
	assert.Equal(t, "Always On", pillars[0].Name().Value())
	assert.Equal(t, "Grow", pillars[1].Name().Value())
	assert.Equal(t, "Transform", pillars[2].Name().Value())
	assert.Equal(t, 3, config.CountActive())
}

func TestStrategyPillarsConfig_ActivePillars(t *testing.T) {
	pillar1 := createPillarWithName("Active1")
	pillar2 := createPillarWithName("Active2")
	inactive := createPillarWithName("Inactive").Deactivate()

	config, _ := NewStrategyPillarsConfig([]StrategyPillar{pillar1, inactive, pillar2})

	active := config.ActivePillars()
	assert.Equal(t, 2, len(active))
	assert.Equal(t, "Active1", active[0].Name().Value())
	assert.Equal(t, "Active2", active[1].Name().Value())
}

func TestStrategyPillarsConfig_FindByID(t *testing.T) {
	pillar1 := createPillarWithName("First")
	pillar2 := createPillarWithName("Second")
	config, _ := NewStrategyPillarsConfig([]StrategyPillar{pillar1, pillar2})

	found, ok := config.FindByID(pillar1.ID())

	assert.True(t, ok)
	assert.Equal(t, "First", found.Name().Value())
}

func TestStrategyPillarsConfig_FindByID_NotFound(t *testing.T) {
	config, _ := NewStrategyPillarsConfig([]StrategyPillar{createPillarWithName("First")})

	_, ok := config.FindByID(NewStrategyPillarID())

	assert.False(t, ok)
}

func TestStrategyPillarsConfig_HasPillarWithName(t *testing.T) {
	pillar := createPillarWithName("Existing")
	config, _ := NewStrategyPillarsConfig([]StrategyPillar{pillar})

	nameExists, _ := NewPillarName("Existing")
	nameCaseInsensitive, _ := NewPillarName("EXISTING")
	nameNotExists, _ := NewPillarName("NotExisting")

	assert.True(t, config.HasPillarWithName(nameExists))
	assert.True(t, config.HasPillarWithName(nameCaseInsensitive))
	assert.False(t, config.HasPillarWithName(nameNotExists))
}

func TestStrategyPillarsConfig_WithAddedPillar(t *testing.T) {
	config := DefaultStrategyPillarsConfig()
	newPillar := createPillarWithName("Innovation")

	updated, err := config.WithAddedPillar(newPillar)

	require.NoError(t, err)
	assert.Equal(t, 4, len(updated.Pillars()))
	assert.Equal(t, "Innovation", updated.Pillars()[3].Name().Value())
	assert.Equal(t, 3, len(config.Pillars()))
}

func TestStrategyPillarsConfig_WithAddedPillar_DuplicateName(t *testing.T) {
	config := DefaultStrategyPillarsConfig()
	duplicate := createPillarWithName("always on")

	_, err := config.WithAddedPillar(duplicate)

	assert.Error(t, err)
	assert.Equal(t, ErrPillarNameDuplicate, err)
}

func TestStrategyPillarsConfig_WithAddedPillar_MaxReached(t *testing.T) {
	pillars := make([]StrategyPillar, 20)
	for i := 0; i < 20; i++ {
		pillars[i] = createPillarWithName("Pillar" + string(rune('A'+i)))
	}
	config, _ := NewStrategyPillarsConfig(pillars)

	_, err := config.WithAddedPillar(createPillarWithName("OneMore"))

	assert.Error(t, err)
	assert.Equal(t, ErrTooManyPillars, err)
}

func TestStrategyPillarsConfig_WithUpdatedPillar(t *testing.T) {
	pillar := createPillarWithName("Original")
	config, _ := NewStrategyPillarsConfig([]StrategyPillar{pillar})
	newName, _ := NewPillarName("Updated")
	newDesc, _ := NewPillarDescription("New description")

	updated, err := config.WithUpdatedPillar(pillar.ID(), newName, newDesc)

	require.NoError(t, err)
	found, _ := updated.FindByID(pillar.ID())
	assert.Equal(t, "Updated", found.Name().Value())
	assert.Equal(t, "New description", found.Description().Value())
}

func TestStrategyPillarsConfig_WithUpdatedPillar_NotFound(t *testing.T) {
	config := DefaultStrategyPillarsConfig()
	newName, _ := NewPillarName("Updated")
	newDesc, _ := NewPillarDescription("")

	_, err := config.WithUpdatedPillar(NewStrategyPillarID(), newName, newDesc)

	assert.Error(t, err)
	assert.Equal(t, ErrPillarNotFound, err)
}

func TestStrategyPillarsConfig_WithUpdatedPillar_DuplicateName(t *testing.T) {
	pillar1 := createPillarWithName("First")
	pillar2 := createPillarWithName("Second")
	config, _ := NewStrategyPillarsConfig([]StrategyPillar{pillar1, pillar2})
	duplicateName, _ := NewPillarName("First")
	desc, _ := NewPillarDescription("")

	_, err := config.WithUpdatedPillar(pillar2.ID(), duplicateName, desc)

	assert.Error(t, err)
	assert.Equal(t, ErrPillarNameDuplicate, err)
}

func TestStrategyPillarsConfig_WithUpdatedPillar_SameNameAllowed(t *testing.T) {
	pillar := createPillarWithName("Original")
	config, _ := NewStrategyPillarsConfig([]StrategyPillar{pillar})
	sameName, _ := NewPillarName("Original")
	newDesc, _ := NewPillarDescription("Updated description only")

	updated, err := config.WithUpdatedPillar(pillar.ID(), sameName, newDesc)

	require.NoError(t, err)
	found, _ := updated.FindByID(pillar.ID())
	assert.Equal(t, "Updated description only", found.Description().Value())
}

func TestStrategyPillarsConfig_WithRemovedPillar(t *testing.T) {
	pillar1 := createPillarWithName("First")
	pillar2 := createPillarWithName("Second")
	config, _ := NewStrategyPillarsConfig([]StrategyPillar{pillar1, pillar2})

	updated, err := config.WithRemovedPillar(pillar1.ID())

	require.NoError(t, err)
	found, ok := updated.FindByID(pillar1.ID())
	assert.True(t, ok)
	assert.False(t, found.IsActive())
	assert.Equal(t, 1, updated.CountActive())
}

func TestStrategyPillarsConfig_WithRemovedPillar_NotFound(t *testing.T) {
	config := DefaultStrategyPillarsConfig()

	_, err := config.WithRemovedPillar(NewStrategyPillarID())

	assert.Error(t, err)
	assert.Equal(t, ErrPillarNotFound, err)
}

func TestStrategyPillarsConfig_WithRemovedPillar_LastActive(t *testing.T) {
	pillar := createPillarWithName("OnlyOne")
	config, _ := NewStrategyPillarsConfig([]StrategyPillar{pillar})

	_, err := config.WithRemovedPillar(pillar.ID())

	assert.Error(t, err)
	assert.Equal(t, ErrCannotRemoveLastActivePillar, err)
}

func TestStrategyPillarsConfig_WithRemovedPillar_AlreadyInactive(t *testing.T) {
	active := createPillarWithName("Active")
	inactive := createPillarWithName("Inactive").Deactivate()
	config, _ := NewStrategyPillarsConfig([]StrategyPillar{active, inactive})

	_, err := config.WithRemovedPillar(inactive.ID())

	assert.Error(t, err)
	assert.Equal(t, ErrPillarAlreadyInactive, err)
}
