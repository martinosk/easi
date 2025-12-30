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

type mockCapabilityRepository struct {
	capability *aggregates.Capability
	getByIDErr error
	saveErr    error
	savedCap   *aggregates.Capability
}

func (m *mockCapabilityRepository) GetByID(ctx context.Context, id string) (*aggregates.Capability, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.capability, nil
}

func (m *mockCapabilityRepository) Save(ctx context.Context, capability *aggregates.Capability) error {
	m.savedCap = capability
	return m.saveErr
}

type mockCapabilityReadModel struct {
	children       []readmodels.CapabilityDTO
	getChildrenErr error
}

func (m *mockCapabilityReadModel) GetChildren(ctx context.Context, parentID string) ([]readmodels.CapabilityDTO, error) {
	if m.getChildrenErr != nil {
		return nil, m.getChildrenErr
	}
	return m.children, nil
}

type capabilityRepository interface {
	GetByID(ctx context.Context, id string) (*aggregates.Capability, error)
	Save(ctx context.Context, capability *aggregates.Capability) error
}

type capabilityReadModelForDelete interface {
	GetChildren(ctx context.Context, parentID string) ([]readmodels.CapabilityDTO, error)
}

type testableDeleteCapabilityHandler struct {
	repository capabilityRepository
	readModel  capabilityReadModelForDelete
}

func newTestableDeleteCapabilityHandler(
	repository capabilityRepository,
	readModel capabilityReadModelForDelete,
) *testableDeleteCapabilityHandler {
	return &testableDeleteCapabilityHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *testableDeleteCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.DeleteCapability)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	children, err := h.readModel.GetChildren(ctx, command.ID)
	if err != nil {
		return err
	}

	if len(children) > 0 {
		return ErrCapabilityHasChildren
	}

	capability, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return err
	}

	if err := capability.Delete(); err != nil {
		return err
	}

	return h.repository.Save(ctx, capability)
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

	mockRepo := &mockCapabilityRepository{
		capability: capability,
	}
	mockReadModel := &mockCapabilityReadModel{
		children: []readmodels.CapabilityDTO{},
	}

	handler := newTestableDeleteCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.DeleteCapability{
		ID: capabilityID,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.NotNil(t, mockRepo.savedCap)
	uncommittedEvents := mockRepo.savedCap.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "CapabilityDeleted", uncommittedEvents[0].EventType())
}

func TestDeleteCapabilityHandler_CapabilityHasChildren_ReturnsError(t *testing.T) {
	capability := createTestCapability(t)
	capabilityID := capability.ID()

	mockRepo := &mockCapabilityRepository{
		capability: capability,
	}
	mockReadModel := &mockCapabilityReadModel{
		children: []readmodels.CapabilityDTO{
			{ID: "child-1", Name: "Child Capability"},
		},
	}

	handler := newTestableDeleteCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.DeleteCapability{
		ID: capabilityID,
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, ErrCapabilityHasChildren, err)

	assert.Nil(t, mockRepo.savedCap)
}

func TestDeleteCapabilityHandler_CapabilityNotFound_ReturnsError(t *testing.T) {
	notFoundErr := errors.New("capability not found")
	mockRepo := &mockCapabilityRepository{
		getByIDErr: notFoundErr,
	}
	mockReadModel := &mockCapabilityReadModel{
		children: []readmodels.CapabilityDTO{},
	}

	handler := newTestableDeleteCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.DeleteCapability{
		ID: "non-existent-id",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, notFoundErr, err)
}

func TestDeleteCapabilityHandler_ReadModelError_ReturnsError(t *testing.T) {
	readModelErr := errors.New("database connection error")
	mockRepo := &mockCapabilityRepository{}
	mockReadModel := &mockCapabilityReadModel{
		getChildrenErr: readModelErr,
	}

	handler := newTestableDeleteCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.DeleteCapability{
		ID: "some-id",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, readModelErr, err)
}

func TestDeleteCapabilityHandler_SaveError_ReturnsError(t *testing.T) {
	capability := createTestCapability(t)
	capabilityID := capability.ID()

	saveErr := errors.New("failed to save")
	mockRepo := &mockCapabilityRepository{
		capability: capability,
		saveErr:    saveErr,
	}
	mockReadModel := &mockCapabilityReadModel{
		children: []readmodels.CapabilityDTO{},
	}

	handler := newTestableDeleteCapabilityHandler(mockRepo, mockReadModel)

	cmd := &commands.DeleteCapability{
		ID: capabilityID,
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, saveErr, err)
}

func TestDeleteCapabilityHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockCapabilityRepository{}
	mockReadModel := &mockCapabilityReadModel{}

	handler := newTestableDeleteCapabilityHandler(mockRepo, mockReadModel)

	invalidCmd := &commands.CreateCapability{
		Name:  "Test",
		Level: "L1",
	}

	err := handler.Handle(context.Background(), invalidCmd)
	assert.Error(t, err)
	assert.Equal(t, cqrs.ErrInvalidCommand, err)
}
