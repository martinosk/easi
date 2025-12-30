package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEnterpriseStrategicImportanceRepository struct {
	savedImportances []*aggregates.EnterpriseStrategicImportance
	saveErr          error
}

func (m *mockEnterpriseStrategicImportanceRepository) Save(ctx context.Context, importance *aggregates.EnterpriseStrategicImportance) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedImportances = append(m.savedImportances, importance)
	return nil
}

type mockEnterpriseCapabilityReadModelForImportance struct {
	existingCapability *readmodels.EnterpriseCapabilityDTO
	getByIDErr         error
}

func (m *mockEnterpriseCapabilityReadModelForImportance) GetByID(ctx context.Context, id string) (*readmodels.EnterpriseCapabilityDTO, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.existingCapability, nil
}

type mockEnterpriseStrategicImportanceReadModel struct {
	existingImportance *readmodels.EnterpriseStrategicImportanceDTO
	getErr             error
}

func (m *mockEnterpriseStrategicImportanceReadModel) GetByCapabilityAndPillar(ctx context.Context, capabilityID, pillarID string) (*readmodels.EnterpriseStrategicImportanceDTO, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.existingImportance, nil
}

type enterpriseStrategicImportanceRepository interface {
	Save(ctx context.Context, importance *aggregates.EnterpriseStrategicImportance) error
}

type enterpriseCapabilityReadModelForImportance interface {
	GetByID(ctx context.Context, id string) (*readmodels.EnterpriseCapabilityDTO, error)
}

type enterpriseStrategicImportanceReadModelForSet interface {
	GetByCapabilityAndPillar(ctx context.Context, capabilityID, pillarID string) (*readmodels.EnterpriseStrategicImportanceDTO, error)
}

type testableSetEnterpriseStrategicImportanceHandler struct {
	repository          enterpriseStrategicImportanceRepository
	capabilityReadModel enterpriseCapabilityReadModelForImportance
	importanceReadModel enterpriseStrategicImportanceReadModelForSet
}

func newTestableSetEnterpriseStrategicImportanceHandler(
	repository enterpriseStrategicImportanceRepository,
	capabilityReadModel enterpriseCapabilityReadModelForImportance,
	importanceReadModel enterpriseStrategicImportanceReadModelForSet,
) *testableSetEnterpriseStrategicImportanceHandler {
	return &testableSetEnterpriseStrategicImportanceHandler{
		repository:          repository,
		capabilityReadModel: capabilityReadModel,
		importanceReadModel: importanceReadModel,
	}
}

func (h *testableSetEnterpriseStrategicImportanceHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.SetEnterpriseStrategicImportance)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	capability, err := h.capabilityReadModel.GetByID(ctx, command.EnterpriseCapabilityID)
	if err != nil {
		return err
	}
	if capability == nil {
		return repositories.ErrEnterpriseCapabilityNotFound
	}

	existing, err := h.importanceReadModel.GetByCapabilityAndPillar(ctx, command.EnterpriseCapabilityID, command.PillarID)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrImportanceAlreadySet
	}

	enterpriseCapabilityID, err := valueobjects.NewEnterpriseCapabilityIDFromString(command.EnterpriseCapabilityID)
	if err != nil {
		return err
	}

	pillarID, err := valueobjects.NewPillarIDFromString(command.PillarID)
	if err != nil {
		return err
	}

	importance, err := valueobjects.NewImportance(command.Importance)
	if err != nil {
		return err
	}

	rationale, err := valueobjects.NewRationale(command.Rationale)
	if err != nil {
		return err
	}

	si, err := aggregates.SetEnterpriseStrategicImportance(aggregates.NewEnterpriseImportanceParams{
		EnterpriseCapabilityID: enterpriseCapabilityID,
		PillarID:               pillarID,
		PillarName:             command.PillarName,
		Importance:             importance,
		Rationale:              rationale,
	})
	if err != nil {
		return err
	}

	command.ID = si.ID()

	return h.repository.Save(ctx, si)
}

