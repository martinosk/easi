package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEnterpriseCapabilityLinkRepository struct {
	savedLinks []*aggregates.EnterpriseCapabilityLink
	saveErr    error
}

func (m *mockEnterpriseCapabilityLinkRepository) Save(ctx context.Context, link *aggregates.EnterpriseCapabilityLink) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedLinks = append(m.savedLinks, link)
	return nil
}

type mockEnterpriseCapabilityRepositoryForLink struct {
	existingCapability *aggregates.EnterpriseCapability
	getByIDErr         error
}

func (m *mockEnterpriseCapabilityRepositoryForLink) GetByID(ctx context.Context, id string) (*aggregates.EnterpriseCapability, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.existingCapability != nil && m.existingCapability.ID() == id {
		return m.existingCapability, nil
	}
	return nil, repositories.ErrEnterpriseCapabilityNotFound
}

type mockEnterpriseCapabilityLinkReadModel struct {
	existingLink *readmodels.EnterpriseCapabilityLinkDTO
	getByIDErr   error
}

func (m *mockEnterpriseCapabilityLinkReadModel) GetByDomainCapabilityID(ctx context.Context, domainCapabilityID string) (*readmodels.EnterpriseCapabilityLinkDTO, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.existingLink, nil
}

type enterpriseCapabilityLinkRepository interface {
	Save(ctx context.Context, link *aggregates.EnterpriseCapabilityLink) error
}

type enterpriseCapabilityRepositoryForLink interface {
	GetByID(ctx context.Context, id string) (*aggregates.EnterpriseCapability, error)
}

type enterpriseCapabilityLinkReadModelForLink interface {
	GetByDomainCapabilityID(ctx context.Context, domainCapabilityID string) (*readmodels.EnterpriseCapabilityLinkDTO, error)
}

type testableLinkCapabilityHandler struct {
	linkRepository       enterpriseCapabilityLinkRepository
	capabilityRepository enterpriseCapabilityRepositoryForLink
	linkReadModel        enterpriseCapabilityLinkReadModelForLink
}

func newTestableLinkCapabilityHandler(
	linkRepository enterpriseCapabilityLinkRepository,
	capabilityRepository enterpriseCapabilityRepositoryForLink,
	linkReadModel enterpriseCapabilityLinkReadModelForLink,
) *testableLinkCapabilityHandler {
	return &testableLinkCapabilityHandler{
		linkRepository:       linkRepository,
		capabilityRepository: capabilityRepository,
		linkReadModel:        linkReadModel,
	}
}

func (h *testableLinkCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.LinkCapability)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	capability, err := h.capabilityRepository.GetByID(ctx, command.EnterpriseCapabilityID)
	if err != nil {
		return err
	}

	existingLink, err := h.linkReadModel.GetByDomainCapabilityID(ctx, command.DomainCapabilityID)
	if err != nil {
		return err
	}
	if existingLink != nil {
		return ErrDomainCapabilityAlreadyLinked
	}

	domainCapabilityID, err := valueobjects.NewDomainCapabilityIDFromString(command.DomainCapabilityID)
	if err != nil {
		return err
	}

	linkedBy, err := valueobjects.NewLinkedBy(command.LinkedBy)
	if err != nil {
		return err
	}

	link, err := aggregates.NewEnterpriseCapabilityLink(capability, domainCapabilityID, linkedBy)
	if err != nil {
		return err
	}

	command.ID = link.ID()

	return h.linkRepository.Save(ctx, link)
}

func TestLinkCapabilityHandler_LinksCapability(t *testing.T) {
	existingCapability := createTestEnterpriseCapability(t, "Enterprise Capability")

	mockLinkRepo := &mockEnterpriseCapabilityLinkRepository{}
	mockCapabilityRepo := &mockEnterpriseCapabilityRepositoryForLink{existingCapability: existingCapability}
	mockLinkReadModel := &mockEnterpriseCapabilityLinkReadModel{existingLink: nil}

	handler := newTestableLinkCapabilityHandler(mockLinkRepo, mockCapabilityRepo, mockLinkReadModel)

	domainCapabilityID := uuid.New().String()
	cmd := &commands.LinkCapability{
		EnterpriseCapabilityID: existingCapability.ID(),
		DomainCapabilityID:     domainCapabilityID,
		LinkedBy:               "user@example.com",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockLinkRepo.savedLinks, 1)
	link := mockLinkRepo.savedLinks[0]
	assert.Equal(t, existingCapability.ID(), link.EnterpriseCapabilityID().Value())
	assert.Equal(t, domainCapabilityID, link.DomainCapabilityID().Value())
	assert.Equal(t, "user@example.com", link.LinkedBy().Value())
}

