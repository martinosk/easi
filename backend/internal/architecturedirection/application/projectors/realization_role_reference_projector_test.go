package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type realizationRoleNameUpdate struct {
	id   string
	name string
}

type mockRealizationRoleReferenceStore struct {
	deletedByCapabilityID  []string
	deletedByComponentID   []string
	cachedCapabilityNames  []realizationRoleNameUpdate
	updatedCapabilityNames []realizationRoleNameUpdate
	cachedComponentNames   []realizationRoleNameUpdate
	updatedComponentNames  []realizationRoleNameUpdate
	err                    error
}

func (m *mockRealizationRoleReferenceStore) DeleteByCapabilityID(_ context.Context, capabilityID string) error {
	if m.err != nil {
		return m.err
	}
	m.deletedByCapabilityID = append(m.deletedByCapabilityID, capabilityID)
	return nil
}

func (m *mockRealizationRoleReferenceStore) DeleteByComponentID(_ context.Context, componentID string) error {
	if m.err != nil {
		return m.err
	}
	m.deletedByComponentID = append(m.deletedByComponentID, componentID)
	return nil
}

func (m *mockRealizationRoleReferenceStore) CacheCapabilityName(_ context.Context, capabilityID, name string) error {
	if m.err != nil {
		return m.err
	}
	m.cachedCapabilityNames = append(m.cachedCapabilityNames, realizationRoleNameUpdate{capabilityID, name})
	return nil
}

func (m *mockRealizationRoleReferenceStore) UpdateCapabilityName(_ context.Context, capabilityID, name string) error {
	if m.err != nil {
		return m.err
	}
	m.updatedCapabilityNames = append(m.updatedCapabilityNames, realizationRoleNameUpdate{capabilityID, name})
	return nil
}

func (m *mockRealizationRoleReferenceStore) CacheComponentName(_ context.Context, componentID, name string) error {
	if m.err != nil {
		return m.err
	}
	m.cachedComponentNames = append(m.cachedComponentNames, realizationRoleNameUpdate{componentID, name})
	return nil
}

func (m *mockRealizationRoleReferenceStore) UpdateComponentName(_ context.Context, componentID, name string) error {
	if m.err != nil {
		return m.err
	}
	m.updatedComponentNames = append(m.updatedComponentNames, realizationRoleNameUpdate{componentID, name})
	return nil
}

func TestRealizationRoleReferenceProjector_CapabilityDeleted_DeletesByCapabilityID(t *testing.T) {
	store := &mockRealizationRoleReferenceStore{}
	projector := NewRealizationRoleReferenceProjector(store)

	id := uuid.New().String()
	payload, _ := json.Marshal(map[string]string{"id": id})

	require.NoError(t, projector.ProjectEvent(context.Background(), cmPL.CapabilityDeleted, payload))

	assert.Equal(t, []string{id}, store.deletedByCapabilityID)
}

func TestRealizationRoleReferenceProjector_ApplicationComponentDeleted_DeletesByComponentID(t *testing.T) {
	store := &mockRealizationRoleReferenceStore{}
	projector := NewRealizationRoleReferenceProjector(store)

	id := uuid.New().String()
	payload, _ := json.Marshal(map[string]string{"id": id})

	require.NoError(t, projector.ProjectEvent(context.Background(), amPL.ApplicationComponentDeleted, payload))

	assert.Equal(t, []string{id}, store.deletedByComponentID)
}

func TestRealizationRoleReferenceProjector_DoesNotSubscribeToSystemRealizationDeleted(t *testing.T) {
	store := &mockRealizationRoleReferenceStore{}
	projector := NewRealizationRoleReferenceProjector(store)

	id := uuid.New().String()
	payload, _ := json.Marshal(map[string]string{"id": id})

	require.NoError(t, projector.ProjectEvent(context.Background(), cmPL.SystemRealizationDeleted, payload))

	assert.Empty(t, store.deletedByCapabilityID, "the deletion reactor owns SystemRealizationDeleted; a read-side delete here would race its pair lookup")
	assert.Empty(t, store.deletedByComponentID)
}

