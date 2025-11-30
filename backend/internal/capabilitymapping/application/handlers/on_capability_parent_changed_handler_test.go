package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	"github.com/stretchr/testify/assert"
)

type mockAssignmentChecker struct {
	assignmentsByCapability map[string][]readmodels.AssignmentDTO
	existingAssignments     map[string]bool
	queryError              error
}

func newMockAssignmentChecker() *mockAssignmentChecker {
	return &mockAssignmentChecker{
		assignmentsByCapability: make(map[string][]readmodels.AssignmentDTO),
		existingAssignments:     make(map[string]bool),
	}
}

func (m *mockAssignmentChecker) GetByCapabilityID(ctx context.Context, capabilityID string) ([]readmodels.AssignmentDTO, error) {
	if m.queryError != nil {
		return nil, m.queryError
	}
	return m.assignmentsByCapability[capabilityID], nil
}

func (m *mockAssignmentChecker) AssignmentExists(ctx context.Context, domainID, capabilityID string) (bool, error) {
	if m.queryError != nil {
		return false, m.queryError
	}
	key := domainID + ":" + capabilityID
	return m.existingAssignments[key], nil
}

type mockCapabilityReader struct {
	capabilities map[string]*readmodels.CapabilityDTO
	queryError   error
}

func newMockCapabilityReader() *mockCapabilityReader {
	return &mockCapabilityReader{
		capabilities: make(map[string]*readmodels.CapabilityDTO),
	}
}

func (m *mockCapabilityReader) GetByID(ctx context.Context, id string) (*readmodels.CapabilityDTO, error) {
	if m.queryError != nil {
		return nil, m.queryError
	}
	return m.capabilities[id], nil
}

func TestOnCapabilityParentChangedHandler_L1ToL2_UnassignsAndReassignsToParent(t *testing.T) {
	commandBus := &mockCommandBus{}
	assignmentRM := newMockAssignmentChecker()
	capabilityRM := newMockCapabilityReader()

	assignmentRM.assignmentsByCapability["cap-child"] = []readmodels.AssignmentDTO{
		{AssignmentID: "assign-1", CapabilityID: "cap-child", BusinessDomainID: "bd-1"},
	}
	capabilityRM.capabilities["cap-parent"] = &readmodels.CapabilityDTO{ID: "cap-parent", Level: "L1"}

	handler := NewOnCapabilityParentChangedHandler(commandBus, assignmentRM, capabilityRM)
	event := events.NewCapabilityParentChanged("cap-child", "", "cap-parent", "L1", "L2")

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	assert.Len(t, commandBus.dispatchedCommands, 2)

	unassignCmd, ok := commandBus.dispatchedCommands[0].(*commands.UnassignCapabilityFromDomain)
	assert.True(t, ok)
	assert.Equal(t, "assign-1", unassignCmd.AssignmentID)

	assignCmd, ok := commandBus.dispatchedCommands[1].(*commands.AssignCapabilityToDomain)
	assert.True(t, ok)
	assert.Equal(t, "bd-1", assignCmd.BusinessDomainID)
	assert.Equal(t, "cap-parent", assignCmd.CapabilityID)
}

func TestOnCapabilityParentChangedHandler_L1ToL3_FindsL1Ancestor(t *testing.T) {
	commandBus := &mockCommandBus{}
	assignmentRM := newMockAssignmentChecker()
	capabilityRM := newMockCapabilityReader()

	assignmentRM.assignmentsByCapability["cap-grandchild"] = []readmodels.AssignmentDTO{
		{AssignmentID: "assign-1", CapabilityID: "cap-grandchild", BusinessDomainID: "bd-1"},
	}
	capabilityRM.capabilities["cap-parent"] = &readmodels.CapabilityDTO{ID: "cap-parent", Level: "L2", ParentID: "cap-grandparent"}
	capabilityRM.capabilities["cap-grandparent"] = &readmodels.CapabilityDTO{ID: "cap-grandparent", Level: "L1"}

	handler := NewOnCapabilityParentChangedHandler(commandBus, assignmentRM, capabilityRM)
	event := events.NewCapabilityParentChanged("cap-grandchild", "", "cap-parent", "L1", "L3")

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	assert.Len(t, commandBus.dispatchedCommands, 2)

	assignCmd, ok := commandBus.dispatchedCommands[1].(*commands.AssignCapabilityToDomain)
	assert.True(t, ok)
	assert.Equal(t, "cap-grandparent", assignCmd.CapabilityID)
}

