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
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockChangeParentRepository struct {
	capabilities map[string]*aggregates.Capability
	saved        *aggregates.Capability
	saveErr      error
	getErr       error
}

func (m *mockChangeParentRepository) GetByID(ctx context.Context, id string) (*aggregates.Capability, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	capability, ok := m.capabilities[id]
	if !ok {
		return nil, repositories.ErrCapabilityNotFound
	}
	return capability, nil
}

func (m *mockChangeParentRepository) Save(ctx context.Context, capability *aggregates.Capability) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saved = capability
	return nil
}

type mockChangeParentCapabilityReadModel struct {
	capabilities map[string]*readmodels.CapabilityDTO
	children     map[string][]readmodels.CapabilityDTO
}

func (m *mockChangeParentCapabilityReadModel) GetChildren(ctx context.Context, parentID string) ([]readmodels.CapabilityDTO, error) {
	return m.children[parentID], nil
}

func (m *mockChangeParentCapabilityReadModel) GetByID(ctx context.Context, id string) (*readmodels.CapabilityDTO, error) {
	cap, ok := m.capabilities[id]
	if !ok {
		return nil, nil
	}
	return cap, nil
}

type mockChangeParentRealizationReadModel struct {
	realizations map[string][]readmodels.RealizationDTO
}

func (m *mockChangeParentRealizationReadModel) GetByCapabilityID(ctx context.Context, capabilityID string) ([]readmodels.RealizationDTO, error) {
	return m.realizations[capabilityID], nil
}

type mockReparentingService struct {
	level valueobjects.CapabilityLevel
}

func (m *mockReparentingService) DetermineNewLevel(ctx context.Context, capabilityID valueobjects.CapabilityID, newParentID valueobjects.CapabilityID, parentLevel valueobjects.CapabilityLevel) (valueobjects.CapabilityLevel, error) {
	return m.level, nil
}

func (m *mockReparentingService) CalculateChildLevel(parentLevel valueobjects.CapabilityLevel) (valueobjects.CapabilityLevel, error) {
	switch parentLevel {
	case valueobjects.LevelL1:
		return valueobjects.LevelL2, nil
	case valueobjects.LevelL2:
		return valueobjects.LevelL3, nil
	case valueobjects.LevelL3:
		return valueobjects.LevelL4, nil
	default:
		return "", aggregates.ErrWouldExceedMaximumDepth
	}
}

func TestChangeCapabilityParentHandler_EmitsInheritanceEvents(t *testing.T) {
	oldParentID := valueobjects.NewCapabilityID()

	capability := newTestCapability(t, oldParentID, valueobjects.LevelL2)
	parent := newTestCapability(t, valueobjects.CapabilityID{}, valueobjects.LevelL1)
	newParentID, err := valueobjects.NewCapabilityIDFromString(parent.ID())
	require.NoError(t, err)

	repo := &mockChangeParentRepository{
		capabilities: map[string]*aggregates.Capability{
			capability.ID(): capability,
			newParentID.Value(): parent,
		},
	}

	capRM := &mockChangeParentCapabilityReadModel{
		capabilities: map[string]*readmodels.CapabilityDTO{
			capability.ID():     {ID: capability.ID(), Name: "Child", ParentID: oldParentID.Value(), Level: "L2"},
			oldParentID.Value(): {ID: oldParentID.Value(), Name: "Old Parent", ParentID: "old-root", Level: "L1"},
			"old-root":          {ID: "old-root", Name: "Old Root", ParentID: "", Level: "L1"},
			newParentID.Value(): {ID: newParentID.Value(), Name: "New Parent", ParentID: "new-root", Level: "L1"},
			"new-root":          {ID: "new-root", Name: "New Root", ParentID: "", Level: "L1"},
		},
		children: map[string][]readmodels.CapabilityDTO{},
	}

	realRM := &mockChangeParentRealizationReadModel{
		realizations: map[string][]readmodels.RealizationDTO{
			capability.ID(): {
				{
					ID:               "real-direct",
					CapabilityID:     capability.ID(),
					ComponentID:      "comp-1",
					ComponentName:    "Component 1",
					RealizationLevel: "Partial",
					Origin:           "Direct",
					LinkedAt:         time.Now().UTC(),
				},
				{
					ID:                  "real-inherited",
					CapabilityID:        capability.ID(),
					ComponentID:         "comp-2",
					ComponentName:       "Component 2",
					RealizationLevel:    "Full",
					Origin:              "Inherited",
					SourceRealizationID: "real-source",
					SourceCapabilityID:  "cap-source",
					SourceCapabilityName:"Source Cap",
					LinkedAt:            time.Now().UTC(),
				},
			},
		},
	}

	handler := NewChangeCapabilityParentHandler(repo, capRM, realRM, &mockReparentingService{level: valueobjects.LevelL2})

	cmd := &commands.ChangeCapabilityParent{CapabilityID: capability.ID(), NewParentID: newParentID.Value()}
	_, err = handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.NotNil(t, repo.saved)
	changes := repo.saved.GetUncommittedChanges()

	assert.Contains(t, eventTypes(changes), "CapabilityParentChanged")
	assert.Contains(t, eventTypes(changes), "CapabilityRealizationsInherited")
	assert.Contains(t, eventTypes(changes), "CapabilityRealizationsUninherited")

	added, ok := findAddedEvent(changes)
	require.True(t, ok)
	assert.Len(t, added.InheritedRealizations, 4)
	assert.Equal(t, newParentID.Value(), added.InheritedRealizations[0].CapabilityID)
	assert.Equal(t, "new-root", added.InheritedRealizations[1].CapabilityID)

	removed, ok := findRemovedEvent(changes)
	require.True(t, ok)
	assert.Len(t, removed.Removals, 2)
	assert.ElementsMatch(t, []string{oldParentID.Value(), "old-root"}, removed.Removals[0].CapabilityIDs)
}

func eventTypes(eventsList []domain.DomainEvent) []string {
	types := make([]string, 0, len(eventsList))
	for _, event := range eventsList {
		types = append(types, event.EventType())
	}
	return types
}

func findAddedEvent(eventsList []domain.DomainEvent) (events.CapabilityRealizationsInherited, bool) {
	for _, event := range eventsList {
		if typed, ok := event.(events.CapabilityRealizationsInherited); ok {
			return typed, true
		}
	}
	return events.CapabilityRealizationsInherited{}, false
}

func findRemovedEvent(eventsList []domain.DomainEvent) (events.CapabilityRealizationsUninherited, bool) {
	for _, event := range eventsList {
		if typed, ok := event.(events.CapabilityRealizationsUninherited); ok {
			return typed, true
		}
	}
	return events.CapabilityRealizationsUninherited{}, false
}

func newTestCapability(t *testing.T, parentID valueobjects.CapabilityID, level valueobjects.CapabilityLevel) *aggregates.Capability {
	name, err := valueobjects.NewCapabilityName("Test")
	require.NoError(t, err)
	description := valueobjects.MustNewDescription("Test")

	capability, err := aggregates.NewCapability(name, description, parentID, level)
	require.NoError(t, err)
	capability.MarkChangesAsCommitted()
	return capability
}
