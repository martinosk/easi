package projectors

import (
	"context"
	"encoding/json"
	"testing"

	authevents "easi/backend/internal/auth/domain/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserNameCacheWriter struct {
	upsertCalls []struct{ id, name, email string }
}

func (m *mockUserNameCacheWriter) Upsert(_ context.Context, id, name, email string) error {
	m.upsertCalls = append(m.upsertCalls, struct{ id, name, email string }{id, name, email})
	return nil
}

func TestUserNameCacheProjector_HandlesUserCreated(t *testing.T) {
	mock := &mockUserNameCacheWriter{}
	projector := NewUserNameCacheProjector(mock)

	event := authevents.NewUserCreated(
		"2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b",
		"alice@example.com",
		"Alice Smith",
		"architect",
		"external-1",
		"invitation-1",
	)

	require.NoError(t, projector.Handle(context.Background(), event))

	require.Len(t, mock.upsertCalls, 1)
	assert.Equal(t, "2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b", mock.upsertCalls[0].id)
	assert.Equal(t, "Alice Smith", mock.upsertCalls[0].name)
	assert.Equal(t, "alice@example.com", mock.upsertCalls[0].email)
}

func TestUserNameCacheProjector_SkipsUserCreatedWithoutID(t *testing.T) {
	mock := &mockUserNameCacheWriter{}
	projector := NewUserNameCacheProjector(mock)

	eventData, err := json.Marshal(struct {
		Name string `json:"name"`
	}{"Alice Smith"})
	require.NoError(t, err)

	require.NoError(t, projector.ProjectEvent(context.Background(), "UserCreated", eventData))

	assert.Empty(t, mock.upsertCalls)
}

func TestUserNameCacheProjector_IgnoresOtherEvents(t *testing.T) {
	mock := &mockUserNameCacheWriter{}
	projector := NewUserNameCacheProjector(mock)

	require.NoError(t, projector.ProjectEvent(context.Background(), "UserRoleChanged", []byte(`{"id":"x"}`)))

	assert.Empty(t, mock.upsertCalls)
}