func TestOnCapabilityParentChangedHandler_L1ToL4_FindsL1Ancestor(t *testing.T) {
	commandBus := &mockCommandBus{}
	assignmentRM := newMockAssignmentChecker()
	capabilityRM := newMockCapabilityReader()

	assignmentRM.assignmentsByCapability["cap-child"] = []readmodels.AssignmentDTO{
		{AssignmentID: "assign-1", CapabilityID: "cap-child", BusinessDomainID: "bd-1"},
	}
	capabilityRM.capabilities["cap-l3"] = &readmodels.CapabilityDTO{ID: "cap-l3", Level: "L3", ParentID: "cap-l2"}
	capabilityRM.capabilities["cap-l2"] = &readmodels.CapabilityDTO{ID: "cap-l2", Level: "L2", ParentID: "cap-l1"}
	capabilityRM.capabilities["cap-l1"] = &readmodels.CapabilityDTO{ID: "cap-l1", Level: "L1"}

	handler := NewOnCapabilityParentChangedHandler(commandBus, assignmentRM, capabilityRM)
	event := events.NewCapabilityParentChanged("cap-child", "", "cap-l3", "L1", "L4")

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	assert.Len(t, commandBus.dispatchedCommands, 2)

	assignCmd, ok := commandBus.dispatchedCommands[1].(*commands.AssignCapabilityToDomain)
	assert.True(t, ok)
	assert.Equal(t, "cap-l1", assignCmd.CapabilityID)
}

func TestOnCapabilityParentChangedHandler_L2ToL3_NoAction(t *testing.T) {
	commandBus := &mockCommandBus{}
	assignmentRM := newMockAssignmentChecker()
	capabilityRM := newMockCapabilityReader()

	handler := NewOnCapabilityParentChangedHandler(commandBus, assignmentRM, capabilityRM)
	event := events.NewCapabilityParentChanged("cap-child", "old-parent", "new-parent", "L2", "L3")

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	assert.Empty(t, commandBus.dispatchedCommands)
}

func TestOnCapabilityParentChangedHandler_L1ToL1_NoAction(t *testing.T) {
	commandBus := &mockCommandBus{}
	assignmentRM := newMockAssignmentChecker()
	capabilityRM := newMockCapabilityReader()

	handler := NewOnCapabilityParentChangedHandler(commandBus, assignmentRM, capabilityRM)
	event := events.NewCapabilityParentChanged("cap-child", "", "", "L1", "L1")

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	assert.Empty(t, commandBus.dispatchedCommands)
}

func TestOnCapabilityParentChangedHandler_NoAssignments_NoCommands(t *testing.T) {
	commandBus := &mockCommandBus{}
	assignmentRM := newMockAssignmentChecker()
	capabilityRM := newMockCapabilityReader()

	handler := NewOnCapabilityParentChangedHandler(commandBus, assignmentRM, capabilityRM)
	event := events.NewCapabilityParentChanged("cap-child", "", "cap-parent", "L1", "L2")

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	assert.Empty(t, commandBus.dispatchedCommands)
}

func TestOnCapabilityParentChangedHandler_ParentAlreadyAssigned_OnlyUnassigns(t *testing.T) {
	commandBus := &mockCommandBus{}
	assignmentRM := newMockAssignmentChecker()
	capabilityRM := newMockCapabilityReader()

	assignmentRM.assignmentsByCapability["cap-child"] = []readmodels.AssignmentDTO{
		{AssignmentID: "assign-1", CapabilityID: "cap-child", BusinessDomainID: "bd-1"},
	}
	assignmentRM.existingAssignments["bd-1:cap-parent"] = true
	capabilityRM.capabilities["cap-parent"] = &readmodels.CapabilityDTO{ID: "cap-parent", Level: "L1"}

	handler := NewOnCapabilityParentChangedHandler(commandBus, assignmentRM, capabilityRM)
	event := events.NewCapabilityParentChanged("cap-child", "", "cap-parent", "L1", "L2")

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	assert.Len(t, commandBus.dispatchedCommands, 1)

	unassignCmd, ok := commandBus.dispatchedCommands[0].(*commands.UnassignCapabilityFromDomain)
	assert.True(t, ok)
	assert.Equal(t, "assign-1", unassignCmd.AssignmentID)
}

