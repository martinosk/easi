package handlers

import (
	"context"
	"errors"
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCascadeRepository struct {
	capabilities map[string]*aggregates.Capability
	saved        []*aggregates.Capability
	getByIDErr   error
	saveErr      error
}

func newMockCascadeRepo() *mockCascadeRepository {
	return &mockCascadeRepository{capabilities: make(map[string]*aggregates.Capability)}
}

func (m *mockCascadeRepository) GetByID(ctx context.Context, id string) (*aggregates.Capability, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	cap, ok := m.capabilities[id]
	if !ok {
		return nil, errors.New("capability not found")
	}
	return cap, nil
}

func (m *mockCascadeRepository) Save(ctx context.Context, capability *aggregates.Capability) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = append(m.saved, capability)
	return nil
}

type mockCascadeHierarchyService struct {
	descendants map[string][]valueobjects.CapabilityID
	err         error
}

func (m *mockCascadeHierarchyService) GetDescendants(ctx context.Context, capabilityID valueobjects.CapabilityID) ([]valueobjects.CapabilityID, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.descendants[capabilityID.Value()], nil
}

type mockCascadeRealizationRM struct {
	realizationsByCapability map[string][]readmodels.RealizationDTO
	realizationsByComponent  map[string][]readmodels.RealizationDTO
	err                      error
}

func (m *mockCascadeRealizationRM) GetByCapabilityID(ctx context.Context, capabilityID string) ([]readmodels.RealizationDTO, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.realizationsByCapability[capabilityID], nil
}

func (m *mockCascadeRealizationRM) GetByComponentID(ctx context.Context, componentID string) ([]readmodels.RealizationDTO, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.realizationsByComponent[componentID], nil
}

type mockCascadeDependencyRM struct {
	outgoing map[string][]readmodels.DependencyDTO
	incoming map[string][]readmodels.DependencyDTO
	err      error
}

func (m *mockCascadeDependencyRM) GetOutgoing(ctx context.Context, capabilityID string) ([]readmodels.DependencyDTO, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.outgoing[capabilityID], nil
}

func (m *mockCascadeDependencyRM) GetIncoming(ctx context.Context, capabilityID string) ([]readmodels.DependencyDTO, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.incoming[capabilityID], nil
}

type mockCascadeCommandBus struct {
	dispatched []cqrs.Command
	err        error
}

func (m *mockCascadeCommandBus) Dispatch(ctx context.Context, cmd cqrs.Command) (cqrs.CommandResult, error) {
	m.dispatched = append(m.dispatched, cmd)
	return cqrs.EmptyResult(), m.err
}

type mockComponentDeleter struct {
	deletedIDs []string
	err        error
}

func (m *mockComponentDeleter) DeleteComponent(ctx context.Context, componentID string) error {
	if m.err != nil {
		return m.err
	}
	m.deletedIDs = append(m.deletedIDs, componentID)
	return nil
}

func cascadeCapabilityWithParent(t *testing.T, level, parentID string) *aggregates.Capability {
	t.Helper()
	name, _ := valueobjects.NewCapabilityName("Capability " + level)
	desc := valueobjects.MustNewDescription("Test")
	lvl, _ := valueobjects.NewCapabilityLevel(level)
	pid, _ := valueobjects.NewCapabilityIDFromString(parentID)
	cap, err := aggregates.NewCapability(name, desc, pid, lvl)
	require.NoError(t, err)
	cap.MarkChangesAsCommitted()
	return cap
}

func defaultCascadeDeps(repo *mockCascadeRepository) CascadeDeleteDeps {
	return CascadeDeleteDeps{
		Repository:       repo,
		HierarchyService: &mockCascadeHierarchyService{descendants: map[string][]valueobjects.CapabilityID{}},
		RealizationRM:    &mockCascadeRealizationRM{},
		DependencyRM:     &mockCascadeDependencyRM{},
		CommandBus:       &mockCascadeCommandBus{},
		CapabilityLookup: &mockDeleteCapabilityLookup{},
		ComponentDeleter: &mockComponentDeleter{},
	}
}

