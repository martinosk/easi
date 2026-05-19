package projectors

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockStaleStore struct {
	marked []string
}

func (m *mockStaleStore) MarkSourceCapabilityStale(_ context.Context, id string) error {
	m.marked = append(m.marked, id)
	return nil
}

func TestStaleReferenceProjector_CapabilityDeleted_MarksStale(t *testing.T) {
	store := &mockStaleStore{}
	projector := NewStaleReferenceProjector(store)

	id := uuid.New().String()
	require.NoError(t, projector.ProjectEvent(context.Background(), "CapabilityDeleted",
		[]byte(`{"id":"`+id+`"}`)))
	assert.Equal(t, []string{id}, store.marked)
}

func TestStaleReferenceProjector_OtherEvents_Ignored(t *testing.T) {
	store := &mockStaleStore{}
	projector := NewStaleReferenceProjector(store)

	require.NoError(t, projector.ProjectEvent(context.Background(), "CapabilityCreated",
		[]byte(`{"id":"x"}`)))
	assert.Empty(t, store.marked)
}
