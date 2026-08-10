package projectors

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/architecturedirection/application/readmodels"
	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	authPL "easi/backend/internal/auth/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
)

type mockCapabilityJourneyReferenceStore struct {
	cachedNames     map[string]string
	capabilityName  []readmodels.CapabilityID
	capabilityStale []readmodels.CapabilityID
	componentName   []readmodels.ComponentID
	componentStale  []readmodels.ComponentID
	domainName      []readmodels.BusinessDomainID
	domainStale     []readmodels.BusinessDomainID
	plannedByName   []string
}

func newMockCapabilityJourneyReferenceStore() *mockCapabilityJourneyReferenceStore {
	return &mockCapabilityJourneyReferenceStore{cachedNames: map[string]string{}}
}

func (m *mockCapabilityJourneyReferenceStore) CacheReferenceName(_ context.Context, entity readmodels.ReferenceEntity, entityID, name string) error {
	m.cachedNames[string(entity)+"|"+entityID] = name
	return nil
}

func (m *mockCapabilityJourneyReferenceStore) UpdateCapabilityName(_ context.Context, capabilityID readmodels.CapabilityID, _ string) error {
	m.capabilityName = append(m.capabilityName, capabilityID)
	return nil
}

func (m *mockCapabilityJourneyReferenceStore) MarkCapabilityStale(_ context.Context, capabilityID readmodels.CapabilityID) error {
	m.capabilityStale = append(m.capabilityStale, capabilityID)
	return nil
}

func (m *mockCapabilityJourneyReferenceStore) UpdateComponentName(_ context.Context, componentID readmodels.ComponentID, _ string) error {
	m.componentName = append(m.componentName, componentID)
	return nil
}

func (m *mockCapabilityJourneyReferenceStore) MarkComponentStale(_ context.Context, componentID readmodels.ComponentID) error {
	m.componentStale = append(m.componentStale, componentID)
	return nil
}

func (m *mockCapabilityJourneyReferenceStore) UpdateDomainName(_ context.Context, domainID readmodels.BusinessDomainID, _ string) error {
	m.domainName = append(m.domainName, domainID)
	return nil
}

func (m *mockCapabilityJourneyReferenceStore) MarkDomainStale(_ context.Context, domainID readmodels.BusinessDomainID) error {
	m.domainStale = append(m.domainStale, domainID)
	return nil
}

func (m *mockCapabilityJourneyReferenceStore) UpdatePlannedByName(_ context.Context, email, _ string) error {
	m.plannedByName = append(m.plannedByName, email)
	return nil
}

func projectJourneyReferenceEvent(t *testing.T, projector *CapabilityJourneyReferenceProjector, eventType string, payload map[string]interface{}) error {
	t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	return projector.ProjectEvent(context.Background(), eventType, data)
}

func TestCapabilityJourneyReferenceProjector_CapabilityCreated_UpdatesName(t *testing.T) {
	store := newMockCapabilityJourneyReferenceStore()
	projector := NewCapabilityJourneyReferenceProjector(store)
	capID := uuid.New().String()

	require.NoError(t, projectJourneyReferenceEvent(t, projector, cmPL.CapabilityCreated, map[string]interface{}{"id": capID, "name": "Booking management"}))

	assert.Equal(t, "Booking management", store.cachedNames["capability|"+capID])
	assert.Contains(t, store.capabilityName, readmodels.CapabilityID(capID))
}

func TestCapabilityJourneyReferenceProjector_CapabilityDeleted_MarksStale(t *testing.T) {
	store := newMockCapabilityJourneyReferenceStore()
	projector := NewCapabilityJourneyReferenceProjector(store)
	capID := uuid.New().String()

	require.NoError(t, projectJourneyReferenceEvent(t, projector, cmPL.CapabilityDeleted, map[string]interface{}{"id": capID}))

	assert.Contains(t, store.capabilityStale, readmodels.CapabilityID(capID))
}

func TestCapabilityJourneyReferenceProjector_ApplicationComponentCreated_UpdatesName(t *testing.T) {
	store := newMockCapabilityJourneyReferenceStore()
	projector := NewCapabilityJourneyReferenceProjector(store)
	compID := uuid.New().String()

	require.NoError(t, projectJourneyReferenceEvent(t, projector, amPL.ApplicationComponentCreated, map[string]interface{}{"id": compID, "name": "Phoenix"}))

	assert.Equal(t, "Phoenix", store.cachedNames["application|"+compID])
	assert.Contains(t, store.componentName, readmodels.ComponentID(compID))
}

func TestCapabilityJourneyReferenceProjector_ApplicationComponentDeleted_MarksStale(t *testing.T) {
	store := newMockCapabilityJourneyReferenceStore()
	projector := NewCapabilityJourneyReferenceProjector(store)
	compID := uuid.New().String()

	require.NoError(t, projectJourneyReferenceEvent(t, projector, amPL.ApplicationComponentDeleted, map[string]interface{}{"id": compID, "name": "Phoenix"}))

	assert.Contains(t, store.componentStale, readmodels.ComponentID(compID))
}

func TestCapabilityJourneyReferenceProjector_BusinessDomainCreated_UpdatesName(t *testing.T) {
	store := newMockCapabilityJourneyReferenceStore()
	projector := NewCapabilityJourneyReferenceProjector(store)
	domID := uuid.New().String()

	require.NoError(t, projectJourneyReferenceEvent(t, projector, cmPL.BusinessDomainCreated, map[string]interface{}{"id": domID, "name": "Group functions"}))

	assert.Equal(t, "Group functions", store.cachedNames["business_domain|"+domID])
	assert.Contains(t, store.domainName, readmodels.BusinessDomainID(domID))
}

func TestCapabilityJourneyReferenceProjector_BusinessDomainDeleted_MarksStale(t *testing.T) {
	store := newMockCapabilityJourneyReferenceStore()
	projector := NewCapabilityJourneyReferenceProjector(store)
	domID := uuid.New().String()

	require.NoError(t, projectJourneyReferenceEvent(t, projector, cmPL.BusinessDomainDeleted, map[string]interface{}{"id": domID}))

	assert.Contains(t, store.domainStale, readmodels.BusinessDomainID(domID))
}

func TestCapabilityJourneyReferenceProjector_UserCreated_CachesPlannedByName(t *testing.T) {
	store := newMockCapabilityJourneyReferenceStore()
	projector := NewCapabilityJourneyReferenceProjector(store)

	require.NoError(t, projectJourneyReferenceEvent(t, projector, authPL.UserCreated, map[string]interface{}{"email": "architect@example.com", "name": "Ada Architect"}))

	assert.Contains(t, store.plannedByName, "architect@example.com")
}

func TestCapabilityJourneyReferenceProjector_NeverSubscribesToSystemRealizationDeleted(t *testing.T) {
	store := newMockCapabilityJourneyReferenceStore()
	projector := NewCapabilityJourneyReferenceProjector(store)

	require.NoError(t, projectJourneyReferenceEvent(t, projector, cmPL.SystemRealizationDeleted, map[string]interface{}{"id": uuid.New().String()}))

	assert.Empty(t, store.capabilityStale)
	assert.Empty(t, store.componentStale)
}

func TestCapabilityJourneyReferenceProjector_UnknownEvent_Ignored(t *testing.T) {
	store := newMockCapabilityJourneyReferenceStore()
	projector := NewCapabilityJourneyReferenceProjector(store)

	require.NoError(t, projector.ProjectEvent(context.Background(), "SomethingElseHappened", []byte(`{}`)))
	assert.Empty(t, store.cachedNames)
}
