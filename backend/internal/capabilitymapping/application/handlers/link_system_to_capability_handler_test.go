package handlers

import (
	"context"
	"testing"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	domainEvents "easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/architecturemodeling"
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
	capability         *aggregates.Capability
	getByIDErr         error
	savedCapabilities  []*aggregates.Capability
	saveErr            error
}

func (m *mockLinkSystemCapabilityRepository) GetByID(ctx context.Context, id string) (*aggregates.Capability, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	return m.capability, nil
}

func (m *mockLinkSystemCapabilityRepository) Save(ctx context.Context, capability *aggregates.Capability) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedCapabilities = append(m.savedCapabilities, capability)
	return nil
}

type mockLinkSystemComponentReadModel struct {
	component *architecturemodeling.ComponentDTO
	getErr    error
}

func (m *mockLinkSystemComponentReadModel) GetByID(ctx context.Context, id string) (*architecturemodeling.ComponentDTO, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.component, nil
}

type mockLinkSystemCapabilityReadModel struct {
	capabilities map[string]*readmodels.CapabilityDTO
	getErr       error
}

func (m *mockLinkSystemCapabilityReadModel) GetByID(ctx context.Context, id string) (*readmodels.CapabilityDTO, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	capability, ok := m.capabilities[id]
	if !ok {
		return nil, nil
	}
	return capability, nil
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

type linkTestFixture struct {
	capRepo      *mockLinkSystemCapabilityRepository
	realRepo     *mockLinkSystemRealizationRepository
	capReadModel *mockLinkSystemCapabilityReadModel
	compReadModel *mockLinkSystemComponentReadModel
	handler      *LinkSystemToCapabilityHandler
}

func setupLinkTest(t *testing.T, capability *aggregates.Capability, component *architecturemodeling.ComponentDTO) *linkTestFixture {
	t.Helper()
	f := &linkTestFixture{
		capRepo:       &mockLinkSystemCapabilityRepository{capability: capability},
		realRepo:      &mockLinkSystemRealizationRepository{},
		capReadModel:  &mockLinkSystemCapabilityReadModel{capabilities: map[string]*readmodels.CapabilityDTO{}},
		compReadModel: &mockLinkSystemComponentReadModel{component: component},
	}
	f.handler = NewLinkSystemToCapabilityHandler(f.realRepo, f.capRepo, f.capReadModel, f.compReadModel)
	return f
}

func TestLinkSystemToCapabilityHandler_CreatesRealization(t *testing.T) {
	l1Capability := createTestCapabilityForLink(t, "L1", "")
	componentID := valueobjects.NewCapabilityID().Value()
	f := setupLinkTest(t, l1Capability, &architecturemodeling.ComponentDTO{ID: componentID, Name: "Test Component"})

	cmd := &commands.LinkSystemToCapability{
		CapabilityID:     l1Capability.ID(),
		ComponentID:      componentID,
		RealizationLevel: "Partial",
		Notes:            "Partially implements capability",
	}

	_, err := f.handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, f.realRepo.savedRealizations, 1, "Handler should create exactly 1 realization")

	realization := f.realRepo.savedRealizations[0]
	assert.Equal(t, l1Capability.ID(), realization.CapabilityID().Value())
	assert.Equal(t, componentID, realization.ComponentID().Value())
	assert.Equal(t, "Partial", realization.RealizationLevel().Value())
}

func TestLinkSystemToCapabilityHandler_ReturnsCreatedID(t *testing.T) {
	l1Capability := createTestCapabilityForLink(t, "L1", "")
	componentID := valueobjects.NewCapabilityID().Value()
	f := setupLinkTest(t, l1Capability, &architecturemodeling.ComponentDTO{ID: componentID, Name: "Test Component"})

	cmd := &commands.LinkSystemToCapability{
		CapabilityID:     l1Capability.ID(),
		ComponentID:      componentID,
		RealizationLevel: "Full",
		Notes:            "",
	}

	result, err := f.handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.NotEmpty(t, result.CreatedID, "Result CreatedID should be set after handling")
	assert.Equal(t, f.realRepo.savedRealizations[0].ID(), result.CreatedID)
}