func TestCascadeDelete_InvalidCommand_ReturnsError(t *testing.T) {
	handler := NewCascadeDeleteCapabilityHandler(defaultCascadeDeps(newMockCascadeRepo()))

	_, err := handler.Handle(context.Background(), &commands.DeleteCapability{ID: "test"})
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestCascadeDelete_CapabilityNotFound_ReturnsError(t *testing.T) {
	repo := newMockCascadeRepo()
	repo.getByIDErr = errors.New("capability not found")

	handler := NewCascadeDeleteCapabilityHandler(defaultCascadeDeps(repo))

	cmd := &commands.CascadeDeleteCapability{ID: "550e8400-e29b-41d4-a716-446655440001"}
	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}

func TestCascadeDelete_HasDescendants_CascadeFalse_ReturnsCascadeRequiredError(t *testing.T) {
	root := createL1Capability(t)
	child := cascadeCapabilityWithParent(t, "L2", root.ID())

	repo := newMockCascadeRepo()
	repo.capabilities[root.ID()] = root

	childID, _ := valueobjects.NewCapabilityIDFromString(child.ID())
	deps := defaultCascadeDeps(repo)
	deps.HierarchyService = &mockCascadeHierarchyService{
		descendants: map[string][]valueobjects.CapabilityID{root.ID(): {childID}},
	}

	handler := NewCascadeDeleteCapabilityHandler(deps)

	cmd := &commands.CascadeDeleteCapability{ID: root.ID(), Cascade: false}
	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, services.ErrCascadeRequiredForChildCapabilities)
}

func TestCascadeDelete_LeafCapability_CascadeFalse_DeletesSuccessfully(t *testing.T) {
	root := createL1Capability(t)

	repo := newMockCascadeRepo()
	repo.capabilities[root.ID()] = root

	handler := NewCascadeDeleteCapabilityHandler(defaultCascadeDeps(repo))

	cmd := &commands.CascadeDeleteCapability{ID: root.ID(), Cascade: false}
	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, repo.saved, 1)
	uncommitted := repo.saved[0].GetUncommittedChanges()
	require.Len(t, uncommitted, 1)
	assert.Equal(t, "CapabilityDeleted", uncommitted[0].EventType())
}

func TestCascadeDelete_WithDescendants_CascadeTrue_DeletesBottomUp(t *testing.T) {
	root := createL1Capability(t)
	child := cascadeCapabilityWithParent(t, "L2", root.ID())
	grandchild := cascadeCapabilityWithParent(t, "L3", child.ID())

	repo := newMockCascadeRepo()
	repo.capabilities[root.ID()] = root
	repo.capabilities[child.ID()] = child
	repo.capabilities[grandchild.ID()] = grandchild

	childID, _ := valueobjects.NewCapabilityIDFromString(child.ID())
	grandchildID, _ := valueobjects.NewCapabilityIDFromString(grandchild.ID())

	deps := defaultCascadeDeps(repo)
	deps.HierarchyService = &mockCascadeHierarchyService{
		descendants: map[string][]valueobjects.CapabilityID{root.ID(): {childID, grandchildID}},
	}

	handler := NewCascadeDeleteCapabilityHandler(deps)

	cmd := &commands.CascadeDeleteCapability{ID: root.ID(), Cascade: true}
	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, repo.saved, 3)

	deletedIDs := make([]string, len(repo.saved))
	for i, cap := range repo.saved {
		deletedIDs[i] = cap.ID()
		assert.Equal(t, "CapabilityDeleted", cap.GetUncommittedChanges()[0].EventType())
	}

	rootIdx := indexOfStr(deletedIDs, root.ID())
	childIdx := indexOfStr(deletedIDs, child.ID())
	grandchildIdx := indexOfStr(deletedIDs, grandchild.ID())
	assert.Greater(t, rootIdx, childIdx)
	assert.Greater(t, childIdx, grandchildIdx)
}

func TestCascadeDelete_DispatchesRealizationDeleteCommands(t *testing.T) {
	root := createL1Capability(t)

	repo := newMockCascadeRepo()
	repo.capabilities[root.ID()] = root

	cmdBus := &mockCascadeCommandBus{}
	deps := defaultCascadeDeps(repo)
	deps.CommandBus = cmdBus
	deps.RealizationRM = &mockCascadeRealizationRM{
		realizationsByCapability: map[string][]readmodels.RealizationDTO{
			root.ID(): {{ID: "real-1", CapabilityID: root.ID(), ComponentID: "comp-1", Origin: "Direct", LinkedAt: time.Now()}},
		},
	}

	handler := NewCascadeDeleteCapabilityHandler(deps)

	_, err := handler.Handle(context.Background(), &commands.CascadeDeleteCapability{ID: root.ID()})
	require.NoError(t, err)

	require.Len(t, cmdBus.dispatched, 1)
	delRealCmd, ok := cmdBus.dispatched[0].(*commands.DeleteSystemRealization)
	require.True(t, ok)
	assert.Equal(t, "real-1", delRealCmd.ID)
}

