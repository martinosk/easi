package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDeleteBusinessDomainRepository struct {
	domain     *aggregates.BusinessDomain
	savedCount int
	getByIDErr error
	saveErr    error
}

func (m *mockDeleteBusinessDomainRepository) GetByID(ctx context.Context, id string) (*aggregates.BusinessDomain, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.domain, nil
}

func (m *mockDeleteBusinessDomainRepository) Save(ctx context.Context, domain *aggregates.BusinessDomain) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedCount++
	return nil
}

type mockDeleteBusinessDomainAssignmentReader struct {
	assignments []readmodels.AssignmentDTO
	getErr      error
}

func (m *mockDeleteBusinessDomainAssignmentReader) GetByDomainID(ctx context.Context, domainID string) ([]readmodels.AssignmentDTO, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.assignments, nil
}

func TestDeleteBusinessDomainHandler_DeletesBusinessDomain(t *testing.T) {
	domain := createTestBusinessDomain(t, "Test Domain", "Description")
	domainID := domain.ID()

	mockRepo := &mockDeleteBusinessDomainRepository{domain: domain}
	mockAssignmentReader := &mockDeleteBusinessDomainAssignmentReader{assignments: []readmodels.AssignmentDTO{}}

	handler := NewDeleteBusinessDomainHandler(mockRepo, mockAssignmentReader)

	cmd := &commands.DeleteBusinessDomain{
		ID: domainID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, 1, mockRepo.savedCount, "Handler should save domain once")
}

func TestDeleteBusinessDomainHandler_DomainHasAssignments_ReturnsError(t *testing.T) {
	domain := createTestBusinessDomain(t, "Test Domain", "Description")
	domainID := domain.ID()

	mockRepo := &mockDeleteBusinessDomainRepository{domain: domain}
	mockAssignmentReader := &mockDeleteBusinessDomainAssignmentReader{
		assignments: []readmodels.AssignmentDTO{
			{AssignmentID: "assignment-1"},
		},
	}

	handler := NewDeleteBusinessDomainHandler(mockRepo, mockAssignmentReader)

	cmd := &commands.DeleteBusinessDomain{
		ID: domainID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrBusinessDomainHasAssignments)
	assert.Equal(t, 0, mockRepo.savedCount, "Should not save when domain has assignments")
}

func TestDeleteBusinessDomainHandler_DomainNotFound_ReturnsError(t *testing.T) {
	mockRepo := &mockDeleteBusinessDomainRepository{
		getByIDErr: repositories.ErrBusinessDomainNotFound,
	}
	mockAssignmentReader := &mockDeleteBusinessDomainAssignmentReader{assignments: []readmodels.AssignmentDTO{}}

	handler := NewDeleteBusinessDomainHandler(mockRepo, mockAssignmentReader)

	cmd := &commands.DeleteBusinessDomain{
		ID: "non-existent",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrBusinessDomainNotFound)
}

func TestDeleteBusinessDomainHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockDeleteBusinessDomainRepository{}
	mockAssignmentReader := &mockDeleteBusinessDomainAssignmentReader{}

	handler := NewDeleteBusinessDomainHandler(mockRepo, mockAssignmentReader)

	invalidCmd := &commands.CreateBusinessDomain{}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestDeleteBusinessDomainHandler_AssignmentReaderError_ReturnsError(t *testing.T) {
	domain := createTestBusinessDomain(t, "Test Domain", "Description")
	domainID := domain.ID()

	mockRepo := &mockDeleteBusinessDomainRepository{domain: domain}
	mockAssignmentReader := &mockDeleteBusinessDomainAssignmentReader{
		getErr: errors.New("database error"),
	}

	handler := NewDeleteBusinessDomainHandler(mockRepo, mockAssignmentReader)

	cmd := &commands.DeleteBusinessDomain{
		ID: domainID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, 0, mockRepo.savedCount)
}
