package projectors

import (
	"context"
	"testing"

	"easi/backend/internal/architecturedirection/application/readmodels"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStaleStore struct {
	marked              []readmodels.CapabilityID
	cachedNames         map[string]string
	capabilityNames     map[string]string
	domainNames         map[string]string
	capabilityDomains   map[string]string
	clearedDomains      []string
	clearedSourceDomain []readmodels.CapabilityID
	sourceDomains       map[readmodels.CapabilityID]string
}

func newMockStaleStore() *mockStaleStore {
	return &mockStaleStore{
		cachedNames:       map[string]string{},
		capabilityNames:   map[string]string{},
		domainNames:       map[string]string{},
		capabilityDomains: map[string]string{},
		sourceDomains:     map[readmodels.CapabilityID]string{},
	}
}

func (m *mockStaleStore) MarkSourceCapabilityStale(_ context.Context, id readmodels.CapabilityID) error {
	m.marked = append(m.marked, id)
	return nil
}

func (m *mockStaleStore) CacheReferenceName(_ context.Context, entityType, entityID, name string) error {
	m.cachedNames[entityType+":"+entityID] = name
	return nil
}

func (m *mockStaleStore) UpdateCapabilityName(_ context.Context, capabilityID readmodels.CapabilityID, name string) error {
	m.capabilityNames[string(capabilityID)] = name
	return nil
}

func (m *mockStaleStore) CacheCapabilityDomain(_ context.Context, capabilityID, businessDomainID string) error {
	m.capabilityDomains[capabilityID] = businessDomainID
	return nil
}

func (m *mockStaleStore) ClearCapabilityDomain(_ context.Context, capabilityID string) error {
	m.clearedDomains = append(m.clearedDomains, capabilityID)
	return nil
}

func (m *mockStaleStore) UpdateSourceCapabilityDomain(_ context.Context, capabilityID readmodels.CapabilityID, businessDomainID string) error {
	m.sourceDomains[capabilityID] = businessDomainID
	return nil
}

func (m *mockStaleStore) ClearSourceCapabilityDomain(_ context.Context, capabilityID readmodels.CapabilityID) error {
	m.clearedSourceDomain = append(m.clearedSourceDomain, capabilityID)
	return nil
}

func (m *mockStaleStore) UpdateBusinessDomainName(_ context.Context, businessDomainID, name string) error {
	m.domainNames[businessDomainID] = name
	return nil
}

func TestStaleReferenceProjector_CapabilityDeleted_MarksStale(t *testing.T) {
	store := newMockStaleStore()
	projector := NewStaleReferenceProjector(store)

	id := uuid.New().String()
	require.NoError(t, projector.ProjectEvent(context.Background(), cmPL.CapabilityDeleted,
		[]byte(`{"id":"`+id+`"}`)))
	assert.Equal(t, []readmodels.CapabilityID{readmodels.CapabilityID(id)}, store.marked)
}

func TestStaleReferenceProjector_CapabilityCreated_CachesAndUpdatesName(t *testing.T) {
	store := newMockStaleStore()
	projector := NewStaleReferenceProjector(store)

	id := uuid.New().String()
	require.NoError(t, projector.ProjectEvent(context.Background(), cmPL.CapabilityCreated,
		[]byte(`{"id":"`+id+`","name":"Payroll"}`)))
	assert.Equal(t, "Payroll", store.cachedNames["capability:"+id])
	assert.Equal(t, "Payroll", store.capabilityNames[id])
}

func TestStaleReferenceProjector_CapabilityUpdated_CachesAndUpdatesName(t *testing.T) {
	store := newMockStaleStore()
	projector := NewStaleReferenceProjector(store)

	id := uuid.New().String()
	require.NoError(t, projector.ProjectEvent(context.Background(), cmPL.CapabilityUpdated,
		[]byte(`{"id":"`+id+`","name":"Payroll (Group)"}`)))
	assert.Equal(t, "Payroll (Group)", store.cachedNames["capability:"+id])
	assert.Equal(t, "Payroll (Group)", store.capabilityNames[id])
}

func TestStaleReferenceProjector_BusinessDomainCreated_CachesAndUpdatesName(t *testing.T) {
	store := newMockStaleStore()
	projector := NewStaleReferenceProjector(store)

	id := uuid.New().String()
	require.NoError(t, projector.ProjectEvent(context.Background(), cmPL.BusinessDomainCreated,
		[]byte(`{"id":"`+id+`","name":"Passenger"}`)))
	assert.Equal(t, "Passenger", store.cachedNames["business_domain:"+id])
	assert.Equal(t, "Passenger", store.domainNames[id])
}

func TestStaleReferenceProjector_BusinessDomainUpdated_CachesAndUpdatesName(t *testing.T) {
	store := newMockStaleStore()
	projector := NewStaleReferenceProjector(store)

	id := uuid.New().String()
	require.NoError(t, projector.ProjectEvent(context.Background(), cmPL.BusinessDomainUpdated,
		[]byte(`{"id":"`+id+`","name":"Passenger Operations"}`)))
	assert.Equal(t, "Passenger Operations", store.cachedNames["business_domain:"+id])
	assert.Equal(t, "Passenger Operations", store.domainNames[id])
}

func TestStaleReferenceProjector_CapabilityAssignedToDomain_CachesAndUpdates(t *testing.T) {
	store := newMockStaleStore()
	projector := NewStaleReferenceProjector(store)

	capID := uuid.New().String()
	domainID := uuid.New().String()
	require.NoError(t, projector.ProjectEvent(context.Background(), cmPL.CapabilityAssignedToDomain,
		[]byte(`{"capabilityId":"`+capID+`","businessDomainId":"`+domainID+`"}`)))
	assert.Equal(t, domainID, store.capabilityDomains[capID])
	assert.Equal(t, domainID, store.sourceDomains[readmodels.CapabilityID(capID)])
}

func TestStaleReferenceProjector_CapabilityUnassignedFromDomain_ClearsAssignment(t *testing.T) {
	store := newMockStaleStore()
	projector := NewStaleReferenceProjector(store)

	capID := uuid.New().String()
	require.NoError(t, projector.ProjectEvent(context.Background(), cmPL.CapabilityUnassignedFromDomain,
		[]byte(`{"capabilityId":"`+capID+`"}`)))
	assert.Contains(t, store.clearedDomains, capID)
	assert.Contains(t, store.clearedSourceDomain, readmodels.CapabilityID(capID))
}

func TestStaleReferenceProjector_UnknownEvent_NoOp(t *testing.T) {
	store := newMockStaleStore()
	projector := NewStaleReferenceProjector(store)

	require.NoError(t, projector.ProjectEvent(context.Background(), "SomeUnrelatedEvent",
		[]byte(`{"id":"x"}`)))
	assert.Empty(t, store.marked)
	assert.Empty(t, store.cachedNames)
}
