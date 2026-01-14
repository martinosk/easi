package aggregates

import (
	"testing"

	"easi/backend/internal/architecturemodeling/domain/entities"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApplicationComponent(t *testing.T) {
	name, err := valueobjects.NewComponentName("User Service")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Handles user authentication and authorization")

	component, err := NewApplicationComponent(name, description)
	require.NoError(t, err)
	assert.NotNil(t, component)
	assert.NotEmpty(t, component.ID())
	assert.Equal(t, name, component.Name())
	assert.Equal(t, description, component.Description())
	assert.NotZero(t, component.CreatedAt())
	assert.Len(t, component.GetUncommittedChanges(), 1)
}

func TestApplicationComponent_RaisesCreatedEvent(t *testing.T) {
	name, err := valueobjects.NewComponentName("User Service")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Test description")

	component, err := NewApplicationComponent(name, description)
	require.NoError(t, err)

	uncommittedEvents := component.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "ApplicationComponentCreated", uncommittedEvents[0].EventType())
}

func TestApplicationComponent_Update(t *testing.T) {
	// Arrange: Create initial component
	name, err := valueobjects.NewComponentName("User Service")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Handles user management")

	component, err := NewApplicationComponent(name, description)
	require.NoError(t, err)

	// Clear uncommitted events to test update event separately
	component.MarkChangesAsCommitted()

	// Act: Update the component
	newName, err := valueobjects.NewComponentName("Enhanced User Service")
	require.NoError(t, err)

	newDescription := valueobjects.MustNewDescription("Handles user management and authentication")

	err = component.Update(newName, newDescription)

	// Assert: Verify update was successful
	require.NoError(t, err)
	assert.Equal(t, newName, component.Name())
	assert.Equal(t, newDescription, component.Description())

	// Verify update event was raised
	uncommittedEvents := component.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "ApplicationComponentUpdated", uncommittedEvents[0].EventType())
}

func TestApplicationComponent_UpdateWithEmptyDescription(t *testing.T) {
	// Arrange: Create initial component
	name, err := valueobjects.NewComponentName("Payment Service")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Processes payments")

	component, err := NewApplicationComponent(name, description)
	require.NoError(t, err)

	component.MarkChangesAsCommitted()

	// Act: Update with empty description (which is allowed)
	newName, err := valueobjects.NewComponentName("Payment Gateway")
	require.NoError(t, err)

	emptyDescription := valueobjects.MustNewDescription("")

	err = component.Update(newName, emptyDescription)

	// Assert: Should succeed
	require.NoError(t, err)
	assert.Equal(t, newName, component.Name())
	assert.Equal(t, emptyDescription, component.Description())
	assert.True(t, component.Description().IsEmpty())
}

func TestLoadApplicationComponentFromHistory(t *testing.T) {
	// Arrange: Create a component and capture its events
	name, err := valueobjects.NewComponentName("Order Service")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Manages order processing")

	originalComponent, err := NewApplicationComponent(name, description)
	require.NoError(t, err)

	events := originalComponent.GetUncommittedChanges()

	// Act: Reconstruct from history
	reconstructedComponent, err := LoadApplicationComponentFromHistory(events)

	// Assert: Reconstructed component matches original
	require.NoError(t, err)
	assert.NotNil(t, reconstructedComponent)
	assert.Equal(t, originalComponent.ID(), reconstructedComponent.ID())
	assert.Equal(t, originalComponent.Name(), reconstructedComponent.Name())
	assert.Equal(t, originalComponent.Description(), reconstructedComponent.Description())
	assert.Equal(t, originalComponent.CreatedAt(), reconstructedComponent.CreatedAt())

	// Verify no uncommitted events on reconstructed aggregate
	assert.Empty(t, reconstructedComponent.GetUncommittedChanges())
}

