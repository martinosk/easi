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

type mockAssignmentRepository struct {
	savedAssignments []*aggregates.BusinessDomainAssignment
	saveErr          error
}

func (m *mockAssignmentRepository) Save(ctx context.Context, assignment *aggregates.BusinessDomainAssignment) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedAssignments = append(m.savedAssignments, assignment)
	return nil
}

type mockBusinessDomainReadModelForAssign struct {
	domain   *readmodels.BusinessDomainDTO
	getErr   error
}

func (m *mockBusinessDomainReadModelForAssign) GetByID(ctx context.Context, id string) (*readmodels.BusinessDomainDTO, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.domain, nil
}

type mockCapabilityReadModelForAssign struct {
	capability *readmodels.CapabilityDTO
	getErr     error
}

func (m *mockCapabilityReadModelForAssign) GetByID(ctx context.Context, id string) (*readmodels.CapabilityDTO, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.capability, nil
}

type mockAssignmentReadModelForAssign struct {
	exists   bool
	checkErr error
}

func (m *mockAssignmentReadModelForAssign) AssignmentExists(ctx context.Context, domainID, capabilityID string) (bool, error) {
	if m.checkErr != nil {
		return false, m.checkErr
	}
	return m.exists, nil
}

type assignmentRepository interface {
	Save(ctx context.Context, assignment *aggregates.BusinessDomainAssignment) error
}

type businessDomainReadModelForAssign interface {
	GetByID(ctx context.Context, id string) (*readmodels.BusinessDomainDTO, error)
}

type capabilityReadModelForAssign interface {
	GetByID(ctx context.Context, id string) (*readmodels.CapabilityDTO, error)
}

type assignmentReadModelForAssign interface {
	AssignmentExists(ctx context.Context, domainID, capabilityID string) (bool, error)
}

type testableAssignCapabilityToDomainHandler struct {
	assignmentRepo   assignmentRepository
	domainReader     businessDomainReadModelForAssign
	capabilityReader capabilityReadModelForAssign
	assignmentReader assignmentReadModelForAssign
}

func newTestableAssignCapabilityToDomainHandler(
	assignmentRepo assignmentRepository,
	domainReader businessDomainReadModelForAssign,
	capabilityReader capabilityReadModelForAssign,
	assignmentReader assignmentReadModelForAssign,
) *testableAssignCapabilityToDomainHandler {
	return &testableAssignCapabilityToDomainHandler{
		assignmentRepo:   assignmentRepo,
		domainReader:     domainReader,
		capabilityReader: capabilityReader,
		assignmentReader: assignmentReader,
	}
}

func (h *testableAssignCapabilityToDomainHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.AssignCapabilityToDomain)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	domain, err := h.domainReader.GetByID(ctx, command.BusinessDomainID)
	if err != nil {
		return err
	}
	if domain == nil {
		return ErrBusinessDomainNotFound
	}

	capability, err := h.capabilityReader.GetByID(ctx, command.CapabilityID)
	if err != nil {
		return err
	}
	if capability == nil {
		return ErrCapabilityNotFound
	}

	if capability.Level != "L1" {
		return ErrOnlyL1CapabilitiesCanBeAssigned
	}

	exists, err := h.assignmentReader.AssignmentExists(ctx, command.BusinessDomainID, command.CapabilityID)
	if err != nil {
		return err
	}
	if exists {
		return ErrAssignmentAlreadyExists
	}

	businessDomainID, err := valueobjects.NewBusinessDomainIDFromString(command.BusinessDomainID)
	if err != nil {
		return err
	}

	capabilityID, err := valueobjects.NewCapabilityIDFromString(command.CapabilityID)
	if err != nil {
		return err
	}

	assignment, err := aggregates.AssignCapabilityToDomain(businessDomainID, capabilityID)
	if err != nil {
		return err
	}

	command.AssignmentID = assignment.ID()

	return h.assignmentRepo.Save(ctx, assignment)
}

func TestAssignCapabilityToDomainHandler_CreatesAssignment(t *testing.T) {
	businessDomainID := valueobjects.NewBusinessDomainID().Value()
	capabilityID := valueobjects.NewCapabilityID().Value()

	mockAssignmentRepo := &mockAssignmentRepository{}
	mockDomainReader := &mockBusinessDomainReadModelForAssign{
		domain: &readmodels.BusinessDomainDTO{ID: businessDomainID, Name: "Test Domain"},
	}
	mockCapabilityReader := &mockCapabilityReadModelForAssign{
		capability: &readmodels.CapabilityDTO{ID: capabilityID, Name: "Test Capability", Level: "L1"},
	}
	mockAssignmentReader := &mockAssignmentReadModelForAssign{exists: false}

	handler := newTestableAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockDomainReader,
		mockCapabilityReader,
		mockAssignmentReader,
	)

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: businessDomainID,
		CapabilityID:     capabilityID,
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockAssignmentRepo.savedAssignments, 1, "Handler should create exactly 1 assignment")
	assert.NotEmpty(t, cmd.AssignmentID, "Command AssignmentID should be set")
}

