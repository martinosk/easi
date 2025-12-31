package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/enterprisearchitecture/application/commands"
	"easi/backend/internal/enterprisearchitecture/domain/aggregates"
	"easi/backend/internal/enterprisearchitecture/domain/valueobjects"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUnlinkRepository struct {
	savedLinks   []*aggregates.EnterpriseCapabilityLink
	existingLink *aggregates.EnterpriseCapabilityLink
	saveErr      error
	getByIDErr   error
}

func (m *mockUnlinkRepository) Save(ctx context.Context, link *aggregates.EnterpriseCapabilityLink) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedLinks = append(m.savedLinks, link)
	return nil
}

func (m *mockUnlinkRepository) GetByID(ctx context.Context, id string) (*aggregates.EnterpriseCapabilityLink, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.existingLink != nil && m.existingLink.ID() == id {
		return m.existingLink, nil
	}
	return nil, repositories.ErrEnterpriseCapabilityLinkNotFound
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

	mockRepo := &mockUnlinkRepository{existingLink: existingLink}

	handler := NewUnlinkCapabilityHandler(mockRepo)

	cmd := &commands.UnlinkCapability{
		LinkID: existingLink.ID(),
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedLinks, 1)
}

func TestUnlinkCapabilityHandler_NonExistentLink_ReturnsError(t *testing.T) {
	mockRepo := &mockUnlinkRepository{getByIDErr: repositories.ErrEnterpriseCapabilityLinkNotFound}

	handler := NewUnlinkCapabilityHandler(mockRepo)

	cmd := &commands.UnlinkCapability{
		LinkID: "non-existent-link-id",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, repositories.ErrEnterpriseCapabilityLinkNotFound)
}

func TestUnlinkCapabilityHandler_RepositoryError_ReturnsError(t *testing.T) {
	existingLink := createTestEnterpriseCapabilityLink(t, "Enterprise Capability")

	mockRepo := &mockUnlinkRepository{
		existingLink: existingLink,
		saveErr:      errors.New("save error"),
	}

	handler := NewUnlinkCapabilityHandler(mockRepo)

	cmd := &commands.UnlinkCapability{
		LinkID: existingLink.ID(),
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}
