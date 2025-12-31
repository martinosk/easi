package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRemoveImportanceRepository struct {
	savedImportances   []*aggregates.EnterpriseStrategicImportance
	existingImportance *aggregates.EnterpriseStrategicImportance
	saveErr            error
	getByIDErr         error
}

func (m *mockRemoveImportanceRepository) Save(ctx context.Context, importance *aggregates.EnterpriseStrategicImportance) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedImportances = append(m.savedImportances, importance)
	return nil
}

func (m *mockRemoveImportanceRepository) GetByID(ctx context.Context, id string) (*aggregates.EnterpriseStrategicImportance, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.existingImportance != nil && m.existingImportance.ID() == id {
		return m.existingImportance, nil
	}
	return nil, repositories.ErrEnterpriseStrategicImportanceNotFound
}

func TestRemoveEnterpriseStrategicImportanceHandler_RemovesRating(t *testing.T) {
	existingImportance := createTestEnterpriseStrategicImportance(t)

	mockRepo := &mockRemoveImportanceRepository{existingImportance: existingImportance}

	handler := NewRemoveEnterpriseStrategicImportanceHandler(mockRepo)

	cmd := &commands.RemoveEnterpriseStrategicImportance{
		ID: existingImportance.ID(),
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedImportances, 1)
}

func TestRemoveEnterpriseStrategicImportanceHandler_RaisesRemovedEvent(t *testing.T) {
	existingImportance := createTestEnterpriseStrategicImportance(t)

	mockRepo := &mockRemoveImportanceRepository{existingImportance: existingImportance}

	handler := NewRemoveEnterpriseStrategicImportanceHandler(mockRepo)

	cmd := &commands.RemoveEnterpriseStrategicImportance{
		ID: existingImportance.ID(),
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	removed := mockRepo.savedImportances[0]
	uncommittedEvents := removed.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "EnterpriseStrategicImportanceRemoved", uncommittedEvents[0].EventType())
}

func TestRemoveEnterpriseStrategicImportanceHandler_NonExistent_ReturnsError(t *testing.T) {
	mockRepo := &mockRemoveImportanceRepository{getByIDErr: repositories.ErrEnterpriseStrategicImportanceNotFound}

	handler := NewRemoveEnterpriseStrategicImportanceHandler(mockRepo)

	cmd := &commands.RemoveEnterpriseStrategicImportance{
		ID: "non-existent-id",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, repositories.ErrEnterpriseStrategicImportanceNotFound)
}

func TestRemoveEnterpriseStrategicImportanceHandler_RepositoryError_ReturnsError(t *testing.T) {
	existingImportance := createTestEnterpriseStrategicImportance(t)

	mockRepo := &mockRemoveImportanceRepository{
		existingImportance: existingImportance,
		saveErr:            errors.New("save error"),
	}

	handler := NewRemoveEnterpriseStrategicImportanceHandler(mockRepo)

	cmd := &commands.RemoveEnterpriseStrategicImportance{
		ID: existingImportance.ID(),
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}