func TestLinkSystemToCapabilityHandler_ComponentNotFound_ReturnsError(t *testing.T) {
	l1Capability := createTestCapabilityForLink(t, "L1", "")
	componentID := valueobjects.NewCapabilityID().Value()
	f := setupLinkTest(t, l1Capability, nil)

	cmd := &commands.LinkSystemToCapability{
		CapabilityID:     l1Capability.ID(),
		ComponentID:      componentID,
		RealizationLevel: "Full",
		Notes:            "",
	}

	_, err := f.handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrComponentNotFound)
}

func TestLinkSystemToCapabilityHandler_CapabilityNotFound_ReturnsError(t *testing.T) {
	componentID := valueobjects.NewCapabilityID().Value()
	capabilityID := valueobjects.NewCapabilityID().Value()
	f := setupLinkTest(t, nil, &architecturemodeling.ComponentDTO{ID: componentID, Name: "Test Component"})
	f.capRepo.getByIDErr = repositories.ErrCapabilityNotFound

	cmd := &commands.LinkSystemToCapability{
		CapabilityID:     capabilityID,
		ComponentID:      componentID,
		RealizationLevel: "Full",
		Notes:            "",
	}

	_, err := f.handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, ErrCapabilityNotFoundForRealization)
}

func TestLinkSystemToCapabilityHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockCapRepo := &mockLinkSystemCapabilityRepository{}
	mockRealRepo := &mockLinkSystemRealizationRepository{}
	mockCompReadModel := &mockLinkSystemComponentReadModel{}

	mockCapReadModel := &mockLinkSystemCapabilityReadModel{capabilities: map[string]*readmodels.CapabilityDTO{}}
	handler := NewLinkSystemToCapabilityHandler(mockRealRepo, mockCapRepo, mockCapReadModel, mockCompReadModel)

	invalidCmd := &commands.DeleteSystemRealization{}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestLinkSystemToCapabilityHandler_EmitsInheritanceEventForAncestors(t *testing.T) {
	parentID := valueobjects.NewCapabilityID().Value()
	rootID := valueobjects.NewCapabilityID().Value()
	l2Capability := createTestCapabilityForLink(t, "L2", parentID)

	componentID := valueobjects.NewCapabilityID().Value()
	mockCapRepo := &mockLinkSystemCapabilityRepository{capability: l2Capability}
	mockRealRepo := &mockLinkSystemRealizationRepository{}
	mockCapReadModel := &mockLinkSystemCapabilityReadModel{
		capabilities: map[string]*readmodels.CapabilityDTO{
			parentID: {ID: parentID, ParentID: rootID},
			rootID:   {ID: rootID, ParentID: ""},
		},
	}
	mockCompReadModel := &mockLinkSystemComponentReadModel{
		component: &architecturemodeling.ComponentDTO{ID: componentID, Name: "Test Component"},
	}

	handler := NewLinkSystemToCapabilityHandler(mockRealRepo, mockCapRepo, mockCapReadModel, mockCompReadModel)
	cmd := &commands.LinkSystemToCapability{CapabilityID: l2Capability.ID(), ComponentID: componentID, RealizationLevel: "Partial"}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)
	require.Len(t, mockCapRepo.savedCapabilities, 1)

	changes := mockCapRepo.savedCapabilities[0].GetUncommittedChanges()
	require.Len(t, changes, 1)

	inherited, ok := changes[0].(domainEvents.CapabilityRealizationsInherited)
	require.True(t, ok)
	require.Len(t, inherited.InheritedRealizations, 2)
	assert.Equal(t, parentID, inherited.InheritedRealizations[0].CapabilityID)
	assert.Equal(t, rootID, inherited.InheritedRealizations[1].CapabilityID)
}
