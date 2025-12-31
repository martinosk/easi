package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
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

type mockDeleteCapabilityReadModel struct {
	children       []readmodels.CapabilityDTO
	getChildrenErr error
}

func (m *mockDeleteCapabilityReadModel) GetChildren(ctx context.Context, parentID string) ([]readmodels.CapabilityDTO, error) {
	if m.getChildrenErr != nil {
		return nil, m.getChildrenErr
	}
	return m.children, nil
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
	mockReadModel := &mockDeleteCapabilityReadModel{
		children: []readmodels.CapabilityDTO{},
	}

	handler := NewDeleteCapabilityHandler(mockRepo, mockReadModel)

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
	mockReadModel := &mockDeleteCapabilityReadModel{
		children: []readmodels.CapabilityDTO{
			{ID: "child-1", Name: "Child Capability"},
		},
	}

	handler := NewDeleteCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.DeleteCapability{
		ID: capabilityID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, ErrCapabilityHasChildren, err)

	assert.Nil(t, mockRepo.savedCap)
}

func TestDeleteCapabilityHandler_CapabilityNotFound_ReturnsError(t *testing.T) {
	notFoundErr := errors.New("capability not found")
	mockRepo := &mockDeleteCapabilityRepository{
		getByIDErr: notFoundErr,
	}
	mockReadModel := &mockDeleteCapabilityReadModel{
		children: []readmodels.CapabilityDTO{},
	}

	handler := NewDeleteCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.DeleteCapability{
		ID: "non-existent-id",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, notFoundErr, err)
}

func TestDeleteCapabilityHandler_ReadModelError_ReturnsError(t *testing.T) {
	readModelErr := errors.New("database connection error")
	mockRepo := &mockDeleteCapabilityRepository{}
	mockReadModel := &mockDeleteCapabilityReadModel{
		getChildrenErr: readModelErr,
	}

	handler := NewDeleteCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.DeleteCapability{
		ID: "some-id",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, readModelErr, err)
}

func TestDeleteCapabilityHandler_SaveError_ReturnsError(t *testing.T) {
	capability := createTestCapability(t)
	capabilityID := capability.ID()

	saveErr := errors.New("failed to save")
	mockRepo := &mockDeleteCapabilityRepository{
		capability: capability,
		saveErr:    saveErr,
	}
	mockReadModel := &mockDeleteCapabilityReadModel{
		children: []readmodels.CapabilityDTO{},
	}

	handler := NewDeleteCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.DeleteCapability{
		ID: capabilityID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, saveErr, err)
}

func TestDeleteCapabilityHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockDeleteCapabilityRepository{}
	mockReadModel := &mockDeleteCapabilityReadModel{}

	handler := NewDeleteCapabilityHandler(mockRepo, mockReadModel)

	invalidCmd := &commands.CreateCapability{
		Name:  "Test",
		Level: "L1",
	}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.Error(t, err)
	assert.Equal(t, cqrs.ErrInvalidCommand, err)
}
