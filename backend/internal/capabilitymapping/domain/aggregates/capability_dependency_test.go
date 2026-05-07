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
	dependency := newCapabilityDependency(t, "Enables", "Digital channels enable customer engagement")

	uncommittedEvents := dependency.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "CapabilityDependencyCreated", uncommittedEvents[0].EventType())
}

func TestCapabilityDependency_Delete(t *testing.T) {
	dependency := newCapabilityDependency(t, "Supports", "Analytics supports decision making")
	dependency.MarkChangesAsCommitted()

	err := dependency.Delete()
	require.NoError(t, err)

	uncommittedEvents := dependency.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "CapabilityDependencyDeleted", uncommittedEvents[0].EventType())
}

func TestCapabilityDependency_LoadFromHistory(t *testing.T) {
	dependency := newCapabilityDependency(t, "Requires", "Test dependency")

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
			dependency := newCapabilityDependency(t, tt.dependencyType, "Test dependency")
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

func TestValidateNewDependency_RejectsInvalidDependencies(t *testing.T) {
	capA := valueobjects.NewCapabilityID()
	capB := valueobjects.NewCapabilityID()
	capC := valueobjects.NewCapabilityID()

	tests := []struct {
		name         string
		source       valueobjects.CapabilityID
		target       valueobjects.CapabilityID
		existingDeps []ExistingDependency
		wantErr      error
	}{
		{
			name:         "duplicate dependency",
			source:       capA,
			target:       capB,
			existingDeps: []ExistingDependency{{SourceID: capA, TargetID: capB}},
			wantErr:      ErrDuplicateDependencyExists,
		},
		{
			name:         "direct circular dependency",
			source:       capB,
			target:       capA,
			existingDeps: []ExistingDependency{{SourceID: capA, TargetID: capB}},
			wantErr:      ErrCircularDependencyDetected,
		},
		{
			name:   "indirect circular dependency",
			source: capC,
			target: capA,
			existingDeps: []ExistingDependency{
				{SourceID: capA, TargetID: capB},
				{SourceID: capB, TargetID: capC},
			},
			wantErr: ErrCircularDependencyDetected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNewDependency(tt.source, tt.target, tt.existingDeps)
			assert.Error(t, err)
			assert.Equal(t, tt.wantErr, err)
		})
	}
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

func newCapabilityDependency(t *testing.T, dependencyTypeName, descriptionText string) *CapabilityDependency {
	t.Helper()

	sourceID := valueobjects.NewCapabilityID()
	targetID := valueobjects.NewCapabilityID()

	dependencyType, err := valueobjects.NewDependencyType(dependencyTypeName)
	require.NoError(t, err)

	description := valueobjects.MustNewDescription(descriptionText)

	dependency, err := NewCapabilityDependency(sourceID, targetID, dependencyType, description)
	require.NoError(t, err)

	return dependency
}
