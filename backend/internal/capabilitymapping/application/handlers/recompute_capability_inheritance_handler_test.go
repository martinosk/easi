package handlers

import (
	"context"
	"testing"
	"time"

	"easi/backend/internal/capabilitymapping/application/commands"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/events"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRecomputeRepo struct {
	capability *aggregates.Capability
	saved      *aggregates.Capability
	err        error
}

func (m *mockRecomputeRepo) GetByID(ctx context.Context, id string) (*aggregates.Capability, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.capability, nil
}

func (m *mockRecomputeRepo) Save(ctx context.Context, capability *aggregates.Capability) error {
	if m.err != nil {
		return m.err
	}
	m.saved = capability
	return nil
}

type mockRecomputeCapRM struct {
	caps map[string]*readmodels.CapabilityDTO
}

func (m *mockRecomputeCapRM) GetByID(ctx context.Context, id string) (*readmodels.CapabilityDTO, error) {
	if c, ok := m.caps[id]; ok {
		return c, nil
	}
	return nil, nil
}

type mockRecomputeRealRM struct {
	realizations       []readmodels.RealizationDTO
	inheritedBySource  map[string][]string
}

func (m *mockRecomputeRealRM) GetByCapabilityID(ctx context.Context, capabilityID string) ([]readmodels.RealizationDTO, error) {
	return m.realizations, nil
}

func (m *mockRecomputeRealRM) GetInheritedCapabilityIDsBySourceRealizationID(ctx context.Context, sourceRealizationID string) ([]string, error) {
	return m.inheritedBySource[sourceRealizationID], nil
}

func newCapabilityForRecompute(t *testing.T, parentID string) *aggregates.Capability {
	name, err := valueobjects.NewCapabilityName("Capability")
	require.NoError(t, err)
	parent := valueobjects.CapabilityID{}
	if parentID != "" {
		parent, err = valueobjects.NewCapabilityIDFromString(parentID)
		require.NoError(t, err)
	}
	capability, err := aggregates.NewCapability(name, valueobjects.MustNewDescription(""), parent, valueobjects.LevelL2)
	require.NoError(t, err)
	capability.MarkChangesAsCommitted()
	return capability
}

func eventNames(eventsList []domain.DomainEvent) []string {
	names := make([]string, 0, len(eventsList))
	for _, e := range eventsList {
		names = append(names, e.EventType())
	}
	return names
}

func TestRecomputeCapabilityInheritanceHandler_EmitsCompensatingEvents(t *testing.T) {
	parentID := valueobjects.NewCapabilityID().Value()
	rootID := valueobjects.NewCapabilityID().Value()
	capability := newCapabilityForRecompute(t, parentID)

	repo := &mockRecomputeRepo{capability: capability}
	capRM := &mockRecomputeCapRM{caps: map[string]*readmodels.CapabilityDTO{
		parentID: {ID: parentID, ParentID: rootID},
		rootID:   {ID: rootID, ParentID: ""},
	}}
	realRM := &mockRecomputeRealRM{
		realizations: []readmodels.RealizationDTO{{
			ID:            "real-1",
			CapabilityID:  capability.ID(),
			ComponentID:   "comp-1",
			ComponentName: "Component A",
			Origin:        "Direct",
			LinkedAt:      time.Now().UTC(),
		}},
		inheritedBySource: map[string][]string{
			"real-1": {parentID, "old-ancestor"},
		},
	}

	handler := NewRecomputeCapabilityInheritanceHandler(repo, capRM, realRM)
	_, err := handler.Handle(context.Background(), &commands.RecomputeCapabilityInheritance{CapabilityID: capability.ID()})
	require.NoError(t, err)
	require.NotNil(t, repo.saved)

	changes := repo.saved.GetUncommittedChanges()
	assert.Contains(t, eventNames(changes), "CapabilityRealizationsInherited")
	assert.Contains(t, eventNames(changes), "CapabilityRealizationsUninherited")

	var inherited events.CapabilityRealizationsInherited
	var uninherited events.CapabilityRealizationsUninherited
	for _, change := range changes {
		if e, ok := change.(events.CapabilityRealizationsInherited); ok {
			inherited = e
		}
		if e, ok := change.(events.CapabilityRealizationsUninherited); ok {
			uninherited = e
		}
	}

	require.Len(t, inherited.InheritedRealizations, 1)
	assert.Equal(t, rootID, inherited.InheritedRealizations[0].CapabilityID)
	require.Len(t, uninherited.Removals, 1)
	assert.Equal(t, "real-1", uninherited.Removals[0].SourceRealizationID)
	assert.Equal(t, []string{"old-ancestor"}, uninherited.Removals[0].CapabilityIDs)
}

func TestRecomputeCapabilityInheritanceHandler_NoOpWhenNoDiff(t *testing.T) {
	parentID := valueobjects.NewCapabilityID().Value()
	capability := newCapabilityForRecompute(t, parentID)

	repo := &mockRecomputeRepo{capability: capability}
	capRM := &mockRecomputeCapRM{caps: map[string]*readmodels.CapabilityDTO{
		parentID: {ID: parentID, ParentID: ""},
	}}
	realRM := &mockRecomputeRealRM{
		realizations: []readmodels.RealizationDTO{{ID: "real-1", Origin: "Direct"}},
		inheritedBySource: map[string][]string{
			"real-1": {parentID},
		},
	}

	handler := NewRecomputeCapabilityInheritanceHandler(repo, capRM, realRM)
	_, err := handler.Handle(context.Background(), &commands.RecomputeCapabilityInheritance{CapabilityID: capability.ID()})
	require.NoError(t, err)
	assert.Nil(t, repo.saved)
}
