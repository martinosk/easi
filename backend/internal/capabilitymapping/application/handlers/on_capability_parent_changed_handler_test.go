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

type parentChangedFixture struct {
	commandBus   *mockCommandBus
	assignmentRM *mockAssignmentChecker
	capabilityRM *mockCapabilityReader
}

func newParentChangedFixture() *parentChangedFixture {
	return &parentChangedFixture{
		commandBus:   &mockCommandBus{},
		assignmentRM: newMockAssignmentChecker(),
		capabilityRM: newMockCapabilityReader(),
	}
}

func (f *parentChangedFixture) handle(event events.CapabilityParentChanged) error {
	handler := NewOnCapabilityParentChangedHandler(f.commandBus, f.assignmentRM, f.capabilityRM)
	return handler.Handle(context.Background(), event)
}

func TestOnCapabilityParentChangedHandler_NoLevelChange(t *testing.T) {
	tests := []struct {
		name  string
		event events.CapabilityParentChanged
	}{
		{
			name:  "L2 to L3 takes no action",
			event: events.NewCapabilityParentChanged("cap-child", "old-parent", "new-parent", "L2", "L3"),
		},
		{
			name:  "L1 to L1 takes no action",
			event: events.NewCapabilityParentChanged("cap-child", "", "", "L1", "L1"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newParentChangedFixture()

			err := f.handle(tt.event)

			assert.NoError(t, err)
			assert.Empty(t, f.commandBus.dispatchedCommands)
		})
	}
}

func TestOnCapabilityParentChangedHandler_NoAssignments_NoCommands(t *testing.T) {
	f := newParentChangedFixture()

	err := f.handle(events.NewCapabilityParentChanged("cap-child", "", "cap-parent", "L1", "L2"))

	assert.NoError(t, err)
	assert.Empty(t, f.commandBus.dispatchedCommands)
}

func TestOnCapabilityParentChangedHandler_FindsL1Ancestor(t *testing.T) {
	tests := []struct {
		name                 string
		childID              string
		parentID             string
		newLevel             string
		capabilities         map[string]*readmodels.CapabilityDTO
		expectedAncestorInID string
	}{
		{
			name:     "L1 to L2 reassigns to direct parent",
			childID:  "cap-child",
			parentID: "cap-parent",
			newLevel: "L2",
			capabilities: map[string]*readmodels.CapabilityDTO{
				"cap-parent": {ID: "cap-parent", Level: "L1"},
			},
			expectedAncestorInID: "cap-parent",
		},
		{
			name:     "L1 to L3 traverses to L1 grandparent",
			childID:  "cap-grandchild",
			parentID: "cap-parent",
			newLevel: "L3",
			capabilities: map[string]*readmodels.CapabilityDTO{
				"cap-parent":      {ID: "cap-parent", Level: "L2", ParentID: "cap-grandparent"},
				"cap-grandparent": {ID: "cap-grandparent", Level: "L1"},
			},
			expectedAncestorInID: "cap-grandparent",
		},
		{
			name:     "L1 to L4 traverses to L1 great-grandparent",
			childID:  "cap-child",
			parentID: "cap-l3",
			newLevel: "L4",
			capabilities: map[string]*readmodels.CapabilityDTO{
				"cap-l3": {ID: "cap-l3", Level: "L3", ParentID: "cap-l2"},
				"cap-l2": {ID: "cap-l2", Level: "L2", ParentID: "cap-l1"},
				"cap-l1": {ID: "cap-l1", Level: "L1"},
			},
			expectedAncestorInID: "cap-l1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newParentChangedFixture()
			f.assignmentRM.assignmentsByCapability[tt.childID] = []readmodels.AssignmentDTO{
				{AssignmentID: "assign-1", CapabilityID: tt.childID, BusinessDomainID: "bd-1"},
			}
			f.capabilityRM.capabilities = tt.capabilities

			err := f.handle(events.NewCapabilityParentChanged(tt.childID, "", tt.parentID, "L1", tt.newLevel))

			assert.NoError(t, err)
			assert.Len(t, f.commandBus.dispatchedCommands, 2)

			unassignCmd, ok := f.commandBus.dispatchedCommands[0].(*commands.UnassignCapabilityFromDomain)
			assert.True(t, ok)
			assert.Equal(t, "assign-1", unassignCmd.AssignmentID)

			assignCmd, ok := f.commandBus.dispatchedCommands[1].(*commands.AssignCapabilityToDomain)
			assert.True(t, ok)
			assert.Equal(t, "bd-1", assignCmd.BusinessDomainID)
			assert.Equal(t, tt.expectedAncestorInID, assignCmd.CapabilityID)
		})
	}
}

