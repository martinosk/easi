package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEnterpriseCapabilityLinkRepositoryForUnlink struct {
	savedLinks   []*aggregates.EnterpriseCapabilityLink
	existingLink *aggregates.EnterpriseCapabilityLink
	saveErr      error
	getByIDErr   error
}

func (m *mockEnterpriseCapabilityLinkRepositoryForUnlink) Save(ctx context.Context, link *aggregates.EnterpriseCapabilityLink) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedLinks = append(m.savedLinks, link)
	return nil
}

func (m *mockEnterpriseCapabilityLinkRepositoryForUnlink) GetByID(ctx context.Context, id string) (*aggregates.EnterpriseCapabilityLink, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.existingLink != nil && m.existingLink.ID() == id {
		return m.existingLink, nil
	}
	return nil, repositories.ErrEnterpriseCapabilityLinkNotFound
}

type enterpriseCapabilityLinkRepositoryForUnlink interface {
	Save(ctx context.Context, link *aggregates.EnterpriseCapabilityLink) error
	GetByID(ctx context.Context, id string) (*aggregates.EnterpriseCapabilityLink, error)
}

type testableUnlinkCapabilityHandler struct {
	repository enterpriseCapabilityLinkRepositoryForUnlink
}

func newTestableUnlinkCapabilityHandler(
	repository enterpriseCapabilityLinkRepositoryForUnlink,
) *testableUnlinkCapabilityHandler {
	return &testableUnlinkCapabilityHandler{
		repository: repository,
	}
}

func (h *testableUnlinkCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.UnlinkCapability)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	link, err := h.repository.GetByID(ctx, command.LinkID)
	if err != nil {
		return err
	}

	if err := link.Unlink(); err != nil {
		return err
	}

	return h.repository.Save(ctx, link)
}

func createTestEnterpriseCapabilityLink(t *testing.T, enterpriseCapabilityName string) *aggregates.EnterpriseCapabilityLink {
	t.Helper()
	capability := createTestEnterpriseCapability(t, enterpriseCapabilityName)

	domainCapabilityID := valueobjects.NewDomainCapabilityID()
	linkedBy, _ := valueobjects.NewLinkedBy("user@example.com")

	link, err := aggregates.NewEnterpriseCapabilityLink(capability, domainCapabilityID, linkedBy)
	require.NoError(t, err)
	link.MarkChangesAsCommitted()
	return link
}

func TestUnlinkCapabilityHandler_UnlinksCapability(t *testing.T) {
	existingLink := createTestEnterpriseCapabilityLink(t, "Enterprise Capability")

	mockRepo := &mockEnterpriseCapabilityLinkRepositoryForUnlink{
		existingLink: existingLink,
	}

	handler := newTestableUnlinkCapabilityHandler(mockRepo)

	cmd := &commands.UnlinkCapability{
		LinkID: existingLink.ID(),
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedLinks, 1)
}

func TestUnlinkCapabilityHandler_NonExistentLink_ReturnsError(t *testing.T) {
	mockRepo := &mockEnterpriseCapabilityLinkRepositoryForUnlink{
		getByIDErr: repositories.ErrEnterpriseCapabilityLinkNotFound,
	}

	handler := newTestableUnlinkCapabilityHandler(mockRepo)

	cmd := &commands.UnlinkCapability{
		LinkID: "non-existent-link-id",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, repositories.ErrEnterpriseCapabilityLinkNotFound)
}

func TestUnlinkCapabilityHandler_RepositoryError_ReturnsError(t *testing.T) {
	existingLink := createTestEnterpriseCapabilityLink(t, "Enterprise Capability")

	mockRepo := &mockEnterpriseCapabilityLinkRepositoryForUnlink{
		existingLink: existingLink,
		saveErr:      errors.New("save error"),
	}

	handler := newTestableUnlinkCapabilityHandler(mockRepo)

	cmd := &commands.UnlinkCapability{
		LinkID: existingLink.ID(),
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}
