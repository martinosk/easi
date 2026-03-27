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

func createCapability(t *testing.T, level string) *aggregates.Capability {
	t.Helper()

	name, err := valueobjects.NewCapabilityName("Test " + level + " Capability")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Test description")

	capLevel, err := valueobjects.NewCapabilityLevel(level)
	require.NoError(t, err)

	var parentID valueobjects.CapabilityID
	if level != "L1" {
		parentID = valueobjects.NewCapabilityID()
	}

	capability, err := aggregates.NewCapability(name, description, parentID, capLevel)
	require.NoError(t, err)
	capability.MarkChangesAsCommitted()

	return capability
}

type assignFixture struct {
	assignmentRepo   *mockAssignCapabilityAssignmentRepository
	capabilityRepo   *mockAssignCapabilityCapabilityRepository
	domainReader     *mockAssignCapabilityDomainReader
	assignmentReader *mockAssignCapabilityAssignmentReader
	handler          *AssignCapabilityToDomainHandler
	businessDomainID string
	capability       *aggregates.Capability
}

func setupAssignFixture(t *testing.T) *assignFixture {
	t.Helper()

	businessDomainID := valueobjects.NewBusinessDomainID().Value()
	capability := createCapability(t, "L1")

	assignmentRepo := &mockAssignCapabilityAssignmentRepository{}
	capabilityRepo := &mockAssignCapabilityCapabilityRepository{capability: capability}
	domainReader := &mockAssignCapabilityDomainReader{
		domain: &readmodels.BusinessDomainDTO{ID: businessDomainID, Name: "Test Domain"},
	}
	assignmentReader := &mockAssignCapabilityAssignmentReader{exists: false}

	handler := NewAssignCapabilityToDomainHandler(
		assignmentRepo,
		capabilityRepo,
		domainReader,
		assignmentReader,
	)

	return &assignFixture{
		assignmentRepo:   assignmentRepo,
		capabilityRepo:   capabilityRepo,
		domainReader:     domainReader,
		assignmentReader: assignmentReader,
		handler:          handler,
		businessDomainID: businessDomainID,
		capability:       capability,
	}
}

func (f *assignFixture) newCommand() *commands.AssignCapabilityToDomain {
	return &commands.AssignCapabilityToDomain{
		BusinessDomainID: f.businessDomainID,
		CapabilityID:     f.capability.ID(),
	}
}

func (f *assignFixture) handle(t *testing.T) (cqrs.CommandResult, error) {
	t.Helper()
	return f.handler.Handle(context.Background(), f.newCommand())
}

func TestAssignCapabilityToDomainHandler_CreatesAssignment(t *testing.T) {
	f := setupAssignFixture(t)

	result, err := f.handle(t)
	require.NoError(t, err)

	require.Len(t, f.assignmentRepo.savedAssignments, 1, "Handler should create exactly 1 assignment")
	assert.NotEmpty(t, result.CreatedID, "Result CreatedID should be set")
}

func TestAssignCapabilityToDomainHandler_BusinessDomainNotFound_ReturnsError(t *testing.T) {
	f := setupAssignFixture(t)
	f.domainReader.domain = nil

	_, err := f.handle(t)
	assert.ErrorIs(t, err, ErrBusinessDomainNotFound)
	assert.Empty(t, f.assignmentRepo.savedAssignments)
}

func TestAssignCapabilityToDomainHandler_CapabilityNotFound_ReturnsError(t *testing.T) {
	f := setupAssignFixture(t)
	f.capabilityRepo.capability = nil
	f.capabilityRepo.getErr = errors.New("capability not found")

	_, err := f.handle(t)
	assert.Error(t, err)
	assert.Empty(t, f.assignmentRepo.savedAssignments)
}

func TestAssignCapabilityToDomainHandler_CapabilityNotL1_ReturnsError(t *testing.T) {
	f := setupAssignFixture(t)
	f.capability = createCapability(t, "L2")
	f.capabilityRepo.capability = f.capability

	_, err := f.handle(t)
	assert.ErrorIs(t, err, aggregates.ErrOnlyL1CanBeAssignedToDomain)
	assert.Empty(t, f.assignmentRepo.savedAssignments)
}

func TestAssignCapabilityToDomainHandler_AssignmentExists_ReturnsError(t *testing.T) {
	f := setupAssignFixture(t)
	f.assignmentReader.exists = true

	_, err := f.handle(t)
	assert.ErrorIs(t, err, ErrAssignmentAlreadyExists)
	assert.Empty(t, f.assignmentRepo.savedAssignments)
}

func TestAssignCapabilityToDomainHandler_InvalidCommand_ReturnsError(t *testing.T) {
	f := setupAssignFixture(t)

	_, err := f.handler.Handle(context.Background(), &commands.DeleteBusinessDomain{})
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestAssignCapabilityToDomainHandler_AssignmentReaderError_ReturnsError(t *testing.T) {
	f := setupAssignFixture(t)
	f.assignmentReader.checkErr = errors.New("database error")

	_, err := f.handle(t)
	assert.Error(t, err)
	assert.Empty(t, f.assignmentRepo.savedAssignments)
}
