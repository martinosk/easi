package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRemoveExpertCapabilityRepository struct {
	capability *aggregates.Capability
	getByIDErr error
	saveErr    error
	savedCap   *aggregates.Capability
}

func (m *mockRemoveExpertCapabilityRepository) GetByID(ctx context.Context, id string) (*aggregates.Capability, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.capability, nil
}

func (m *mockRemoveExpertCapabilityRepository) Save(ctx context.Context, capability *aggregates.Capability) error {
	m.savedCap = capability
	return m.saveErr
}

func createCapabilityWithExpert(t *testing.T) *aggregates.Capability {
	t.Helper()

	name, err := valueobjects.NewCapabilityName("Customer Management")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Test description")

	level, err := valueobjects.NewCapabilityLevel("L1")
	require.NoError(t, err)

	var parentID valueobjects.CapabilityID

	capability, err := aggregates.NewCapability(name, description, parentID, level)
	require.NoError(t, err)

	expert := valueobjects.MustNewExpert("Alice Smith", "Product Owner", "alice@example.com", time.Now().UTC())
	err = capability.AddExpert(expert)
	require.NoError(t, err)

	capability.MarkChangesAsCommitted()

	return capability
}

func TestRemoveCapabilityExpertHandler_Success(t *testing.T) {
	capability := createCapabilityWithExpert(t)
	capabilityID := capability.ID()

	mockRepo := &mockRemoveExpertCapabilityRepository{
		capability: capability,
	}

	handler := NewRemoveCapabilityExpertHandler(mockRepo)

	cmd := &commands.RemoveCapabilityExpert{
		CapabilityID: capabilityID,
		ExpertName:   "Alice Smith",
		ExpertRole:   "Product Owner",
		ContactInfo:  "alice@example.com",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.NotNil(t, mockRepo.savedCap)
	uncommittedEvents := mockRepo.savedCap.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "CapabilityExpertRemoved", uncommittedEvents[0].EventType())
}

func TestRemoveCapabilityExpertHandler_ExpertNoLongerInList(t *testing.T) {
	capability := createCapabilityWithExpert(t)
	capabilityID := capability.ID()

	mockRepo := &mockRemoveExpertCapabilityRepository{
		capability: capability,
	}

	handler := NewRemoveCapabilityExpertHandler(mockRepo)

	cmd := &commands.RemoveCapabilityExpert{
		CapabilityID: capabilityID,
		ExpertName:   "Alice Smith",
		ExpertRole:   "Product Owner",
		ContactInfo:  "alice@example.com",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.Empty(t, mockRepo.savedCap.Experts())
}

func TestRemoveCapabilityExpertHandler_CapabilityNotFound_ReturnsError(t *testing.T) {
	notFoundErr := errors.New("capability not found")
	mockRepo := &mockRemoveExpertCapabilityRepository{
		getByIDErr: notFoundErr,
	}

	handler := NewRemoveCapabilityExpertHandler(mockRepo)

	cmd := &commands.RemoveCapabilityExpert{
		CapabilityID: "550e8400-e29b-41d4-a716-446655440000",
		ExpertName:   "Alice Smith",
		ExpertRole:   "Product Owner",
		ContactInfo:  "alice@example.com",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, notFoundErr, err)
}

func TestRemoveCapabilityExpertHandler_InvalidExpertData_ReturnsError(t *testing.T) {
	capability := createCapabilityWithExpert(t)
	capabilityID := capability.ID()

	mockRepo := &mockRemoveExpertCapabilityRepository{
		capability: capability,
	}

	handler := NewRemoveCapabilityExpertHandler(mockRepo)

	cmd := &commands.RemoveCapabilityExpert{
		CapabilityID: capabilityID,
		ExpertName:   "",
		ExpertRole:   "Product Owner",
		ContactInfo:  "alice@example.com",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, valueobjects.ErrExpertNameEmpty, err)
}

func TestRemoveCapabilityExpertHandler_SaveError_ReturnsError(t *testing.T) {
	capability := createCapabilityWithExpert(t)
	capabilityID := capability.ID()

	saveErr := errors.New("failed to save")
	mockRepo := &mockRemoveExpertCapabilityRepository{
		capability: capability,
		saveErr:    saveErr,
	}

	handler := NewRemoveCapabilityExpertHandler(mockRepo)

	cmd := &commands.RemoveCapabilityExpert{
		CapabilityID: capabilityID,
		ExpertName:   "Alice Smith",
		ExpertRole:   "Product Owner",
		ContactInfo:  "alice@example.com",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, saveErr, err)
}

func TestRemoveCapabilityExpertHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockRemoveExpertCapabilityRepository{}

	handler := NewRemoveCapabilityExpertHandler(mockRepo)

	invalidCmd := &commands.AddCapabilityExpert{
		CapabilityID: "test-id",
		ExpertName:   "Alice Smith",
		ExpertRole:   "Product Owner",
		ContactInfo:  "alice@example.com",
	}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.Error(t, err)
	assert.Equal(t, cqrs.ErrInvalidCommand, err)
}

func TestRemoveCapabilityExpertHandler_RoleRemainsAvailable_WhenOtherExpertUsesIt(t *testing.T) {
	name, err := valueobjects.NewCapabilityName("Customer Management")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Test description")

	level, err := valueobjects.NewCapabilityLevel("L1")
	require.NoError(t, err)

	var parentID valueobjects.CapabilityID

	capability, err := aggregates.NewCapability(name, description, parentID, level)
	require.NoError(t, err)

	expert1 := valueobjects.MustNewExpert("Alice Smith", "Product Owner", "alice@example.com", time.Now().UTC())
	expert2 := valueobjects.MustNewExpert("Bob Jones", "Product Owner", "bob@example.com", time.Now().UTC())
	_ = capability.AddExpert(expert1)
	_ = capability.AddExpert(expert2)
	capability.MarkChangesAsCommitted()

	mockRepo := &mockRemoveExpertCapabilityRepository{
		capability: capability,
	}

	handler := NewRemoveCapabilityExpertHandler(mockRepo)

	cmd := &commands.RemoveCapabilityExpert{
		CapabilityID: capability.ID(),
		ExpertName:   "Alice Smith",
		ExpertRole:   "Product Owner",
		ContactInfo:  "alice@example.com",
	}

	_, err = handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	experts := mockRepo.savedCap.Experts()
	require.Len(t, experts, 1)
	assert.Equal(t, "Bob Jones", experts[0].Name())
	assert.Equal(t, "Product Owner", experts[0].Role())
}