func TestLoadApplicationComponentFromHistory_WithUpdateEvents(t *testing.T) {
	// Arrange: Create a component, update it, and capture all events
	name, err := valueobjects.NewComponentName("Notification Service")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Sends notifications")

	component, err := NewApplicationComponent(name, description)
	require.NoError(t, err)

	// Update the component
	updatedName, err := valueobjects.NewComponentName("Enhanced Notification Service")
	require.NoError(t, err)

	updatedDescription := valueobjects.MustNewDescription("Sends notifications via email and SMS")

	component.Update(updatedName, updatedDescription)

	allEvents := component.GetUncommittedChanges()

	// Act: Reconstruct from complete history
	reconstructedComponent, err := LoadApplicationComponentFromHistory(allEvents)

	// Assert: Should reflect the updated state
	require.NoError(t, err)
	assert.Equal(t, component.ID(), reconstructedComponent.ID())
	assert.Equal(t, updatedName, reconstructedComponent.Name())
	assert.Equal(t, updatedDescription, reconstructedComponent.Description())
	assert.Equal(t, component.CreatedAt(), reconstructedComponent.CreatedAt())
}

func TestApplicationComponent_EmptyDescription(t *testing.T) {
	// Arrange
	name, err := valueobjects.NewComponentName("API Gateway")
	require.NoError(t, err)

	emptyDescription := valueobjects.MustNewDescription("")

	// Act
	component, err := NewApplicationComponent(name, emptyDescription)

	// Assert: Empty description is allowed
	require.NoError(t, err)
	assert.NotNil(t, component)
	assert.True(t, component.Description().IsEmpty())
	assert.Equal(t, "", component.Description().Value())
}

func TestApplicationComponent_AddExpert(t *testing.T) {
	name, _ := valueobjects.NewComponentName("Customer Portal")
	description := valueobjects.MustNewDescription("Portal for customers")

	component, err := NewApplicationComponent(name, description)
	require.NoError(t, err)

	component.MarkChangesAsCommitted()

	expert, err := entities.NewExpert("Alice Smith", "Product Owner", "alice@example.com")
	require.NoError(t, err)

	err = component.AddExpert(expert)
	require.NoError(t, err)

	experts := component.Experts()
	assert.Len(t, experts, 1)
	assert.Equal(t, "Alice Smith", experts[0].Name().Value())
	assert.Equal(t, "Product Owner", experts[0].Role().Value())
	assert.Equal(t, "alice@example.com", experts[0].Contact().Value())

	uncommittedEvents := component.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "ApplicationComponentExpertAdded", uncommittedEvents[0].EventType())
}

func TestApplicationComponent_RemoveExpert(t *testing.T) {
	name, _ := valueobjects.NewComponentName("Customer Portal")
	description := valueobjects.MustNewDescription("Portal for customers")

	component, err := NewApplicationComponent(name, description)
	require.NoError(t, err)

	expert, err := entities.NewExpert("Alice Smith", "Product Owner", "alice@example.com")
	require.NoError(t, err)

	component.AddExpert(expert)
	component.MarkChangesAsCommitted()

	assert.Len(t, component.Experts(), 1)

	err = component.RemoveExpert("Alice Smith", "Product Owner", "alice@example.com")
	require.NoError(t, err)

	assert.Len(t, component.Experts(), 0)

	uncommittedEvents := component.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "ApplicationComponentExpertRemoved", uncommittedEvents[0].EventType())
}

func TestApplicationComponent_AddMultipleExperts(t *testing.T) {
	name, _ := valueobjects.NewComponentName("Customer Portal")
	description := valueobjects.MustNewDescription("Portal for customers")

	component, err := NewApplicationComponent(name, description)
	require.NoError(t, err)

	expert1, _ := entities.NewExpert("Alice Smith", "Product Owner", "alice@example.com")
	expert2, _ := entities.NewExpert("Bob Johnson", "Tech Lead", "bob@example.com")

	component.AddExpert(expert1)
	component.AddExpert(expert2)

	experts := component.Experts()
	assert.Len(t, experts, 2)
}

func TestLoadApplicationComponentFromHistory_WithExpertEvents(t *testing.T) {
	name, _ := valueobjects.NewComponentName("Customer Portal")
	description := valueobjects.MustNewDescription("Portal for customers")

	component, err := NewApplicationComponent(name, description)
	require.NoError(t, err)

	expert, _ := entities.NewExpert("Alice Smith", "Product Owner", "alice@example.com")
	component.AddExpert(expert)

	allEvents := component.GetUncommittedChanges()

	reconstructed, err := LoadApplicationComponentFromHistory(allEvents)
	require.NoError(t, err)

	assert.Len(t, reconstructed.Experts(), 1)
	assert.Equal(t, "Alice Smith", reconstructed.Experts()[0].Name().Value())
}
