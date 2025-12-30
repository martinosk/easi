package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEnterpriseCapabilityRepositoryForUpdate struct {
	savedCapabilities  []*aggregates.EnterpriseCapability
	existingCapability *aggregates.EnterpriseCapability
	saveErr            error
	getByIDErr         error
}

func (m *mockEnterpriseCapabilityRepositoryForUpdate) Save(ctx context.Context, capability *aggregates.EnterpriseCapability) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedCapabilities = append(m.savedCapabilities, capability)
	return nil
}

func (m *mockEnterpriseCapabilityRepositoryForUpdate) GetByID(ctx context.Context, id string) (*aggregates.EnterpriseCapability, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.existingCapability != nil && m.existingCapability.ID() == id {
		return m.existingCapability, nil
	}
	return nil, repositories.ErrEnterpriseCapabilityNotFound
}

type mockEnterpriseCapabilityReadModelForUpdate struct {
	nameExists      bool
	checkErr        error
	excludedIDCheck string
}

func (m *mockEnterpriseCapabilityReadModelForUpdate) NameExists(ctx context.Context, name, excludeID string) (bool, error) {
	m.excludedIDCheck = excludeID
	if m.checkErr != nil {
		return false, m.checkErr
	}
	return m.nameExists, nil
}

type enterpriseCapabilityRepositoryForUpdate interface {
	Save(ctx context.Context, capability *aggregates.EnterpriseCapability) error
	GetByID(ctx context.Context, id string) (*aggregates.EnterpriseCapability, error)
}

type testableUpdateEnterpriseCapabilityHandler struct {
	repository enterpriseCapabilityRepositoryForUpdate
	readModel  enterpriseCapabilityReadModelForCreate
}

func newTestableUpdateEnterpriseCapabilityHandler(
	repository enterpriseCapabilityRepositoryForUpdate,
	readModel enterpriseCapabilityReadModelForCreate,
) *testableUpdateEnterpriseCapabilityHandler {
	return &testableUpdateEnterpriseCapabilityHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *testableUpdateEnterpriseCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdateEnterpriseCapability)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	exists, err := h.readModel.NameExists(ctx, command.Name, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}
	if exists {
		return cqrs.EmptyResult(), ErrEnterpriseCapabilityNameExists
	}

	capability, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	name, err := valueobjects.NewEnterpriseCapabilityName(command.Name)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	description, err := valueobjects.NewDescription(command.Description)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	category, err := valueobjects.NewCategory(command.Category)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := capability.Update(name, description, category); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, capability); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.EmptyResult(), nil
}

func createTestEnterpriseCapability(t *testing.T, name string) *aggregates.EnterpriseCapability {
	t.Helper()
	capName, _ := valueobjects.NewEnterpriseCapabilityName(name)
	description := valueobjects.MustNewDescription("Test description")
	category, _ := valueobjects.NewCategory("Test")

	capability, err := aggregates.NewEnterpriseCapability(capName, description, category)
	require.NoError(t, err)
	capability.MarkChangesAsCommitted()
	return capability
}

func TestUpdateEnterpriseCapabilityHandler_UpdatesCapability(t *testing.T) {
	existingCapability := createTestEnterpriseCapability(t, "Old Name")

	mockRepo := &mockEnterpriseCapabilityRepositoryForUpdate{
		existingCapability: existingCapability,
	}
	mockReadModel := &mockEnterpriseCapabilityReadModelForUpdate{nameExists: false}

	handler := newTestableUpdateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateEnterpriseCapability{
		ID:          existingCapability.ID(),
		Name:        "New Name",
		Description: "New Description",
		Category:    "New Category",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedCapabilities, 1)
	updated := mockRepo.savedCapabilities[0]
	assert.Equal(t, "New Name", updated.Name().Value())
	assert.Equal(t, "New Description", updated.Description().Value())
	assert.Equal(t, "New Category", updated.Category().Value())
}

func TestUpdateEnterpriseCapabilityHandler_ExcludesSelfFromDuplicateCheck(t *testing.T) {
	existingCapability := createTestEnterpriseCapability(t, "Existing Name")

	mockRepo := &mockEnterpriseCapabilityRepositoryForUpdate{
		existingCapability: existingCapability,
	}
	mockReadModel := &mockEnterpriseCapabilityReadModelForUpdate{nameExists: false}

	handler := newTestableUpdateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateEnterpriseCapability{
		ID:          existingCapability.ID(),
		Name:        "Existing Name",
		Description: "Updated Description",
		Category:    "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, existingCapability.ID(), mockReadModel.excludedIDCheck)
}

func TestUpdateEnterpriseCapabilityHandler_DuplicateName_ReturnsError(t *testing.T) {
	existingCapability := createTestEnterpriseCapability(t, "Existing Name")

	mockRepo := &mockEnterpriseCapabilityRepositoryForUpdate{
		existingCapability: existingCapability,
	}
	mockReadModel := &mockEnterpriseCapabilityReadModelForUpdate{nameExists: true}

	handler := newTestableUpdateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateEnterpriseCapability{
		ID:          existingCapability.ID(),
		Name:        "Duplicate Name",
		Description: "Should fail",
		Category:    "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrEnterpriseCapabilityNameExists)
	assert.Empty(t, mockRepo.savedCapabilities)
}

func TestUpdateEnterpriseCapabilityHandler_NonExistent_ReturnsError(t *testing.T) {
	mockRepo := &mockEnterpriseCapabilityRepositoryForUpdate{
		getByIDErr: repositories.ErrEnterpriseCapabilityNotFound,
	}
	mockReadModel := &mockEnterpriseCapabilityReadModelForUpdate{nameExists: false}

	handler := newTestableUpdateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateEnterpriseCapability{
		ID:          "non-existent-id",
		Name:        "New Name",
		Description: "",
		Category:    "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, repositories.ErrEnterpriseCapabilityNotFound)
}

func TestUpdateEnterpriseCapabilityHandler_ReadModelError_ReturnsError(t *testing.T) {
	existingCapability := createTestEnterpriseCapability(t, "Existing Name")

	mockRepo := &mockEnterpriseCapabilityRepositoryForUpdate{
		existingCapability: existingCapability,
	}
	mockReadModel := &mockEnterpriseCapabilityReadModelForUpdate{checkErr: errors.New("database error")}

	handler := newTestableUpdateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateEnterpriseCapability{
		ID:          existingCapability.ID(),
		Name:        "New Name",
		Description: "",
		Category:    "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockRepo.savedCapabilities)
}

func TestUpdateEnterpriseCapabilityHandler_RepositoryError_ReturnsError(t *testing.T) {
	existingCapability := createTestEnterpriseCapability(t, "Existing Name")

	mockRepo := &mockEnterpriseCapabilityRepositoryForUpdate{
		existingCapability: existingCapability,
		saveErr:            errors.New("save error"),
	}
	mockReadModel := &mockEnterpriseCapabilityReadModelForUpdate{nameExists: false}

	handler := newTestableUpdateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateEnterpriseCapability{
		ID:          existingCapability.ID(),
		Name:        "New Name",
		Description: "",
		Category:    "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}