func TestOnCapabilityParentChangedHandler_L1AncestorAlreadyAssigned_OnlyUnassigns(t *testing.T) {
	commandBus := &mockCommandBus{}
	assignmentRM := newMockAssignmentChecker()
	capabilityRM := newMockCapabilityReader()

	assignmentRM.assignmentsByCapability["cap-child"] = []readmodels.AssignmentDTO{
		{AssignmentID: "assign-1", CapabilityID: "cap-child", BusinessDomainID: "bd-1"},
	}
	assignmentRM.existingAssignments["bd-1:cap-l1"] = true
	capabilityRM.capabilities["cap-l2"] = &readmodels.CapabilityDTO{ID: "cap-l2", Level: "L2", ParentID: "cap-l1"}
	capabilityRM.capabilities["cap-l1"] = &readmodels.CapabilityDTO{ID: "cap-l1", Level: "L1"}

	handler := NewOnCapabilityParentChangedHandler(commandBus, assignmentRM, capabilityRM)
	event := events.NewCapabilityParentChanged("cap-child", "", "cap-l2", "L1", "L3")

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	assert.Len(t, commandBus.dispatchedCommands, 1)

	unassignCmd, ok := commandBus.dispatchedCommands[0].(*commands.UnassignCapabilityFromDomain)
	assert.True(t, ok)
	assert.Equal(t, "assign-1", unassignCmd.AssignmentID)
}

func TestOnCapabilityParentChangedHandler_MultipleDomains_HandlesAll(t *testing.T) {
	commandBus := &mockCommandBus{}
	assignmentRM := newMockAssignmentChecker()
	capabilityRM := newMockCapabilityReader()

	assignmentRM.assignmentsByCapability["cap-child"] = []readmodels.AssignmentDTO{
		{AssignmentID: "assign-1", CapabilityID: "cap-child", BusinessDomainID: "bd-1"},
		{AssignmentID: "assign-2", CapabilityID: "cap-child", BusinessDomainID: "bd-2"},
		{AssignmentID: "assign-3", CapabilityID: "cap-child", BusinessDomainID: "bd-3"},
	}
	assignmentRM.existingAssignments["bd-2:cap-parent"] = true
	capabilityRM.capabilities["cap-parent"] = &readmodels.CapabilityDTO{ID: "cap-parent", Level: "L1"}

	handler := NewOnCapabilityParentChangedHandler(commandBus, assignmentRM, capabilityRM)
	event := events.NewCapabilityParentChanged("cap-child", "", "cap-parent", "L1", "L2")

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	assert.Len(t, commandBus.dispatchedCommands, 5)

	unassignCount := 0
	assignCount := 0
	assignedDomains := []string{}

	for _, cmd := range commandBus.dispatchedCommands {
		switch c := cmd.(type) {
		case *commands.UnassignCapabilityFromDomain:
			unassignCount++
		case *commands.AssignCapabilityToDomain:
			assignCount++
			assignedDomains = append(assignedDomains, c.BusinessDomainID)
		}
	}

	assert.Equal(t, 3, unassignCount)
	assert.Equal(t, 2, assignCount)
	assert.Contains(t, assignedDomains, "bd-1")
	assert.NotContains(t, assignedDomains, "bd-2")
	assert.Contains(t, assignedDomains, "bd-3")
}

func TestOnCapabilityParentChangedHandler_ReadModelError_ReturnsError(t *testing.T) {
	commandBus := &mockCommandBus{}
	assignmentRM := newMockAssignmentChecker()
	capabilityRM := newMockCapabilityReader()
	assignmentRM.queryError = errors.New("database error")

	handler := NewOnCapabilityParentChangedHandler(commandBus, assignmentRM, capabilityRM)
	event := events.NewCapabilityParentChanged("cap-child", "", "cap-parent", "L1", "L2")

	err := handler.Handle(context.Background(), event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}
