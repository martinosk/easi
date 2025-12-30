package handlers

import (
	"context"
	"testing"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRealizationRepositoryForDelete struct {
	realizations map[string]*aggregates.CapabilityRealization
	savedIDs     []string
	getErr       error
	saveErr      error
}

func newMockRealizationRepositoryForDelete() *mockRealizationRepositoryForDelete {
	return &mockRealizationRepositoryForDelete{
		realizations: make(map[string]*aggregates.CapabilityRealization),
	}
}

func (m *mockRealizationRepositoryForDelete) GetByID(ctx context.Context, id string) (*aggregates.CapabilityRealization, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if realization, ok := m.realizations[id]; ok {
		return realization, nil
	}
	return nil, repositories.ErrRealizationNotFound
}

func (m *mockRealizationRepositoryForDelete) Save(ctx context.Context, realization *aggregates.CapabilityRealization) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedIDs = append(m.savedIDs, realization.ID())
	return nil
}

func (m *mockRealizationRepositoryForDelete) addRealization(realization *aggregates.CapabilityRealization) {
	m.realizations[realization.ID()] = realization
}

type realizationRepositoryForDelete interface {
	GetByID(ctx context.Context, id string) (*aggregates.CapabilityRealization, error)
	Save(ctx context.Context, realization *aggregates.CapabilityRealization) error
}

type testableDeleteSystemRealizationHandler struct {
	repository realizationRepositoryForDelete
}

func newTestableDeleteSystemRealizationHandler(
	repository realizationRepositoryForDelete,
) *testableDeleteSystemRealizationHandler {
	return &testableDeleteSystemRealizationHandler{
		repository: repository,
	}
}

func (h *testableDeleteSystemRealizationHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.DeleteSystemRealization)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	realization, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	if err := realization.Delete(); err != nil {
		return err
	}

	return h.repository.Save(ctx, realization)
}

func createRealization(t *testing.T) *aggregates.CapabilityRealization {
	t.Helper()

	capabilityID := valueobjects.NewCapabilityID()
	componentID, err := valueobjects.NewComponentIDFromString(valueobjects.NewCapabilityID().Value())
	require.NoError(t, err)
	level, _ := valueobjects.NewRealizationLevel("Full")
	notes := valueobjects.MustNewDescription("Test notes")

	realization, err := aggregates.NewCapabilityRealization(capabilityID, componentID, level, notes)
	require.NoError(t, err)
	realization.MarkChangesAsCommitted()

	return realization
}

func TestDeleteSystemRealizationHandler_DeletesRealization(t *testing.T) {
	realization := createRealization(t)
	realizationID := realization.ID()

	mockRepo := newMockRealizationRepositoryForDelete()
	mockRepo.addRealization(realization)

	handler := newTestableDeleteSystemRealizationHandler(mockRepo)

	cmd := &commands.DeleteSystemRealization{
		ID: realizationID,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.Contains(t, mockRepo.savedIDs, realizationID, "Realization should be saved (deleted via event)")
}

func TestDeleteSystemRealizationHandler_RealizationNotFound_ReturnsError(t *testing.T) {
	mockRepo := newMockRealizationRepositoryForDelete()

	handler := newTestableDeleteSystemRealizationHandler(mockRepo)

	cmd := &commands.DeleteSystemRealization{
		ID: "non-existent-id",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, repositories.ErrRealizationNotFound)
}

func TestDeleteSystemRealizationHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := newMockRealizationRepositoryForDelete()

	handler := newTestableDeleteSystemRealizationHandler(mockRepo)

	invalidCmd := &commands.LinkSystemToCapability{}

	err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}
