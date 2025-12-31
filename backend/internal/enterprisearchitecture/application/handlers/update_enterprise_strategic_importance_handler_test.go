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

type mockUpdateImportanceRepository struct {
	savedImportances   []*aggregates.EnterpriseStrategicImportance
	existingImportance *aggregates.EnterpriseStrategicImportance
	saveErr            error
	getByIDErr         error
}

func (m *mockUpdateImportanceRepository) Save(ctx context.Context, importance *aggregates.EnterpriseStrategicImportance) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedImportances = append(m.savedImportances, importance)
	return nil
}

func (m *mockUpdateImportanceRepository) GetByID(ctx context.Context, id string) (*aggregates.EnterpriseStrategicImportance, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.existingImportance != nil && m.existingImportance.ID() == id {
		return m.existingImportance, nil
	}
	return nil, repositories.ErrEnterpriseStrategicImportanceNotFound
}

func createTestEnterpriseStrategicImportance(t *testing.T) *aggregates.EnterpriseStrategicImportance {
	t.Helper()

	enterpriseCapabilityID := valueobjects.NewEnterpriseCapabilityID()
	pillarID := valueobjects.NewPillarID()
	importance, _ := valueobjects.NewImportance(3)
	rationale, _ := valueobjects.NewRationale("Initial rationale")

	si, err := aggregates.SetEnterpriseStrategicImportance(aggregates.NewEnterpriseImportanceParams{
		EnterpriseCapabilityID: enterpriseCapabilityID,
		PillarID:               pillarID,
		PillarName:             "Test Pillar",
		Importance:             importance,
		Rationale:              rationale,
	})
	require.NoError(t, err)
	si.MarkChangesAsCommitted()
	return si
}

func TestUpdateEnterpriseStrategicImportanceHandler_UpdatesImportance(t *testing.T) {
	existingImportance := createTestEnterpriseStrategicImportance(t)

	mockRepo := &mockUpdateImportanceRepository{existingImportance: existingImportance}

	handler := NewUpdateEnterpriseStrategicImportanceHandler(mockRepo)

	cmd := &commands.UpdateEnterpriseStrategicImportance{
		ID:         existingImportance.ID(),
		Importance: 5,
		Rationale:  "Updated rationale",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedImportances, 1)
	updated := mockRepo.savedImportances[0]
	assert.Equal(t, 5, updated.Importance().Value())
	assert.Equal(t, "Updated rationale", updated.Rationale().Value())
}

func TestUpdateEnterpriseStrategicImportanceHandler_RaisesUpdatedEvent(t *testing.T) {
	existingImportance := createTestEnterpriseStrategicImportance(t)

	mockRepo := &mockUpdateImportanceRepository{existingImportance: existingImportance}

	handler := NewUpdateEnterpriseStrategicImportanceHandler(mockRepo)

	cmd := &commands.UpdateEnterpriseStrategicImportance{
		ID:         existingImportance.ID(),
		Importance: 4,
		Rationale:  "Updated rationale",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	updated := mockRepo.savedImportances[0]
	uncommittedEvents := updated.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "EnterpriseStrategicImportanceUpdated", uncommittedEvents[0].EventType())
}

func TestUpdateEnterpriseStrategicImportanceHandler_NonExistent_ReturnsError(t *testing.T) {
	mockRepo := &mockUpdateImportanceRepository{getByIDErr: repositories.ErrEnterpriseStrategicImportanceNotFound}

	handler := NewUpdateEnterpriseStrategicImportanceHandler(mockRepo)

	cmd := &commands.UpdateEnterpriseStrategicImportance{
		ID:         "non-existent-id",
		Importance: 3,
		Rationale:  "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, repositories.ErrEnterpriseStrategicImportanceNotFound)
}

func TestUpdateEnterpriseStrategicImportanceHandler_InvalidImportance_ReturnsError(t *testing.T) {
	existingImportance := createTestEnterpriseStrategicImportance(t)

	mockRepo := &mockUpdateImportanceRepository{existingImportance: existingImportance}

	handler := NewUpdateEnterpriseStrategicImportanceHandler(mockRepo)

	cmd := &commands.UpdateEnterpriseStrategicImportance{
		ID:         existingImportance.ID(),
		Importance: 0,
		Rationale:  "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockRepo.savedImportances)
}

func TestUpdateEnterpriseStrategicImportanceHandler_RepositoryError_ReturnsError(t *testing.T) {
	existingImportance := createTestEnterpriseStrategicImportance(t)

	mockRepo := &mockUpdateImportanceRepository{
		existingImportance: existingImportance,
		saveErr:            errors.New("save error"),
	}

	handler := NewUpdateEnterpriseStrategicImportanceHandler(mockRepo)

	cmd := &commands.UpdateEnterpriseStrategicImportance{
		ID:         existingImportance.ID(),
		Importance: 4,
		Rationale:  "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}