func TestAssignCapabilityToDomainHandler_BusinessDomainNotFound_ReturnsError(t *testing.T) {
	mockAssignmentRepo := &mockAssignmentRepository{}
	mockDomainReader := &mockBusinessDomainReadModelForAssign{domain: nil}
	mockCapabilityReader := &mockCapabilityReadModelForAssign{
		capability: &readmodels.CapabilityDTO{ID: "cap-456", Level: "L1"},
	}
	mockAssignmentReader := &mockAssignmentReadModelForAssign{exists: false}

	handler := newTestableAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockDomainReader,
		mockCapabilityReader,
		mockAssignmentReader,
	)

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: "bd-nonexistent",
		CapabilityID:     "cap-456",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrBusinessDomainNotFound)
	assert.Empty(t, mockAssignmentRepo.savedAssignments)
}

func TestAssignCapabilityToDomainHandler_CapabilityNotFound_ReturnsError(t *testing.T) {
	mockAssignmentRepo := &mockAssignmentRepository{}
	mockDomainReader := &mockBusinessDomainReadModelForAssign{
		domain: &readmodels.BusinessDomainDTO{ID: "bd-123"},
	}
	mockCapabilityReader := &mockCapabilityReadModelForAssign{capability: nil}
	mockAssignmentReader := &mockAssignmentReadModelForAssign{exists: false}

	handler := newTestableAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockDomainReader,
		mockCapabilityReader,
		mockAssignmentReader,
	)

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: "bd-123",
		CapabilityID:     "cap-nonexistent",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrCapabilityNotFound)
	assert.Empty(t, mockAssignmentRepo.savedAssignments)
}

func TestAssignCapabilityToDomainHandler_CapabilityNotL1_ReturnsError(t *testing.T) {
	businessDomainID := valueobjects.NewBusinessDomainID().Value()
	capabilityID := valueobjects.NewCapabilityID().Value()

	mockAssignmentRepo := &mockAssignmentRepository{}
	mockDomainReader := &mockBusinessDomainReadModelForAssign{
		domain: &readmodels.BusinessDomainDTO{ID: businessDomainID},
	}
	mockCapabilityReader := &mockCapabilityReadModelForAssign{
		capability: &readmodels.CapabilityDTO{ID: capabilityID, Level: "L2"},
	}
	mockAssignmentReader := &mockAssignmentReadModelForAssign{exists: false}

	handler := newTestableAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockDomainReader,
		mockCapabilityReader,
		mockAssignmentReader,
	)

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: businessDomainID,
		CapabilityID:     capabilityID,
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrOnlyL1CapabilitiesCanBeAssigned)
	assert.Empty(t, mockAssignmentRepo.savedAssignments)
}

func TestAssignCapabilityToDomainHandler_AssignmentExists_ReturnsError(t *testing.T) {
	businessDomainID := valueobjects.NewBusinessDomainID().Value()
	capabilityID := valueobjects.NewCapabilityID().Value()

	mockAssignmentRepo := &mockAssignmentRepository{}
	mockDomainReader := &mockBusinessDomainReadModelForAssign{
		domain: &readmodels.BusinessDomainDTO{ID: businessDomainID},
	}
	mockCapabilityReader := &mockCapabilityReadModelForAssign{
		capability: &readmodels.CapabilityDTO{ID: capabilityID, Level: "L1"},
	}
	mockAssignmentReader := &mockAssignmentReadModelForAssign{exists: true}

	handler := newTestableAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockDomainReader,
		mockCapabilityReader,
		mockAssignmentReader,
	)

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: businessDomainID,
		CapabilityID:     capabilityID,
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrAssignmentAlreadyExists)
	assert.Empty(t, mockAssignmentRepo.savedAssignments)
}

func TestAssignCapabilityToDomainHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockAssignmentRepo := &mockAssignmentRepository{}
	mockDomainReader := &mockBusinessDomainReadModelForAssign{}
	mockCapabilityReader := &mockCapabilityReadModelForAssign{}
	mockAssignmentReader := &mockAssignmentReadModelForAssign{}

	handler := newTestableAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockDomainReader,
		mockCapabilityReader,
		mockAssignmentReader,
	)

	invalidCmd := &commands.DeleteBusinessDomain{}

	err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestAssignCapabilityToDomainHandler_AssignmentReaderError_ReturnsError(t *testing.T) {
	businessDomainID := valueobjects.NewBusinessDomainID().Value()
	capabilityID := valueobjects.NewCapabilityID().Value()

	mockAssignmentRepo := &mockAssignmentRepository{}
	mockDomainReader := &mockBusinessDomainReadModelForAssign{
		domain: &readmodels.BusinessDomainDTO{ID: businessDomainID},
	}
	mockCapabilityReader := &mockCapabilityReadModelForAssign{
		capability: &readmodels.CapabilityDTO{ID: capabilityID, Level: "L1"},
	}
	mockAssignmentReader := &mockAssignmentReadModelForAssign{checkErr: errors.New("database error")}

	handler := newTestableAssignCapabilityToDomainHandler(
		mockAssignmentRepo,
		mockDomainReader,
		mockCapabilityReader,
		mockAssignmentReader,
	)

	cmd := &commands.AssignCapabilityToDomain{
		BusinessDomainID: businessDomainID,
		CapabilityID:     capabilityID,
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockAssignmentRepo.savedAssignments)
}
