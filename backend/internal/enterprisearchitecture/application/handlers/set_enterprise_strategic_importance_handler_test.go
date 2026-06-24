package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

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

func existingCapabilityReadModel(capabilityID string) *mockSetImportanceCapabilityReadModel {
	return &mockSetImportanceCapabilityReadModel{
		existingCapability: &readmodels.EnterpriseCapabilityDTO{ID: capabilityID},
	}
}

func setImportanceCommand(capabilityID string, importance int) *commands.SetEnterpriseStrategicImportance {
	return &commands.SetEnterpriseStrategicImportance{
		EnterpriseCapabilityID: capabilityID,
		PillarID:               uuid.New().String(),
		PillarName:             "Test Pillar",
		Importance:             importance,
	}
}

func runSetImportance(
	repo *mockSetImportanceRepository,
	capabilityReadModel *mockSetImportanceCapabilityReadModel,
	importanceReadModel *mockSetImportanceReadModel,
	cmd *commands.SetEnterpriseStrategicImportance,
) (cqrs.CommandResult, error) {
	handler := NewSetEnterpriseStrategicImportanceHandler(repo, capabilityReadModel, importanceReadModel)
	return handler.Handle(context.Background(), cmd)
}

func TestSetEnterpriseStrategicImportanceHandler_SetsImportance(t *testing.T) {
	repo := &mockSetImportanceRepository{}
	capabilityID := uuid.New().String()
	pillarID := uuid.New().String()
	cmd := &commands.SetEnterpriseStrategicImportance{
		EnterpriseCapabilityID: capabilityID,
		PillarID:               pillarID,
		PillarName:             "Strategic Pillar 1",
		Importance:             4,
		Rationale:              "Critical for business operations",
	}

	_, err := runSetImportance(repo, existingCapabilityReadModel(capabilityID), &mockSetImportanceReadModel{}, cmd)
	require.NoError(t, err)

	require.Len(t, repo.savedImportances, 1)
	importance := repo.savedImportances[0]
	assert.Equal(t, capabilityID, importance.EnterpriseCapabilityID().Value())
	assert.Equal(t, pillarID, importance.PillarID().Value())
	assert.Equal(t, 4, importance.Importance().Value())
	assert.Equal(t, "Critical for business operations", importance.Rationale().Value())
}

func TestSetEnterpriseStrategicImportanceHandler_ReturnsCreatedID(t *testing.T) {
	repo := &mockSetImportanceRepository{}
	capabilityID := uuid.New().String()

	result, err := runSetImportance(repo, existingCapabilityReadModel(capabilityID), &mockSetImportanceReadModel{}, setImportanceCommand(capabilityID, 3))
	require.NoError(t, err)

	assert.NotEmpty(t, result.CreatedID)
	assert.Equal(t, repo.savedImportances[0].ID(), result.CreatedID)
}

func TestSetEnterpriseStrategicImportanceHandler_RaisesSetEvent(t *testing.T) {
	repo := &mockSetImportanceRepository{}
	capabilityID := uuid.New().String()

	_, err := runSetImportance(repo, existingCapabilityReadModel(capabilityID), &mockSetImportanceReadModel{}, setImportanceCommand(capabilityID, 5))
	require.NoError(t, err)

	uncommittedEvents := repo.savedImportances[0].GetUncommittedChanges()
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
			repo := &mockSetImportanceRepository{}
			capabilityID := uuid.New().String()

			_, err := runSetImportance(repo, existingCapabilityReadModel(capabilityID), &mockSetImportanceReadModel{}, setImportanceCommand(capabilityID, tc.importance))
			if tc.shouldFail {
				assert.Error(t, err)
				assert.Empty(t, repo.savedImportances)
			} else {
				assert.NoError(t, err)
				require.Len(t, repo.savedImportances, 1)
			}
		})
	}
}

func TestSetEnterpriseStrategicImportanceHandler_ErrorCases(t *testing.T) {
	testCases := []struct {
		name                string
		repo                *mockSetImportanceRepository
		capabilityReadModel *mockSetImportanceCapabilityReadModel
		importanceReadModel *mockSetImportanceReadModel
		wantErrIs           error
		wantNoSaves         bool
	}{
		{
			name:                "non-existent capability",
			repo:                &mockSetImportanceRepository{},
			capabilityReadModel: &mockSetImportanceCapabilityReadModel{existingCapability: nil},
			importanceReadModel: &mockSetImportanceReadModel{},
			wantErrIs:           repositories.ErrEnterpriseCapabilityNotFound,
		},
		{
			name:                "importance already set",
			repo:                &mockSetImportanceRepository{},
			capabilityReadModel: existingCapabilityReadModel("set-cap-id"),
			importanceReadModel: &mockSetImportanceReadModel{
				existingImportance: &readmodels.EnterpriseStrategicImportanceDTO{ID: "existing-id"},
			},
			wantErrIs:   ErrImportanceAlreadySet,
			wantNoSaves: true,
		},
		{
			name:                "read model error",
			repo:                &mockSetImportanceRepository{},
			capabilityReadModel: &mockSetImportanceCapabilityReadModel{getByIDErr: errors.New("database error")},
			importanceReadModel: &mockSetImportanceReadModel{},
		},
		{
			name:                "repository error",
			repo:                &mockSetImportanceRepository{saveErr: errors.New("save error")},
			capabilityReadModel: existingCapabilityReadModel("save-cap-id"),
			importanceReadModel: &mockSetImportanceReadModel{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := setImportanceCommand(uuid.New().String(), 3)

			_, err := runSetImportance(tc.repo, tc.capabilityReadModel, tc.importanceReadModel, cmd)

			if tc.wantErrIs != nil {
				assert.ErrorIs(t, err, tc.wantErrIs)
			} else {
				assert.Error(t, err)
			}
			if tc.wantNoSaves {
				assert.Empty(t, tc.repo.savedImportances)
			}
		})
	}
}
