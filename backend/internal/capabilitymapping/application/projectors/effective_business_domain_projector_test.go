package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEffectiveBDStore struct {
	upserted         []readmodels.CMEffectiveBusinessDomainDTO
	deleted          []string
	rows             map[string]*readmodels.CMEffectiveBusinessDomainDTO
	l1SubtreeUpdates []l1SubtreeUpdate
}

type l1SubtreeUpdate struct {
	l1CapabilityID string
	bdID           string
	bdName         string
}

func (m *mockEffectiveBDStore) Upsert(ctx context.Context, dto readmodels.CMEffectiveBusinessDomainDTO) error {
	m.upserted = append(m.upserted, dto)
	if m.rows == nil {
		m.rows = make(map[string]*readmodels.CMEffectiveBusinessDomainDTO)
	}
	copied := dto
	m.rows[dto.CapabilityID] = &copied
	return nil
}

func (m *mockEffectiveBDStore) Delete(ctx context.Context, capabilityID string) error {
	m.deleted = append(m.deleted, capabilityID)
	delete(m.rows, capabilityID)
	return nil
}

func (m *mockEffectiveBDStore) GetByCapabilityID(ctx context.Context, capabilityID string) (*readmodels.CMEffectiveBusinessDomainDTO, error) {
	if m.rows == nil {
		return nil, nil
	}
	dto, ok := m.rows[capabilityID]
	if !ok {
		return nil, nil
	}
	return dto, nil
}

func (m *mockEffectiveBDStore) UpdateBusinessDomainForL1Subtree(ctx context.Context, l1CapabilityID string, bdID string, bdName string) error {
	m.l1SubtreeUpdates = append(m.l1SubtreeUpdates, l1SubtreeUpdate{l1CapabilityID, bdID, bdName})
	for _, dto := range m.rows {
		if dto.L1CapabilityID == l1CapabilityID {
			dto.BusinessDomainID = bdID
			dto.BusinessDomainName = bdName
		}
	}
	return nil
}

type mockBDNameProvider struct {
	domains map[string]*readmodels.BusinessDomainDTO
	err    error
}

func (m *mockBDNameProvider) GetByID(ctx context.Context, id string) (*readmodels.BusinessDomainDTO, error) {
	if m != nil && m.err != nil {
		return nil, m.err
	}
	if m == nil || m.domains == nil {
		return nil, nil
	}
	dto, ok := m.domains[id]
	if !ok {
		return nil, nil
	}
	return dto, nil
}

type mockCapabilityChildProvider struct {
	children map[string][]readmodels.CapabilityDTO
	err      error
}

func (m *mockCapabilityChildProvider) GetChildren(ctx context.Context, parentID string) ([]readmodels.CapabilityDTO, error) {
	if m != nil && m.err != nil {
		return nil, m.err
	}
	if m == nil || m.children == nil {
		return nil, nil
	}
	return m.children[parentID], nil
}

func TestEffectiveBusinessDomainProjector_CapabilityCreated_ParentLookupError_ReturnsError(t *testing.T) {
	store := &mockEffectiveBDStore{}
	store.rows = map[string]*readmodels.CMEffectiveBusinessDomainDTO{}
	projector := NewEffectiveBusinessDomainProjector(store, nil, nil)

	eventData, err := json.Marshal(events.NewCapabilityCreated("child-l2", "Sub-Payments", "", "parent-l1", "L2"))
	require.NoError(t, err)

	storeGetErr := errors.New("db read failed")
	storeWithErr := &mockEffectiveBDStoreWithGetErr{mockEffectiveBDStore: *store, getErr: storeGetErr}
	projector = NewEffectiveBusinessDomainProjector(storeWithErr, nil, nil)

	err = projector.ProjectEvent(context.Background(), "CapabilityCreated", eventData)
	assert.Error(t, err)
}

func TestEffectiveBusinessDomainProjector_CapabilityAssignedToDomain_BusinessDomainLookupError_ReturnsError(t *testing.T) {
	store := &mockEffectiveBDStore{
		rows: map[string]*readmodels.CMEffectiveBusinessDomainDTO{
			"l1-root": {CapabilityID: "l1-root", L1CapabilityID: "l1-root", BusinessDomainID: "bd-99", BusinessDomainName: "Marketing"},
		},
	}
	bdProvider := &mockBDNameProvider{err: errors.New("lookup failed")}
	projector := NewEffectiveBusinessDomainProjector(store, bdProvider, nil)

	eventData, err := json.Marshal(events.NewCapabilityAssignedToDomain("a-1", "bd-99", "l1-root"))
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "CapabilityAssignedToDomain", eventData)
	assert.Error(t, err)
}

type mockEffectiveBDStoreWithGetErr struct {
	mockEffectiveBDStore
	getErr error
}

func (m *mockEffectiveBDStoreWithGetErr) GetByCapabilityID(ctx context.Context, capabilityID string) (*readmodels.CMEffectiveBusinessDomainDTO, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.mockEffectiveBDStore.GetByCapabilityID(ctx, capabilityID)
}