func TestCascadeDelete_DispatchesDependencyDeleteCommands(t *testing.T) {
	root := createL1Capability(t)

	repo := newMockCascadeRepo()
	repo.capabilities[root.ID()] = root

	cmdBus := &mockCascadeCommandBus{}
	deps := defaultCascadeDeps(repo)
	deps.CommandBus = cmdBus
	deps.DependencyRM = &mockCascadeDependencyRM{
		outgoing: map[string][]readmodels.DependencyDTO{
			root.ID(): {{ID: "dep-1", SourceCapabilityID: root.ID(), TargetCapabilityID: "other-cap"}},
		},
		incoming: map[string][]readmodels.DependencyDTO{},
	}

	handler := NewCascadeDeleteCapabilityHandler(deps)

	_, err := handler.Handle(context.Background(), &commands.CascadeDeleteCapability{ID: root.ID()})
	require.NoError(t, err)

	require.Len(t, cmdBus.dispatched, 1)
	delDepCmd, ok := cmdBus.dispatched[0].(*commands.DeleteCapabilityDependency)
	require.True(t, ok)
	assert.Equal(t, "dep-1", delDepCmd.ID)
}

func TestCascadeDelete_DeduplicatesDependencies(t *testing.T) {
	root := createL1Capability(t)

	repo := newMockCascadeRepo()
	repo.capabilities[root.ID()] = root

	cmdBus := &mockCascadeCommandBus{}
	deps := defaultCascadeDeps(repo)
	deps.CommandBus = cmdBus
	deps.DependencyRM = &mockCascadeDependencyRM{
		outgoing: map[string][]readmodels.DependencyDTO{
			root.ID(): {{ID: "dep-1", SourceCapabilityID: root.ID(), TargetCapabilityID: "other-cap"}},
		},
		incoming: map[string][]readmodels.DependencyDTO{
			root.ID(): {{ID: "dep-1", SourceCapabilityID: "other-cap", TargetCapabilityID: root.ID()}},
		},
	}

	handler := NewCascadeDeleteCapabilityHandler(deps)

	_, err := handler.Handle(context.Background(), &commands.CascadeDeleteCapability{ID: root.ID()})
	require.NoError(t, err)

	require.Len(t, cmdBus.dispatched, 1)
}

func TestCascadeDelete_DeleteRealisingApplications_DeletesExclusiveComponents(t *testing.T) {
	root := createL1Capability(t)

	repo := newMockCascadeRepo()
	repo.capabilities[root.ID()] = root

	deleter := &mockComponentDeleter{}
	deps := defaultCascadeDeps(repo)
	deps.ComponentDeleter = deleter
	deps.RealizationRM = &mockCascadeRealizationRM{
		realizationsByCapability: map[string][]readmodels.RealizationDTO{
			root.ID(): {
				{ID: "real-1", CapabilityID: root.ID(), ComponentID: "comp-exclusive", Origin: "Direct", LinkedAt: time.Now()},
				{ID: "real-2", CapabilityID: root.ID(), ComponentID: "comp-shared", Origin: "Direct", LinkedAt: time.Now()},
			},
		},
		realizationsByComponent: map[string][]readmodels.RealizationDTO{
			"comp-exclusive": {{ID: "real-1", CapabilityID: root.ID(), ComponentID: "comp-exclusive", Origin: "Direct"}},
			"comp-shared": {
				{ID: "real-2", CapabilityID: root.ID(), ComponentID: "comp-shared", Origin: "Direct"},
				{ID: "real-3", CapabilityID: "other-cap", ComponentID: "comp-shared", Origin: "Direct"},
			},
		},
	}

	handler := NewCascadeDeleteCapabilityHandler(deps)

	cmd := &commands.CascadeDeleteCapability{ID: root.ID(), DeleteRealisingApplications: true}
	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, deleter.deletedIDs, 1)
	assert.Equal(t, "comp-exclusive", deleter.deletedIDs[0])
}

func TestCascadeDelete_DeleteRealisingApplications_False_SkipsComponentDeletion(t *testing.T) {
	root := createL1Capability(t)

	repo := newMockCascadeRepo()
	repo.capabilities[root.ID()] = root

	deleter := &mockComponentDeleter{}
	deps := defaultCascadeDeps(repo)
	deps.ComponentDeleter = deleter
	deps.RealizationRM = &mockCascadeRealizationRM{
		realizationsByCapability: map[string][]readmodels.RealizationDTO{
			root.ID(): {{ID: "real-1", CapabilityID: root.ID(), ComponentID: "comp-exclusive", Origin: "Direct", LinkedAt: time.Now()}},
		},
		realizationsByComponent: map[string][]readmodels.RealizationDTO{
			"comp-exclusive": {{ID: "real-1", CapabilityID: root.ID(), ComponentID: "comp-exclusive", Origin: "Direct"}},
		},
	}

	handler := NewCascadeDeleteCapabilityHandler(deps)

	cmd := &commands.CascadeDeleteCapability{ID: root.ID(), DeleteRealisingApplications: false}
	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	assert.Empty(t, deleter.deletedIDs)
}

func indexOfStr(strs []string, target string) int {
	for i, s := range strs {
		if s == target {
			return i
		}
	}
	return -1
}
