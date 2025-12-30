package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockBusinessDomainRepositoryForDelete struct {
	domain     *aggregates.BusinessDomain
	savedCount int
	getByIDErr error
	saveErr    error
}

func (m *mockBusinessDomainRepositoryForDelete) GetByID(ctx context.Context, id string) (*aggregates.BusinessDomain, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.domain, nil
}

func (m *mockBusinessDomainRepositoryForDelete) Save(ctx context.Context, domain *aggregates.BusinessDomain) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedCount++
	return nil
}

type mockAssignmentReadModelForDelete struct {
	assignments []readmodels.AssignmentDTO
	getErr      error
}

func (m *mockAssignmentReadModelForDelete) GetByDomainID(ctx context.Context, domainID string) ([]readmodels.AssignmentDTO, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.assignments, nil
}

type businessDomainRepositoryForDelete interface {
	GetByID(ctx context.Context, id string) (*aggregates.BusinessDomain, error)
	Save(ctx context.Context, domain *aggregates.BusinessDomain) error
}

type assignmentReadModelForDelete interface {
	GetByDomainID(ctx context.Context, domainID string) ([]readmodels.AssignmentDTO, error)
}

type testableDeleteBusinessDomainHandler struct {
	repository       businessDomainRepositoryForDelete
	assignmentReader assignmentReadModelForDelete
}

func newTestableDeleteBusinessDomainHandler(
	repository businessDomainRepositoryForDelete,
	assignmentReader assignmentReadModelForDelete,
) *testableDeleteBusinessDomainHandler {
	return &testableDeleteBusinessDomainHandler{
		repository:       repository,
		assignmentReader: assignmentReader,
	}
}

func (h *testableDeleteBusinessDomainHandler) Handle(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	command, ok := cmd.(*commands.DeleteBusinessDomain)
	if !ok {
		return cqrs.EmptyResult(), cqrs.ErrInvalidCommand
	}

	assignments, err := h.assignmentReader.GetByDomainID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if len(assignments) > 0 {
		return cqrs.EmptyResult(), ErrBusinessDomainHasAssignments
	}

	domain, err := h.repository.GetByID(ctx, command.ID)
	if err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := domain.Delete(); err != nil {
		return cqrs.EmptyResult(), err
	}

	if err := h.repository.Save(ctx, domain); err != nil {
		return cqrs.EmptyResult(), err
	}
	return cqrs.EmptyResult(), nil
}

func TestDeleteBusinessDomainHandler_DeletesBusinessDomain(t *testing.T) {
	domain := createTestBusinessDomain(t, "Test Domain", "Description")
	domainID := domain.ID()

	mockRepo := &mockBusinessDomainRepositoryForDelete{domain: domain}
	mockAssignmentReader := &mockAssignmentReadModelForDelete{assignments: []readmodels.AssignmentDTO{}}

	handler := newTestableDeleteBusinessDomainHandler(mockRepo, mockAssignmentReader)

	cmd := &commands.DeleteBusinessDomain{
		ID: domainID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.Equal(t, 1, mockRepo.savedCount, "Handler should save domain once")
}

func TestDeleteBusinessDomainHandler_DomainHasAssignments_ReturnsError(t *testing.T) {
	domain := createTestBusinessDomain(t, "Test Domain", "Description")
	domainID := domain.ID()

	mockRepo := &mockBusinessDomainRepositoryForDelete{domain: domain}
	mockAssignmentReader := &mockAssignmentReadModelForDelete{
		assignments: []readmodels.AssignmentDTO{
			{AssignmentID: "assignment-1"},
		},
	}

	handler := newTestableDeleteBusinessDomainHandler(mockRepo, mockAssignmentReader)

	cmd := &commands.DeleteBusinessDomain{
		ID: domainID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrBusinessDomainHasAssignments)
	assert.Equal(t, 0, mockRepo.savedCount, "Should not save when domain has assignments")
}

func TestDeleteBusinessDomainHandler_DomainNotFound_ReturnsError(t *testing.T) {
	mockRepo := &mockBusinessDomainRepositoryForDelete{
		getByIDErr: repositories.ErrBusinessDomainNotFound,
	}
	mockAssignmentReader := &mockAssignmentReadModelForDelete{assignments: []readmodels.AssignmentDTO{}}

	handler := newTestableDeleteBusinessDomainHandler(mockRepo, mockAssignmentReader)

	cmd := &commands.DeleteBusinessDomain{
		ID: "non-existent",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, repositories.ErrBusinessDomainNotFound)
}

func TestDeleteBusinessDomainHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockBusinessDomainRepositoryForDelete{}
	mockAssignmentReader := &mockAssignmentReadModelForDelete{}

	handler := newTestableDeleteBusinessDomainHandler(mockRepo, mockAssignmentReader)

	invalidCmd := &commands.CreateBusinessDomain{}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestDeleteBusinessDomainHandler_AssignmentReaderError_ReturnsError(t *testing.T) {
	domain := createTestBusinessDomain(t, "Test Domain", "Description")
	domainID := domain.ID()

	mockRepo := &mockBusinessDomainRepositoryForDelete{domain: domain}
	mockAssignmentReader := &mockAssignmentReadModelForDelete{
		getErr: errors.New("database error"),
	}

	handler := newTestableDeleteBusinessDomainHandler(mockRepo, mockAssignmentReader)

	cmd := &commands.DeleteBusinessDomain{
		ID: domainID,
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Equal(t, 0, mockRepo.savedCount)
}