func TestSetEnterpriseStrategicImportanceHandler_SetsImportance(t *testing.T) {
	mockRepo := &mockEnterpriseStrategicImportanceRepository{}
	mockCapabilityReadModel := &mockEnterpriseCapabilityReadModelForImportance{
		existingCapability: &readmodels.EnterpriseCapabilityDTO{ID: uuid.New().String()},
	}
	mockImportanceReadModel := &mockEnterpriseStrategicImportanceReadModel{existingImportance: nil}

	handler := newTestableSetEnterpriseStrategicImportanceHandler(mockRepo, mockCapabilityReadModel, mockImportanceReadModel)

	capabilityID := uuid.New().String()
	pillarID := uuid.New().String()
	cmd := &commands.SetEnterpriseStrategicImportance{
		EnterpriseCapabilityID: capabilityID,
		PillarID:               pillarID,
		PillarName:             "Strategic Pillar 1",
		Importance:             4,
		Rationale:              "Critical for business operations",
	}

	mockCapabilityReadModel.existingCapability.ID = capabilityID

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedImportances, 1)
	importance := mockRepo.savedImportances[0]
	assert.Equal(t, capabilityID, importance.EnterpriseCapabilityID().Value())
	assert.Equal(t, pillarID, importance.PillarID().Value())
	assert.Equal(t, 4, importance.Importance().Value())
	assert.Equal(t, "Critical for business operations", importance.Rationale().Value())
}

