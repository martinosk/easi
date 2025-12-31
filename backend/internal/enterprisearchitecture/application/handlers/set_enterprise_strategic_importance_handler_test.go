package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSetImportanceRepository struct {
	savedImportances []*aggregates.EnterpriseStrategicImportance
	saveErr          error
}

func (m *mockSetImportanceRepository) Save(ctx context.Context, importance *aggregates.EnterpriseStrategicImportance) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedImportances = append(m.savedImportances, importance)
	return nil
}

type mockSetImportanceCapabilityReadModel struct {
	existingCapability *readmodels.EnterpriseCapabilityDTO
	getByIDErr         error
}

func (m *mockSetImportanceCapabilityReadModel) GetByID(ctx context.Context, id string) (*readmodels.EnterpriseCapabilityDTO, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.existingCapability, nil
}

type mockSetImportanceReadModel struct {
	existingImportance *readmodels.EnterpriseStrategicImportanceDTO
	getErr             error
}

func (m *mockSetImportanceReadModel) GetByCapabilityAndPillar(ctx context.Context, capabilityID, pillarID string) (*readmodels.EnterpriseStrategicImportanceDTO, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.existingImportance, nil
}

func TestSetEnterpriseStrategicImportanceHandler_SetsImportance(t *testing.T) {
	mockRepo := &mockSetImportanceRepository{}
	capabilityID := uuid.New().String()
	mockCapabilityReadModel := &mockSetImportanceCapabilityReadModel{
		existingCapability: &readmodels.EnterpriseCapabilityDTO{ID: capabilityID},
	}
	mockImportanceReadModel := &mockSetImportanceReadModel{existingImportance: nil}

	handler := NewSetEnterpriseStrategicImportanceHandler(mockRepo, mockCapabilityReadModel, mockImportanceReadModel)

	pillarID := uuid.New().String()
	cmd := &commands.SetEnterpriseStrategicImportance{
		EnterpriseCapabilityID: capabilityID,
		PillarID:               pillarID,
		PillarName:             "Strategic Pillar 1",
		Importance:             4,
		Rationale:              "Critical for business operations",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedImportances, 1)
	importance := mockRepo.savedImportances[0]
	assert.Equal(t, capabilityID, importance.EnterpriseCapabilityID().Value())
	assert.Equal(t, pillarID, importance.PillarID().Value())
	assert.Equal(t, 4, importance.Importance().Value())
	assert.Equal(t, "Critical for business operations", importance.Rationale().Value())
}

func TestSetEnterpriseStrategicImportanceHandler_ReturnsCreatedID(t *testing.T) {
	mockRepo := &mockSetImportanceRepository{}
	capabilityID := uuid.New().String()
	mockCapabilityReadModel := &mockSetImportanceCapabilityReadModel{
		existingCapability: &readmodels.EnterpriseCapabilityDTO{ID: capabilityID},
	}
	mockImportanceReadModel := &mockSetImportanceReadModel{existingImportance: nil}

	handler := NewSetEnterpriseStrategicImportanceHandler(mockRepo, mockCapabilityReadModel, mockImportanceReadModel)

	pillarID := uuid.New().String()
	cmd := &commands.SetEnterpriseStrategicImportance{
		EnterpriseCapabilityID: capabilityID,
		PillarID:               pillarID,
		PillarName:             "Strategic Pillar 1",
		Importance:             3,
		Rationale:              "",
	}

	result, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.NotEmpty(t, result.CreatedID)
	assert.Equal(t, mockRepo.savedImportances[0].ID(), result.CreatedID)
}

func TestSetEnterpriseStrategicImportanceHandler_RaisesSetEvent(t *testing.T) {
	mockRepo := &mockSetImportanceRepository{}
	capabilityID := uuid.New().String()
	mockCapabilityReadModel := &mockSetImportanceCapabilityReadModel{
		existingCapability: &readmodels.EnterpriseCapabilityDTO{ID: capabilityID},
	}
	mockImportanceReadModel := &mockSetImportanceReadModel{existingImportance: nil}

	handler := NewSetEnterpriseStrategicImportanceHandler(mockRepo, mockCapabilityReadModel, mockImportanceReadModel)

	pillarID := uuid.New().String()
	cmd := &commands.SetEnterpriseStrategicImportance{
		EnterpriseCapabilityID: capabilityID,
		PillarID:               pillarID,
		PillarName:             "Strategic Pillar 1",
		Importance:             5,
		Rationale:              "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	si := mockRepo.savedImportances[0]
	uncommittedEvents := si.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "EnterpriseStrategicImportanceSet", uncommittedEvents[0].EventType())
}

