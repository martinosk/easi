package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDeleteCapabilityRepository struct {
	capability *aggregates.Capability
	getByIDErr error
	saveErr    error
	savedCap   *aggregates.Capability
}

func (m *mockDeleteCapabilityRepository) GetByID(ctx context.Context, id string) (*aggregates.Capability, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.capability, nil
}

func (m *mockDeleteCapabilityRepository) Save(ctx context.Context, capability *aggregates.Capability) error {
	m.savedCap = capability
	return m.saveErr
}

type mockCapabilityDeletionService struct {
	canDeleteErr error
}

func (m *mockCapabilityDeletionService) CanDelete(ctx context.Context, capabilityID valueobjects.CapabilityID) error {
	return m.canDeleteErr
}

func createTestCapability(t *testing.T) *aggregates.Capability {
	t.Helper()

	name, err := valueobjects.NewCapabilityName("Test Capability")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Test description")

	level, err := valueobjects.NewCapabilityLevel("L1")
	require.NoError(t, err)

	var parentID valueobjects.CapabilityID

	capability, err := aggregates.NewCapability(name, description, parentID, level)
	require.NoError(t, err)
	capability.MarkChangesAsCommitted()

	return capability
}

func TestDeleteCapabilityHandler_Success(t *testing.T) {
	capability := createTestCapability(t)
	capabilityID := capability.ID()

	mockRepo := &mockDeleteCapabilityRepository{
		capability: capability,
	}
	mockDeletionService := &mockCapabilityDeletionService{}

	handler := NewDeleteCapabilityHandler(mockRepo, mockDeletionService)

	cmd := &commands.DeleteCapability{
		ID: capabilityID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.NotNil(t, mockRepo.savedCap)
	uncommittedEvents := mockRepo.savedCap.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "CapabilityDeleted", uncommittedEvents[0].EventType())
}

func TestDeleteCapabilityHandler_CapabilityHasChildren_ReturnsError(t *testing.T) {
	capability := createTestCapability(t)
	capabilityID := capability.ID()

	mockRepo := &mockDeleteCapabilityRepository{
		capability: capability,
	}
	mockDeletionService := &mockCapabilityDeletionService{
		canDeleteErr: services.ErrCapabilityHasChildren,
	}

	handler := NewDeleteCapabilityHandler(mockRepo, mockDeletionService)

	cmd := &commands.DeleteCapability{
		ID: capabilityID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, services.ErrCapabilityHasChildren, err)

	assert.Nil(t, mockRepo.savedCap)
}

func TestDeleteCapabilityHandler_CapabilityNotFound_ReturnsError(t *testing.T) {
	notFoundErr := errors.New("capability not found")
	mockRepo := &mockDeleteCapabilityRepository{
		getByIDErr: notFoundErr,
	}
	mockDeletionService := &mockCapabilityDeletionService{}

	handler := NewDeleteCapabilityHandler(mockRepo, mockDeletionService)

	cmd := &commands.DeleteCapability{
		ID: "550e8400-e29b-41d4-a716-446655440000",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, notFoundErr, err)
}

func TestDeleteCapabilityHandler_DeletionServiceError_ReturnsError(t *testing.T) {
	capability := createTestCapability(t)
	capabilityID := capability.ID()

	serviceErr := errors.New("database connection error")
	mockRepo := &mockDeleteCapabilityRepository{
		capability: capability,
	}
	mockDeletionService := &mockCapabilityDeletionService{
		canDeleteErr: serviceErr,
	}

	handler := NewDeleteCapabilityHandler(mockRepo, mockDeletionService)

	cmd := &commands.DeleteCapability{
		ID: capabilityID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, serviceErr, err)
}

func TestDeleteCapabilityHandler_SaveError_ReturnsError(t *testing.T) {
	capability := createTestCapability(t)
	capabilityID := capability.ID()

	saveErr := errors.New("failed to save")
	mockRepo := &mockDeleteCapabilityRepository{
		capability: capability,
		saveErr:    saveErr,
	}
	mockDeletionService := &mockCapabilityDeletionService{}

	handler := NewDeleteCapabilityHandler(mockRepo, mockDeletionService)

	cmd := &commands.DeleteCapability{
		ID: capabilityID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, saveErr, err)
}

func TestDeleteCapabilityHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockDeleteCapabilityRepository{}
	mockDeletionService := &mockCapabilityDeletionService{}

	handler := NewDeleteCapabilityHandler(mockRepo, mockDeletionService)

	invalidCmd := &commands.CreateCapability{
		Name:  "Test",
		Level: "L1",
	}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.Error(t, err)
	assert.Equal(t, cqrs.ErrInvalidCommand, err)
}
