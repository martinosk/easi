package projectors

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	authPL "easi/backend/internal/auth/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type timeAssessmentNameUpdate struct {
	id   string
	name string
}

type mockTimeAssessmentReferenceStore struct {
	deletedByCapabilityID  []string
	deletedByComponentID   []string
	cachedCapabilityNames  []timeAssessmentNameUpdate
	updatedCapabilityNames []timeAssessmentNameUpdate
	cachedComponentNames   []timeAssessmentNameUpdate
	updatedComponentNames  []timeAssessmentNameUpdate
	cachedUserNames        []timeAssessmentNameUpdate
	err                    error
}

func (m *mockTimeAssessmentReferenceStore) DeleteByCapabilityID(_ context.Context, capabilityID string) error {
	if m.err != nil {
		return m.err
	}
	m.deletedByCapabilityID = append(m.deletedByCapabilityID, capabilityID)
	return nil
}

func (m *mockTimeAssessmentReferenceStore) DeleteByComponentID(_ context.Context, componentID string) error {
	if m.err != nil {
		return m.err
	}
	m.deletedByComponentID = append(m.deletedByComponentID, componentID)
	return nil
}

func (m *mockTimeAssessmentReferenceStore) CacheCapabilityName(_ context.Context, capabilityID, name string) error {
	if m.err != nil {
		return m.err
	}
	m.cachedCapabilityNames = append(m.cachedCapabilityNames, timeAssessmentNameUpdate{capabilityID, name})
	return nil
}

func (m *mockTimeAssessmentReferenceStore) UpdateCapabilityName(_ context.Context, capabilityID, name string) error {
	if m.err != nil {
		return m.err
	}
	m.updatedCapabilityNames = append(m.updatedCapabilityNames, timeAssessmentNameUpdate{capabilityID, name})
	return nil
}

func (m *mockTimeAssessmentReferenceStore) CacheComponentName(_ context.Context, componentID, name string) error {
	if m.err != nil {
		return m.err
	}
	m.cachedComponentNames = append(m.cachedComponentNames, timeAssessmentNameUpdate{componentID, name})
	return nil
}

func (m *mockTimeAssessmentReferenceStore) UpdateComponentName(_ context.Context, componentID, name string) error {
	if m.err != nil {
		return m.err
	}
	m.updatedComponentNames = append(m.updatedComponentNames, timeAssessmentNameUpdate{componentID, name})
	return nil
}

func (m *mockTimeAssessmentReferenceStore) CacheUserName(_ context.Context, email, name string) error {
	if m.err != nil {
		return m.err
	}
	m.cachedUserNames = append(m.cachedUserNames, timeAssessmentNameUpdate{email, name})
	return nil
}

func TestTimeAssessmentReferenceProjector_DoesNotSubscribeToSystemRealizationDeleted(t *testing.T) {
	store := &mockTimeAssessmentReferenceStore{}
	projector := NewTimeAssessmentReferenceProjector(store)

	id := uuid.New().String()
	payload, _ := json.Marshal(map[string]string{"id": id})

	require.NoError(t, projector.ProjectEvent(context.Background(), cmPL.SystemRealizationDeleted, payload))

	assert.Empty(t, store.deletedByCapabilityID, "the deletion reactor owns SystemRealizationDeleted; a read-side delete here would race its pair lookup")
	assert.Empty(t, store.deletedByComponentID)
}

func TestTimeAssessmentReferenceProjector_CapabilityDeleted_DeletesByCapabilityID(t *testing.T) {
	store := &mockTimeAssessmentReferenceStore{}
	projector := NewTimeAssessmentReferenceProjector(store)

	id := uuid.New().String()
	payload, _ := json.Marshal(map[string]string{"id": id})

	require.NoError(t, projector.ProjectEvent(context.Background(), cmPL.CapabilityDeleted, payload))

	assert.Equal(t, []string{id}, store.deletedByCapabilityID)
}

func TestTimeAssessmentReferenceProjector_ApplicationComponentDeleted_DeletesByComponentID(t *testing.T) {
	store := &mockTimeAssessmentReferenceStore{}
	projector := NewTimeAssessmentReferenceProjector(store)

	id := uuid.New().String()
	payload, _ := json.Marshal(map[string]string{"id": id})

	require.NoError(t, projector.ProjectEvent(context.Background(), amPL.ApplicationComponentDeleted, payload))

	assert.Equal(t, []string{id}, store.deletedByComponentID)
}

