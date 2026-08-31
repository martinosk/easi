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
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
)

type mockReferenceCacheStore struct {
	saved   map[string]string
	removed []string
}

func newMockReferenceCacheStore() *mockReferenceCacheStore {
	return &mockReferenceCacheStore{saved: map[string]string{}}
}

func (m *mockReferenceCacheStore) SaveReference(_ context.Context, entity readmodels.ReferenceEntity, entityID, name string) error {
	m.saved[string(entity)+"|"+entityID] = name
	return nil
}

func (m *mockReferenceCacheStore) RemoveReference(_ context.Context, entity readmodels.ReferenceEntity, entityID string) error {
	m.removed = append(m.removed, string(entity)+"|"+entityID)
	return nil
}

func projectReferenceCacheEvent(t *testing.T, projector *ReferenceCacheProjector, eventType string, payload map[string]interface{}) error {
	t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	return projector.ProjectEvent(context.Background(), eventType, data)
}

func TestReferenceCacheProjector_ApplicationComponentCreated_CachesComponent(t *testing.T) {
	store := newMockReferenceCacheStore()
	projector := NewReferenceCacheProjector(store)
	componentID := uuid.New().String()

	require.NoError(t, projectReferenceCacheEvent(t, projector, amPL.ApplicationComponentCreated,
		map[string]interface{}{"id": componentID, "name": "Phoenix"}))

	assert.Equal(t, "Phoenix", store.saved["application|"+componentID])
	assert.Empty(t, store.removed)
}

func TestReferenceCacheProjector_ApplicationComponentUpdated_CachesNewName(t *testing.T) {
	store := newMockReferenceCacheStore()
	projector := NewReferenceCacheProjector(store)
	componentID := uuid.New().String()

	require.NoError(t, projectReferenceCacheEvent(t, projector, amPL.ApplicationComponentUpdated,
		map[string]interface{}{"id": componentID, "name": "Phoenix v2"}))

	assert.Equal(t, "Phoenix v2", store.saved["application|"+componentID])
}

func TestReferenceCacheProjector_ApplicationComponentDeleted_RemovesComponent(t *testing.T) {
	store := newMockReferenceCacheStore()
	projector := NewReferenceCacheProjector(store)
	componentID := uuid.New().String()

	require.NoError(t, projectReferenceCacheEvent(t, projector, amPL.ApplicationComponentDeleted,
		map[string]interface{}{"id": componentID, "name": "Phoenix"}))

	assert.Equal(t, []string{"application|" + componentID}, store.removed)
	assert.Empty(t, store.saved)
}

func TestReferenceCacheProjector_BusinessDomainCreated_CachesDomain(t *testing.T) {
	store := newMockReferenceCacheStore()
	projector := NewReferenceCacheProjector(store)
	domainID := uuid.New().String()

	require.NoError(t, projectReferenceCacheEvent(t, projector, cmPL.BusinessDomainCreated,
		map[string]interface{}{"id": domainID, "name": "Group functions"}))

	assert.Equal(t, "Group functions", store.saved["business_domain|"+domainID])
}

func TestReferenceCacheProjector_BusinessDomainUpdated_CachesNewName(t *testing.T) {
	store := newMockReferenceCacheStore()
	projector := NewReferenceCacheProjector(store)
	domainID := uuid.New().String()

	require.NoError(t, projectReferenceCacheEvent(t, projector, cmPL.BusinessDomainUpdated,
		map[string]interface{}{"id": domainID, "name": "Group Functions"}))

	assert.Equal(t, "Group Functions", store.saved["business_domain|"+domainID])
}

func TestReferenceCacheProjector_BusinessDomainDeleted_RemovesDomain(t *testing.T) {
	store := newMockReferenceCacheStore()
	projector := NewReferenceCacheProjector(store)
	domainID := uuid.New().String()

	require.NoError(t, projectReferenceCacheEvent(t, projector, cmPL.BusinessDomainDeleted,
		map[string]interface{}{"id": domainID}))

	assert.Equal(t, []string{"business_domain|" + domainID}, store.removed)
}

func TestReferenceCacheProjector_EmptyID_Ignored(t *testing.T) {
	store := newMockReferenceCacheStore()
	projector := NewReferenceCacheProjector(store)

	require.NoError(t, projectReferenceCacheEvent(t, projector, amPL.ApplicationComponentCreated,
		map[string]interface{}{"name": "Nameless"}))

	assert.Empty(t, store.saved)
	assert.Empty(t, store.removed)
}

func TestReferenceCacheProjector_UnknownEvent_Ignored(t *testing.T) {
	store := newMockReferenceCacheStore()
	projector := NewReferenceCacheProjector(store)

	require.NoError(t, projector.ProjectEvent(context.Background(), "SomethingElseHappened", []byte(`{}`)))

	assert.Empty(t, store.saved)
	assert.Empty(t, store.removed)
}
