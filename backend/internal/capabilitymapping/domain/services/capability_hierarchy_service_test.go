package services

import (
	"context"
	"testing"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockCapabilityLookup struct {
	capabilities map[string]*CapabilityInfo
	children     map[string][]valueobjects.CapabilityID
}

func newMockCapabilityLookup() *mockCapabilityLookup {
	return &mockCapabilityLookup{
		capabilities: make(map[string]*CapabilityInfo),
		children:     make(map[string][]valueobjects.CapabilityID),
	}
}

func (m *mockCapabilityLookup) GetCapabilityInfo(ctx context.Context, id valueobjects.CapabilityID) (*CapabilityInfo, error) {
	info, ok := m.capabilities[id.Value()]
	if !ok {
		return nil, nil
	}
	return info, nil
}

func (m *mockCapabilityLookup) GetChildren(ctx context.Context, parentID valueobjects.CapabilityID) ([]valueobjects.CapabilityID, error) {
	children, ok := m.children[parentID.Value()]
	if !ok {
		return nil, nil
	}
	return children, nil
}

func (m *mockCapabilityLookup) addCapability(id valueobjects.CapabilityID, level valueobjects.CapabilityLevel, parentID valueobjects.CapabilityID) {
	m.capabilities[id.Value()] = &CapabilityInfo{
		ID:       id,
		Level:    level,
		ParentID: parentID,
	}
}

func (m *mockCapabilityLookup) addChild(parentID, childID valueobjects.CapabilityID) {
	m.children[parentID.Value()] = append(m.children[parentID.Value()], childID)
}

type hierarchyHarness struct {
	lookup  *mockCapabilityLookup
	service CapabilityHierarchyService
}

func newHierarchyHarness() *hierarchyHarness {
	lookup := newMockCapabilityLookup()
	return &hierarchyHarness{
		lookup:  lookup,
		service: NewCapabilityHierarchyService(lookup),
	}
}

func (h *hierarchyHarness) addRootL1() valueobjects.CapabilityID {
	l1ID := valueobjects.NewCapabilityID()
	h.lookup.addCapability(l1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	return l1ID
}

func (h *hierarchyHarness) addChildAt(level valueobjects.CapabilityLevel, parentID valueobjects.CapabilityID) valueobjects.CapabilityID {
	childID := valueobjects.NewCapabilityID()
	h.lookup.addCapability(childID, level, parentID)
	h.lookup.addChild(parentID, childID)
	return childID
}

func TestCapabilityHierarchyService_FindL1Ancestor(t *testing.T) {
	tests := []struct {
		name  string
		build func(h *hierarchyHarness) (target, expected valueobjects.CapabilityID)
	}{
		{
			name: "L1 returns self",
			build: func(h *hierarchyHarness) (valueobjects.CapabilityID, valueobjects.CapabilityID) {
				l1ID := h.addRootL1()
				return l1ID, l1ID
			},
		},
		{
			name: "L2 returns L1 parent",
			build: func(h *hierarchyHarness) (valueobjects.CapabilityID, valueobjects.CapabilityID) {
				l1ID := h.addRootL1()
				l2ID := h.addChildAt(valueobjects.LevelL2, l1ID)
				return l2ID, l1ID
			},
		},
		{
			name: "L3 returns L1 ancestor",
			build: func(h *hierarchyHarness) (valueobjects.CapabilityID, valueobjects.CapabilityID) {
				l1ID := h.addRootL1()
				l2ID := h.addChildAt(valueobjects.LevelL2, l1ID)
				l3ID := h.addChildAt(valueobjects.LevelL3, l2ID)
				return l3ID, l1ID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHierarchyHarness()
			target, expected := tt.build(h)

			result, err := h.service.FindL1Ancestor(context.Background(), target)
			require.NoError(t, err)
			assert.Equal(t, expected.Value(), result.Value())
		})
	}
}

func TestCapabilityHierarchyService_FindL1Ancestor_CapabilityNotFound(t *testing.T) {
	h := newHierarchyHarness()

	unknownID := valueobjects.NewCapabilityID()

	_, err := h.service.FindL1Ancestor(context.Background(), unknownID)
	assert.Error(t, err)
	assert.Equal(t, ErrCapabilityNotFound, err)
}

func TestCapabilityHierarchyService_GetDescendants_NoChildren(t *testing.T) {
	h := newHierarchyHarness()

	l1ID := h.addRootL1()

	descendants, err := h.service.GetDescendants(context.Background(), l1ID)
	require.NoError(t, err)
	assert.Empty(t, descendants)
}

func TestCapabilityHierarchyService_GetDescendants_WithChildren(t *testing.T) {
	h := newHierarchyHarness()

	l1ID := h.addRootL1()
	l2ID := h.addChildAt(valueobjects.LevelL2, l1ID)
	l3ID := h.addChildAt(valueobjects.LevelL3, l2ID)

	descendants, err := h.service.GetDescendants(context.Background(), l1ID)
	require.NoError(t, err)
	assert.Len(t, descendants, 2)
	assert.Contains(t, descendantIDs(descendants), l2ID.Value())
	assert.Contains(t, descendantIDs(descendants), l3ID.Value())
}

func TestCapabilityHierarchyService_ValidateHierarchyChange(t *testing.T) {
	tests := []struct {
		name    string
		build   func(h *hierarchyHarness) (capID, parentID valueobjects.CapabilityID)
		wantErr error
	}{
		{
			name: "empty parent succeeds",
			build: func(h *hierarchyHarness) (valueobjects.CapabilityID, valueobjects.CapabilityID) {
				return valueobjects.NewCapabilityID(), valueobjects.CapabilityID{}
			},
		},
		{
			name: "self reference fails",
			build: func(h *hierarchyHarness) (valueobjects.CapabilityID, valueobjects.CapabilityID) {
				capID := valueobjects.NewCapabilityID()
				return capID, capID
			},
			wantErr: ErrWouldCreateCircularHierarchy,
		},
		{
			name: "descendant as parent fails",
			build: func(h *hierarchyHarness) (valueobjects.CapabilityID, valueobjects.CapabilityID) {
				l1ID := h.addRootL1()
				l2ID := h.addChildAt(valueobjects.LevelL2, l1ID)
				return l1ID, l2ID
			},
			wantErr: ErrWouldCreateCircularHierarchy,
		},
		{
			name: "valid parent succeeds",
			build: func(h *hierarchyHarness) (valueobjects.CapabilityID, valueobjects.CapabilityID) {
				l1ID := h.addRootL1()
				l2ID := h.addChildAt(valueobjects.LevelL2, l1ID)
				otherL1ID := h.addRootL1()
				return l2ID, otherL1ID
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHierarchyHarness()
			capID, parentID := tt.build(h)

			err := h.service.ValidateHierarchyChange(context.Background(), capID, parentID)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func descendantIDs(descendants []valueobjects.CapabilityID) []string {
	ids := make([]string, len(descendants))
	for i, d := range descendants {
		ids[i] = d.Value()
	}
	return ids
}
