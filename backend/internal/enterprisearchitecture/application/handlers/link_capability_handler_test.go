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

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestEnterpriseCapability(t *testing.T, name string) *aggregates.EnterpriseCapability {
	t.Helper()
	capName, err := valueobjects.NewEnterpriseCapabilityName(name)
	require.NoError(t, err)
	description, err := valueobjects.NewDescription("Test description")
	require.NoError(t, err)
	category, err := valueobjects.NewCategory("Test")
	require.NoError(t, err)

	capability, err := aggregates.NewEnterpriseCapability(capName, description, category)
	require.NoError(t, err)
	capability.MarkChangesAsCommitted()
	return capability
}

type mockLinkRepository struct {
	savedLinks []*aggregates.EnterpriseCapabilityLink
	saveErr    error
}

func (m *mockLinkRepository) Save(ctx context.Context, link *aggregates.EnterpriseCapabilityLink) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedLinks = append(m.savedLinks, link)
	return nil
}

type mockCapabilityRepository struct {
	existingCapability *aggregates.EnterpriseCapability
	getByIDErr         error
}

func (m *mockCapabilityRepository) GetByID(ctx context.Context, id string) (*aggregates.EnterpriseCapability, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	if m.existingCapability != nil && m.existingCapability.ID() == id {
		return m.existingCapability, nil
	}
	return nil, repositories.ErrEnterpriseCapabilityNotFound
}

type mockLinkReadModel struct {
	existingLink      *readmodels.EnterpriseCapabilityLinkDTO
	getByIDErr        error
	hierarchyConflict *readmodels.HierarchyConflict
	hierarchyErr      error
}

func (m *mockLinkReadModel) GetByDomainCapabilityID(ctx context.Context, domainCapabilityID string) (*readmodels.EnterpriseCapabilityLinkDTO, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.existingLink, nil
}

func (m *mockLinkReadModel) CheckHierarchyConflict(ctx context.Context, domainCapabilityID string, targetEnterpriseCapabilityID string) (*readmodels.HierarchyConflict, error) {
	if m.hierarchyErr != nil {
		return nil, m.hierarchyErr
	}
	return m.hierarchyConflict, nil
}

func linkCmd(enterpriseCapID, domainCapID string) *commands.LinkCapability {
	return &commands.LinkCapability{
		EnterpriseCapabilityID: enterpriseCapID,
		DomainCapabilityID:     domainCapID,
		LinkedBy:               "user@example.com",
	}
}

func TestLinkCapabilityHandler_SuccessfulLink(t *testing.T) {
	capability := createTestEnterpriseCapability(t, "Enterprise Capability")
	linkRepo := &mockLinkRepository{}
	handler := NewLinkCapabilityHandler(
		linkRepo,
		&mockCapabilityRepository{existingCapability: capability},
		&mockLinkReadModel{},
	)

	domainCapID := uuid.New().String()
	result, err := handler.Handle(context.Background(), linkCmd(capability.ID(), domainCapID))
	require.NoError(t, err)

	require.Len(t, linkRepo.savedLinks, 1)
	link := linkRepo.savedLinks[0]
	assert.Equal(t, capability.ID(), link.EnterpriseCapabilityID().Value())
	assert.Equal(t, domainCapID, link.DomainCapabilityID().Value())
	assert.Equal(t, "user@example.com", link.LinkedBy().Value())
	assert.NotEmpty(t, result.CreatedID)
	assert.Equal(t, link.ID(), result.CreatedID)
}

type errorTestCase struct {
	linkRepo    *mockLinkRepository
	capRepo     *mockCapabilityRepository
	readModel   *mockLinkReadModel
	domainCapID string
	expectedErr error
}

