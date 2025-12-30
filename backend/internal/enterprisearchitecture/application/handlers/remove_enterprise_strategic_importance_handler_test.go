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

type mockEnterpriseStrategicImportanceRepositoryForRemove struct {
	savedImportances   []*aggregates.EnterpriseStrategicImportance
	existingImportance *aggregates.EnterpriseStrategicImportance
	saveErr            error
	getByIDErr         error
}

func (m *mockEnterpriseStrategicImportanceRepositoryForRemove) Save(ctx context.Context, importance *aggregates.EnterpriseStrategicImportance) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedImportances = append(m.savedImportances, importance)
	return nil
}

func (m *mockEnterpriseStrategicImportanceRepositoryForRemove) GetByID(ctx context.Context, id string) (*aggregates.EnterpriseStrategicImportance, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.existingImportance != nil && m.existingImportance.ID() == id {
		return m.existingImportance, nil
	}
	return nil, repositories.ErrEnterpriseStrategicImportanceNotFound
}

type enterpriseStrategicImportanceRepositoryForRemove interface {
	Save(ctx context.Context, importance *aggregates.EnterpriseStrategicImportance) error
	GetByID(ctx context.Context, id string) (*aggregates.EnterpriseStrategicImportance, error)
}

type testableRemoveEnterpriseStrategicImportanceHandler struct {
	repository enterpriseStrategicImportanceRepositoryForRemove
}

func newTestableRemoveEnterpriseStrategicImportanceHandler(
	repository enterpriseStrategicImportanceRepositoryForRemove,
) *testableRemoveEnterpriseStrategicImportanceHandler {
	return &testableRemoveEnterpriseStrategicImportanceHandler{
		repository: repository,
	}
}

func (h *testableRemoveEnterpriseStrategicImportanceHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.RemoveEnterpriseStrategicImportance)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	si, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	if err := si.Remove(); err != nil {
		return err
	}

	return h.repository.Save(ctx, si)
}

func TestRemoveEnterpriseStrategicImportanceHandler_RemovesRating(t *testing.T) {
	existingImportance := createTestEnterpriseStrategicImportance(t)

	mockRepo := &mockEnterpriseStrategicImportanceRepositoryForRemove{
		existingImportance: existingImportance,
	}

	handler := newTestableRemoveEnterpriseStrategicImportanceHandler(mockRepo)

	cmd := &commands.RemoveEnterpriseStrategicImportance{
		ID: existingImportance.ID(),
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedImportances, 1)
}

func TestRemoveEnterpriseStrategicImportanceHandler_RaisesRemovedEvent(t *testing.T) {
	existingImportance := createTestEnterpriseStrategicImportance(t)

	mockRepo := &mockEnterpriseStrategicImportanceRepositoryForRemove{
		existingImportance: existingImportance,
	}

	handler := newTestableRemoveEnterpriseStrategicImportanceHandler(mockRepo)

	cmd := &commands.RemoveEnterpriseStrategicImportance{
		ID: existingImportance.ID(),
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	removed := mockRepo.savedImportances[0]
	uncommittedEvents := removed.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "EnterpriseStrategicImportanceRemoved", uncommittedEvents[0].EventType())
}

func TestRemoveEnterpriseStrategicImportanceHandler_NonExistent_ReturnsError(t *testing.T) {
	mockRepo := &mockEnterpriseStrategicImportanceRepositoryForRemove{
		getByIDErr: repositories.ErrEnterpriseStrategicImportanceNotFound,
	}

	handler := newTestableRemoveEnterpriseStrategicImportanceHandler(mockRepo)

	cmd := &commands.RemoveEnterpriseStrategicImportance{
		ID: "non-existent-id",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, repositories.ErrEnterpriseStrategicImportanceNotFound)
}

func TestRemoveEnterpriseStrategicImportanceHandler_RepositoryError_ReturnsError(t *testing.T) {
	existingImportance := createTestEnterpriseStrategicImportance(t)

	mockRepo := &mockEnterpriseStrategicImportanceRepositoryForRemove{
		existingImportance: existingImportance,
		saveErr:            errors.New("save error"),
	}

	handler := newTestableRemoveEnterpriseStrategicImportanceHandler(mockRepo)

	cmd := &commands.RemoveEnterpriseStrategicImportance{
		ID: existingImportance.ID(),
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}
