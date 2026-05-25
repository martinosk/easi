package projectors

import (
	"context"
	"encoding/json"
	"testing"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type nameRecord struct {
	id   string
	name string
}

type mockStaleApplicationStore struct {
	stale   []string
	cached  []nameRecord
	updated []nameRecord
	err     error
}

func (m *mockStaleApplicationStore) MarkApplicationStale(_ context.Context, applicationID string) error {
	if m.err != nil {
		return m.err
	}
	m.stale = append(m.stale, applicationID)
	return nil
}

func (m *mockStaleApplicationStore) CacheApplicationName(_ context.Context, applicationID, name string) error {
	m.cached = append(m.cached, nameRecord{id: applicationID, name: name})
	return nil
}

func (m *mockStaleApplicationStore) UpdateApplicationName(_ context.Context, applicationID, name string) error {
	m.updated = append(m.updated, nameRecord{id: applicationID, name: name})
	return nil
}

func TestStaleApplicationProjector_NameCacheAndUpdate(t *testing.T) {
	tests := []struct {
		label     string
		eventType string
		appName   string
	}{
		{"created event caches and updates", amPL.ApplicationComponentCreated, "Acme ERP"},
		{"updated event caches and updates", amPL.ApplicationComponentUpdated, "Acme ERP (Cloud)"},
	}
	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			store := &mockStaleApplicationStore{}
			projector := NewStaleApplicationProjector(store)
			id := uuid.New().String()
			payload, _ := json.Marshal(map[string]any{"id": id, "name": tt.appName})

			require.NoError(t, projector.ProjectEvent(context.Background(), tt.eventType, payload))

			want := nameRecord{id: id, name: tt.appName}
			assert.Equal(t, []nameRecord{want}, store.cached)
			assert.Equal(t, []nameRecord{want}, store.updated)
		})
	}
}

func TestStaleApplicationProjector_MarksDeletedApplicationStale(t *testing.T) {
	store := &mockStaleApplicationStore{}
	projector := NewStaleApplicationProjector(store)
	id := uuid.New().String()
	payload, _ := json.Marshal(map[string]any{"id": id})

	require.NoError(t, projector.ProjectEvent(context.Background(), amPL.ApplicationComponentDeleted, payload))

	assert.Equal(t, []string{id}, store.stale)
}

func TestStaleApplicationProjector_UnknownEvent_NoOp(t *testing.T) {
	store := &mockStaleApplicationStore{}
	projector := NewStaleApplicationProjector(store)

	require.NoError(t, projector.ProjectEvent(context.Background(), "SomeUnrelatedEvent", []byte(`{}`)))

	assert.Empty(t, store.stale)
	assert.Empty(t, store.cached)
}

func TestStaleApplicationProjector_EmptyID_NoOp(t *testing.T) {
	store := &mockStaleApplicationStore{}
	projector := NewStaleApplicationProjector(store)
	payload, _ := json.Marshal(map[string]any{"id": ""})

	require.NoError(t, projector.ProjectEvent(context.Background(), amPL.ApplicationComponentDeleted, payload))

	assert.Empty(t, store.stale)
}
