package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEnterpriseCapabilityRepositoryForDelete struct {
	savedCapabilities  []*aggregates.EnterpriseCapability
	existingCapability *aggregates.EnterpriseCapability
	saveErr            error
	getByIDErr         error
}

func (m *mockEnterpriseCapabilityRepositoryForDelete) Save(ctx context.Context, capability *aggregates.EnterpriseCapability) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedCapabilities = append(m.savedCapabilities, capability)
	return nil
}

func (m *mockEnterpriseCapabilityRepositoryForDelete) GetByID(ctx context.Context, id string) (*aggregates.EnterpriseCapability, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.existingCapability != nil && m.existingCapability.ID() == id {
		return m.existingCapability, nil
	}
	return nil, repositories.ErrEnterpriseCapabilityNotFound
}

type enterpriseCapabilityRepositoryForDelete interface {
	Save(ctx context.Context, capability *aggregates.EnterpriseCapability) error
	GetByID(ctx context.Context, id string) (*aggregates.EnterpriseCapability, error)
}

type testableDeleteEnterpriseCapabilityHandler struct {
	repository enterpriseCapabilityRepositoryForDelete
}

func newTestableDeleteEnterpriseCapabilityHandler(
	repository enterpriseCapabilityRepositoryForDelete,
) *testableDeleteEnterpriseCapabilityHandler {
	return &testableDeleteEnterpriseCapabilityHandler{
		repository: repository,
	}
}

func (h *testableDeleteEnterpriseCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.DeleteEnterpriseCapability)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	capability, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	if err := capability.Delete(); err != nil {
		return err
	}

	return h.repository.Save(ctx, capability)
}

func TestDeleteEnterpriseCapabilityHandler_DeletesCapability(t *testing.T) {
	existingCapability := createTestEnterpriseCapability(t, "To Be Deleted")

	mockRepo := &mockEnterpriseCapabilityRepositoryForDelete{
		existingCapability: existingCapability,
	}

	handler := newTestableDeleteEnterpriseCapabilityHandler(mockRepo)

	cmd := &commands.DeleteEnterpriseCapability{
		ID: existingCapability.ID(),
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedCapabilities, 1)
	deleted := mockRepo.savedCapabilities[0]
	assert.False(t, deleted.IsActive())
}

func TestDeleteEnterpriseCapabilityHandler_NonExistent_ReturnsError(t *testing.T) {
	mockRepo := &mockEnterpriseCapabilityRepositoryForDelete{
		getByIDErr: repositories.ErrEnterpriseCapabilityNotFound,
	}

	handler := newTestableDeleteEnterpriseCapabilityHandler(mockRepo)

	cmd := &commands.DeleteEnterpriseCapability{
		ID: "non-existent-id",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, repositories.ErrEnterpriseCapabilityNotFound)
}

func TestDeleteEnterpriseCapabilityHandler_RepositoryError_ReturnsError(t *testing.T) {
	existingCapability := createTestEnterpriseCapability(t, "To Be Deleted")

	mockRepo := &mockEnterpriseCapabilityRepositoryForDelete{
		existingCapability: existingCapability,
		saveErr:            errors.New("save error"),
	}

	handler := newTestableDeleteEnterpriseCapabilityHandler(mockRepo)

	cmd := &commands.DeleteEnterpriseCapability{
		ID: existingCapability.ID(),
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}