func projectEvent(t *testing.T, p *EffectiveBusinessDomainProjector, eventType string, event any) {
	t.Helper()
	eventData, err := json.Marshal(event)
	require.NoError(t, err)
	err = p.ProjectEvent(context.Background(), eventType, eventData)
	require.NoError(t, err)
}

func TestEffectiveBusinessDomainProjector_CapabilityCreated(t *testing.T) {
	t.Run("child inherits parent BD", func(t *testing.T) {
		store := &mockEffectiveBDStore{
			rows: map[string]*readmodels.CMEffectiveBusinessDomainDTO{
				"parent-l1": {
					CapabilityID:       "parent-l1",
					L1CapabilityID:     "parent-l1",
					BusinessDomainID:   "bd-1",
					BusinessDomainName: "Finance",
				},
			},
		}
		projector := NewEffectiveBusinessDomainProjector(store, nil, nil)

		projectEvent(t, projector, "CapabilityCreated",
			events.NewCapabilityCreated("child-l2", "Sub-Payments", "", "parent-l1", "L2"))

		require.Len(t, store.upserted, 1)
		assert.Equal(t, "child-l2", store.upserted[0].CapabilityID)
		assert.Equal(t, "parent-l1", store.upserted[0].L1CapabilityID)
		assert.Equal(t, "bd-1", store.upserted[0].BusinessDomainID)
		assert.Equal(t, "Finance", store.upserted[0].BusinessDomainName)
	})

	tests := []struct {
		name  string
		event events.CapabilityCreated
	}{
		{"L1 defaults to self", events.NewCapabilityCreated("cap-1", "Payments", "desc", "", "L1")},
		{"non-L1 without parent defaults to self", events.NewCapabilityCreated("orphan-l2", "Orphan", "", "", "L2")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockEffectiveBDStore{}
			projector := NewEffectiveBusinessDomainProjector(store, nil, nil)

			projectEvent(t, projector, "CapabilityCreated", tt.event)

			require.Len(t, store.upserted, 1)
			assert.Equal(t, tt.event.ID, store.upserted[0].CapabilityID)
			assert.Equal(t, tt.event.ID, store.upserted[0].L1CapabilityID)
			assert.Empty(t, store.upserted[0].BusinessDomainID)
			assert.Empty(t, store.upserted[0].BusinessDomainName)
		})
	}
}

func TestEffectiveBusinessDomainProjector_CapabilityDeleted(t *testing.T) {
	store := &mockEffectiveBDStore{
		rows: map[string]*readmodels.CMEffectiveBusinessDomainDTO{
			"cap-1": {CapabilityID: "cap-1", L1CapabilityID: "cap-1"},
		},
	}
	projector := NewEffectiveBusinessDomainProjector(store, nil, nil)

	projectEvent(t, projector, "CapabilityDeleted", events.NewCapabilityDeleted("cap-1"))

	require.Len(t, store.deleted, 1)
	assert.Equal(t, "cap-1", store.deleted[0])
}

func TestEffectiveBusinessDomainProjector_DomainAssignment(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		event     any
		wantL1    string
		wantBDID  string
	}{
		{"assign updates L1 subtree", "CapabilityAssignedToDomain",
			events.NewCapabilityAssignedToDomain("a-1", "bd-99", "l1-root"), "l1-root", "bd-99"},
		{"unassign clears BD for L1 subtree", "CapabilityUnassignedFromDomain",
			events.NewCapabilityUnassignedFromDomain("a-1", "bd-99", "l1-root"), "l1-root", ""},
		{"assign via child uses L1", "CapabilityAssignedToDomain",
			events.NewCapabilityAssignedToDomain("a-2", "bd-99", "l2-child"), "l1-root", "bd-99"},
		{"unassign via child uses L1", "CapabilityUnassignedFromDomain",
			events.NewCapabilityUnassignedFromDomain("a-2", "bd-99", "l2-child"), "l1-root", ""},
		{"assign ignores non-existent", "CapabilityAssignedToDomain",
			events.NewCapabilityAssignedToDomain("a-3", "bd-1", "non-existent"), "", ""},
		{"unassign ignores non-existent", "CapabilityUnassignedFromDomain",
			events.NewCapabilityUnassignedFromDomain("a-3", "bd-1", "non-existent"), "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockEffectiveBDStore{
				rows: map[string]*readmodels.CMEffectiveBusinessDomainDTO{
					"l1-root":  {CapabilityID: "l1-root", L1CapabilityID: "l1-root", BusinessDomainID: "bd-99", BusinessDomainName: "Marketing"},
					"l2-child": {CapabilityID: "l2-child", L1CapabilityID: "l1-root", BusinessDomainID: "bd-99", BusinessDomainName: "Marketing"},
				},
			}
			bdProvider := &mockBDNameProvider{
				domains: map[string]*readmodels.BusinessDomainDTO{"bd-99": {ID: "bd-99", Name: "Marketing"}},
			}
			projector := NewEffectiveBusinessDomainProjector(store, bdProvider, nil)

			projectEvent(t, projector, tt.eventType, tt.event)

			if tt.wantL1 == "" {
				assert.Empty(t, store.l1SubtreeUpdates)
				return
			}
			require.Len(t, store.l1SubtreeUpdates, 1)
			assert.Equal(t, tt.wantL1, store.l1SubtreeUpdates[0].l1CapabilityID)
			assert.Equal(t, tt.wantBDID, store.l1SubtreeUpdates[0].bdID)
		})
	}
}

