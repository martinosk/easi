package projectors

import (
	"context"
	"encoding/json"
	"testing"

	archEvents "easi/backend/internal/architecturemodeling/domain/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockComponentCacheWriter struct {
	upsertCalls []struct{ id, name string }
	deleteCalls []string
}

func (m *mockComponentCacheWriter) Upsert(ctx context.Context, id, name string) error {
	m.upsertCalls = append(m.upsertCalls, struct{ id, name string }{id, name})
	return nil
}

func (m *mockComponentCacheWriter) Delete(ctx context.Context, id string) error {
	m.deleteCalls = append(m.deleteCalls, id)
	return nil
}

func TestComponentCacheProjector_HandlesApplicationComponentCreated(t *testing.T) {
	mock := &mockComponentCacheWriter{}
	projector := NewComponentCacheProjector(mock)

	event := archEvents.ApplicationComponentCreated{
		ID:   "comp-123",
		Name: "Test Component",
	}
	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "ApplicationComponentCreated", eventData)
	require.NoError(t, err)

	require.Len(t, mock.upsertCalls, 1)
	assert.Equal(t, "comp-123", mock.upsertCalls[0].id)
	assert.Equal(t, "Test Component", mock.upsertCalls[0].name)
}

func TestComponentCacheProjector_HandlesApplicationComponentUpdated(t *testing.T) {
	mock := &mockComponentCacheWriter{}
	projector := NewComponentCacheProjector(mock)

	event := archEvents.ApplicationComponentUpdated{
		ID:   "comp-123",
		Name: "Updated Component Name",
	}
	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "ApplicationComponentUpdated", eventData)
	require.NoError(t, err)

	require.Len(t, mock.upsertCalls, 1)
	assert.Equal(t, "comp-123", mock.upsertCalls[0].id)
	assert.Equal(t, "Updated Component Name", mock.upsertCalls[0].name)
}

func TestComponentCacheProjector_HandlesApplicationComponentDeleted(t *testing.T) {
	mock := &mockComponentCacheWriter{}
	projector := NewComponentCacheProjector(mock)

	event := archEvents.ApplicationComponentDeleted{
		ID:   "comp-123",
		Name: "Test Component",
	}
	eventData, err := json.Marshal(event)
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "ApplicationComponentDeleted", eventData)
	require.NoError(t, err)

	require.Len(t, mock.deleteCalls, 1)
	assert.Equal(t, "comp-123", mock.deleteCalls[0])
}

func TestComponentCacheProjector_IgnoresUnknownEvents(t *testing.T) {
	mock := &mockComponentCacheWriter{}
	projector := NewComponentCacheProjector(mock)

	err := projector.ProjectEvent(context.Background(), "SomeOtherEvent", []byte("{}"))
	require.NoError(t, err)

	assert.Empty(t, mock.upsertCalls)
	assert.Empty(t, mock.deleteCalls)
}