func (tc errorTestCase) run(t *testing.T) {
	t.Helper()
	domainCapID := tc.domainCapID
	if domainCapID == "" {
		domainCapID = uuid.New().String()
	}
	linkRepo := tc.linkRepo
	if linkRepo == nil {
		linkRepo = &mockLinkRepository{}
	}
	readModel := tc.readModel
	if readModel == nil {
		readModel = &mockLinkReadModel{}
	}
	enterpriseCapID := "non-existent-id"
	if tc.capRepo.existingCapability != nil {
		enterpriseCapID = tc.capRepo.existingCapability.ID()
	}
	handler := NewLinkCapabilityHandler(linkRepo, tc.capRepo, readModel)
	_, err := handler.Handle(context.Background(), linkCmd(enterpriseCapID, domainCapID))
	assert.Error(t, err)
	if tc.expectedErr != nil {
		assert.ErrorIs(t, err, tc.expectedErr)
	}
	assert.Empty(t, linkRepo.savedLinks)
}

func TestLinkCapabilityHandler_ValidationErrors(t *testing.T) {
	activeCapability := createTestEnterpriseCapability(t, "Active Capability")
	inactiveCapability := createTestEnterpriseCapability(t, "Inactive Capability")
	inactiveCapability.Delete()
	inactiveCapability.MarkChangesAsCommitted()

	t.Run("inactive capability", func(t *testing.T) {
		errorTestCase{capRepo: &mockCapabilityRepository{existingCapability: inactiveCapability}, expectedErr: aggregates.ErrCannotLinkInactiveCapability}.run(t)
	})
	t.Run("duplicate link", func(t *testing.T) {
		errorTestCase{capRepo: &mockCapabilityRepository{existingCapability: activeCapability}, readModel: &mockLinkReadModel{existingLink: &readmodels.EnterpriseCapabilityLinkDTO{ID: "existing"}}, expectedErr: ErrDomainCapabilityAlreadyLinked}.run(t)
	})
	t.Run("non-existent capability", func(t *testing.T) {
		errorTestCase{capRepo: &mockCapabilityRepository{getByIDErr: repositories.ErrEnterpriseCapabilityNotFound}, expectedErr: repositories.ErrEnterpriseCapabilityNotFound}.run(t)
	})
	t.Run("invalid domain capability ID", func(t *testing.T) {
		errorTestCase{capRepo: &mockCapabilityRepository{existingCapability: activeCapability}, domainCapID: "invalid-uuid"}.run(t)
	})
}

func TestLinkCapabilityHandler_InfrastructureErrors(t *testing.T) {
	activeCapability := createTestEnterpriseCapability(t, "Active Capability")

	t.Run("read model error", func(t *testing.T) {
		errorTestCase{capRepo: &mockCapabilityRepository{existingCapability: activeCapability}, readModel: &mockLinkReadModel{getByIDErr: errors.New("database error")}}.run(t)
	})
	t.Run("save error", func(t *testing.T) {
		errorTestCase{linkRepo: &mockLinkRepository{saveErr: errors.New("save error")}, capRepo: &mockCapabilityRepository{existingCapability: activeCapability}}.run(t)
	})
	t.Run("hierarchy check error", func(t *testing.T) {
		errorTestCase{capRepo: &mockCapabilityRepository{existingCapability: activeCapability}, readModel: &mockLinkReadModel{hierarchyErr: errors.New("database error")}}.run(t)
	})
}

func TestLinkCapabilityHandler_HierarchyConflicts(t *testing.T) {
	tests := []struct {
		name        string
		isAncestor  bool
		expectedErr error
	}{
		{"ancestor linked to different capability", true, ErrAncestorLinkedToDifferent},
		{"descendant linked to different capability", false, ErrDescendantLinkedToDifferent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capability := createTestEnterpriseCapability(t, "Enterprise Capability")
			linkRepo := &mockLinkRepository{}
			handler := NewLinkCapabilityHandler(
				linkRepo,
				&mockCapabilityRepository{existingCapability: capability},
				&mockLinkReadModel{
					hierarchyConflict: &readmodels.HierarchyConflict{
						ConflictingCapabilityID:   uuid.New().String(),
						ConflictingCapabilityName: "Conflicting Capability",
						LinkedToCapabilityID:      uuid.New().String(),
						LinkedToCapabilityName:    "Different Enterprise Capability",
						IsAncestor:                tt.isAncestor,
					},
				},
			)

			_, err := handler.Handle(context.Background(), linkCmd(capability.ID(), uuid.New().String()))
			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Empty(t, linkRepo.savedLinks)
		})
	}
}