func TestEffectiveBusinessDomainProjector_CapabilityParentChanged(t *testing.T) {
	t.Run("subtree inherits new parent BD", func(t *testing.T) {
		store := &mockEffectiveBDStore{
			rows: map[string]*readmodels.CMEffectiveBusinessDomainDTO{
				"l1-a": {CapabilityID: "l1-a", L1CapabilityID: "l1-a", BusinessDomainID: "bd-old", BusinessDomainName: "OldDomain"},
				"l2-b": {CapabilityID: "l2-b", L1CapabilityID: "l1-a", BusinessDomainID: "bd-old", BusinessDomainName: "OldDomain"},
				"l3-c": {CapabilityID: "l3-c", L1CapabilityID: "l1-a", BusinessDomainID: "bd-old", BusinessDomainName: "OldDomain"},
				"l1-d": {CapabilityID: "l1-d", L1CapabilityID: "l1-d", BusinessDomainID: "bd-new", BusinessDomainName: "NewDomain"},
			},
		}
		capProvider := &mockCapabilityChildProvider{
			children: map[string][]readmodels.CapabilityDTO{
				"l2-b": {{ID: "l3-c", Level: "L3"}},
				"l3-c": {},
			},
		}
		projector := NewEffectiveBusinessDomainProjector(store, nil, capProvider)

		projectEvent(t, projector, "CapabilityParentChanged",
			events.NewCapabilityParentChanged("l2-b", "l1-a", "l1-d", "L2", "L2"))

		l2b := store.rows["l2-b"]
		require.NotNil(t, l2b)
		assert.Equal(t, "l1-d", l2b.L1CapabilityID)
		assert.Equal(t, "bd-new", l2b.BusinessDomainID)
		assert.Equal(t, "NewDomain", l2b.BusinessDomainName)

		l3c := store.rows["l3-c"]
		require.NotNil(t, l3c)
		assert.Equal(t, "l1-d", l3c.L1CapabilityID)
		assert.Equal(t, "bd-new", l3c.BusinessDomainID)
		assert.Equal(t, "NewDomain", l3c.BusinessDomainName)

		l1a := store.rows["l1-a"]
		assert.Equal(t, "bd-old", l1a.BusinessDomainID, "L1-A should not be affected")
	})

	t.Run("becomes L1 clears BD", func(t *testing.T) {
		store := &mockEffectiveBDStore{
			rows: map[string]*readmodels.CMEffectiveBusinessDomainDTO{
				"cap-x": {CapabilityID: "cap-x", L1CapabilityID: "old-l1", BusinessDomainID: "bd-1", BusinessDomainName: "Finance"},
			},
		}
		capProvider := &mockCapabilityChildProvider{
			children: map[string][]readmodels.CapabilityDTO{
				"cap-x": {},
			},
		}
		projector := NewEffectiveBusinessDomainProjector(store, nil, capProvider)

		projectEvent(t, projector, "CapabilityParentChanged",
			events.NewCapabilityParentChanged("cap-x", "old-parent", "", "L2", "L1"))

		capX := store.rows["cap-x"]
		require.NotNil(t, capX)
		assert.Equal(t, "cap-x", capX.L1CapabilityID, "Should become its own L1")
		assert.Empty(t, capX.BusinessDomainID, "Should have no BD since no assignment exists for new L1")
	})
}

func TestEffectiveBusinessDomainProjector_CapabilityLevelChanged_ToL1(t *testing.T) {
	store := &mockEffectiveBDStore{
		rows: map[string]*readmodels.CMEffectiveBusinessDomainDTO{
			"cap-x":  {CapabilityID: "cap-x", L1CapabilityID: "old-l1", BusinessDomainID: "bd-1", BusinessDomainName: "Finance"},
			"cap-ch": {CapabilityID: "cap-ch", L1CapabilityID: "old-l1", BusinessDomainID: "bd-1", BusinessDomainName: "Finance"},
		},
	}
	capProvider := &mockCapabilityChildProvider{
		children: map[string][]readmodels.CapabilityDTO{
			"cap-x":  {{ID: "cap-ch", Level: "L2"}},
			"cap-ch": {},
		},
	}
	projector := NewEffectiveBusinessDomainProjector(store, nil, capProvider)

	projectEvent(t, projector, "CapabilityLevelChanged",
		events.NewCapabilityLevelChanged("cap-x", "L2", "L1"))

	capX := store.rows["cap-x"]
	require.NotNil(t, capX)
	assert.Equal(t, "cap-x", capX.L1CapabilityID, "Should become its own L1")
	assert.Empty(t, capX.BusinessDomainID)

	capCh := store.rows["cap-ch"]
	require.NotNil(t, capCh)
	assert.Equal(t, "cap-x", capCh.L1CapabilityID, "Child should inherit new L1")
}
