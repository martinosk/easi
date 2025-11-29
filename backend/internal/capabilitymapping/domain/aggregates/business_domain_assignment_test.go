package aggregates

import (
	"testing"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssignCapabilityToDomain(t *testing.T) {
	businessDomainID, err := valueobjects.NewBusinessDomainIDFromString("550e8400-e29b-41d4-a716-446655440000")
	require.NoError(t, err)

	capabilityID, err := valueobjects.NewCapabilityIDFromString("550e8400-e29b-41d4-a716-446655440001")
	require.NoError(t, err)

	assignment, err := AssignCapabilityToDomain(businessDomainID, capabilityID)
	require.NoError(t, err)
	assert.NotNil(t, assignment)
	assert.NotEmpty(t, assignment.ID())
	assert.Equal(t, businessDomainID, assignment.BusinessDomainID())
	assert.Equal(t, capabilityID, assignment.CapabilityID())
	assert.NotZero(t, assignment.AssignedAt())
	assert.Len(t, assignment.GetUncommittedChanges(), 1)
}

func TestAssignCapabilityToDomain_RaisesEvent(t *testing.T) {
	businessDomainID, err := valueobjects.NewBusinessDomainIDFromString("550e8400-e29b-41d4-a716-446655440000")
	require.NoError(t, err)

	capabilityID, err := valueobjects.NewCapabilityIDFromString("550e8400-e29b-41d4-a716-446655440001")
	require.NoError(t, err)

	assignment, err := AssignCapabilityToDomain(businessDomainID, capabilityID)
	require.NoError(t, err)

	uncommittedEvents := assignment.GetUncommittedChanges()
	assert.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "CapabilityAssignedToDomain", uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	assert.Equal(t, assignment.ID(), eventData["id"])
	assert.Equal(t, businessDomainID.Value(), eventData["businessDomainId"])
	assert.Equal(t, capabilityID.Value(), eventData["capabilityId"])
	assert.NotNil(t, eventData["assignedAt"])
}

func TestUnassignCapabilityFromDomain(t *testing.T) {
	assignment := createAssignment(t)
	assignment.MarkChangesAsCommitted()

	err := assignment.Unassign()
	require.NoError(t, err)

	uncommittedEvents := assignment.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "CapabilityUnassignedFromDomain", uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	assert.Equal(t, assignment.ID(), eventData["id"])
	assert.NotNil(t, eventData["unassignedAt"])
}

func TestUnassignCapabilityFromDomain_PreservesState(t *testing.T) {
	assignment := createAssignment(t)
	assignment.MarkChangesAsCommitted()

	originalID := assignment.ID()
	originalBusinessDomainID := assignment.BusinessDomainID().Value()
	originalCapabilityID := assignment.CapabilityID().Value()

	err := assignment.Unassign()
	require.NoError(t, err)

	assert.Equal(t, originalID, assignment.ID())
	assert.Equal(t, originalBusinessDomainID, assignment.BusinessDomainID().Value())
	assert.Equal(t, originalCapabilityID, assignment.CapabilityID().Value())
}

func TestBusinessDomainAssignment_LoadFromHistory(t *testing.T) {
	businessDomainID, err := valueobjects.NewBusinessDomainIDFromString("550e8400-e29b-41d4-a716-446655440000")
	require.NoError(t, err)

	capabilityID, err := valueobjects.NewCapabilityIDFromString("550e8400-e29b-41d4-a716-446655440001")
	require.NoError(t, err)

	assignment, err := AssignCapabilityToDomain(businessDomainID, capabilityID)
	require.NoError(t, err)

	events := assignment.GetUncommittedChanges()

	loadedAssignment, err := LoadBusinessDomainAssignmentFromHistory(events)
	require.NoError(t, err)
	assert.NotNil(t, loadedAssignment)
	assert.Equal(t, assignment.ID(), loadedAssignment.ID())
	assert.Equal(t, assignment.BusinessDomainID().Value(), loadedAssignment.BusinessDomainID().Value())
	assert.Equal(t, assignment.CapabilityID().Value(), loadedAssignment.CapabilityID().Value())
}

func TestBusinessDomainAssignment_LoadFromHistoryWithUnassign(t *testing.T) {
	assignment := createAssignment(t)

	err := assignment.Unassign()
	require.NoError(t, err)

	allEvents := assignment.GetUncommittedChanges()
	require.Len(t, allEvents, 2)

	loadedAssignment, err := LoadBusinessDomainAssignmentFromHistory(allEvents)
	require.NoError(t, err)
	assert.Equal(t, assignment.ID(), loadedAssignment.ID())
	assert.Equal(t, assignment.BusinessDomainID().Value(), loadedAssignment.BusinessDomainID().Value())
	assert.Equal(t, assignment.CapabilityID().Value(), loadedAssignment.CapabilityID().Value())
}

func createAssignment(t *testing.T) *BusinessDomainAssignment {
	t.Helper()

	businessDomainID, err := valueobjects.NewBusinessDomainIDFromString("550e8400-e29b-41d4-a716-446655440000")
	require.NoError(t, err)

	capabilityID, err := valueobjects.NewCapabilityIDFromString("550e8400-e29b-41d4-a716-446655440001")
	require.NoError(t, err)

	assignment, err := AssignCapabilityToDomain(businessDomainID, capabilityID)
	require.NoError(t, err)

	return assignment
}