type timeAssessmentNameChangeCase struct {
	entity     string
	eventTypes []string
	entityName string
	cached     func(*mockTimeAssessmentReferenceStore) []timeAssessmentNameUpdate
	updated    func(*mockTimeAssessmentReferenceStore) []timeAssessmentNameUpdate
}

func TestTimeAssessmentReferenceProjector_NameChange_CachesAndUpdates(t *testing.T) {
	cases := []timeAssessmentNameChangeCase{
		{
			entity:     "capability",
			eventTypes: []string{cmPL.CapabilityCreated, cmPL.CapabilityUpdated},
			entityName: "Booking management",
			cached:     func(s *mockTimeAssessmentReferenceStore) []timeAssessmentNameUpdate { return s.cachedCapabilityNames },
			updated:    func(s *mockTimeAssessmentReferenceStore) []timeAssessmentNameUpdate { return s.updatedCapabilityNames },
		},
		{
			entity:     "component",
			eventTypes: []string{amPL.ApplicationComponentCreated, amPL.ApplicationComponentUpdated},
			entityName: "Seabook",
			cached:     func(s *mockTimeAssessmentReferenceStore) []timeAssessmentNameUpdate { return s.cachedComponentNames },
			updated:    func(s *mockTimeAssessmentReferenceStore) []timeAssessmentNameUpdate { return s.updatedComponentNames },
		},
	}

	for _, tc := range cases {
		for _, eventType := range tc.eventTypes {
			t.Run(tc.entity+"/"+eventType, func(t *testing.T) {
				store := &mockTimeAssessmentReferenceStore{}
				projector := NewTimeAssessmentReferenceProjector(store)

				id := uuid.New().String()
				payload, _ := json.Marshal(map[string]string{"id": id, "name": tc.entityName})

				require.NoError(t, projector.ProjectEvent(context.Background(), eventType, payload))

				want := []timeAssessmentNameUpdate{{id, tc.entityName}}
				assert.Equal(t, want, tc.cached(store))
				assert.Equal(t, want, tc.updated(store))
			})
		}
	}
}

func TestTimeAssessmentReferenceProjector_UserCreated_CachesDisplayName(t *testing.T) {
	store := &mockTimeAssessmentReferenceStore{}
	projector := NewTimeAssessmentReferenceProjector(store)

	payload, _ := json.Marshal(map[string]string{"email": "architect@example.com", "name": "Ada Architect"})

	require.NoError(t, projector.ProjectEvent(context.Background(), authPL.UserCreated, payload))

	assert.Equal(t, []timeAssessmentNameUpdate{{"architect@example.com", "Ada Architect"}}, store.cachedUserNames)
}

func TestTimeAssessmentReferenceProjector_UnknownEvent_NoOp(t *testing.T) {
	store := &mockTimeAssessmentReferenceStore{}
	projector := NewTimeAssessmentReferenceProjector(store)

	err := projector.ProjectEvent(context.Background(), "SomeUnrelatedEvent", []byte(`{}`))

	require.NoError(t, err)
	assert.Empty(t, store.deletedByCapabilityID)
	assert.Empty(t, store.deletedByComponentID)
}

func TestTimeAssessmentReferenceProjector_EmptyID_NoOp(t *testing.T) {
	store := &mockTimeAssessmentReferenceStore{}
	projector := NewTimeAssessmentReferenceProjector(store)

	payload, _ := json.Marshal(map[string]string{"id": ""})

	require.NoError(t, projector.ProjectEvent(context.Background(), cmPL.CapabilityDeleted, payload))

	assert.Empty(t, store.deletedByCapabilityID)
}

func TestTimeAssessmentReferenceProjector_ErrorPropagation(t *testing.T) {
	store := &mockTimeAssessmentReferenceStore{err: errors.New("db down")}
	projector := NewTimeAssessmentReferenceProjector(store)

	payload, _ := json.Marshal(map[string]string{"id": uuid.New().String()})
	err := projector.ProjectEvent(context.Background(), cmPL.CapabilityDeleted, payload)

	assert.Error(t, err)
}

func TestTimeAssessmentReferenceProjector_InvalidJSON_ReturnsError(t *testing.T) {
	store := &mockTimeAssessmentReferenceStore{}
	projector := NewTimeAssessmentReferenceProjector(store)

	err := projector.ProjectEvent(context.Background(), cmPL.CapabilityDeleted, []byte("invalid"))
	assert.Error(t, err)
}
