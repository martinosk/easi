package aggregates

import (
	"testing"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCapabilityDependency(t *testing.T) {
	sourceID := valueobjects.NewCapabilityID()
	targetID := valueobjects.NewCapabilityID()

	dependencyType, err := valueobjects.NewDependencyType("Requires")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Payment processing requires customer management")

	dependency, err := NewCapabilityDependency(sourceID, targetID, dependencyType, description)
	require.NoError(t, err)
	assert.NotNil(t, dependency)
	assert.NotEmpty(t, dependency.ID())
	assert.Equal(t, sourceID, dependency.SourceCapabilityID())
	assert.Equal(t, targetID, dependency.TargetCapabilityID())
	assert.Equal(t, dependencyType, dependency.DependencyType())
	assert.Equal(t, description, dependency.Description())
	assert.NotZero(t, dependency.CreatedAt())
	assert.Len(t, dependency.GetUncommittedChanges(), 1)
}

func TestNewCapabilityDependency_SelfDependency_ShouldFail(t *testing.T) {
	sourceID := valueobjects.NewCapabilityID()

	dependencyType, err := valueobjects.NewDependencyType("Requires")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Test")

	dependency, err := NewCapabilityDependency(sourceID, sourceID, dependencyType, description)
	assert.Error(t, err)
	assert.Nil(t, dependency)
	assert.Equal(t, ErrCannotCreateSelfDependency, err)
}

func TestCapabilityDependency_RaisesCreatedEvent(t *testing.T) {
	sourceID := valueobjects.NewCapabilityID()
	targetID := valueobjects.NewCapabilityID()

	dependencyType, err := valueobjects.NewDependencyType("Enables")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Digital channels enable customer engagement")

	dependency, err := NewCapabilityDependency(sourceID, targetID, dependencyType, description)
	require.NoError(t, err)

	uncommittedEvents := dependency.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "CapabilityDependencyCreated", uncommittedEvents[0].EventType())
}

func TestCapabilityDependency_Delete(t *testing.T) {
	sourceID := valueobjects.NewCapabilityID()
	targetID := valueobjects.NewCapabilityID()

	dependencyType, err := valueobjects.NewDependencyType("Supports")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Analytics supports decision making")

	dependency, err := NewCapabilityDependency(sourceID, targetID, dependencyType, description)
	require.NoError(t, err)

	dependency.MarkChangesAsCommitted()

	err = dependency.Delete()
	require.NoError(t, err)

	uncommittedEvents := dependency.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "CapabilityDependencyDeleted", uncommittedEvents[0].EventType())
}

func TestCapabilityDependency_LoadFromHistory(t *testing.T) {
	sourceID := valueobjects.NewCapabilityID()
	targetID := valueobjects.NewCapabilityID()

	dependencyType, err := valueobjects.NewDependencyType("Requires")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Test dependency")

	dependency, err := NewCapabilityDependency(sourceID, targetID, dependencyType, description)
	require.NoError(t, err)

	events := dependency.GetUncommittedChanges()

	loadedDependency, err := LoadCapabilityDependencyFromHistory(events)
	require.NoError(t, err)
	assert.NotNil(t, loadedDependency)
	assert.Equal(t, dependency.ID(), loadedDependency.ID())
	assert.Equal(t, dependency.SourceCapabilityID().Value(), loadedDependency.SourceCapabilityID().Value())
	assert.Equal(t, dependency.TargetCapabilityID().Value(), loadedDependency.TargetCapabilityID().Value())
	assert.Equal(t, dependency.DependencyType().Value(), loadedDependency.DependencyType().Value())
	assert.Equal(t, dependency.Description().Value(), loadedDependency.Description().Value())
}

func TestCapabilityDependency_AllDependencyTypes(t *testing.T) {
	tests := []struct {
		name           string
		dependencyType string
	}{
		{"Requires type", "Requires"},
		{"Enables type", "Enables"},
		{"Supports type", "Supports"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceID := valueobjects.NewCapabilityID()
			targetID := valueobjects.NewCapabilityID()

			dependencyType, err := valueobjects.NewDependencyType(tt.dependencyType)
			require.NoError(t, err)

			description := valueobjects.MustNewDescription("Test dependency")

			dependency, err := NewCapabilityDependency(sourceID, targetID, dependencyType, description)
			require.NoError(t, err)
			assert.NotNil(t, dependency)
			assert.Equal(t, tt.dependencyType, dependency.DependencyType().Value())
		})
	}
}

func TestValidateNewDependency_Valid(t *testing.T) {
	sourceID := valueobjects.NewCapabilityID()
	targetID := valueobjects.NewCapabilityID()

	err := ValidateNewDependency(sourceID, targetID, nil)
	assert.NoError(t, err)
}

func TestValidateNewDependency_SelfDependency_Fails(t *testing.T) {
	sourceID := valueobjects.NewCapabilityID()

	err := ValidateNewDependency(sourceID, sourceID, nil)
	assert.Error(t, err)
	assert.Equal(t, ErrCannotCreateSelfDependency, err)
}

func TestValidateNewDependency_DuplicateDependency_Fails(t *testing.T) {
	sourceID := valueobjects.NewCapabilityID()
	targetID := valueobjects.NewCapabilityID()

	existingDeps := []ExistingDependency{
		{SourceID: sourceID, TargetID: targetID},
	}

	err := ValidateNewDependency(sourceID, targetID, existingDeps)
	assert.Error(t, err)
	assert.Equal(t, ErrDuplicateDependencyExists, err)
}

func TestValidateNewDependency_DirectCircularDependency_Fails(t *testing.T) {
	capA := valueobjects.NewCapabilityID()
	capB := valueobjects.NewCapabilityID()

	existingDeps := []ExistingDependency{
		{SourceID: capA, TargetID: capB},
	}

	err := ValidateNewDependency(capB, capA, existingDeps)
	assert.Error(t, err)
	assert.Equal(t, ErrCircularDependencyDetected, err)
}

func TestValidateNewDependency_IndirectCircularDependency_Fails(t *testing.T) {
	capA := valueobjects.NewCapabilityID()
	capB := valueobjects.NewCapabilityID()
	capC := valueobjects.NewCapabilityID()

	existingDeps := []ExistingDependency{
		{SourceID: capA, TargetID: capB},
		{SourceID: capB, TargetID: capC},
	}

	err := ValidateNewDependency(capC, capA, existingDeps)
	assert.Error(t, err)
	assert.Equal(t, ErrCircularDependencyDetected, err)
}

func TestValidateNewDependency_NoCircularWithDifferentTarget(t *testing.T) {
	capA := valueobjects.NewCapabilityID()
	capB := valueobjects.NewCapabilityID()
	capC := valueobjects.NewCapabilityID()
	capD := valueobjects.NewCapabilityID()

	existingDeps := []ExistingDependency{
		{SourceID: capA, TargetID: capB},
		{SourceID: capB, TargetID: capC},
	}

	err := ValidateNewDependency(capD, capA, existingDeps)
	assert.NoError(t, err)
}
