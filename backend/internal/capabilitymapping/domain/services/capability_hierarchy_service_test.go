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

func TestCapabilityHierarchyService_FindL1Ancestor_L1_ReturnsSelf(t *testing.T) {
	lookup := newMockCapabilityLookup()
	service := NewCapabilityHierarchyService(lookup)

	l1ID := valueobjects.NewCapabilityID()
	lookup.addCapability(l1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})

	result, err := service.FindL1Ancestor(context.Background(), l1ID)
	require.NoError(t, err)
	assert.Equal(t, l1ID.Value(), result.Value())
}

func TestCapabilityHierarchyService_FindL1Ancestor_L2_ReturnsL1Parent(t *testing.T) {
	lookup := newMockCapabilityLookup()
	service := NewCapabilityHierarchyService(lookup)

	l1ID := valueobjects.NewCapabilityID()
	l2ID := valueobjects.NewCapabilityID()

	lookup.addCapability(l1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	lookup.addCapability(l2ID, valueobjects.LevelL2, l1ID)

	result, err := service.FindL1Ancestor(context.Background(), l2ID)
	require.NoError(t, err)
	assert.Equal(t, l1ID.Value(), result.Value())
}

func TestCapabilityHierarchyService_FindL1Ancestor_L3_ReturnsL1Ancestor(t *testing.T) {
	lookup := newMockCapabilityLookup()
	service := NewCapabilityHierarchyService(lookup)

	l1ID := valueobjects.NewCapabilityID()
	l2ID := valueobjects.NewCapabilityID()
	l3ID := valueobjects.NewCapabilityID()

	lookup.addCapability(l1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	lookup.addCapability(l2ID, valueobjects.LevelL2, l1ID)
	lookup.addCapability(l3ID, valueobjects.LevelL3, l2ID)

	result, err := service.FindL1Ancestor(context.Background(), l3ID)
	require.NoError(t, err)
	assert.Equal(t, l1ID.Value(), result.Value())
}

func TestCapabilityHierarchyService_FindL1Ancestor_CapabilityNotFound(t *testing.T) {
	lookup := newMockCapabilityLookup()
	service := NewCapabilityHierarchyService(lookup)

	unknownID := valueobjects.NewCapabilityID()

	_, err := service.FindL1Ancestor(context.Background(), unknownID)
	assert.Error(t, err)
	assert.Equal(t, ErrCapabilityNotFound, err)
}

func TestCapabilityHierarchyService_GetDescendants_NoChildren(t *testing.T) {
	lookup := newMockCapabilityLookup()
	service := NewCapabilityHierarchyService(lookup)

	l1ID := valueobjects.NewCapabilityID()
	lookup.addCapability(l1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})

	descendants, err := service.GetDescendants(context.Background(), l1ID)
	require.NoError(t, err)
	assert.Empty(t, descendants)
}

func TestCapabilityHierarchyService_GetDescendants_WithChildren(t *testing.T) {
	lookup := newMockCapabilityLookup()
	service := NewCapabilityHierarchyService(lookup)

	l1ID := valueobjects.NewCapabilityID()
	l2ID := valueobjects.NewCapabilityID()
	l3ID := valueobjects.NewCapabilityID()

	lookup.addCapability(l1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	lookup.addCapability(l2ID, valueobjects.LevelL2, l1ID)
	lookup.addCapability(l3ID, valueobjects.LevelL3, l2ID)

	lookup.addChild(l1ID, l2ID)
	lookup.addChild(l2ID, l3ID)

	descendants, err := service.GetDescendants(context.Background(), l1ID)
	require.NoError(t, err)
	assert.Len(t, descendants, 2)
	assert.Contains(t, descendantIDs(descendants), l2ID.Value())
	assert.Contains(t, descendantIDs(descendants), l3ID.Value())
}

func TestCapabilityHierarchyService_ValidateHierarchyChange_EmptyParent_Succeeds(t *testing.T) {
	lookup := newMockCapabilityLookup()
	service := NewCapabilityHierarchyService(lookup)

	capID := valueobjects.NewCapabilityID()

	err := service.ValidateHierarchyChange(context.Background(), capID, valueobjects.CapabilityID{})
	assert.NoError(t, err)
}

func TestCapabilityHierarchyService_ValidateHierarchyChange_SelfReference_Fails(t *testing.T) {
	lookup := newMockCapabilityLookup()
	service := NewCapabilityHierarchyService(lookup)

	capID := valueobjects.NewCapabilityID()

	err := service.ValidateHierarchyChange(context.Background(), capID, capID)
	assert.Error(t, err)
	assert.Equal(t, ErrWouldCreateCircularHierarchy, err)
}

func TestCapabilityHierarchyService_ValidateHierarchyChange_DescendantAsParent_Fails(t *testing.T) {
	lookup := newMockCapabilityLookup()
	service := NewCapabilityHierarchyService(lookup)

	l1ID := valueobjects.NewCapabilityID()
	l2ID := valueobjects.NewCapabilityID()

	lookup.addCapability(l1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	lookup.addCapability(l2ID, valueobjects.LevelL2, l1ID)

	lookup.addChild(l1ID, l2ID)

	err := service.ValidateHierarchyChange(context.Background(), l1ID, l2ID)
	assert.Error(t, err)
	assert.Equal(t, ErrWouldCreateCircularHierarchy, err)
}

func TestCapabilityHierarchyService_ValidateHierarchyChange_ValidParent_Succeeds(t *testing.T) {
	lookup := newMockCapabilityLookup()
	service := NewCapabilityHierarchyService(lookup)

	l1ID := valueobjects.NewCapabilityID()
	l2ID := valueobjects.NewCapabilityID()
	otherL1ID := valueobjects.NewCapabilityID()

	lookup.addCapability(l1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})
	lookup.addCapability(l2ID, valueobjects.LevelL2, l1ID)
	lookup.addCapability(otherL1ID, valueobjects.LevelL1, valueobjects.CapabilityID{})

	lookup.addChild(l1ID, l2ID)

	err := service.ValidateHierarchyChange(context.Background(), l2ID, otherL1ID)
	assert.NoError(t, err)
}

func descendantIDs(descendants []valueobjects.CapabilityID) []string {
	ids := make([]string, len(descendants))
	for i, d := range descendants {
		ids[i] = d.Value()
	}
	return ids
}
