package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
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

type mockBusinessDomainDeletionService struct {
	canDeleteErr error
}

func (m *mockBusinessDomainDeletionService) CanDelete(ctx context.Context, domainID valueobjects.BusinessDomainID) error {
	return m.canDeleteErr
}

func TestDeleteBusinessDomainHandler_DeletesBusinessDomain(t *testing.T) {
	domain := createTestBusinessDomain(t, "Test Domain", "Description")
	domainID := domain.ID()

	mockRepo := &mockDeleteBusinessDomainRepository{domain: domain}
	mockDeletionService := &mockBusinessDomainDeletionService{}

	handler := NewDeleteBusinessDomainHandler(mockRepo, mockDeletionService)

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
	mockDeletionService := &mockBusinessDomainDeletionService{
		canDeleteErr: services.ErrBusinessDomainHasAssignments,
	}

	handler := NewDeleteBusinessDomainHandler(mockRepo, mockDeletionService)

	cmd := &commands.DeleteBusinessDomain{
		ID: domainID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, services.ErrBusinessDomainHasAssignments)
	assert.Equal(t, 0, mockRepo.savedCount, "Should not save when domain has assignments")
}

func TestDeleteBusinessDomainHandler_DomainNotFound_ReturnsError(t *testing.T) {
	mockRepo := &mockDeleteBusinessDomainRepository{
		getByIDErr: repositories.ErrBusinessDomainNotFound,
	}
	mockDeletionService := &mockBusinessDomainDeletionService{}

	handler := NewDeleteBusinessDomainHandler(mockRepo, mockDeletionService)

	cmd := &commands.DeleteBusinessDomain{
		ID: valueobjects.NewBusinessDomainID().Value(),
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrBusinessDomainNotFound)
}

func TestDeleteBusinessDomainHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockDeleteBusinessDomainRepository{}
	mockDeletionService := &mockBusinessDomainDeletionService{}

	handler := NewDeleteBusinessDomainHandler(mockRepo, mockDeletionService)

	invalidCmd := &commands.CreateBusinessDomain{}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestDeleteBusinessDomainHandler_DeletionServiceError_ReturnsError(t *testing.T) {
	domain := createTestBusinessDomain(t, "Test Domain", "Description")
	domainID := domain.ID()

	mockRepo := &mockDeleteBusinessDomainRepository{domain: domain}
	mockDeletionService := &mockBusinessDomainDeletionService{
		canDeleteErr: errors.New("database error"),
	}

	handler := NewDeleteBusinessDomainHandler(mockRepo, mockDeletionService)

	cmd := &commands.DeleteBusinessDomain{
		ID: domainID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, 0, mockRepo.savedCount)
}
