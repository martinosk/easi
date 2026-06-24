package services

import (
	"context"
	"testing"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAssignmentLookup struct {
	assignments   map[string][]AssignmentInfo
	existingPairs map[string]bool
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

type assignmentServiceHarness struct {
	lookup   *mockAssignmentLookup
	executor *mockCommandExecutor
	service  BusinessDomainAssignmentService
}

func newAssignmentServiceHarness() *assignmentServiceHarness {
	lookup := newMockAssignmentLookup()
	executor := newMockCommandExecutor()
	return &assignmentServiceHarness{
		lookup:   lookup,
		executor: executor,
		service:  NewBusinessDomainAssignmentService(lookup, executor, &mockHierarchyService{}),
	}
}

func TestBusinessDomainAssignmentService_ReassignToL1Ancestor_NoAssignments(t *testing.T) {
	h := newAssignmentServiceHarness()

	capabilityID := valueobjects.NewCapabilityID()
	newL1ID := valueobjects.NewCapabilityID()

	err := h.service.ReassignToL1Ancestor(context.Background(), capabilityID, newL1ID)
	require.NoError(t, err)
	assert.Empty(t, h.executor.unassignedIDs)
	assert.Empty(t, h.executor.assignedPairs)
}

func TestBusinessDomainAssignmentService_ReassignToL1Ancestor_UnassignsOldAndAssignsNew(t *testing.T) {
	h := newAssignmentServiceHarness()

	capabilityID := valueobjects.NewCapabilityID()
	domainID := valueobjects.NewBusinessDomainID()
	newL1ID := valueobjects.NewCapabilityID()

	h.lookup.addAssignment(capabilityID, "assignment-1", domainID)
	h.lookup.setAssignmentExists(domainID, newL1ID, false)

	err := h.service.ReassignToL1Ancestor(context.Background(), capabilityID, newL1ID)
	require.NoError(t, err)

	assert.Len(t, h.executor.unassignedIDs, 1)
	assert.Equal(t, "assignment-1", h.executor.unassignedIDs[0])

	assert.Len(t, h.executor.assignedPairs, 1)
	assert.Equal(t, domainID.Value(), h.executor.assignedPairs[0].DomainID.Value())
	assert.Equal(t, newL1ID.Value(), h.executor.assignedPairs[0].CapabilityID.Value())
}

func TestBusinessDomainAssignmentService_ReassignToL1Ancestor_SkipsIfL1AlreadyAssigned(t *testing.T) {
	h := newAssignmentServiceHarness()

	capabilityID := valueobjects.NewCapabilityID()
	domainID := valueobjects.NewBusinessDomainID()
	newL1ID := valueobjects.NewCapabilityID()

	h.lookup.addAssignment(capabilityID, "assignment-1", domainID)
	h.lookup.setAssignmentExists(domainID, newL1ID, true)

	err := h.service.ReassignToL1Ancestor(context.Background(), capabilityID, newL1ID)
	require.NoError(t, err)

	assert.Len(t, h.executor.unassignedIDs, 1)
	assert.Empty(t, h.executor.assignedPairs)
}

func TestBusinessDomainAssignmentService_UnassignAllForCapability_NoAssignments(t *testing.T) {
	h := newAssignmentServiceHarness()

	capabilityID := valueobjects.NewCapabilityID()

	err := h.service.UnassignAllForCapability(context.Background(), capabilityID)
	require.NoError(t, err)
	assert.Empty(t, h.executor.unassignedIDs)
}

func TestBusinessDomainAssignmentService_UnassignAllForCapability_UnassignsAll(t *testing.T) {
	h := newAssignmentServiceHarness()

	capabilityID := valueobjects.NewCapabilityID()
	domainID1 := valueobjects.NewBusinessDomainID()
	domainID2 := valueobjects.NewBusinessDomainID()

	h.lookup.addAssignment(capabilityID, "assignment-1", domainID1)
	h.lookup.addAssignment(capabilityID, "assignment-2", domainID2)

	err := h.service.UnassignAllForCapability(context.Background(), capabilityID)
	require.NoError(t, err)

	assert.Len(t, h.executor.unassignedIDs, 2)
	assert.Contains(t, h.executor.unassignedIDs, "assignment-1")
	assert.Contains(t, h.executor.unassignedIDs, "assignment-2")
}
