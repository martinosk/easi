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

func TestDeleteBusinessDomainHandler_WithExistingDomain(t *testing.T) {
	tests := []struct {
		name              string
		canDeleteErr      error
		expectErr         bool
		expectedIs        error
		expectedSaveCount int
		msg               string
	}{
		{
			name:              "deletes business domain",
			expectedSaveCount: 1,
			msg:               "Handler should save domain once",
		},
		{
			name:              "domain has assignments",
			canDeleteErr:      services.ErrBusinessDomainHasAssignments,
			expectErr:         true,
			expectedIs:        services.ErrBusinessDomainHasAssignments,
			expectedSaveCount: 0,
			msg:               "Should not save when domain has assignments",
		},
		{
			name:              "deletion service error",
			canDeleteErr:      errors.New("database error"),
			expectErr:         true,
			expectedSaveCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain := createTestBusinessDomain(t, "Test Domain", "Description")
			mockRepo := &mockDeleteBusinessDomainRepository{domain: domain}
			mockDeletionService := &mockBusinessDomainDeletionService{canDeleteErr: tt.canDeleteErr}
			handler := NewDeleteBusinessDomainHandler(mockRepo, mockDeletionService)

			cmd := &commands.DeleteBusinessDomain{ID: domain.ID()}

			_, err := handler.Handle(context.Background(), cmd)

			if tt.expectErr {
				if tt.expectedIs != nil {
					assert.ErrorIs(t, err, tt.expectedIs)
				} else {
					assert.Error(t, err)
				}
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expectedSaveCount, mockRepo.savedCount, tt.msg)
		})
	}
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