func TestLinkCapabilityHandler_SetsCommandID(t *testing.T) {
	existingCapability := createTestEnterpriseCapability(t, "Enterprise Capability")

	mockLinkRepo := &mockEnterpriseCapabilityLinkRepository{}
	mockCapabilityRepo := &mockEnterpriseCapabilityRepositoryForLink{existingCapability: existingCapability}
	mockLinkReadModel := &mockEnterpriseCapabilityLinkReadModel{existingLink: nil}

	handler := newTestableLinkCapabilityHandler(mockLinkRepo, mockCapabilityRepo, mockLinkReadModel)

	domainCapabilityID := uuid.New().String()
	cmd := &commands.LinkCapability{
		EnterpriseCapabilityID: existingCapability.ID(),
		DomainCapabilityID:     domainCapabilityID,
		LinkedBy:               "user@example.com",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.NotEmpty(t, cmd.ID)
	assert.Equal(t, mockLinkRepo.savedLinks[0].ID(), cmd.ID)
}

func TestLinkCapabilityHandler_InactiveCapability_ReturnsError(t *testing.T) {
	existingCapability := createTestEnterpriseCapability(t, "Enterprise Capability")
	existingCapability.Delete()
	existingCapability.MarkChangesAsCommitted()

	mockLinkRepo := &mockEnterpriseCapabilityLinkRepository{}
	mockCapabilityRepo := &mockEnterpriseCapabilityRepositoryForLink{existingCapability: existingCapability}
	mockLinkReadModel := &mockEnterpriseCapabilityLinkReadModel{existingLink: nil}

	handler := newTestableLinkCapabilityHandler(mockLinkRepo, mockCapabilityRepo, mockLinkReadModel)

	domainCapabilityID := uuid.New().String()
	cmd := &commands.LinkCapability{
		EnterpriseCapabilityID: existingCapability.ID(),
		DomainCapabilityID:     domainCapabilityID,
		LinkedBy:               "user@example.com",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, aggregates.ErrCannotLinkInactiveCapability)
	assert.Empty(t, mockLinkRepo.savedLinks)
}

func TestLinkCapabilityHandler_DuplicateLink_ReturnsError(t *testing.T) {
	existingCapability := createTestEnterpriseCapability(t, "Enterprise Capability")

	mockLinkRepo := &mockEnterpriseCapabilityLinkRepository{}
	mockCapabilityRepo := &mockEnterpriseCapabilityRepositoryForLink{existingCapability: existingCapability}
	mockLinkReadModel := &mockEnterpriseCapabilityLinkReadModel{
		existingLink: &readmodels.EnterpriseCapabilityLinkDTO{ID: "existing-link-id"},
	}

	handler := newTestableLinkCapabilityHandler(mockLinkRepo, mockCapabilityRepo, mockLinkReadModel)

	domainCapabilityID := uuid.New().String()
	cmd := &commands.LinkCapability{
		EnterpriseCapabilityID: existingCapability.ID(),
		DomainCapabilityID:     domainCapabilityID,
		LinkedBy:               "user@example.com",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrDomainCapabilityAlreadyLinked)
	assert.Empty(t, mockLinkRepo.savedLinks)
}

func TestLinkCapabilityHandler_NonExistentCapability_ReturnsError(t *testing.T) {
	mockLinkRepo := &mockEnterpriseCapabilityLinkRepository{}
	mockCapabilityRepo := &mockEnterpriseCapabilityRepositoryForLink{
		getByIDErr: repositories.ErrEnterpriseCapabilityNotFound,
	}
	mockLinkReadModel := &mockEnterpriseCapabilityLinkReadModel{}

	handler := newTestableLinkCapabilityHandler(mockLinkRepo, mockCapabilityRepo, mockLinkReadModel)

	domainCapabilityID := uuid.New().String()
	cmd := &commands.LinkCapability{
		EnterpriseCapabilityID: "non-existent-id",
		DomainCapabilityID:     domainCapabilityID,
		LinkedBy:               "user@example.com",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, repositories.ErrEnterpriseCapabilityNotFound)
}

func TestLinkCapabilityHandler_InvalidDomainCapabilityID_ReturnsError(t *testing.T) {
	existingCapability := createTestEnterpriseCapability(t, "Enterprise Capability")

	mockLinkRepo := &mockEnterpriseCapabilityLinkRepository{}
	mockCapabilityRepo := &mockEnterpriseCapabilityRepositoryForLink{existingCapability: existingCapability}
	mockLinkReadModel := &mockEnterpriseCapabilityLinkReadModel{existingLink: nil}

	handler := newTestableLinkCapabilityHandler(mockLinkRepo, mockCapabilityRepo, mockLinkReadModel)

	cmd := &commands.LinkCapability{
		EnterpriseCapabilityID: existingCapability.ID(),
		DomainCapabilityID:     "invalid-uuid",
		LinkedBy:               "user@example.com",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockLinkRepo.savedLinks)
}

func TestLinkCapabilityHandler_ReadModelError_ReturnsError(t *testing.T) {
	existingCapability := createTestEnterpriseCapability(t, "Enterprise Capability")

	mockLinkRepo := &mockEnterpriseCapabilityLinkRepository{}
	mockCapabilityRepo := &mockEnterpriseCapabilityRepositoryForLink{existingCapability: existingCapability}
	mockLinkReadModel := &mockEnterpriseCapabilityLinkReadModel{getByIDErr: errors.New("database error")}

	handler := newTestableLinkCapabilityHandler(mockLinkRepo, mockCapabilityRepo, mockLinkReadModel)

	domainCapabilityID := uuid.New().String()
	cmd := &commands.LinkCapability{
		EnterpriseCapabilityID: existingCapability.ID(),
		DomainCapabilityID:     domainCapabilityID,
		LinkedBy:               "user@example.com",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockLinkRepo.savedLinks)
}

func TestLinkCapabilityHandler_RepositoryError_ReturnsError(t *testing.T) {
	existingCapability := createTestEnterpriseCapability(t, "Enterprise Capability")

	mockLinkRepo := &mockEnterpriseCapabilityLinkRepository{saveErr: errors.New("save error")}
	mockCapabilityRepo := &mockEnterpriseCapabilityRepositoryForLink{existingCapability: existingCapability}
	mockLinkReadModel := &mockEnterpriseCapabilityLinkReadModel{existingLink: nil}

	handler := newTestableLinkCapabilityHandler(mockLinkRepo, mockCapabilityRepo, mockLinkReadModel)

	domainCapabilityID := uuid.New().String()
	cmd := &commands.LinkCapability{
		EnterpriseCapabilityID: existingCapability.ID(),
		DomainCapabilityID:     domainCapabilityID,
		LinkedBy:               "user@example.com",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}
