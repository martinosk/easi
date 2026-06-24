package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCreateCapabilityRepository struct {
	savedCapabilities []*aggregates.EnterpriseCapability
	saveErr           error
}

func (m *mockCreateCapabilityRepository) Save(ctx context.Context, capability *aggregates.EnterpriseCapability) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedCapabilities = append(m.savedCapabilities, capability)
	return nil
}

type mockCreateCapabilityReadModel struct {
	nameExists bool
	checkErr   error
}

func (m *mockCreateCapabilityReadModel) NameExists(ctx context.Context, name, excludeID string) (bool, error) {
	if m.checkErr != nil {
		return false, m.checkErr
	}
	return m.nameExists, nil
}

func runCreateCapability(repo *mockCreateCapabilityRepository, readModel *mockCreateCapabilityReadModel, cmd *commands.CreateEnterpriseCapability) (cqrs.CommandResult, error) {
	handler := NewCreateEnterpriseCapabilityHandler(repo, readModel)
	return handler.Handle(context.Background(), cmd)
}

func TestCreateEnterpriseCapabilityHandler_CreatesCapability(t *testing.T) {
	repo := &mockCreateCapabilityRepository{}
	cmd := &commands.CreateEnterpriseCapability{
		Name:        "Payroll Management",
		Description: "Manages employee payroll and compensation",
		Category:    "HR",
	}

	_, err := runCreateCapability(repo, &mockCreateCapabilityReadModel{nameExists: false}, cmd)
	require.NoError(t, err)

	require.Len(t, repo.savedCapabilities, 1)
	capability := repo.savedCapabilities[0]
	assert.Equal(t, "Payroll Management", capability.Name().Value())
	assert.Equal(t, "Manages employee payroll and compensation", capability.Description().Value())
	assert.Equal(t, "HR", capability.Category().Value())
}

func TestCreateEnterpriseCapabilityHandler_ReturnsCreatedID(t *testing.T) {
	repo := &mockCreateCapabilityRepository{}
	cmd := &commands.CreateEnterpriseCapability{Name: "Order Processing"}

	result, err := runCreateCapability(repo, &mockCreateCapabilityReadModel{nameExists: false}, cmd)
	require.NoError(t, err)

	assert.NotEmpty(t, result.CreatedID)
	assert.Equal(t, repo.savedCapabilities[0].ID(), result.CreatedID)
}

func TestCreateEnterpriseCapabilityHandler_HandlesOptionalDescriptionAndCategory(t *testing.T) {
	repo := &mockCreateCapabilityRepository{}
	cmd := &commands.CreateEnterpriseCapability{Name: "Minimal Capability"}

	_, err := runCreateCapability(repo, &mockCreateCapabilityReadModel{nameExists: false}, cmd)
	require.NoError(t, err)

	capability := repo.savedCapabilities[0]
	assert.Equal(t, "Minimal Capability", capability.Name().Value())
	assert.Empty(t, capability.Description().Value())
	assert.Empty(t, capability.Category().Value())
}

func TestCreateEnterpriseCapabilityHandler_ErrorCases(t *testing.T) {
	testCases := []struct {
		name        string
		repo        *mockCreateCapabilityRepository
		readModel   *mockCreateCapabilityReadModel
		cmd         *commands.CreateEnterpriseCapability
		wantErrIs   error
		wantNoSaves bool
	}{
		{
			name:        "name already exists",
			repo:        &mockCreateCapabilityRepository{},
			readModel:   &mockCreateCapabilityReadModel{nameExists: true},
			cmd:         &commands.CreateEnterpriseCapability{Name: "Duplicate Name", Description: "Should fail"},
			wantErrIs:   ErrEnterpriseCapabilityNameExists,
			wantNoSaves: true,
		},
		{
			name:        "invalid name",
			repo:        &mockCreateCapabilityRepository{},
			readModel:   &mockCreateCapabilityReadModel{nameExists: false},
			cmd:         &commands.CreateEnterpriseCapability{Name: "", Description: "Invalid name"},
			wantNoSaves: true,
		},
		{
			name:        "read model error",
			repo:        &mockCreateCapabilityRepository{},
			readModel:   &mockCreateCapabilityReadModel{checkErr: errors.New("database error")},
			cmd:         &commands.CreateEnterpriseCapability{Name: "Test Capability", Description: "Test"},
			wantNoSaves: true,
		},
		{
			name:      "repository error",
			repo:      &mockCreateCapabilityRepository{saveErr: errors.New("save error")},
			readModel: &mockCreateCapabilityReadModel{nameExists: false},
			cmd:       &commands.CreateEnterpriseCapability{Name: "Test Capability", Description: "Test"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := runCreateCapability(tc.repo, tc.readModel, tc.cmd)

			if tc.wantErrIs != nil {
				assert.ErrorIs(t, err, tc.wantErrIs)
			} else {
				assert.Error(t, err)
			}
			if tc.wantNoSaves {
				assert.Empty(t, tc.repo.savedCapabilities)
			}
		})
	}
}