func TestSetEnterpriseStrategicImportanceHandler_ValidatesImportanceRange(t *testing.T) {
	testCases := []struct {
		name       string
		importance int
		shouldFail bool
	}{
		{"value 0 fails", 0, true},
		{"value 1 succeeds", 1, false},
		{"value 5 succeeds", 5, false},
		{"value 6 fails", 6, true},
		{"negative fails", -1, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mockSetImportanceRepository{}
			capabilityID := uuid.New().String()
			mockCapabilityReadModel := &mockSetImportanceCapabilityReadModel{
				existingCapability: &readmodels.EnterpriseCapabilityDTO{ID: capabilityID},
			}
			mockImportanceReadModel := &mockSetImportanceReadModel{existingImportance: nil}

			handler := NewSetEnterpriseStrategicImportanceHandler(mockRepo, mockCapabilityReadModel, mockImportanceReadModel)

			pillarID := uuid.New().String()
			cmd := &commands.SetEnterpriseStrategicImportance{
				EnterpriseCapabilityID: capabilityID,
				PillarID:               pillarID,
				PillarName:             "Test Pillar",
				Importance:             tc.importance,
				Rationale:              "",
			}

			_, err := handler.Handle(context.Background(), cmd)
			if tc.shouldFail {
				assert.Error(t, err)
				assert.Empty(t, mockRepo.savedImportances)
			} else {
				assert.NoError(t, err)
				require.Len(t, mockRepo.savedImportances, 1)
			}
		})
	}
}

func TestSetEnterpriseStrategicImportanceHandler_NonExistentCapability_ReturnsError(t *testing.T) {
	mockRepo := &mockSetImportanceRepository{}
	mockCapabilityReadModel := &mockSetImportanceCapabilityReadModel{existingCapability: nil}
	mockImportanceReadModel := &mockSetImportanceReadModel{existingImportance: nil}

	handler := NewSetEnterpriseStrategicImportanceHandler(mockRepo, mockCapabilityReadModel, mockImportanceReadModel)

	cmd := &commands.SetEnterpriseStrategicImportance{
		EnterpriseCapabilityID: uuid.New().String(),
		PillarID:               uuid.New().String(),
		PillarName:             "Test Pillar",
		Importance:             3,
		Rationale:              "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, repositories.ErrEnterpriseCapabilityNotFound)
}

func TestSetEnterpriseStrategicImportanceHandler_AlreadySet_ReturnsError(t *testing.T) {
	mockRepo := &mockSetImportanceRepository{}
	capabilityID := uuid.New().String()
	mockCapabilityReadModel := &mockSetImportanceCapabilityReadModel{
		existingCapability: &readmodels.EnterpriseCapabilityDTO{ID: capabilityID},
	}
	mockImportanceReadModel := &mockSetImportanceReadModel{
		existingImportance: &readmodels.EnterpriseStrategicImportanceDTO{ID: "existing-id"},
	}

	handler := NewSetEnterpriseStrategicImportanceHandler(mockRepo, mockCapabilityReadModel, mockImportanceReadModel)

	cmd := &commands.SetEnterpriseStrategicImportance{
		EnterpriseCapabilityID: capabilityID,
		PillarID:               uuid.New().String(),
		PillarName:             "Test Pillar",
		Importance:             3,
		Rationale:              "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrImportanceAlreadySet)
	assert.Empty(t, mockRepo.savedImportances)
}

func TestSetEnterpriseStrategicImportanceHandler_ReadModelError_ReturnsError(t *testing.T) {
	mockRepo := &mockSetImportanceRepository{}
	mockCapabilityReadModel := &mockSetImportanceCapabilityReadModel{getByIDErr: errors.New("database error")}
	mockImportanceReadModel := &mockSetImportanceReadModel{}

	handler := NewSetEnterpriseStrategicImportanceHandler(mockRepo, mockCapabilityReadModel, mockImportanceReadModel)

	cmd := &commands.SetEnterpriseStrategicImportance{
		EnterpriseCapabilityID: uuid.New().String(),
		PillarID:               uuid.New().String(),
		PillarName:             "Test Pillar",
		Importance:             3,
		Rationale:              "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}

func TestSetEnterpriseStrategicImportanceHandler_RepositoryError_ReturnsError(t *testing.T) {
	mockRepo := &mockSetImportanceRepository{saveErr: errors.New("save error")}
	capabilityID := uuid.New().String()
	mockCapabilityReadModel := &mockSetImportanceCapabilityReadModel{
		existingCapability: &readmodels.EnterpriseCapabilityDTO{ID: capabilityID},
	}
	mockImportanceReadModel := &mockSetImportanceReadModel{existingImportance: nil}

	handler := NewSetEnterpriseStrategicImportanceHandler(mockRepo, mockCapabilityReadModel, mockImportanceReadModel)

	cmd := &commands.SetEnterpriseStrategicImportance{
		EnterpriseCapabilityID: capabilityID,
		PillarID:               uuid.New().String(),
		PillarName:             "Test Pillar",
		Importance:             3,
		Rationale:              "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}
