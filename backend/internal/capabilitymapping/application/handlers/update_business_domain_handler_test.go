package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUpdateBusinessDomainRepository struct {
	domain     *aggregates.BusinessDomain
	savedCount int
	getByIDErr error
	saveErr    error
}

func (m *mockUpdateBusinessDomainRepository) GetByID(ctx context.Context, id string) (*aggregates.BusinessDomain, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.domain, nil
}

func (m *mockUpdateBusinessDomainRepository) Save(ctx context.Context, domain *aggregates.BusinessDomain) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedCount++
	return nil
}

type mockUpdateBusinessDomainReadModel struct {
	nameExists bool
	checkErr   error
}

func (m *mockUpdateBusinessDomainReadModel) NameExists(ctx context.Context, name, excludeID string) (bool, error) {
	if m.checkErr != nil {
		return false, m.checkErr
	}
	return m.nameExists, nil
}

func createTestBusinessDomain(t *testing.T, name, description string) *aggregates.BusinessDomain {
	t.Helper()

	domainName, err := valueobjects.NewDomainName(name)
	require.NoError(t, err)

	desc := valueobjects.MustNewDescription(description)

	domain, err := aggregates.NewBusinessDomain(domainName, desc)
	require.NoError(t, err)
	domain.MarkChangesAsCommitted()

	return domain
}

func TestUpdateBusinessDomainHandler_UpdatesBusinessDomain(t *testing.T) {
	domain := createTestBusinessDomain(t, "Original Name", "Original Description")
	domainID := domain.ID()

	mockRepo := &mockUpdateBusinessDomainRepository{domain: domain}
	mockReadModel := &mockUpdateBusinessDomainReadModel{nameExists: false}

	handler := NewUpdateBusinessDomainHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateBusinessDomain{
		ID:          domainID,
		Name:        "Updated Name",
		Description: "Updated Description",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, 1, mockRepo.savedCount, "Handler should save domain once")
	assert.Equal(t, "Updated Name", domain.Name().Value())
	assert.Equal(t, "Updated Description", domain.Description().Value())
}

func TestUpdateBusinessDomainHandler_NameExistsForOtherDomain_ReturnsError(t *testing.T) {
	domain := createTestBusinessDomain(t, "Original Name", "Description")
	domainID := domain.ID()

	mockRepo := &mockUpdateBusinessDomainRepository{domain: domain}
	mockReadModel := &mockUpdateBusinessDomainReadModel{nameExists: true}

	handler := NewUpdateBusinessDomainHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateBusinessDomain{
		ID:          domainID,
		Name:        "Duplicate Name",
		Description: "Description",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrBusinessDomainNameExists)
	assert.Equal(t, 0, mockRepo.savedCount, "Should not save when name exists")
}

func TestUpdateBusinessDomainHandler_DomainNotFound_ReturnsError(t *testing.T) {
	mockRepo := &mockUpdateBusinessDomainRepository{
		getByIDErr: repositories.ErrBusinessDomainNotFound,
	}
	mockReadModel := &mockUpdateBusinessDomainReadModel{nameExists: false}

	handler := NewUpdateBusinessDomainHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateBusinessDomain{
		ID:          "non-existent",
		Name:        "Name",
		Description: "Description",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrBusinessDomainNotFound)
}

func TestUpdateBusinessDomainHandler_InvalidName_ReturnsError(t *testing.T) {
	domain := createTestBusinessDomain(t, "Original Name", "Description")
	domainID := domain.ID()

	mockRepo := &mockUpdateBusinessDomainRepository{domain: domain}
	mockReadModel := &mockUpdateBusinessDomainReadModel{nameExists: false}

	handler := NewUpdateBusinessDomainHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateBusinessDomain{
		ID:          domainID,
		Name:        "",
		Description: "Description",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, 0, mockRepo.savedCount, "Should not save with invalid name")
}

func TestUpdateBusinessDomainHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockUpdateBusinessDomainRepository{}
	mockReadModel := &mockUpdateBusinessDomainReadModel{}

	handler := NewUpdateBusinessDomainHandler(mockRepo, mockReadModel)

	invalidCmd := &commands.DeleteBusinessDomain{}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestUpdateBusinessDomainHandler_ReadModelError_ReturnsError(t *testing.T) {
	domain := createTestBusinessDomain(t, "Original Name", "Description")
	domainID := domain.ID()

	mockRepo := &mockUpdateBusinessDomainRepository{domain: domain}
	mockReadModel := &mockUpdateBusinessDomainReadModel{checkErr: errors.New("database error")}

	handler := NewUpdateBusinessDomainHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateBusinessDomain{
		ID:          domainID,
		Name:        "New Name",
		Description: "Description",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, 0, mockRepo.savedCount)
}
