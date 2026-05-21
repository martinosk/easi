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
	marked []readmodels.CapabilityID
}

func (m *mockStaleStore) MarkSourceCapabilityStale(_ context.Context, id readmodels.CapabilityID) error {
	m.marked = append(m.marked, id)
	return nil
}

func TestStaleReferenceProjector_CapabilityDeleted_MarksStale(t *testing.T) {
	store := &mockStaleStore{}
	projector := NewStaleReferenceProjector(store)

	id := uuid.New().String()
	require.NoError(t, projector.ProjectEvent(context.Background(), cmPL.CapabilityDeleted,
		[]byte(`{"id":"`+id+`"}`)))
	assert.Equal(t, []readmodels.CapabilityID{readmodels.CapabilityID(id)}, store.marked)
}

func TestStaleReferenceProjector_OtherEvents_Ignored(t *testing.T) {
	store := &mockStaleStore{}
	projector := NewStaleReferenceProjector(store)

	require.NoError(t, projector.ProjectEvent(context.Background(), cmPL.CapabilityCreated,
		[]byte(`{"id":"x"}`)))
	assert.Empty(t, store.marked)
}
