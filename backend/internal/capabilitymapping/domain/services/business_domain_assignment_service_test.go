package services

import (
	"context"
	"testing"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAssignmentLookup struct {
	assignments     map[string][]AssignmentInfo
	existingPairs   map[string]bool
}

func newMockAssignmentLookup() *mockAssignmentLookup {
	return &mockAssignmentLookup{
		assignments:   make(map[string][]AssignmentInfo),
		existingPairs: make(map[string]bool),
	}
}

func (m *mockAssignmentLookup) GetByCapabilityID(ctx context.Context, capabilityID valueobjects.CapabilityID) ([]AssignmentInfo, error) {
	return m.assignments[capabilityID.Value()], nil
}

func (m *mockAssignmentLookup) AssignmentExists(ctx context.Context, domainID valueobjects.BusinessDomainID, capabilityID valueobjects.CapabilityID) (bool, error) {
	key := domainID.Value() + ":" + capabilityID.Value()
	return m.existingPairs[key], nil
}

func (m *mockAssignmentLookup) addAssignment(capabilityID valueobjects.CapabilityID, assignmentID string, domainID valueobjects.BusinessDomainID) {
	m.assignments[capabilityID.Value()] = append(m.assignments[capabilityID.Value()], AssignmentInfo{
		AssignmentID:     assignmentID,
		BusinessDomainID: domainID,
		CapabilityID:     capabilityID,
	})
}

func (m *mockAssignmentLookup) setAssignmentExists(domainID valueobjects.BusinessDomainID, capabilityID valueobjects.CapabilityID, exists bool) {
	key := domainID.Value() + ":" + capabilityID.Value()
	m.existingPairs[key] = exists
}

type mockCommandExecutor struct {
	unassignedIDs []string
	assignedPairs []struct {
		DomainID     valueobjects.BusinessDomainID
		CapabilityID valueobjects.CapabilityID
	}
}

func newMockCommandExecutor() *mockCommandExecutor {
	return &mockCommandExecutor{}
}

func (m *mockCommandExecutor) Unassign(ctx context.Context, assignmentID string) error {
	m.unassignedIDs = append(m.unassignedIDs, assignmentID)
	return nil
}

func (m *mockCommandExecutor) Assign(ctx context.Context, domainID valueobjects.BusinessDomainID, capabilityID valueobjects.CapabilityID) error {
	m.assignedPairs = append(m.assignedPairs, struct {
		DomainID     valueobjects.BusinessDomainID
		CapabilityID valueobjects.CapabilityID
	}{domainID, capabilityID})
	return nil
}

type mockHierarchyService struct{}

func (m *mockHierarchyService) FindL1Ancestor(ctx context.Context, capabilityID valueobjects.CapabilityID) (valueobjects.CapabilityID, error) {
	return capabilityID, nil
}

func (m *mockHierarchyService) GetDescendants(ctx context.Context, capabilityID valueobjects.CapabilityID) ([]valueobjects.CapabilityID, error) {
	return nil, nil
}

func (m *mockHierarchyService) ValidateHierarchyChange(ctx context.Context, capabilityID valueobjects.CapabilityID, newParentID valueobjects.CapabilityID) error {
	return nil
}

func TestBusinessDomainAssignmentService_ReassignToL1Ancestor_NoAssignments(t *testing.T) {
	lookup := newMockAssignmentLookup()
	executor := newMockCommandExecutor()
	hierarchy := &mockHierarchyService{}
	service := NewBusinessDomainAssignmentService(lookup, executor, hierarchy)

	capabilityID := valueobjects.NewCapabilityID()
	newL1ID := valueobjects.NewCapabilityID()

	err := service.ReassignToL1Ancestor(context.Background(), capabilityID, newL1ID)
	require.NoError(t, err)
	assert.Empty(t, executor.unassignedIDs)
	assert.Empty(t, executor.assignedPairs)
}

func TestBusinessDomainAssignmentService_ReassignToL1Ancestor_UnassignsOldAndAssignsNew(t *testing.T) {
	lookup := newMockAssignmentLookup()
	executor := newMockCommandExecutor()
	hierarchy := &mockHierarchyService{}
	service := NewBusinessDomainAssignmentService(lookup, executor, hierarchy)

	capabilityID := valueobjects.NewCapabilityID()
	domainID := valueobjects.NewBusinessDomainID()
	newL1ID := valueobjects.NewCapabilityID()

	lookup.addAssignment(capabilityID, "assignment-1", domainID)
	lookup.setAssignmentExists(domainID, newL1ID, false)

	err := service.ReassignToL1Ancestor(context.Background(), capabilityID, newL1ID)
	require.NoError(t, err)

	assert.Len(t, executor.unassignedIDs, 1)
	assert.Equal(t, "assignment-1", executor.unassignedIDs[0])

	assert.Len(t, executor.assignedPairs, 1)
	assert.Equal(t, domainID.Value(), executor.assignedPairs[0].DomainID.Value())
	assert.Equal(t, newL1ID.Value(), executor.assignedPairs[0].CapabilityID.Value())
}

func TestBusinessDomainAssignmentService_ReassignToL1Ancestor_SkipsIfL1AlreadyAssigned(t *testing.T) {
	lookup := newMockAssignmentLookup()
	executor := newMockCommandExecutor()
	hierarchy := &mockHierarchyService{}
	service := NewBusinessDomainAssignmentService(lookup, executor, hierarchy)

	capabilityID := valueobjects.NewCapabilityID()
	domainID := valueobjects.NewBusinessDomainID()
	newL1ID := valueobjects.NewCapabilityID()

	lookup.addAssignment(capabilityID, "assignment-1", domainID)
	lookup.setAssignmentExists(domainID, newL1ID, true)

	err := service.ReassignToL1Ancestor(context.Background(), capabilityID, newL1ID)
	require.NoError(t, err)

	assert.Len(t, executor.unassignedIDs, 1)
	assert.Empty(t, executor.assignedPairs)
}

func TestBusinessDomainAssignmentService_UnassignAllForCapability_NoAssignments(t *testing.T) {
	lookup := newMockAssignmentLookup()
	executor := newMockCommandExecutor()
	hierarchy := &mockHierarchyService{}
	service := NewBusinessDomainAssignmentService(lookup, executor, hierarchy)

	capabilityID := valueobjects.NewCapabilityID()

	err := service.UnassignAllForCapability(context.Background(), capabilityID)
	require.NoError(t, err)
	assert.Empty(t, executor.unassignedIDs)
}

func TestBusinessDomainAssignmentService_UnassignAllForCapability_UnassignsAll(t *testing.T) {
	lookup := newMockAssignmentLookup()
	executor := newMockCommandExecutor()
	hierarchy := &mockHierarchyService{}
	service := NewBusinessDomainAssignmentService(lookup, executor, hierarchy)

	capabilityID := valueobjects.NewCapabilityID()
	domainID1 := valueobjects.NewBusinessDomainID()
	domainID2 := valueobjects.NewBusinessDomainID()

	lookup.addAssignment(capabilityID, "assignment-1", domainID1)
	lookup.addAssignment(capabilityID, "assignment-2", domainID2)

	err := service.UnassignAllForCapability(context.Background(), capabilityID)
	require.NoError(t, err)

	assert.Len(t, executor.unassignedIDs, 2)
	assert.Contains(t, executor.unassignedIDs, "assignment-1")
	assert.Contains(t, executor.unassignedIDs, "assignment-2")
}
