package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/architecturedirection/application/commands"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	"easi/backend/internal/architecturedirection/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

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

func runUpdateCapability(repo *mockUpdateCapabilityRepository, readModel *mockUpdateCapabilityReadModel, cmd *commands.UpdateEnterpriseCapability) (cqrs.CommandResult, error) {
	handler := NewUpdateEnterpriseCapabilityHandler(repo, readModel)
	return handler.Handle(context.Background(), cmd)
}

func TestUpdateEnterpriseCapabilityHandler_UpdatesCapability(t *testing.T) {
	existingCapability := createTestCapability(t, "Old Name")
	repo := &mockUpdateCapabilityRepository{existingCapability: existingCapability}

	cmd := &commands.UpdateEnterpriseCapability{
		ID:          existingCapability.ID(),
		Name:        "New Name",
		Description: "New Description",
		Category:    "New Category",
	}

	_, err := runUpdateCapability(repo, &mockUpdateCapabilityReadModel{nameExists: false}, cmd)
	require.NoError(t, err)

	require.Len(t, repo.savedCapabilities, 1)
	updated := repo.savedCapabilities[0]
	assert.Equal(t, "New Name", updated.Name().Value())
	assert.Equal(t, "New Description", updated.Description().Value())
	assert.Equal(t, "New Category", updated.Category().Value())
}

func TestUpdateEnterpriseCapabilityHandler_ExcludesSelfFromDuplicateCheck(t *testing.T) {
	existingCapability := createTestCapability(t, "Existing Name")
	repo := &mockUpdateCapabilityRepository{existingCapability: existingCapability}
	readModel := &mockUpdateCapabilityReadModel{nameExists: false}

	cmd := &commands.UpdateEnterpriseCapability{
		ID:          existingCapability.ID(),
		Name:        "Existing Name",
		Description: "Updated Description",
	}

	_, err := runUpdateCapability(repo, readModel, cmd)
	require.NoError(t, err)

	assert.Equal(t, existingCapability.ID(), readModel.excludedIDCheck)
}

func TestUpdateEnterpriseCapabilityHandler_ErrorCases(t *testing.T) {
	testCases := []struct {
		name        string
		repo        func(c *aggregates.EnterpriseCapability) *mockUpdateCapabilityRepository
		readModel   *mockUpdateCapabilityReadModel
		id          string
		wantErrIs   error
		wantNoSaves bool
	}{
		{
			name: "duplicate name",
			repo: func(c *aggregates.EnterpriseCapability) *mockUpdateCapabilityRepository {
				return &mockUpdateCapabilityRepository{existingCapability: c}
			},
			readModel:   &mockUpdateCapabilityReadModel{nameExists: true},
			wantErrIs:   ErrEnterpriseCapabilityNameExists,
			wantNoSaves: true,
		},
		{
			name: "non-existent capability",
			repo: func(c *aggregates.EnterpriseCapability) *mockUpdateCapabilityRepository {
				return &mockUpdateCapabilityRepository{getByIDErr: repositories.ErrEnterpriseCapabilityNotFound}
			},
			readModel: &mockUpdateCapabilityReadModel{nameExists: false},
			id:        "non-existent-id",
			wantErrIs: repositories.ErrEnterpriseCapabilityNotFound,
		},
		{
			name: "read model error",
			repo: func(c *aggregates.EnterpriseCapability) *mockUpdateCapabilityRepository {
				return &mockUpdateCapabilityRepository{existingCapability: c}
			},
			readModel:   &mockUpdateCapabilityReadModel{checkErr: errors.New("database error")},
			wantNoSaves: true,
		},
		{
			name: "repository error",
			repo: func(c *aggregates.EnterpriseCapability) *mockUpdateCapabilityRepository {
				return &mockUpdateCapabilityRepository{existingCapability: c, saveErr: errors.New("save error")}
			},
			readModel: &mockUpdateCapabilityReadModel{nameExists: false},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			existingCapability := createTestCapability(t, "Existing Name")
			repo := tc.repo(existingCapability)

			id := tc.id
			if id == "" {
				id = existingCapability.ID()
			}
			cmd := &commands.UpdateEnterpriseCapability{ID: id, Name: "New Name"}

			_, err := runUpdateCapability(repo, tc.readModel, cmd)
			if tc.wantErrIs != nil {
				assert.ErrorIs(t, err, tc.wantErrIs)
			} else {
				assert.Error(t, err)
			}
			if tc.wantNoSaves {
				assert.Empty(t, repo.savedCapabilities)
			}
		})
	}
}
