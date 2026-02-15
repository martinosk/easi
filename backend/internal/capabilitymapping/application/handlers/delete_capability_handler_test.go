package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
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

type mockDeleteRealizationReadModel struct {
	realizations []readmodels.RealizationDTO
	err          error
}

func (m *mockDeleteRealizationReadModel) GetByCapabilityID(ctx context.Context, capabilityID string) ([]readmodels.RealizationDTO, error) {
	return m.realizations, m.err
}

type mockDeleteCapabilityLookup struct {
	capabilities map[string]*readmodels.CapabilityDTO
}

func (m *mockDeleteCapabilityLookup) GetByID(ctx context.Context, id string) (*readmodels.CapabilityDTO, error) {
	if cap, ok := m.capabilities[id]; ok {
		return cap, nil
	}
	return nil, nil
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

func createTestCapabilityWithParent(t *testing.T, parentIDStr string) *aggregates.Capability {
	t.Helper()

	name, err := valueobjects.NewCapabilityName("Child Capability")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Child description")

	level, err := valueobjects.NewCapabilityLevel("L2")
	require.NoError(t, err)

	parentID, err := valueobjects.NewCapabilityIDFromString(parentIDStr)
	require.NoError(t, err)

	capability, err := aggregates.NewCapability(name, description, parentID, level)
	require.NoError(t, err)
	capability.MarkChangesAsCommitted()

	return capability
}

func newDeleteHandler(repo *mockDeleteCapabilityRepository, deletionSvc *mockCapabilityDeletionService, realizationRM *mockDeleteRealizationReadModel, capLookup *mockDeleteCapabilityLookup) *DeleteCapabilityHandler {
	return NewDeleteCapabilityHandler(repo, deletionSvc, realizationRM, capLookup)
}

func TestDeleteCapabilityHandler_Success(t *testing.T) {
	capability := createTestCapability(t)
	capabilityID := capability.ID()

	mockRepo := &mockDeleteCapabilityRepository{capability: capability}
	handler := newDeleteHandler(mockRepo, &mockCapabilityDeletionService{}, &mockDeleteRealizationReadModel{}, &mockDeleteCapabilityLookup{})

	cmd := &commands.DeleteCapability{ID: capabilityID}

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

	mockRepo := &mockDeleteCapabilityRepository{capability: capability}
	handler := newDeleteHandler(mockRepo, &mockCapabilityDeletionService{canDeleteErr: services.ErrCapabilityHasChildren}, &mockDeleteRealizationReadModel{}, &mockDeleteCapabilityLookup{})

	cmd := &commands.DeleteCapability{ID: capabilityID}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, services.ErrCapabilityHasChildren, err)
	assert.Nil(t, mockRepo.savedCap)
}

func TestDeleteCapabilityHandler_CapabilityNotFound_ReturnsError(t *testing.T) {
	notFoundErr := errors.New("capability not found")
	mockRepo := &mockDeleteCapabilityRepository{getByIDErr: notFoundErr}
	handler := newDeleteHandler(mockRepo, &mockCapabilityDeletionService{}, &mockDeleteRealizationReadModel{}, &mockDeleteCapabilityLookup{})

	cmd := &commands.DeleteCapability{ID: "550e8400-e29b-41d4-a716-446655440000"}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, notFoundErr, err)
}

func TestDeleteCapabilityHandler_DeletionServiceError_ReturnsError(t *testing.T) {
	capability := createTestCapability(t)
	capabilityID := capability.ID()

	serviceErr := errors.New("database connection error")
	mockRepo := &mockDeleteCapabilityRepository{capability: capability}
	handler := newDeleteHandler(mockRepo, &mockCapabilityDeletionService{canDeleteErr: serviceErr}, &mockDeleteRealizationReadModel{}, &mockDeleteCapabilityLookup{})

	cmd := &commands.DeleteCapability{ID: capabilityID}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, serviceErr, err)
}

func TestDeleteCapabilityHandler_SaveError_ReturnsError(t *testing.T) {
	capability := createTestCapability(t)
	capabilityID := capability.ID()

	saveErr := errors.New("failed to save")
	mockRepo := &mockDeleteCapabilityRepository{capability: capability, saveErr: saveErr}
	handler := newDeleteHandler(mockRepo, &mockCapabilityDeletionService{}, &mockDeleteRealizationReadModel{}, &mockDeleteCapabilityLookup{})

	cmd := &commands.DeleteCapability{ID: capabilityID}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, saveErr, err)
}

func TestDeleteCapabilityHandler_InvalidCommand_ReturnsError(t *testing.T) {
	handler := newDeleteHandler(&mockDeleteCapabilityRepository{}, &mockCapabilityDeletionService{}, &mockDeleteRealizationReadModel{}, &mockDeleteCapabilityLookup{})

	invalidCmd := &commands.CreateCapability{Name: "Test", Level: "L1"}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.Error(t, err)
	assert.Equal(t, cqrs.ErrInvalidCommand, err)
}

