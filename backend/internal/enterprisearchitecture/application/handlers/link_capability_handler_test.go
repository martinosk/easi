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

func TestLinkCapabilityHandler_LinksCapability(t *testing.T) {
	existingCapability := createTestEnterpriseCapability(t, "Enterprise Capability")

	mockLinkRepo := &mockLinkRepository{}
	mockCapabilityRepo := &mockCapabilityRepository{existingCapability: existingCapability}
	mockReadModel := &mockLinkReadModel{existingLink: nil}

	handler := NewLinkCapabilityHandler(mockLinkRepo, mockCapabilityRepo, mockReadModel)

	domainCapabilityID := uuid.New().String()
	cmd := &commands.LinkCapability{
		EnterpriseCapabilityID: existingCapability.ID(),
		DomainCapabilityID:     domainCapabilityID,
		LinkedBy:               "user@example.com",
	}

	result, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockLinkRepo.savedLinks, 1)
	link := mockLinkRepo.savedLinks[0]
	assert.Equal(t, existingCapability.ID(), link.EnterpriseCapabilityID().Value())
	assert.Equal(t, domainCapabilityID, link.DomainCapabilityID().Value())
	assert.Equal(t, "user@example.com", link.LinkedBy().Value())
	assert.Equal(t, link.ID(), result.CreatedID)
}

func TestLinkCapabilityHandler_ReturnsCreatedID(t *testing.T) {
	existingCapability := createTestEnterpriseCapability(t, "Enterprise Capability")

	mockLinkRepo := &mockLinkRepository{}
	mockCapabilityRepo := &mockCapabilityRepository{existingCapability: existingCapability}
	mockReadModel := &mockLinkReadModel{existingLink: nil}

	handler := NewLinkCapabilityHandler(mockLinkRepo, mockCapabilityRepo, mockReadModel)

	domainCapabilityID := uuid.New().String()
	cmd := &commands.LinkCapability{
		EnterpriseCapabilityID: existingCapability.ID(),
		DomainCapabilityID:     domainCapabilityID,
		LinkedBy:               "user@example.com",
	}

	result, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.NotEmpty(t, result.CreatedID)
	assert.Equal(t, mockLinkRepo.savedLinks[0].ID(), result.CreatedID)
}

func TestLinkCapabilityHandler_InactiveCapability_ReturnsError(t *testing.T) {
	existingCapability := createTestEnterpriseCapability(t, "Enterprise Capability")
	existingCapability.Delete()
	existingCapability.MarkChangesAsCommitted()

	mockLinkRepo := &mockLinkRepository{}
	mockCapabilityRepo := &mockCapabilityRepository{existingCapability: existingCapability}
	mockReadModel := &mockLinkReadModel{existingLink: nil}

	handler := NewLinkCapabilityHandler(mockLinkRepo, mockCapabilityRepo, mockReadModel)

	domainCapabilityID := uuid.New().String()
	cmd := &commands.LinkCapability{
		EnterpriseCapabilityID: existingCapability.ID(),
		DomainCapabilityID:     domainCapabilityID,
		LinkedBy:               "user@example.com",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, aggregates.ErrCannotLinkInactiveCapability)
	assert.Empty(t, mockLinkRepo.savedLinks)
}

