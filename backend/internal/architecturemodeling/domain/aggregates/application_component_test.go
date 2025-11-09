package aggregates

import (
	"testing"

	"github.com/easi/backend/internal/architecturemodeling/domain/valueobjects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApplicationComponent(t *testing.T) {
	name, err := valueobjects.NewComponentName("User Service")
	require.NoError(t, err)

	description := valueobjects.NewDescription("Handles user authentication and authorization")

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

	description := valueobjects.NewDescription("Test description")

	component, err := NewApplicationComponent(name, description)
	require.NoError(t, err)

	uncommittedEvents := component.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "ApplicationComponentCreated", uncommittedEvents[0].EventType())
}
