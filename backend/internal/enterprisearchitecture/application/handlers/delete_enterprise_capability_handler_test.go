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

type mockDeleteCapabilityRepository struct {
	savedCapabilities  []*aggregates.EnterpriseCapability
	existingCapability *aggregates.EnterpriseCapability
	saveErr            error
	getByIDErr         error
}

func (m *mockDeleteCapabilityRepository) Save(ctx context.Context, capability *aggregates.EnterpriseCapability) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedCapabilities = append(m.savedCapabilities, capability)
	return nil
}

func (m *mockDeleteCapabilityRepository) GetByID(ctx context.Context, id string) (*aggregates.EnterpriseCapability, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.existingCapability != nil && m.existingCapability.ID() == id {
		return m.existingCapability, nil
	}
	return nil, repositories.ErrEnterpriseCapabilityNotFound
}

func createDeleteTestCapability(t *testing.T, name string) *aggregates.EnterpriseCapability {
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

func TestDeleteEnterpriseCapabilityHandler_DeletesCapability(t *testing.T) {
	existingCapability := createDeleteTestCapability(t, "To Be Deleted")

	mockRepo := &mockDeleteCapabilityRepository{existingCapability: existingCapability}

	handler := NewDeleteEnterpriseCapabilityHandler(mockRepo)

	cmd := &commands.DeleteEnterpriseCapability{
		ID: existingCapability.ID(),
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedCapabilities, 1)
	deleted := mockRepo.savedCapabilities[0]
	assert.False(t, deleted.IsActive())
}

func TestDeleteEnterpriseCapabilityHandler_NonExistent_ReturnsError(t *testing.T) {
	mockRepo := &mockDeleteCapabilityRepository{getByIDErr: repositories.ErrEnterpriseCapabilityNotFound}

	handler := NewDeleteEnterpriseCapabilityHandler(mockRepo)

	cmd := &commands.DeleteEnterpriseCapability{
		ID: "non-existent-id",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, repositories.ErrEnterpriseCapabilityNotFound)
}

func TestDeleteEnterpriseCapabilityHandler_RepositoryError_ReturnsError(t *testing.T) {
	existingCapability := createDeleteTestCapability(t, "To Be Deleted")

	mockRepo := &mockDeleteCapabilityRepository{
		existingCapability: existingCapability,
		saveErr:            errors.New("save error"),
	}

	handler := NewDeleteEnterpriseCapabilityHandler(mockRepo)

	cmd := &commands.DeleteEnterpriseCapability{
		ID: existingCapability.ID(),
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}
