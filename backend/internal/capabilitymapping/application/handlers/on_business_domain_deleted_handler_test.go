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

type mockAssignmentReadModelForDomain struct {
	assignmentsByDomain []readmodels.AssignmentDTO
	queryError          error
}

func (m *mockAssignmentReadModelForDomain) GetByDomainID(ctx context.Context, domainID string) ([]readmodels.AssignmentDTO, error) {
	if m.queryError != nil {
		return nil, m.queryError
	}
	return m.assignmentsByDomain, nil
}

func (m *mockAssignmentReadModelForDomain) GetByCapabilityID(ctx context.Context, capabilityID string) ([]readmodels.AssignmentDTO, error) {
	return nil, nil
}

func TestOnBusinessDomainDeletedHandler_NoAssignments_NoCommandsDispatched(t *testing.T) {
	commandBus := &mockCommandBus{}
	readModel := &mockAssignmentReadModelForDomain{
		assignmentsByDomain: []readmodels.AssignmentDTO{},
	}

	handler := NewOnBusinessDomainDeletedHandler(commandBus, readModel)
	event := events.NewBusinessDomainDeleted("bd-123")

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	assert.Empty(t, commandBus.dispatchedCommands)
}

func TestOnBusinessDomainDeletedHandler_OneAssignment_DispatchesUnassignCommand(t *testing.T) {
	commandBus := &mockCommandBus{}
	assignment := readmodels.AssignmentDTO{
		AssignmentID:     "assign-1",
		CapabilityID:     "cap-456",
		BusinessDomainID: "bd-123",
	}
	readModel := &mockAssignmentReadModelForDomain{
		assignmentsByDomain: []readmodels.AssignmentDTO{assignment},
	}

	handler := NewOnBusinessDomainDeletedHandler(commandBus, readModel)
	event := events.NewBusinessDomainDeleted("bd-123")

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	assert.Len(t, commandBus.dispatchedCommands, 1)

	cmd, ok := commandBus.dispatchedCommands[0].(*commands.UnassignCapabilityFromDomain)
	assert.True(t, ok)
	assert.Equal(t, "assign-1", cmd.AssignmentID)
}

func TestOnBusinessDomainDeletedHandler_MultipleAssignments_DispatchesAllUnassignCommands(t *testing.T) {
	commandBus := &mockCommandBus{}
	assignments := []readmodels.AssignmentDTO{
		{AssignmentID: "assign-1", CapabilityID: "cap-456", BusinessDomainID: "bd-123"},
		{AssignmentID: "assign-2", CapabilityID: "cap-789", BusinessDomainID: "bd-123"},
		{AssignmentID: "assign-3", CapabilityID: "cap-101", BusinessDomainID: "bd-123"},
	}
	readModel := &mockAssignmentReadModelForDomain{
		assignmentsByDomain: assignments,
	}

	handler := NewOnBusinessDomainDeletedHandler(commandBus, readModel)
	event := events.NewBusinessDomainDeleted("bd-123")

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

func TestOnBusinessDomainDeletedHandler_ReadModelError_ReturnsError(t *testing.T) {
	commandBus := &mockCommandBus{}
	readModel := &mockAssignmentReadModelForDomain{
		queryError: errors.New("database error"),
	}

	handler := NewOnBusinessDomainDeletedHandler(commandBus, readModel)
	event := events.NewBusinessDomainDeleted("bd-123")

	err := handler.Handle(context.Background(), event)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database error")
}

func TestOnBusinessDomainDeletedHandler_CommandDispatchError_ContinuesWithOtherCommands(t *testing.T) {
	commandBus := &mockCommandBus{
		dispatchError: errors.New("command dispatch failed"),
	}
	assignments := []readmodels.AssignmentDTO{
		{AssignmentID: "assign-1", CapabilityID: "cap-456", BusinessDomainID: "bd-123"},
		{AssignmentID: "assign-2", CapabilityID: "cap-789", BusinessDomainID: "bd-123"},
	}
	readModel := &mockAssignmentReadModelForDomain{
		assignmentsByDomain: assignments,
	}

	handler := NewOnBusinessDomainDeletedHandler(commandBus, readModel)
	event := events.NewBusinessDomainDeleted("bd-123")

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
}

func TestOnBusinessDomainDeletedHandler_IsIdempotent(t *testing.T) {
	commandBus := &mockCommandBus{}
	assignment := readmodels.AssignmentDTO{
		AssignmentID:     "assign-1",
		CapabilityID:     "cap-456",
		BusinessDomainID: "bd-123",
	}
	readModel := &mockAssignmentReadModelForDomain{
		assignmentsByDomain: []readmodels.AssignmentDTO{assignment},
	}

	handler := NewOnBusinessDomainDeletedHandler(commandBus, readModel)
	event := events.NewBusinessDomainDeleted("bd-123")

	err1 := handler.Handle(context.Background(), event)
	err2 := handler.Handle(context.Background(), event)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Len(t, commandBus.dispatchedCommands, 2)
}
