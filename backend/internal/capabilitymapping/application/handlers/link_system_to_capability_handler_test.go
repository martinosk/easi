package handlers

import (
	"context"
	"testing"

	archReadModels "easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockLinkSystemRealizationRepository struct {
	savedRealizations []*aggregates.CapabilityRealization
	saveErr           error
}

func (m *mockLinkSystemRealizationRepository) Save(ctx context.Context, realization *aggregates.CapabilityRealization) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedRealizations = append(m.savedRealizations, realization)
	return nil
}

type mockLinkSystemCapabilityRepository struct {
	capability *aggregates.Capability
	getByIDErr error
}

func (m *mockLinkSystemCapabilityRepository) GetByID(ctx context.Context, id string) (*aggregates.Capability, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.capability, nil
}

type mockLinkSystemComponentReadModel struct {
	component *archReadModels.ApplicationComponentDTO
	getErr    error
}

func (m *mockLinkSystemComponentReadModel) GetByID(ctx context.Context, id string) (*archReadModels.ApplicationComponentDTO, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.component, nil
}

func createTestCapabilityForLink(t *testing.T, level string, parentID string) *aggregates.Capability {
	t.Helper()

	name, err := valueobjects.NewCapabilityName("Test Capability")
	require.NoError(t, err)

	description := valueobjects.MustNewDescription("Test description")

	capLevel, err := valueobjects.NewCapabilityLevel(level)
	require.NoError(t, err)

	var parent valueobjects.CapabilityID
	if parentID != "" {
		parent, err = valueobjects.NewCapabilityIDFromString(parentID)
		require.NoError(t, err)
	}

	capability, err := aggregates.NewCapability(name, description, parent, capLevel)
	require.NoError(t, err)
	capability.MarkChangesAsCommitted()

	return capability
}

func TestLinkSystemToCapabilityHandler_CreatesRealization(t *testing.T) {
	l1Capability := createTestCapabilityForLink(t, "L1", "")
	l1CapabilityID := l1Capability.ID()

	componentID := valueobjects.NewCapabilityID().Value()

	mockCapRepo := &mockLinkSystemCapabilityRepository{
		capability: l1Capability,
	}
	mockRealRepo := &mockLinkSystemRealizationRepository{}
	mockCompReadModel := &mockLinkSystemComponentReadModel{
		component: &archReadModels.ApplicationComponentDTO{
			ID:   componentID,
			Name: "Test Component",
		},
	}

	handler := NewLinkSystemToCapabilityHandler(mockRealRepo, mockCapRepo, mockCompReadModel)

	cmd := &commands.LinkSystemToCapability{
		CapabilityID:     l1CapabilityID,
		ComponentID:      componentID,
		RealizationLevel: "Partial",
		Notes:            "Partially implements capability",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRealRepo.savedRealizations, 1, "Handler should create exactly 1 realization")

	realization := mockRealRepo.savedRealizations[0]
	assert.Equal(t, l1CapabilityID, realization.CapabilityID().Value())
	assert.Equal(t, componentID, realization.ComponentID().Value())
	assert.Equal(t, "Partial", realization.RealizationLevel().Value())
}

func TestLinkSystemToCapabilityHandler_ReturnsCreatedID(t *testing.T) {
	l1Capability := createTestCapabilityForLink(t, "L1", "")
	l1CapabilityID := l1Capability.ID()

	componentID := valueobjects.NewCapabilityID().Value()

	mockCapRepo := &mockLinkSystemCapabilityRepository{
		capability: l1Capability,
	}
	mockRealRepo := &mockLinkSystemRealizationRepository{}
	mockCompReadModel := &mockLinkSystemComponentReadModel{
		component: &archReadModels.ApplicationComponentDTO{
			ID:   componentID,
			Name: "Test Component",
		},
	}

	handler := NewLinkSystemToCapabilityHandler(mockRealRepo, mockCapRepo, mockCompReadModel)

	cmd := &commands.LinkSystemToCapability{
		CapabilityID:     l1CapabilityID,
		ComponentID:      componentID,
		RealizationLevel: "Full",
		Notes:            "",
	}

	result, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.NotEmpty(t, result.CreatedID, "Result CreatedID should be set after handling")
	assert.Equal(t, mockRealRepo.savedRealizations[0].ID(), result.CreatedID)
}

func TestLinkSystemToCapabilityHandler_ComponentNotFound_ReturnsError(t *testing.T) {
	l1Capability := createTestCapabilityForLink(t, "L1", "")
	l1CapabilityID := l1Capability.ID()

	componentID := valueobjects.NewCapabilityID().Value()

	mockCapRepo := &mockLinkSystemCapabilityRepository{
		capability: l1Capability,
	}
	mockRealRepo := &mockLinkSystemRealizationRepository{}
	mockCompReadModel := &mockLinkSystemComponentReadModel{
		component: nil,
	}

	handler := NewLinkSystemToCapabilityHandler(mockRealRepo, mockCapRepo, mockCompReadModel)

	cmd := &commands.LinkSystemToCapability{
		CapabilityID:     l1CapabilityID,
		ComponentID:      componentID,
		RealizationLevel: "Full",
		Notes:            "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrComponentNotFound)
}

func TestLinkSystemToCapabilityHandler_CapabilityNotFound_ReturnsError(t *testing.T) {
	componentID := valueobjects.NewCapabilityID().Value()
	capabilityID := valueobjects.NewCapabilityID().Value()

	mockCapRepo := &mockLinkSystemCapabilityRepository{
		getByIDErr: repositories.ErrCapabilityNotFound,
	}
	mockRealRepo := &mockLinkSystemRealizationRepository{}
	mockCompReadModel := &mockLinkSystemComponentReadModel{
		component: &archReadModels.ApplicationComponentDTO{
			ID:   componentID,
			Name: "Test Component",
		},
	}

	handler := NewLinkSystemToCapabilityHandler(mockRealRepo, mockCapRepo, mockCompReadModel)

	cmd := &commands.LinkSystemToCapability{
		CapabilityID:     capabilityID,
		ComponentID:      componentID,
		RealizationLevel: "Full",
		Notes:            "",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrCapabilityNotFoundForRealization)
}

func TestLinkSystemToCapabilityHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockCapRepo := &mockLinkSystemCapabilityRepository{}
	mockRealRepo := &mockLinkSystemRealizationRepository{}
	mockCompReadModel := &mockLinkSystemComponentReadModel{}

	handler := NewLinkSystemToCapabilityHandler(mockRealRepo, mockCapRepo, mockCompReadModel)

	invalidCmd := &commands.DeleteSystemRealization{}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}
