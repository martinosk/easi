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

func createRealization(t *testing.T) *aggregates.CapabilityRealization {
	t.Helper()

	capabilityID := valueobjects.NewCapabilityID()
	componentID, err := valueobjects.NewComponentIDFromString(valueobjects.NewCapabilityID().Value())
	require.NoError(t, err)
	level, _ := valueobjects.NewRealizationLevel("Full")
	notes := valueobjects.MustNewDescription("Test notes")

	realization, err := aggregates.NewCapabilityRealization(capabilityID, componentID, "Test Component", level, notes)
	require.NoError(t, err)
	realization.MarkChangesAsCommitted()

	return realization
}

func TestDeleteSystemRealizationHandler_DeletesRealization(t *testing.T) {
	realization := createRealization(t)
	realizationID := realization.ID()

	mockRepo := newMockRealizationRepositoryForDelete()
	mockRepo.addRealization(realization)

	handler := NewDeleteSystemRealizationHandler(mockRepo)

	cmd := &commands.DeleteSystemRealization{
		ID: realizationID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.Contains(t, mockRepo.savedIDs, realizationID, "Realization should be saved (deleted via event)")
}

func TestDeleteSystemRealizationHandler_RealizationNotFound_ReturnsError(t *testing.T) {
	mockRepo := newMockRealizationRepositoryForDelete()

	handler := NewDeleteSystemRealizationHandler(mockRepo)

	cmd := &commands.DeleteSystemRealization{
		ID: "non-existent-id",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, repositories.ErrRealizationNotFound)
}

func TestDeleteSystemRealizationHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := newMockRealizationRepositoryForDelete()

	handler := NewDeleteSystemRealizationHandler(mockRepo)

	invalidCmd := &commands.LinkSystemToCapability{}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}
