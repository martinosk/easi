package projectors

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStaleApplicationStore struct {
	stale []string
	err   error
}

func (m *mockStaleApplicationStore) MarkApplicationStale(_ context.Context, applicationID string) error {
	if m.err != nil {
		return m.err
	}
	m.stale = append(m.stale, applicationID)
	return nil
}

func TestStaleApplicationProjector_MarksDeletedApplicationStale(t *testing.T) {
	store := &mockStaleApplicationStore{}
	projector := NewStaleApplicationProjector(store)

	id := uuid.New().String()
	payload, _ := json.Marshal(map[string]interface{}{"id": id})
	err := projector.ProjectEvent(context.Background(), "ApplicationComponentDeleted", payload)

	require.NoError(t, err)
	assert.Equal(t, []string{id}, store.stale)
}

func TestStaleApplicationProjector_UnknownEvent_NoOp(t *testing.T) {
	store := &mockStaleApplicationStore{}
	projector := NewStaleApplicationProjector(store)

	err := projector.ProjectEvent(context.Background(), "SomeUnrelatedEvent", []byte(`{}`))

	require.NoError(t, err)
	assert.Empty(t, store.stale)
}

func TestStaleApplicationProjector_EmptyID_NoOp(t *testing.T) {
	store := &mockStaleApplicationStore{}
	projector := NewStaleApplicationProjector(store)

	payload, _ := json.Marshal(map[string]interface{}{"id": ""})
	err := projector.ProjectEvent(context.Background(), "ApplicationComponentDeleted", payload)

	require.NoError(t, err)
	assert.Empty(t, store.stale)
}
