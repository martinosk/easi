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

type mockRealizationCacheStore struct {
	saved               []readmodels.DirectRealizationDTO
	removed             []readmodels.RealizationID
	removedByCapability []readmodels.CapabilityID
	removedByComponent  []readmodels.ComponentID
}

func (m *mockRealizationCacheStore) SaveDirectRealization(_ context.Context, dto readmodels.DirectRealizationDTO) error {
	m.saved = append(m.saved, dto)
	return nil
}

func (m *mockRealizationCacheStore) RemoveRealization(_ context.Context, realizationID readmodels.RealizationID) error {
	m.removed = append(m.removed, realizationID)
	return nil
}

func (m *mockRealizationCacheStore) RemoveRealizationsOfCapability(_ context.Context, capabilityID readmodels.CapabilityID) error {
	m.removedByCapability = append(m.removedByCapability, capabilityID)
	return nil
}

func (m *mockRealizationCacheStore) RemoveRealizationsOfComponent(_ context.Context, componentID readmodels.ComponentID) error {
	m.removedByComponent = append(m.removedByComponent, componentID)
	return nil
}

func projectRealizationCacheEvent(t *testing.T, projector *RealizationCacheProjector, eventType string, payload map[string]interface{}) error {
	t.Helper()
	data, err := json.Marshal(payload)
	require.NoError(t, err)
	return projector.ProjectEvent(context.Background(), eventType, data)
}

func TestRealizationCacheProjector_SystemLinkedToCapability_CachesDirectRealization(t *testing.T) {
	store := &mockRealizationCacheStore{}
	projector := NewRealizationCacheProjector(store)
	realizationID, capabilityID, componentID := uuid.New().String(), uuid.New().String(), uuid.New().String()

	require.NoError(t, projectRealizationCacheEvent(t, projector, cmPL.SystemLinkedToCapability, map[string]interface{}{
		"id": realizationID, "capabilityId": capabilityID, "componentId": componentID,
	}))

	assert.Equal(t, []readmodels.DirectRealizationDTO{{
		RealizationID: readmodels.RealizationID(realizationID),
		CapabilityID:  readmodels.CapabilityID(capabilityID),
		ComponentID:   readmodels.ComponentID(componentID),
	}}, store.saved)
}

func TestRealizationCacheProjector_SystemLinkedToCapability_IncompletePayloadIgnored(t *testing.T) {
	store := &mockRealizationCacheStore{}
	projector := NewRealizationCacheProjector(store)

	require.NoError(t, projectRealizationCacheEvent(t, projector, cmPL.SystemLinkedToCapability, map[string]interface{}{
		"id": uuid.New().String(), "capabilityId": uuid.New().String(),
	}))

	assert.Empty(t, store.saved)
}

func TestRealizationCacheProjector_SystemRealizationDeleted_RemovesRealization(t *testing.T) {
	store := &mockRealizationCacheStore{}
	projector := NewRealizationCacheProjector(store)
	realizationID := uuid.New().String()

	require.NoError(t, projectRealizationCacheEvent(t, projector, cmPL.SystemRealizationDeleted,
		map[string]interface{}{"id": realizationID}))

	assert.Equal(t, []readmodels.RealizationID{readmodels.RealizationID(realizationID)}, store.removed)
}

func TestRealizationCacheProjector_CapabilityDeleted_RemovesRealizationsOfCapability(t *testing.T) {
	store := &mockRealizationCacheStore{}
	projector := NewRealizationCacheProjector(store)
	capabilityID := uuid.New().String()

	require.NoError(t, projectRealizationCacheEvent(t, projector, cmPL.CapabilityDeleted,
		map[string]interface{}{"id": capabilityID}))

	assert.Equal(t, []readmodels.CapabilityID{readmodels.CapabilityID(capabilityID)}, store.removedByCapability)
}

func TestRealizationCacheProjector_ApplicationComponentDeleted_RemovesRealizationsOfComponent(t *testing.T) {
	store := &mockRealizationCacheStore{}
	projector := NewRealizationCacheProjector(store)
	componentID := uuid.New().String()

	require.NoError(t, projectRealizationCacheEvent(t, projector, amPL.ApplicationComponentDeleted,
		map[string]interface{}{"id": componentID, "name": "Phoenix"}))

	assert.Equal(t, []readmodels.ComponentID{readmodels.ComponentID(componentID)}, store.removedByComponent)
}

func TestRealizationCacheProjector_EmptyID_Ignored(t *testing.T) {
	store := &mockRealizationCacheStore{}
	projector := NewRealizationCacheProjector(store)

	require.NoError(t, projectRealizationCacheEvent(t, projector, cmPL.SystemRealizationDeleted, map[string]interface{}{"id": ""}))

	assert.Empty(t, store.removed)
}

func TestRealizationCacheProjector_UnknownEvent_Ignored(t *testing.T) {
	store := &mockRealizationCacheStore{}
	projector := NewRealizationCacheProjector(store)

	require.NoError(t, projector.ProjectEvent(context.Background(), "SomethingElseHappened", []byte(`{}`)))

	assert.Empty(t, store.saved)
	assert.Empty(t, store.removed)
	assert.Empty(t, store.removedByCapability)
	assert.Empty(t, store.removedByComponent)
}