func TestLinkCapabilityHandler_ErrorCases(t *testing.T) {
	tests := []struct {
		name               string
		setupCapability    func(t *testing.T) *aggregates.EnterpriseCapability
		linkRepoConfig     *mockLinkRepository
		capRepoConfig      func(*aggregates.EnterpriseCapability) *mockCapabilityRepository
		readModelConfig    *mockLinkReadModel
		domainCapabilityID string
		expectedErr        error
		checkErrorIs       bool
	}{
		{
			name:            "duplicate link returns error",
			setupCapability: func(t *testing.T) *aggregates.EnterpriseCapability { return createTestEnterpriseCapability(t, "Enterprise Capability") },
			linkRepoConfig:  &mockLinkRepository{},
			capRepoConfig:   func(cap *aggregates.EnterpriseCapability) *mockCapabilityRepository { return &mockCapabilityRepository{existingCapability: cap} },
			readModelConfig: &mockLinkReadModel{existingLink: &readmodels.EnterpriseCapabilityLinkDTO{ID: "existing-link-id"}},
			expectedErr:     ErrDomainCapabilityAlreadyLinked,
			checkErrorIs:    true,
		},
		{
			name:            "non-existent capability returns error",
			setupCapability: nil,
			linkRepoConfig:  &mockLinkRepository{},
			capRepoConfig:   func(_ *aggregates.EnterpriseCapability) *mockCapabilityRepository { return &mockCapabilityRepository{getByIDErr: repositories.ErrEnterpriseCapabilityNotFound} },
			readModelConfig: &mockLinkReadModel{},
			expectedErr:     repositories.ErrEnterpriseCapabilityNotFound,
			checkErrorIs:    true,
		},
		{
			name:               "invalid domain capability ID returns error",
			setupCapability:    func(t *testing.T) *aggregates.EnterpriseCapability { return createTestEnterpriseCapability(t, "Enterprise Capability") },
			linkRepoConfig:     &mockLinkRepository{},
			capRepoConfig:      func(cap *aggregates.EnterpriseCapability) *mockCapabilityRepository { return &mockCapabilityRepository{existingCapability: cap} },
			readModelConfig:    &mockLinkReadModel{existingLink: nil},
			domainCapabilityID: "invalid-uuid",
			checkErrorIs:       false,
		},
		{
			name:            "read model error returns error",
			setupCapability: func(t *testing.T) *aggregates.EnterpriseCapability { return createTestEnterpriseCapability(t, "Enterprise Capability") },
			linkRepoConfig:  &mockLinkRepository{},
			capRepoConfig:   func(cap *aggregates.EnterpriseCapability) *mockCapabilityRepository { return &mockCapabilityRepository{existingCapability: cap} },
			readModelConfig: &mockLinkReadModel{getByIDErr: errors.New("database error")},
			checkErrorIs:    false,
		},
		{
			name:            "repository save error returns error",
			setupCapability: func(t *testing.T) *aggregates.EnterpriseCapability { return createTestEnterpriseCapability(t, "Enterprise Capability") },
			linkRepoConfig:  &mockLinkRepository{saveErr: errors.New("save error")},
			capRepoConfig:   func(cap *aggregates.EnterpriseCapability) *mockCapabilityRepository { return &mockCapabilityRepository{existingCapability: cap} },
			readModelConfig: &mockLinkReadModel{existingLink: nil},
			checkErrorIs:    false,
		},
		{
			name:            "hierarchy check error returns error",
			setupCapability: func(t *testing.T) *aggregates.EnterpriseCapability { return createTestEnterpriseCapability(t, "Enterprise Capability") },
			linkRepoConfig:  &mockLinkRepository{},
			capRepoConfig:   func(cap *aggregates.EnterpriseCapability) *mockCapabilityRepository { return &mockCapabilityRepository{existingCapability: cap} },
			readModelConfig: &mockLinkReadModel{existingLink: nil, hierarchyErr: errors.New("database error")},
			checkErrorIs:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capability *aggregates.EnterpriseCapability
			if tt.setupCapability != nil {
				capability = tt.setupCapability(t)
			}

			domainCapabilityID := tt.domainCapabilityID
			if domainCapabilityID == "" {
				domainCapabilityID = uuid.New().String()
			}

			enterpriseCapID := "non-existent-id"
			if capability != nil {
				enterpriseCapID = capability.ID()
			}

			handler := NewLinkCapabilityHandler(tt.linkRepoConfig, tt.capRepoConfig(capability), tt.readModelConfig)

			cmd := &commands.LinkCapability{
				EnterpriseCapabilityID: enterpriseCapID,
				DomainCapabilityID:     domainCapabilityID,
				LinkedBy:               "user@example.com",
			}

			_, err := handler.Handle(context.Background(), cmd)
			assert.Error(t, err)
			if tt.checkErrorIs && tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			}
			assert.Empty(t, tt.linkRepoConfig.savedLinks)
		})
	}
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
			existingCapability := createTestEnterpriseCapability(t, "Enterprise Capability")

			mockLinkRepo := &mockLinkRepository{}
			mockCapabilityRepo := &mockCapabilityRepository{existingCapability: existingCapability}
			mockReadModel := &mockLinkReadModel{
				existingLink: nil,
				hierarchyConflict: &readmodels.HierarchyConflict{
					ConflictingCapabilityID:   uuid.New().String(),
					ConflictingCapabilityName: "Conflicting Capability",
					LinkedToCapabilityID:      uuid.New().String(),
					LinkedToCapabilityName:    "Different Enterprise Capability",
					IsAncestor:                tt.isAncestor,
				},
			}

			handler := NewLinkCapabilityHandler(mockLinkRepo, mockCapabilityRepo, mockReadModel)

			cmd := &commands.LinkCapability{
				EnterpriseCapabilityID: existingCapability.ID(),
				DomainCapabilityID:     uuid.New().String(),
				LinkedBy:               "user@example.com",
			}

			_, err := handler.Handle(context.Background(), cmd)
			assert.ErrorIs(t, err, tt.expectedErr)
			assert.Empty(t, mockLinkRepo.savedLinks)
		})
	}
}

func TestLinkCapabilityHandler_NoHierarchyConflict_Succeeds(t *testing.T) {
	existingCapability := createTestEnterpriseCapability(t, "Enterprise Capability")

	mockLinkRepo := &mockLinkRepository{}
	mockCapabilityRepo := &mockCapabilityRepository{existingCapability: existingCapability}
	mockReadModel := &mockLinkReadModel{existingLink: nil, hierarchyConflict: nil}

	handler := NewLinkCapabilityHandler(mockLinkRepo, mockCapabilityRepo, mockReadModel)

	domainCapabilityID := uuid.New().String()
	cmd := &commands.LinkCapability{
		EnterpriseCapabilityID: existingCapability.ID(),
		DomainCapabilityID:     domainCapabilityID,
		LinkedBy:               "user@example.com",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockLinkRepo.savedLinks, 1)
	link := mockLinkRepo.savedLinks[0]
	assert.Equal(t, existingCapability.ID(), link.EnterpriseCapabilityID().Value())
	assert.Equal(t, domainCapabilityID, link.DomainCapabilityID().Value())
}
