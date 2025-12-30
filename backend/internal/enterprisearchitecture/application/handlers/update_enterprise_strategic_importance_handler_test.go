package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEnterpriseStrategicImportanceRepositoryForUpdate struct {
	savedImportances   []*aggregates.EnterpriseStrategicImportance
	existingImportance *aggregates.EnterpriseStrategicImportance
	saveErr            error
	getByIDErr         error
}

func (m *mockEnterpriseStrategicImportanceRepositoryForUpdate) Save(ctx context.Context, importance *aggregates.EnterpriseStrategicImportance) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedImportances = append(m.savedImportances, importance)
	return nil
}

func (m *mockEnterpriseStrategicImportanceRepositoryForUpdate) GetByID(ctx context.Context, id string) (*aggregates.EnterpriseStrategicImportance, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.existingImportance != nil && m.existingImportance.ID() == id {
		return m.existingImportance, nil
	}
	return nil, repositories.ErrEnterpriseStrategicImportanceNotFound
}

type enterpriseStrategicImportanceRepositoryForUpdate interface {
	Save(ctx context.Context, importance *aggregates.EnterpriseStrategicImportance) error
	GetByID(ctx context.Context, id string) (*aggregates.EnterpriseStrategicImportance, error)
}

type testableUpdateEnterpriseStrategicImportanceHandler struct {
	repository enterpriseStrategicImportanceRepositoryForUpdate
}

func newTestableUpdateEnterpriseStrategicImportanceHandler(
	repository enterpriseStrategicImportanceRepositoryForUpdate,
) *testableUpdateEnterpriseStrategicImportanceHandler {
	return &testableUpdateEnterpriseStrategicImportanceHandler{
		repository: repository,
	}
}

func (h *testableUpdateEnterpriseStrategicImportanceHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.UpdateEnterpriseStrategicImportance)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	si, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	importance, err := valueobjects.NewImportance(command.Importance)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	rationale, err := valueobjects.NewRationale(command.Rationale)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := si.Update(importance, rationale); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, si); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.EmptyResult(), nil
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

	mockRepo := &mockEnterpriseStrategicImportanceRepositoryForUpdate{
		existingImportance: existingImportance,
	}

	handler := newTestableUpdateEnterpriseStrategicImportanceHandler(mockRepo)

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

	mockRepo := &mockEnterpriseStrategicImportanceRepositoryForUpdate{
		existingImportance: existingImportance,
	}

	handler := newTestableUpdateEnterpriseStrategicImportanceHandler(mockRepo)

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
	mockRepo := &mockEnterpriseStrategicImportanceRepositoryForUpdate{
		getByIDErr: repositories.ErrEnterpriseStrategicImportanceNotFound,
	}

	handler := newTestableUpdateEnterpriseStrategicImportanceHandler(mockRepo)

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

	mockRepo := &mockEnterpriseStrategicImportanceRepositoryForUpdate{
		existingImportance: existingImportance,
	}

	handler := newTestableUpdateEnterpriseStrategicImportanceHandler(mockRepo)

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

	mockRepo := &mockEnterpriseStrategicImportanceRepositoryForUpdate{
		existingImportance: existingImportance,
		saveErr:            errors.New("save error"),
	}

	handler := newTestableUpdateEnterpriseStrategicImportanceHandler(mockRepo)

	cmd := &commands.UpdateEnterpriseStrategicImportance{
		ID:         existingImportance.ID(),
		Importance: 4,
		Rationale:  "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}