type realizationRoleNameChangeCase struct {
	entity     string
	eventTypes []string
	entityName string
	cached     func(*mockRealizationRoleReferenceStore) []realizationRoleNameUpdate
	updated    func(*mockRealizationRoleReferenceStore) []realizationRoleNameUpdate
}

func TestRealizationRoleReferenceProjector_NameChange_CachesAndUpdates(t *testing.T) {
	cases := []realizationRoleNameChangeCase{
		{
			entity:     "capability",
			eventTypes: []string{cmPL.CapabilityCreated, cmPL.CapabilityUpdated},
			entityName: "Booking management",
			cached:     func(s *mockRealizationRoleReferenceStore) []realizationRoleNameUpdate { return s.cachedCapabilityNames },
			updated: func(s *mockRealizationRoleReferenceStore) []realizationRoleNameUpdate {
				return s.updatedCapabilityNames
			},
		},
		{
			entity:     "component",
			eventTypes: []string{amPL.ApplicationComponentCreated, amPL.ApplicationComponentUpdated},
			entityName: "Seabook",
			cached:     func(s *mockRealizationRoleReferenceStore) []realizationRoleNameUpdate { return s.cachedComponentNames },
			updated:    func(s *mockRealizationRoleReferenceStore) []realizationRoleNameUpdate { return s.updatedComponentNames },
		},
	}

	for _, tc := range cases {
		for _, eventType := range tc.eventTypes {
			t.Run(tc.entity+"/"+eventType, func(t *testing.T) {
				store := &mockRealizationRoleReferenceStore{}
				projector := NewRealizationRoleReferenceProjector(store)

				id := uuid.New().String()
				payload, _ := json.Marshal(map[string]string{"id": id, "name": tc.entityName})

				require.NoError(t, projector.ProjectEvent(context.Background(), eventType, payload))

				want := []realizationRoleNameUpdate{{id, tc.entityName}}
				assert.Equal(t, want, tc.cached(store))
				assert.Equal(t, want, tc.updated(store))
			})
		}
	}
}

func TestRealizationRoleReferenceProjector_UnknownEvent_NoOp(t *testing.T) {
	store := &mockRealizationRoleReferenceStore{}
	projector := NewRealizationRoleReferenceProjector(store)

	err := projector.ProjectEvent(context.Background(), "SomeUnrelatedEvent", []byte(`{}`))

	require.NoError(t, err)
	assert.Empty(t, store.deletedByCapabilityID)
	assert.Empty(t, store.deletedByComponentID)
}

func TestRealizationRoleReferenceProjector_EmptyID_NoOp(t *testing.T) {
	store := &mockRealizationRoleReferenceStore{}
	projector := NewRealizationRoleReferenceProjector(store)

	payload, _ := json.Marshal(map[string]string{"id": ""})

	require.NoError(t, projector.ProjectEvent(context.Background(), cmPL.CapabilityDeleted, payload))

	assert.Empty(t, store.deletedByCapabilityID)
}

func TestRealizationRoleReferenceProjector_ErrorPropagation(t *testing.T) {
	store := &mockRealizationRoleReferenceStore{err: errors.New("db down")}
	projector := NewRealizationRoleReferenceProjector(store)

	payload, _ := json.Marshal(map[string]string{"id": uuid.New().String()})
	err := projector.ProjectEvent(context.Background(), cmPL.CapabilityDeleted, payload)

	assert.Error(t, err)
}

func TestRealizationRoleReferenceProjector_InvalidJSON_ReturnsError(t *testing.T) {
	store := &mockRealizationRoleReferenceStore{}
	projector := NewRealizationRoleReferenceProjector(store)

	err := projector.ProjectEvent(context.Background(), cmPL.CapabilityDeleted, []byte("invalid"))
	assert.Error(t, err)
}