func TestDeleteCapabilityHandler_WithRealisations_RemovesInheritedRealisations(t *testing.T) {
	parentCapabilityID := "550e8400-e29b-41d4-a716-446655440001"
	childCapability := createTestCapabilityWithParent(t, parentCapabilityID)
	realizationID := "550e8400-e29b-41d4-a716-446655440099"

	mockRepo := &mockDeleteCapabilityRepository{capability: childCapability}
	mockRealizationRM := &mockDeleteRealizationReadModel{
		realizations: []readmodels.RealizationDTO{
			{
				ID:               realizationID,
				CapabilityID:     childCapability.ID(),
				ComponentID:      "comp-1",
				ComponentName:    "App 1",
				RealizationLevel: "Full",
				Origin:           "Direct",
				LinkedAt:         time.Now(),
			},
		},
	}
	mockCapLookup := &mockDeleteCapabilityLookup{
		capabilities: map[string]*readmodels.CapabilityDTO{
			parentCapabilityID: {ID: parentCapabilityID, Name: "Parent Capability", ParentID: ""},
		},
	}

	handler := newDeleteHandler(mockRepo, &mockCapabilityDeletionService{}, mockRealizationRM, mockCapLookup)

	cmd := &commands.DeleteCapability{ID: childCapability.ID()}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.NotNil(t, mockRepo.savedCap)
	uncommittedEvents := mockRepo.savedCap.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 2)
	assert.Equal(t, "CapabilityDeleted", uncommittedEvents[0].EventType())
	assert.Equal(t, "CapabilityRealizationsUninherited", uncommittedEvents[1].EventType())

	uninheritedEvent, ok := uncommittedEvents[1].(events.CapabilityRealizationsUninherited)
	require.True(t, ok)
	require.Len(t, uninheritedEvent.Removals, 1)
	assert.Equal(t, realizationID, uninheritedEvent.Removals[0].SourceRealizationID)
	assert.Contains(t, uninheritedEvent.Removals[0].CapabilityIDs, parentCapabilityID)
}

func TestDeleteCapabilityHandler_WithMultipleAncestors_RemovesFromAllAncestors(t *testing.T) {
	grandparentID := "550e8400-e29b-41d4-a716-446655440001"
	parentID := "550e8400-e29b-41d4-a716-446655440002"
	childCapability := createTestCapabilityWithParent(t, parentID)
	realizationID := "550e8400-e29b-41d4-a716-446655440099"

	mockRepo := &mockDeleteCapabilityRepository{capability: childCapability}
	mockRealizationRM := &mockDeleteRealizationReadModel{
		realizations: []readmodels.RealizationDTO{
			{
				ID:               realizationID,
				CapabilityID:     childCapability.ID(),
				ComponentID:      "comp-1",
				Origin:           "Direct",
				LinkedAt:         time.Now(),
			},
		},
	}
	mockCapLookup := &mockDeleteCapabilityLookup{
		capabilities: map[string]*readmodels.CapabilityDTO{
			parentID:      {ID: parentID, Name: "Parent", ParentID: grandparentID},
			grandparentID: {ID: grandparentID, Name: "Grandparent", ParentID: ""},
		},
	}

	handler := newDeleteHandler(mockRepo, &mockCapabilityDeletionService{}, mockRealizationRM, mockCapLookup)

	cmd := &commands.DeleteCapability{ID: childCapability.ID()}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	uncommittedEvents := mockRepo.savedCap.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 2)

	uninheritedEvent, ok := uncommittedEvents[1].(events.CapabilityRealizationsUninherited)
	require.True(t, ok)
	require.Len(t, uninheritedEvent.Removals, 1)
	assert.Contains(t, uninheritedEvent.Removals[0].CapabilityIDs, parentID)
	assert.Contains(t, uninheritedEvent.Removals[0].CapabilityIDs, grandparentID)
}

func TestDeleteCapabilityHandler_NoParent_NoInheritanceCleanup(t *testing.T) {
	capability := createTestCapability(t)

	mockRepo := &mockDeleteCapabilityRepository{capability: capability}
	mockRealizationRM := &mockDeleteRealizationReadModel{
		realizations: []readmodels.RealizationDTO{
			{
				ID:           "r1",
				CapabilityID: capability.ID(),
				ComponentID:  "comp-1",
				Origin:       "Direct",
				LinkedAt:     time.Now(),
			},
		},
	}

	handler := newDeleteHandler(mockRepo, &mockCapabilityDeletionService{}, mockRealizationRM, &mockDeleteCapabilityLookup{})

	cmd := &commands.DeleteCapability{ID: capability.ID()}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	uncommittedEvents := mockRepo.savedCap.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "CapabilityDeleted", uncommittedEvents[0].EventType())
}

func TestDeleteCapabilityHandler_NoRealisations_NoInheritanceCleanup(t *testing.T) {
	parentID := "550e8400-e29b-41d4-a716-446655440001"
	childCapability := createTestCapabilityWithParent(t, parentID)

	mockRepo := &mockDeleteCapabilityRepository{capability: childCapability}
	mockRealizationRM := &mockDeleteRealizationReadModel{realizations: nil}

	handler := newDeleteHandler(mockRepo, &mockCapabilityDeletionService{}, mockRealizationRM, &mockDeleteCapabilityLookup{})

	cmd := &commands.DeleteCapability{ID: childCapability.ID()}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	uncommittedEvents := mockRepo.savedCap.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "CapabilityDeleted", uncommittedEvents[0].EventType())
}
