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

	description := valueobjects.NewDescription("Payment processing requires customer management")

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

	description := valueobjects.NewDescription("Test")

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

	description := valueobjects.NewDescription("Digital channels enable customer engagement")

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

	description := valueobjects.NewDescription("Analytics supports decision making")

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

	description := valueobjects.NewDescription("Test dependency")

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

			description := valueobjects.NewDescription("Test dependency")

			dependency, err := NewCapabilityDependency(sourceID, targetID, dependencyType, description)
			require.NoError(t, err)
			assert.NotNil(t, dependency)
			assert.Equal(t, tt.dependencyType, dependency.DependencyType().Value())
		})
	}
}
