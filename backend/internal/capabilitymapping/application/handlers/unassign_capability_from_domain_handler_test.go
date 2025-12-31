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

type mockUnassignCapabilityRepository struct {
	assignment *aggregates.BusinessDomainAssignment
	savedCount int
	getByIDErr error
	saveErr    error
}

func (m *mockUnassignCapabilityRepository) GetByID(ctx context.Context, id string) (*aggregates.BusinessDomainAssignment, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.assignment, nil
}

func (m *mockUnassignCapabilityRepository) Save(ctx context.Context, assignment *aggregates.BusinessDomainAssignment) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedCount++
	return nil
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

	mockRepo := &mockUnassignCapabilityRepository{assignment: assignment}

	handler := NewUnassignCapabilityFromDomainHandler(mockRepo)

	cmd := &commands.UnassignCapabilityFromDomain{
		AssignmentID: assignmentID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, 1, mockRepo.savedCount, "Handler should save assignment once")
}

func TestUnassignCapabilityFromDomainHandler_AssignmentNotFound_ReturnsError(t *testing.T) {
	mockRepo := &mockUnassignCapabilityRepository{
		getByIDErr: repositories.ErrAssignmentNotFound,
	}

	handler := NewUnassignCapabilityFromDomainHandler(mockRepo)

	cmd := &commands.UnassignCapabilityFromDomain{
		AssignmentID: "non-existent",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, repositories.ErrAssignmentNotFound)
}

func TestUnassignCapabilityFromDomainHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockUnassignCapabilityRepository{}

	handler := NewUnassignCapabilityFromDomainHandler(mockRepo)

	invalidCmd := &commands.DeleteBusinessDomain{}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}
