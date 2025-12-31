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

type mockAssignCapabilityCapabilityReader struct {
	capability *readmodels.CapabilityDTO
	getErr     error
}

func (m *mockAssignCapabilityCapabilityReader) GetByID(ctx context.Context, id string) (*readmodels.CapabilityDTO, error) {
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

func TestAssignCapabilityToDomainHandler_CreatesAssignment(t *testing.T) {
	businessDomainID := valueobjects.NewBusinessDomainID().Value()
	capabilityID := valueobjects.NewCapabilityID().Value()

	mockAssignmentRepo := &mockAssignCapabilityAssignmentRepository{}
	mockDomainReader := &mockAssignCapabilityDomainReader{
		domain: &readmodels.BusinessDomainDTO{ID: businessDomainID, Name: "Test Domain"},
	}
	mockCapabilityReader := &mockAssignCapabilityCapabilityReader{
		capability: &readmodels.CapabilityDTO{ID: capabilityID, Name: "Test Capability", Level: "L1"},
	}
	mockAssignmentReader := &mockAssignCapabilityAssignmentReader{exists: false}

	handler := NewAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockDomainReader,
		mockCapabilityReader,
		mockAssignmentReader,
	)

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: businessDomainID,
		CapabilityID:     capabilityID,
	}

	result, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockAssignmentRepo.savedAssignments, 1, "Handler should create exactly 1 assignment")
	assert.NotEmpty(t, result.CreatedID, "Result CreatedID should be set")
}

func TestAssignCapabilityToDomainHandler_BusinessDomainNotFound_ReturnsError(t *testing.T) {
	mockAssignmentRepo := &mockAssignCapabilityAssignmentRepository{}
	mockDomainReader := &mockAssignCapabilityDomainReader{domain: nil}
	mockCapabilityReader := &mockAssignCapabilityCapabilityReader{
		capability: &readmodels.CapabilityDTO{ID: "cap-456", Level: "L1"},
	}
	mockAssignmentReader := &mockAssignCapabilityAssignmentReader{exists: false}

	handler := NewAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockDomainReader,
		mockCapabilityReader,
		mockAssignmentReader,
	)

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: "bd-nonexistent",
		CapabilityID:     "cap-456",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrBusinessDomainNotFound)
	assert.Empty(t, mockAssignmentRepo.savedAssignments)
}

func TestAssignCapabilityToDomainHandler_CapabilityNotFound_ReturnsError(t *testing.T) {
	mockAssignmentRepo := &mockAssignCapabilityAssignmentRepository{}
	mockDomainReader := &mockAssignCapabilityDomainReader{
		domain: &readmodels.BusinessDomainDTO{ID: "bd-123"},
	}
	mockCapabilityReader := &mockAssignCapabilityCapabilityReader{capability: nil}
	mockAssignmentReader := &mockAssignCapabilityAssignmentReader{exists: false}

	handler := NewAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockDomainReader,
		mockCapabilityReader,
		mockAssignmentReader,
	)

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: "bd-123",
		CapabilityID:     "cap-nonexistent",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrCapabilityNotFound)
	assert.Empty(t, mockAssignmentRepo.savedAssignments)
}

func TestAssignCapabilityToDomainHandler_CapabilityNotL1_ReturnsError(t *testing.T) {
	businessDomainID := valueobjects.NewBusinessDomainID().Value()
	capabilityID := valueobjects.NewCapabilityID().Value()

	mockAssignmentRepo := &mockAssignCapabilityAssignmentRepository{}
	mockDomainReader := &mockAssignCapabilityDomainReader{
		domain: &readmodels.BusinessDomainDTO{ID: businessDomainID},
	}
	mockCapabilityReader := &mockAssignCapabilityCapabilityReader{
		capability: &readmodels.CapabilityDTO{ID: capabilityID, Level: "L2"},
	}
	mockAssignmentReader := &mockAssignCapabilityAssignmentReader{exists: false}

	handler := NewAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockDomainReader,
		mockCapabilityReader,
		mockAssignmentReader,
	)

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: businessDomainID,
		CapabilityID:     capabilityID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrOnlyL1CapabilitiesCanBeAssigned)
	assert.Empty(t, mockAssignmentRepo.savedAssignments)
}

func TestAssignCapabilityToDomainHandler_AssignmentExists_ReturnsError(t *testing.T) {
	businessDomainID := valueobjects.NewBusinessDomainID().Value()
	capabilityID := valueobjects.NewCapabilityID().Value()

	mockAssignmentRepo := &mockAssignCapabilityAssignmentRepository{}
	mockDomainReader := &mockAssignCapabilityDomainReader{
		domain: &readmodels.BusinessDomainDTO{ID: businessDomainID},
	}
	mockCapabilityReader := &mockAssignCapabilityCapabilityReader{
		capability: &readmodels.CapabilityDTO{ID: capabilityID, Level: "L1"},
	}
	mockAssignmentReader := &mockAssignCapabilityAssignmentReader{exists: true}

	handler := NewAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockDomainReader,
		mockCapabilityReader,
		mockAssignmentReader,
	)

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: businessDomainID,
		CapabilityID:     capabilityID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrAssignmentAlreadyExists)
	assert.Empty(t, mockAssignmentRepo.savedAssignments)
}

func TestAssignCapabilityToDomainHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockAssignmentRepo := &mockAssignCapabilityAssignmentRepository{}
	mockDomainReader := &mockAssignCapabilityDomainReader{}
	mockCapabilityReader := &mockAssignCapabilityCapabilityReader{}
	mockAssignmentReader := &mockAssignCapabilityAssignmentReader{}

	handler := NewAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockDomainReader,
		mockCapabilityReader,
		mockAssignmentReader,
	)

	invalidCmd := &commands.DeleteBusinessDomain{}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestAssignCapabilityToDomainHandler_AssignmentReaderError_ReturnsError(t *testing.T) {
	businessDomainID := valueobjects.NewBusinessDomainID().Value()
	capabilityID := valueobjects.NewCapabilityID().Value()

	mockAssignmentRepo := &mockAssignCapabilityAssignmentRepository{}
	mockDomainReader := &mockAssignCapabilityDomainReader{
		domain: &readmodels.BusinessDomainDTO{ID: businessDomainID},
	}
	mockCapabilityReader := &mockAssignCapabilityCapabilityReader{
		capability: &readmodels.CapabilityDTO{ID: capabilityID, Level: "L1"},
	}
	mockAssignmentReader := &mockAssignCapabilityAssignmentReader{checkErr: errors.New("database error")}

	handler := NewAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockDomainReader,
		mockCapabilityReader,
		mockAssignmentReader,
	)

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: businessDomainID,
		CapabilityID:     capabilityID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockAssignmentRepo.savedAssignments)
}
