package handlers

import (
	"context"
	"testing"

	archReadModels "easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRealizationRepository struct {
	savedRealizations []*aggregates.CapabilityRealization
	saveErr           error
}

func (m *mockRealizationRepository) Save(ctx context.Context, realization *aggregates.CapabilityRealization) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedRealizations = append(m.savedRealizations, realization)
	return nil
}

type mockCapabilityRepositoryForLink struct {
	capability *aggregates.Capability
	getByIDErr error
}

func (m *mockCapabilityRepositoryForLink) GetByID(ctx context.Context, id string) (*aggregates.Capability, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.capability, nil
}

type mockComponentReadModel struct {
	component *archReadModels.ApplicationComponentDTO
	getErr    error
}

func (m *mockComponentReadModel) GetByID(ctx context.Context, id string) (*archReadModels.ApplicationComponentDTO, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.component, nil
}

type capabilityRepositoryForLink interface {
	GetByID(ctx context.Context, id string) (*aggregates.Capability, error)
}

type realizationRepository interface {
	Save(ctx context.Context, realization *aggregates.CapabilityRealization) error
}

type componentReadModelForLink interface {
	GetByID(ctx context.Context, id string) (*archReadModels.ApplicationComponentDTO, error)
}

type testableLinkSystemToCapabilityHandler struct {
	realizationRepository realizationRepository
	capabilityRepository  capabilityRepositoryForLink
	componentReadModel    componentReadModelForLink
}

func newTestableLinkSystemToCapabilityHandler(
	realizationRepository realizationRepository,
	capabilityRepository capabilityRepositoryForLink,
	componentReadModel componentReadModelForLink,
) *testableLinkSystemToCapabilityHandler {
	return &testableLinkSystemToCapabilityHandler{
		realizationRepository: realizationRepository,
		capabilityRepository:  capabilityRepository,
		componentReadModel:    componentReadModel,
	}
}

func (h *testableLinkSystemToCapabilityHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.LinkSystemToCapability)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	capabilityID, err := valueobjects.NewCapabilityIDFromString(command.CapabilityID)
	if err != nil {
		return err
	}

	componentID, err := valueobjects.NewComponentIDFromString(command.ComponentID)
	if err != nil {
		return err
	}

	_, err = h.capabilityRepository.GetByID(ctx, capabilityID.Value())
	if err != nil {
		return ErrCapabilityNotFoundForRealization
	}

	component, err := h.componentReadModel.GetByID(ctx, componentID.Value())
	if err != nil {
		return err
	}
	if component == nil {
		return ErrComponentNotFound
	}

	realizationLevel, err := valueobjects.NewRealizationLevel(command.RealizationLevel)
	if err != nil {
		return err
	}

	notes := valueobjects.NewDescription(command.Notes)

	realization, err := aggregates.NewCapabilityRealization(
		capabilityID,
		componentID,
		realizationLevel,
		notes,
	)
	if err != nil {
		return err
	}

	command.ID = realization.ID()

	return h.realizationRepository.Save(ctx, realization)
}

func createTestCapabilityForLink(t *testing.T, level string, parentID string) *aggregates.Capability {
	t.Helper()

	name, err := valueobjects.NewCapabilityName("Test Capability")
	require.NoError(t, err)

	description := valueobjects.NewDescription("Test description")

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

	mockCapRepo := &mockCapabilityRepositoryForLink{
		capability: l1Capability,
	}
	mockRealRepo := &mockRealizationRepository{}
	mockCompReadModel := &mockComponentReadModel{
		component: &archReadModels.ApplicationComponentDTO{
			ID:   componentID,
			Name: "Test Component",
		},
	}

	handler := newTestableLinkSystemToCapabilityHandler(mockRealRepo, mockCapRepo, mockCompReadModel)

	cmd := &commands.LinkSystemToCapability{
		CapabilityID:     l1CapabilityID,
		ComponentID:      componentID,
		RealizationLevel: "Partial",
		Notes:            "Partially implements capability",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRealRepo.savedRealizations, 1, "Handler should create exactly 1 realization")

	realization := mockRealRepo.savedRealizations[0]
	assert.Equal(t, l1CapabilityID, realization.CapabilityID().Value())
	assert.Equal(t, componentID, realization.ComponentID().Value())
	assert.Equal(t, "Partial", realization.RealizationLevel().Value())
}

func TestLinkSystemToCapabilityHandler_SetsCommandID(t *testing.T) {
	l1Capability := createTestCapabilityForLink(t, "L1", "")
	l1CapabilityID := l1Capability.ID()

	componentID := valueobjects.NewCapabilityID().Value()

	mockCapRepo := &mockCapabilityRepositoryForLink{
		capability: l1Capability,
	}
	mockRealRepo := &mockRealizationRepository{}
	mockCompReadModel := &mockComponentReadModel{
		component: &archReadModels.ApplicationComponentDTO{
			ID:   componentID,
			Name: "Test Component",
		},
	}

	handler := newTestableLinkSystemToCapabilityHandler(mockRealRepo, mockCapRepo, mockCompReadModel)

	cmd := &commands.LinkSystemToCapability{
		CapabilityID:     l1CapabilityID,
		ComponentID:      componentID,
		RealizationLevel: "Full",
		Notes:            "",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.NotEmpty(t, cmd.ID, "Command ID should be set after handling")
	assert.Equal(t, mockRealRepo.savedRealizations[0].ID(), cmd.ID)
}

func TestLinkSystemToCapabilityHandler_ComponentNotFound_ReturnsError(t *testing.T) {
	l1Capability := createTestCapabilityForLink(t, "L1", "")
	l1CapabilityID := l1Capability.ID()

	componentID := valueobjects.NewCapabilityID().Value()

	mockCapRepo := &mockCapabilityRepositoryForLink{
		capability: l1Capability,
	}
	mockRealRepo := &mockRealizationRepository{}
	mockCompReadModel := &mockComponentReadModel{
		component: nil,
	}

	handler := newTestableLinkSystemToCapabilityHandler(mockRealRepo, mockCapRepo, mockCompReadModel)

	cmd := &commands.LinkSystemToCapability{
		CapabilityID:     l1CapabilityID,
		ComponentID:      componentID,
		RealizationLevel: "Full",
		Notes:            "",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrComponentNotFound)
}

func TestLinkSystemToCapabilityHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockCapRepo := &mockCapabilityRepositoryForLink{}
	mockRealRepo := &mockRealizationRepository{}
	mockCompReadModel := &mockComponentReadModel{}

	handler := newTestableLinkSystemToCapabilityHandler(mockRealRepo, mockCapRepo, mockCompReadModel)

	invalidCmd := &commands.DeleteSystemRealization{}

	err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}
