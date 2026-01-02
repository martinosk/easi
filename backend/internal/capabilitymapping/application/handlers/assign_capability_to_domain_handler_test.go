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

type mockAssignCapabilityAssignmentRepository struct {
	savedAssignments []*aggregates.BusinessDomainAssignment
	saveErr          error
}

func (m *mockAssignCapabilityAssignmentRepository) Save(ctx context.Context, assignment *aggregates.BusinessDomainAssignment) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedAssignments = append(m.savedAssignments, assignment)
	return nil
}

type mockAssignCapabilityDomainReader struct {
	domain *readmodels.BusinessDomainDTO
	getErr error
}

func (m *mockAssignCapabilityDomainReader) GetByID(ctx context.Context, id string) (*readmodels.BusinessDomainDTO, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.domain, nil
}

type mockAssignCapabilityCapabilityRepository struct {
	capability *aggregates.Capability
	getErr     error
}

func (m *mockAssignCapabilityCapabilityRepository) GetByID(ctx context.Context, id string) (*aggregates.Capability, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.capability, nil
}

type mockAssignCapabilityAssignmentReader struct {
	exists   bool
	checkErr error
}

func (m *mockAssignCapabilityAssignmentReader) AssignmentExists(ctx context.Context, domainID, capabilityID string) (bool, error) {
	if m.checkErr != nil {
		return false, m.checkErr
	}
	return m.exists, nil
}

func createL1Capability(t *testing.T) *aggregates.Capability {
	t.Helper()

	name, err := valueobjects.NewCapabilityName("Test L1 Capability")
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

func createL2Capability(t *testing.T) *aggregates.Capability {
	t.Helper()

	name, err := valueobjects.NewCapabilityName("Test L2 Capability")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Test description")

	level, err := valueobjects.NewCapabilityLevel("L2")
	require.NoError(t, err)

	parentID := valueobjects.NewCapabilityID()

	capability, err := aggregates.NewCapability(name, description, parentID, level)
	require.NoError(t, err)
	capability.MarkChangesAsCommitted()

	return capability
}

func TestAssignCapabilityToDomainHandler_CreatesAssignment(t *testing.T) {
	businessDomainID := valueobjects.NewBusinessDomainID().Value()
	capability := createL1Capability(t)

	mockAssignmentRepo := &mockAssignCapabilityAssignmentRepository{}
	mockCapabilityRepo := &mockAssignCapabilityCapabilityRepository{
		capability: capability,
	}
	mockDomainReader := &mockAssignCapabilityDomainReader{
		domain: &readmodels.BusinessDomainDTO{ID: businessDomainID, Name: "Test Domain"},
	}
	mockAssignmentReader := &mockAssignCapabilityAssignmentReader{exists: false}

	handler := NewAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockCapabilityRepo,
		mockDomainReader,
		mockAssignmentReader,
	)

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: businessDomainID,
		CapabilityID:     capability.ID(),
	}

	result, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockAssignmentRepo.savedAssignments, 1, "Handler should create exactly 1 assignment")
	assert.NotEmpty(t, result.CreatedID, "Result CreatedID should be set")
}

func TestAssignCapabilityToDomainHandler_BusinessDomainNotFound_ReturnsError(t *testing.T) {
	capability := createL1Capability(t)

	mockAssignmentRepo := &mockAssignCapabilityAssignmentRepository{}
	mockCapabilityRepo := &mockAssignCapabilityCapabilityRepository{
		capability: capability,
	}
	mockDomainReader := &mockAssignCapabilityDomainReader{domain: nil}
	mockAssignmentReader := &mockAssignCapabilityAssignmentReader{exists: false}

	handler := NewAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockCapabilityRepo,
		mockDomainReader,
		mockAssignmentReader,
	)

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: valueobjects.NewBusinessDomainID().Value(),
		CapabilityID:     capability.ID(),
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrBusinessDomainNotFound)
	assert.Empty(t, mockAssignmentRepo.savedAssignments)
}