func TestSetEnterpriseStrategicImportanceHandler_SetsCommandID(t *testing.T) {
	mockRepo := &mockEnterpriseStrategicImportanceRepository{}
	capabilityID := uuid.New().String()
	mockCapabilityReadModel := &mockEnterpriseCapabilityReadModelForImportance{
		existingCapability: &readmodels.EnterpriseCapabilityDTO{ID: capabilityID},
	}
	mockImportanceReadModel := &mockEnterpriseStrategicImportanceReadModel{existingImportance: nil}

	handler := newTestableSetEnterpriseStrategicImportanceHandler(mockRepo, mockCapabilityReadModel, mockImportanceReadModel)

	pillarID := uuid.New().String()
	cmd := &commands.SetEnterpriseStrategicImportance{
		EnterpriseCapabilityID: capabilityID,
		PillarID:               pillarID,
		PillarName:             "Strategic Pillar 1",
		Importance:             3,
		Rationale:              "",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.NotEmpty(t, cmd.ID)
	assert.Equal(t, mockRepo.savedImportances[0].ID(), cmd.ID)
}

func TestSetEnterpriseStrategicImportanceHandler_RaisesSetEvent(t *testing.T) {
	mockRepo := &mockEnterpriseStrategicImportanceRepository{}
	capabilityID := uuid.New().String()
	mockCapabilityReadModel := &mockEnterpriseCapabilityReadModelForImportance{
		existingCapability: &readmodels.EnterpriseCapabilityDTO{ID: capabilityID},
	}
	mockImportanceReadModel := &mockEnterpriseStrategicImportanceReadModel{existingImportance: nil}

	handler := newTestableSetEnterpriseStrategicImportanceHandler(mockRepo, mockCapabilityReadModel, mockImportanceReadModel)

	pillarID := uuid.New().String()
	cmd := &commands.SetEnterpriseStrategicImportance{
		EnterpriseCapabilityID: capabilityID,
		PillarID:               pillarID,
		PillarName:             "Strategic Pillar 1",
		Importance:             5,
		Rationale:              "",
	}

	err := handler.Handle(context.Background(), cmd)
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
			mockRepo := &mockEnterpriseStrategicImportanceRepository{}
			capabilityID := uuid.New().String()
			mockCapabilityReadModel := &mockEnterpriseCapabilityReadModelForImportance{
				existingCapability: &readmodels.EnterpriseCapabilityDTO{ID: capabilityID},
			}
			mockImportanceReadModel := &mockEnterpriseStrategicImportanceReadModel{existingImportance: nil}

			handler := newTestableSetEnterpriseStrategicImportanceHandler(mockRepo, mockCapabilityReadModel, mockImportanceReadModel)

			pillarID := uuid.New().String()
			cmd := &commands.SetEnterpriseStrategicImportance{
				EnterpriseCapabilityID: capabilityID,
				PillarID:               pillarID,
				PillarName:             "Test Pillar",
				Importance:             tc.importance,
				Rationale:              "",
			}

			err := handler.Handle(context.Background(), cmd)
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
	mockRepo := &mockEnterpriseStrategicImportanceRepository{}
	mockCapabilityReadModel := &mockEnterpriseCapabilityReadModelForImportance{
		existingCapability: nil,
	}
	mockImportanceReadModel := &mockEnterpriseStrategicImportanceReadModel{existingImportance: nil}

	handler := newTestableSetEnterpriseStrategicImportanceHandler(mockRepo, mockCapabilityReadModel, mockImportanceReadModel)

	cmd := &commands.SetEnterpriseStrategicImportance{
		EnterpriseCapabilityID: uuid.New().String(),
		PillarID:               uuid.New().String(),
		PillarName:             "Test Pillar",
		Importance:             3,
		Rationale:              "",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, repositories.ErrEnterpriseCapabilityNotFound)
}

func TestSetEnterpriseStrategicImportanceHandler_AlreadySet_ReturnsError(t *testing.T) {
	mockRepo := &mockEnterpriseStrategicImportanceRepository{}
	capabilityID := uuid.New().String()
	mockCapabilityReadModel := &mockEnterpriseCapabilityReadModelForImportance{
		existingCapability: &readmodels.EnterpriseCapabilityDTO{ID: capabilityID},
	}
	mockImportanceReadModel := &mockEnterpriseStrategicImportanceReadModel{
		existingImportance: &readmodels.EnterpriseStrategicImportanceDTO{ID: "existing-id"},
	}

	handler := newTestableSetEnterpriseStrategicImportanceHandler(mockRepo, mockCapabilityReadModel, mockImportanceReadModel)

	cmd := &commands.SetEnterpriseStrategicImportance{
		EnterpriseCapabilityID: capabilityID,
		PillarID:               uuid.New().String(),
		PillarName:             "Test Pillar",
		Importance:             3,
		Rationale:              "",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrImportanceAlreadySet)
	assert.Empty(t, mockRepo.savedImportances)
}

func TestSetEnterpriseStrategicImportanceHandler_ReadModelError_ReturnsError(t *testing.T) {
	mockRepo := &mockEnterpriseStrategicImportanceRepository{}
	mockCapabilityReadModel := &mockEnterpriseCapabilityReadModelForImportance{
		getByIDErr: errors.New("database error"),
	}
	mockImportanceReadModel := &mockEnterpriseStrategicImportanceReadModel{}

	handler := newTestableSetEnterpriseStrategicImportanceHandler(mockRepo, mockCapabilityReadModel, mockImportanceReadModel)

	cmd := &commands.SetEnterpriseStrategicImportance{
		EnterpriseCapabilityID: uuid.New().String(),
		PillarID:               uuid.New().String(),
		PillarName:             "Test Pillar",
		Importance:             3,
		Rationale:              "",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}

func TestSetEnterpriseStrategicImportanceHandler_RepositoryError_ReturnsError(t *testing.T) {
	mockRepo := &mockEnterpriseStrategicImportanceRepository{saveErr: errors.New("save error")}
	capabilityID := uuid.New().String()
	mockCapabilityReadModel := &mockEnterpriseCapabilityReadModelForImportance{
		existingCapability: &readmodels.EnterpriseCapabilityDTO{ID: capabilityID},
	}
	mockImportanceReadModel := &mockEnterpriseStrategicImportanceReadModel{existingImportance: nil}

	handler := newTestableSetEnterpriseStrategicImportanceHandler(mockRepo, mockCapabilityReadModel, mockImportanceReadModel)

	cmd := &commands.SetEnterpriseStrategicImportance{
		EnterpriseCapabilityID: capabilityID,
		PillarID:               uuid.New().String(),
		PillarName:             "Test Pillar",
		Importance:             3,
		Rationale:              "",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}
