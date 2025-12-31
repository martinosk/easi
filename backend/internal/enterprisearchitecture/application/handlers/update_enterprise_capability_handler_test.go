package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUpdateCapabilityRepository struct {
	savedCapabilities  []*aggregates.EnterpriseCapability
	existingCapability *aggregates.EnterpriseCapability
	saveErr            error
	getByIDErr         error
}

func (m *mockUpdateCapabilityRepository) Save(ctx context.Context, capability *aggregates.EnterpriseCapability) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedCapabilities = append(m.savedCapabilities, capability)
	return nil
}

func (m *mockUpdateCapabilityRepository) GetByID(ctx context.Context, id string) (*aggregates.EnterpriseCapability, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.existingCapability != nil && m.existingCapability.ID() == id {
		return m.existingCapability, nil
	}
	return nil, repositories.ErrEnterpriseCapabilityNotFound
}

type mockUpdateCapabilityReadModel struct {
	nameExists      bool
	checkErr        error
	excludedIDCheck string
}

func (m *mockUpdateCapabilityReadModel) NameExists(ctx context.Context, name, excludeID string) (bool, error) {
	m.excludedIDCheck = excludeID
	if m.checkErr != nil {
		return false, m.checkErr
	}
	return m.nameExists, nil
}

func createTestCapability(t *testing.T, name string) *aggregates.EnterpriseCapability {
	t.Helper()
	capName, err := valueobjects.NewEnterpriseCapabilityName(name)
	require.NoError(t, err)
	description, err := valueobjects.NewDescription("Test description")
	require.NoError(t, err)
	category, err := valueobjects.NewCategory("Test")
	require.NoError(t, err)

	capability, err := aggregates.NewEnterpriseCapability(capName, description, category)
	require.NoError(t, err)
	capability.MarkChangesAsCommitted()
	return capability
}

func TestUpdateEnterpriseCapabilityHandler_UpdatesCapability(t *testing.T) {
	existingCapability := createTestCapability(t, "Old Name")

	mockRepo := &mockUpdateCapabilityRepository{existingCapability: existingCapability}
	mockReadModel := &mockUpdateCapabilityReadModel{nameExists: false}

	handler := NewUpdateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

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
	existingCapability := createTestCapability(t, "Existing Name")

	mockRepo := &mockUpdateCapabilityRepository{existingCapability: existingCapability}
	mockReadModel := &mockUpdateCapabilityReadModel{nameExists: false}

	handler := NewUpdateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

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
	existingCapability := createTestCapability(t, "Existing Name")

	mockRepo := &mockUpdateCapabilityRepository{existingCapability: existingCapability}
	mockReadModel := &mockUpdateCapabilityReadModel{nameExists: true}

	handler := NewUpdateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

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
	mockRepo := &mockUpdateCapabilityRepository{getByIDErr: repositories.ErrEnterpriseCapabilityNotFound}
	mockReadModel := &mockUpdateCapabilityReadModel{nameExists: false}

	handler := NewUpdateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

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
	existingCapability := createTestCapability(t, "Existing Name")

	mockRepo := &mockUpdateCapabilityRepository{existingCapability: existingCapability}
	mockReadModel := &mockUpdateCapabilityReadModel{checkErr: errors.New("database error")}

	handler := NewUpdateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

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
	existingCapability := createTestCapability(t, "Existing Name")

	mockRepo := &mockUpdateCapabilityRepository{
		existingCapability: existingCapability,
		saveErr:            errors.New("save error"),
	}
	mockReadModel := &mockUpdateCapabilityReadModel{nameExists: false}

	handler := NewUpdateEnterpriseCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.UpdateEnterpriseCapability{
		ID:          existingCapability.ID(),
		Name:        "New Name",
		Description: "",
		Category:    "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}