func TestAssignCapabilityToDomainHandler_CapabilityNotFound_ReturnsError(t *testing.T) {
	businessDomainID := valueobjects.NewBusinessDomainID().Value()

	mockAssignmentRepo := &mockAssignCapabilityAssignmentRepository{}
	mockCapabilityRepo := &mockAssignCapabilityCapabilityRepository{
		getErr: errors.New("capability not found"),
	}
	mockDomainReader := &mockAssignCapabilityDomainReader{
		domain: &readmodels.BusinessDomainDTO{ID: businessDomainID},
	}
	mockAssignmentReader := &mockAssignCapabilityAssignmentReader{exists: false}

	handler := NewAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockCapabilityRepo,
		mockDomainReader,
		mockAssignmentReader,
	)

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: businessDomainID,
		CapabilityID:     valueobjects.NewCapabilityID().Value(),
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockAssignmentRepo.savedAssignments)
}

func TestAssignCapabilityToDomainHandler_CapabilityNotL1_ReturnsError(t *testing.T) {
	businessDomainID := valueobjects.NewBusinessDomainID().Value()
	capability := createL2Capability(t)

	mockAssignmentRepo := &mockAssignCapabilityAssignmentRepository{}
	mockCapabilityRepo := &mockAssignCapabilityCapabilityRepository{
		capability: capability,
	}
	mockDomainReader := &mockAssignCapabilityDomainReader{
		domain: &readmodels.BusinessDomainDTO{ID: businessDomainID},
	}
	mockAssignmentReader := &mockAssignCapabilityAssignmentReader{exists: false}

	handler := NewAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockCapabilityRepo,
		mockDomainReader,
		mockAssignmentReader,
	)

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: businessDomainID,
		CapabilityID:     capability.ID(),
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, aggregates.ErrOnlyL1CanBeAssignedToDomain)
	assert.Empty(t, mockAssignmentRepo.savedAssignments)
}

func TestAssignCapabilityToDomainHandler_AssignmentExists_ReturnsError(t *testing.T) {
	businessDomainID := valueobjects.NewBusinessDomainID().Value()
	capability := createL1Capability(t)

	mockAssignmentRepo := &mockAssignCapabilityAssignmentRepository{}
	mockCapabilityRepo := &mockAssignCapabilityCapabilityRepository{
		capability: capability,
	}
	mockDomainReader := &mockAssignCapabilityDomainReader{
		domain: &readmodels.BusinessDomainDTO{ID: businessDomainID},
	}
	mockAssignmentReader := &mockAssignCapabilityAssignmentReader{exists: true}

	handler := NewAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockCapabilityRepo,
		mockDomainReader,
		mockAssignmentReader,
	)

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: businessDomainID,
		CapabilityID:     capability.ID(),
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrAssignmentAlreadyExists)
	assert.Empty(t, mockAssignmentRepo.savedAssignments)
}

func TestAssignCapabilityToDomainHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockAssignmentRepo := &mockAssignCapabilityAssignmentRepository{}
	mockCapabilityRepo := &mockAssignCapabilityCapabilityRepository{}
	mockDomainReader := &mockAssignCapabilityDomainReader{}
	mockAssignmentReader := &mockAssignCapabilityAssignmentReader{}

	handler := NewAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockCapabilityRepo,
		mockDomainReader,
		mockAssignmentReader,
	)

	invalidCmd := &commands.DeleteBusinessDomain{}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestAssignCapabilityToDomainHandler_AssignmentReaderError_ReturnsError(t *testing.T) {
	businessDomainID := valueobjects.NewBusinessDomainID().Value()
	capability := createL1Capability(t)

	mockAssignmentRepo := &mockAssignCapabilityAssignmentRepository{}
	mockCapabilityRepo := &mockAssignCapabilityCapabilityRepository{
		capability: capability,
	}
	mockDomainReader := &mockAssignCapabilityDomainReader{
		domain: &readmodels.BusinessDomainDTO{ID: businessDomainID},
	}
	mockAssignmentReader := &mockAssignCapabilityAssignmentReader{checkErr: errors.New("database error")}

	handler := NewAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockCapabilityRepo,
		mockDomainReader,
		mockAssignmentReader,
	)

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: businessDomainID,
		CapabilityID:     capability.ID(),
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockAssignmentRepo.savedAssignments)
}
