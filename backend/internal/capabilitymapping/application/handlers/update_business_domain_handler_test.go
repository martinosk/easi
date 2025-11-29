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

type mockBusinessDomainRepositoryForUpdate struct {
	domain     *aggregates.BusinessDomain
	savedCount int
	getByIDErr error
	saveErr    error
}

func (m *mockBusinessDomainRepositoryForUpdate) GetByID(ctx context.Context, id string) (*aggregates.BusinessDomain, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.domain, nil
}

func (m *mockBusinessDomainRepositoryForUpdate) Save(ctx context.Context, domain *aggregates.BusinessDomain) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedCount++
	return nil
}

type mockBusinessDomainReadModelForUpdate struct {
	nameExists bool
	checkErr   error
}

func (m *mockBusinessDomainReadModelForUpdate) NameExists(ctx context.Context, name, excludeID string) (bool, error) {
	if m.checkErr != nil {
		return false, m.checkErr
	}
	return m.nameExists, nil
}

type businessDomainRepositoryForUpdate interface {
	GetByID(ctx context.Context, id string) (*aggregates.BusinessDomain, error)
	Save(ctx context.Context, domain *aggregates.BusinessDomain) error
}

type businessDomainReadModelForUpdate interface {
	NameExists(ctx context.Context, name, excludeID string) (bool, error)
}

type testableUpdateBusinessDomainHandler struct {
	repository businessDomainRepositoryForUpdate
	readModel  businessDomainReadModelForUpdate
}

func newTestableUpdateBusinessDomainHandler(
	repository businessDomainRepositoryForUpdate,
	readModel businessDomainReadModelForUpdate,
) *testableUpdateBusinessDomainHandler {
	return &testableUpdateBusinessDomainHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *testableUpdateBusinessDomainHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.UpdateBusinessDomain)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	domain, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	exists, err := h.readModel.NameExists(ctx, command.Name, command.ID)
	if err != nil {
		return err
	}
	if exists {
		return ErrBusinessDomainNameExists
	}

	name, err := valueobjects.NewDomainName(command.Name)
	if err != nil {
		return err
	}

	description := valueobjects.NewDescription(command.Description)

	if err := domain.Update(name, description); err != nil {
		return err
	}

	return h.repository.Save(ctx, domain)
}

func createTestBusinessDomain(t *testing.T, name, description string) *aggregates.BusinessDomain {
	t.Helper()

	domainName, err := valueobjects.NewDomainName(name)
	require.NoError(t, err)

	desc := valueobjects.NewDescription(description)

	domain, err := aggregates.NewBusinessDomain(domainName, desc)
	require.NoError(t, err)
	domain.MarkChangesAsCommitted()

	return domain
}

func TestUpdateBusinessDomainHandler_UpdatesBusinessDomain(t *testing.T) {
	domain := createTestBusinessDomain(t, "Original Name", "Original Description")
	domainID := domain.ID()

	mockRepo := &mockBusinessDomainRepositoryForUpdate{domain: domain}
	mockReadModel := &mockBusinessDomainReadModelForUpdate{nameExists: false}

	handler := newTestableUpdateBusinessDomainHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateBusinessDomain{
		ID:          domainID,
		Name:        "Updated Name",
		Description: "Updated Description",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, 1, mockRepo.savedCount, "Handler should save domain once")
	assert.Equal(t, "Updated Name", domain.Name().Value())
	assert.Equal(t, "Updated Description", domain.Description().Value())
}

func TestUpdateBusinessDomainHandler_NameExistsForOtherDomain_ReturnsError(t *testing.T) {
	domain := createTestBusinessDomain(t, "Original Name", "Description")
	domainID := domain.ID()

	mockRepo := &mockBusinessDomainRepositoryForUpdate{domain: domain}
	mockReadModel := &mockBusinessDomainReadModelForUpdate{nameExists: true}

	handler := newTestableUpdateBusinessDomainHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateBusinessDomain{
		ID:          domainID,
		Name:        "Duplicate Name",
		Description: "Description",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrBusinessDomainNameExists)
	assert.Equal(t, 0, mockRepo.savedCount, "Should not save when name exists")
}

func TestUpdateBusinessDomainHandler_DomainNotFound_ReturnsError(t *testing.T) {
	mockRepo := &mockBusinessDomainRepositoryForUpdate{
		getByIDErr: repositories.ErrBusinessDomainNotFound,
	}
	mockReadModel := &mockBusinessDomainReadModelForUpdate{nameExists: false}

	handler := newTestableUpdateBusinessDomainHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateBusinessDomain{
		ID:          "non-existent",
		Name:        "Name",
		Description: "Description",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, repositories.ErrBusinessDomainNotFound)
}

func TestUpdateBusinessDomainHandler_InvalidName_ReturnsError(t *testing.T) {
	domain := createTestBusinessDomain(t, "Original Name", "Description")
	domainID := domain.ID()

	mockRepo := &mockBusinessDomainRepositoryForUpdate{domain: domain}
	mockReadModel := &mockBusinessDomainReadModelForUpdate{nameExists: false}

	handler := newTestableUpdateBusinessDomainHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateBusinessDomain{
		ID:          domainID,
		Name:        "",
		Description: "Description",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, 0, mockRepo.savedCount, "Should not save with invalid name")
}

func TestUpdateBusinessDomainHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockBusinessDomainRepositoryForUpdate{}
	mockReadModel := &mockBusinessDomainReadModelForUpdate{}

	handler := newTestableUpdateBusinessDomainHandler(mockRepo, mockReadModel)

	invalidCmd := &commands.DeleteBusinessDomain{}

	err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestUpdateBusinessDomainHandler_ReadModelError_ReturnsError(t *testing.T) {
	domain := createTestBusinessDomain(t, "Original Name", "Description")
	domainID := domain.ID()

	mockRepo := &mockBusinessDomainRepositoryForUpdate{domain: domain}
	mockReadModel := &mockBusinessDomainReadModelForUpdate{checkErr: errors.New("database error")}

	handler := newTestableUpdateBusinessDomainHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateBusinessDomain{
		ID:          domainID,
		Name:        "New Name",
		Description: "Description",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, 0, mockRepo.savedCount)
}
