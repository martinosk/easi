package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/shared/cqrs"
	"github.com/stretchr/testify/assert"
)

type mockCommandBus struct {
	dispatchedCommands []cqrs.Command
	dispatchError      error
}

func (m *mockCommandBus) Register(commandName string, handler cqrs.CommandHandler) {
}

func (m *mockCommandBus) Dispatch(ctx context.Context, command cqrs.Command) error {
	if m.dispatchError != nil {
		return m.dispatchError
	}
	m.dispatchedCommands = append(m.dispatchedCommands, command)
	return nil
}

type mockAssignmentReadModel struct {
	assignmentsByCapability []readmodels.AssignmentDTO
	queryError              error
}

func (m *mockAssignmentReadModel) GetByCapabilityID(ctx context.Context, capabilityID string) ([]readmodels.AssignmentDTO, error) {
	if m.queryError != nil {
		return nil, m.queryError
	}
	return m.assignmentsByCapability, nil
}

func (m *mockAssignmentReadModel) GetByDomainID(ctx context.Context, domainID string) ([]readmodels.AssignmentDTO, error) {
	return nil, nil
}

func TestOnCapabilityDeletedHandler_NoAssignments_NoCommandsDispatched(t *testing.T) {
	commandBus := &mockCommandBus{}
	readModel := &mockAssignmentReadModel{
		assignmentsByCapability: []readmodels.AssignmentDTO{},
	}

	handler := NewOnCapabilityDeletedHandler(commandBus, readModel)
	event := events.NewCapabilityDeleted("cap-123")

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	assert.Empty(t, commandBus.dispatchedCommands)
}

func TestOnCapabilityDeletedHandler_OneAssignment_DispatchesUnassignCommand(t *testing.T) {
	commandBus := &mockCommandBus{}
	assignment := readmodels.AssignmentDTO{
		AssignmentID:     "assign-1",
		CapabilityID:     "cap-123",
		BusinessDomainID: "bd-456",
	}
	readModel := &mockAssignmentReadModel{
		assignmentsByCapability: []readmodels.AssignmentDTO{assignment},
	}

	handler := NewOnCapabilityDeletedHandler(commandBus, readModel)
	event := events.NewCapabilityDeleted("cap-123")

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	assert.Len(t, commandBus.dispatchedCommands, 1)

	cmd, ok := commandBus.dispatchedCommands[0].(*commands.UnassignCapabilityFromDomain)
	assert.True(t, ok)
	assert.Equal(t, "assign-1", cmd.AssignmentID)
}

func TestOnCapabilityDeletedHandler_MultipleAssignments_DispatchesAllUnassignCommands(t *testing.T) {
	commandBus := &mockCommandBus{}
	assignments := []readmodels.AssignmentDTO{
		{AssignmentID: "assign-1", CapabilityID: "cap-123", BusinessDomainID: "bd-456"},
		{AssignmentID: "assign-2", CapabilityID: "cap-123", BusinessDomainID: "bd-789"},
		{AssignmentID: "assign-3", CapabilityID: "cap-123", BusinessDomainID: "bd-101"},
	}
	readModel := &mockAssignmentReadModel{
		assignmentsByCapability: assignments,
	}

	handler := NewOnCapabilityDeletedHandler(commandBus, readModel)
	event := events.NewCapabilityDeleted("cap-123")

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	assert.Len(t, commandBus.dispatchedCommands, 3)

	assignmentIDs := []string{}
	for _, cmd := range commandBus.dispatchedCommands {
		unassignCmd, ok := cmd.(*commands.UnassignCapabilityFromDomain)
		assert.True(t, ok)
		assignmentIDs = append(assignmentIDs, unassignCmd.AssignmentID)
	}

	assert.Contains(t, assignmentIDs, "assign-1")
	assert.Contains(t, assignmentIDs, "assign-2")
	assert.Contains(t, assignmentIDs, "assign-3")
}

func TestOnCapabilityDeletedHandler_ReadModelError_ReturnsError(t *testing.T) {
	commandBus := &mockCommandBus{}
	readModel := &mockAssignmentReadModel{
		queryError: errors.New("database error"),
	}

	handler := NewOnCapabilityDeletedHandler(commandBus, readModel)
	event := events.NewCapabilityDeleted("cap-123")

	err := handler.Handle(context.Background(), event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestOnCapabilityDeletedHandler_CommandDispatchError_ContinuesWithOtherCommands(t *testing.T) {
	commandBus := &mockCommandBus{
		dispatchError: errors.New("command dispatch failed"),
	}
	assignments := []readmodels.AssignmentDTO{
		{AssignmentID: "assign-1", CapabilityID: "cap-123", BusinessDomainID: "bd-456"},
		{AssignmentID: "assign-2", CapabilityID: "cap-123", BusinessDomainID: "bd-789"},
	}
	readModel := &mockAssignmentReadModel{
		assignmentsByCapability: assignments,
	}

	handler := NewOnCapabilityDeletedHandler(commandBus, readModel)
	event := events.NewCapabilityDeleted("cap-123")

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
}

func TestOnCapabilityDeletedHandler_IsIdempotent(t *testing.T) {
	commandBus := &mockCommandBus{}
	assignment := readmodels.AssignmentDTO{
		AssignmentID:     "assign-1",
		CapabilityID:     "cap-123",
		BusinessDomainID: "bd-456",
	}
	readModel := &mockAssignmentReadModel{
		assignmentsByCapability: []readmodels.AssignmentDTO{assignment},
	}

	handler := NewOnCapabilityDeletedHandler(commandBus, readModel)
	event := events.NewCapabilityDeleted("cap-123")

	err1 := handler.Handle(context.Background(), event)
	err2 := handler.Handle(context.Background(), event)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Len(t, commandBus.dispatchedCommands, 2)
}
