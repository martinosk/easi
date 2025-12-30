package handlers

import (
	"context"
	"testing"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAssignmentRepositoryForUnassign struct {
	assignment *aggregates.BusinessDomainAssignment
	savedCount int
	getByIDErr error
	saveErr    error
}

func (m *mockAssignmentRepositoryForUnassign) GetByID(ctx context.Context, id string) (*aggregates.BusinessDomainAssignment, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.assignment, nil
}

func (m *mockAssignmentRepositoryForUnassign) Save(ctx context.Context, assignment *aggregates.BusinessDomainAssignment) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedCount++
	return nil
}

type assignmentRepositoryForUnassign interface {
	GetByID(ctx context.Context, id string) (*aggregates.BusinessDomainAssignment, error)
	Save(ctx context.Context, assignment *aggregates.BusinessDomainAssignment) error
}

type testableUnassignCapabilityFromDomainHandler struct {
	repository assignmentRepositoryForUnassign
}

func newTestableUnassignCapabilityFromDomainHandler(
	repository assignmentRepositoryForUnassign,
) *testableUnassignCapabilityFromDomainHandler {
	return &testableUnassignCapabilityFromDomainHandler{
		repository: repository,
	}
}

func (h *testableUnassignCapabilityFromDomainHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UnassignCapabilityFromDomain)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	assignment, err := h.repository.GetByID(ctx, command.AssignmentID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := assignment.Unassign(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, assignment); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.EmptyResult(), nil
}

func createTestAssignment(t *testing.T) *aggregates.BusinessDomainAssignment {
	t.Helper()

	businessDomainID := valueobjects.NewBusinessDomainID()
	capabilityID := valueobjects.NewCapabilityID()

	assignment, err := aggregates.AssignCapabilityToDomain(businessDomainID, capabilityID)
	require.NoError(t, err)
	assignment.MarkChangesAsCommitted()

	return assignment
}

func TestUnassignCapabilityFromDomainHandler_UnassignsCapability(t *testing.T) {
	assignment := createTestAssignment(t)
	assignmentID := assignment.ID()

	mockRepo := &mockAssignmentRepositoryForUnassign{assignment: assignment}

	handler := newTestableUnassignCapabilityFromDomainHandler(mockRepo)

	cmd := &commands.UnassignCapabilityFromDomain{
		AssignmentID: assignmentID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, 1, mockRepo.savedCount, "Handler should save assignment once")
}

func TestUnassignCapabilityFromDomainHandler_AssignmentNotFound_ReturnsError(t *testing.T) {
	mockRepo := &mockAssignmentRepositoryForUnassign{
		getByIDErr: repositories.ErrAssignmentNotFound,
	}

	handler := newTestableUnassignCapabilityFromDomainHandler(mockRepo)

	cmd := &commands.UnassignCapabilityFromDomain{
		AssignmentID: "non-existent",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, repositories.ErrAssignmentNotFound)
}

func TestUnassignCapabilityFromDomainHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockAssignmentRepositoryForUnassign{}

	handler := newTestableUnassignCapabilityFromDomainHandler(mockRepo)

	invalidCmd := &commands.DeleteBusinessDomain{}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}