func TestOnCapabilityParentChangedHandler_AncestorAlreadyAssigned_OnlyUnassigns(t *testing.T) {
	tests := []struct {
		name                string
		childID             string
		parentID            string
		newLevel            string
		capabilities        map[string]*readmodels.CapabilityDTO
		existingAssignments map[string]bool
	}{
		{
			name:     "direct parent already assigned",
			childID:  "cap-child",
			parentID: "cap-parent",
			newLevel: "L2",
			capabilities: map[string]*readmodels.CapabilityDTO{
				"cap-parent": {ID: "cap-parent", Level: "L1"},
			},
			existingAssignments: map[string]bool{"bd-1:cap-parent": true},
		},
		{
			name:     "L1 ancestor already assigned",
			childID:  "cap-child",
			parentID: "cap-l2",
			newLevel: "L3",
			capabilities: map[string]*readmodels.CapabilityDTO{
				"cap-l2": {ID: "cap-l2", Level: "L2", ParentID: "cap-l1"},
				"cap-l1": {ID: "cap-l1", Level: "L1"},
			},
			existingAssignments: map[string]bool{"bd-1:cap-l1": true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newParentChangedFixture()
			f.assignmentRM.assignmentsByCapability[tt.childID] = []readmodels.AssignmentDTO{
				{AssignmentID: "assign-1", CapabilityID: tt.childID, BusinessDomainID: "bd-1"},
			}
			f.assignmentRM.existingAssignments = tt.existingAssignments
			f.capabilityRM.capabilities = tt.capabilities

			err := f.handle(events.NewCapabilityParentChanged(tt.childID, "", tt.parentID, "L1", tt.newLevel))

			assert.NoError(t, err)
			assert.Len(t, f.commandBus.dispatchedCommands, 1)

			unassignCmd, ok := f.commandBus.dispatchedCommands[0].(*commands.UnassignCapabilityFromDomain)
			assert.True(t, ok)
			assert.Equal(t, "assign-1", unassignCmd.AssignmentID)
		})
	}
}

func TestOnCapabilityParentChangedHandler_MultipleDomains_HandlesAll(t *testing.T) {
	f := newParentChangedFixture()
	f.assignmentRM.assignmentsByCapability["cap-child"] = []readmodels.AssignmentDTO{
		{AssignmentID: "assign-1", CapabilityID: "cap-child", BusinessDomainID: "bd-1"},
		{AssignmentID: "assign-2", CapabilityID: "cap-child", BusinessDomainID: "bd-2"},
		{AssignmentID: "assign-3", CapabilityID: "cap-child", BusinessDomainID: "bd-3"},
	}
	f.assignmentRM.existingAssignments["bd-2:cap-parent"] = true
	f.capabilityRM.capabilities["cap-parent"] = &readmodels.CapabilityDTO{ID: "cap-parent", Level: "L1"}

	err := f.handle(events.NewCapabilityParentChanged("cap-child", "", "cap-parent", "L1", "L2"))

	assert.NoError(t, err)
	assert.Len(t, f.commandBus.dispatchedCommands, 5)

	unassignCount := 0
	assignCount := 0
	assignedDomains := []string{}

	for _, cmd := range f.commandBus.dispatchedCommands {
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
	f := newParentChangedFixture()
	f.assignmentRM.queryError = errors.New("database error")

	err := f.handle(events.NewCapabilityParentChanged("cap-child", "", "cap-parent", "L1", "L2"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}
