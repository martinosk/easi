package projectors

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserNameCacheWriter struct {
	upsertCalls []struct{ id, name, email string }
}

func (m *mockUserNameCacheWriter) Upsert(ctx context.Context, id, name, email string) error {
	m.upsertCalls = append(m.upsertCalls, struct{ id, name, email string }{id, name, email})
	return nil
}

func TestUserNameCacheProjector_HandlesUserCreated(t *testing.T) {
	mock := &mockUserNameCacheWriter{}
	projector := NewUserNameCacheProjector(mock)

	eventData, err := json.Marshal(struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}{"2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b", "Alice Smith", "alice@example.com"})
	require.NoError(t, err)

	err = projector.ProjectEvent(context.Background(), "UserCreated", eventData)
	require.NoError(t, err)

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

	err = projector.ProjectEvent(context.Background(), "UserCreated", eventData)
	require.NoError(t, err)

	assert.Empty(t, mock.upsertCalls)
}

func TestUserNameCacheProjector_IgnoresOtherEvents(t *testing.T) {
	mock := &mockUserNameCacheWriter{}
	projector := NewUserNameCacheProjector(mock)

	err := projector.ProjectEvent(context.Background(), "UserRoleChanged", []byte(`{"id":"x"}`))
	require.NoError(t, err)

	assert.Empty(t, mock.upsertCalls)
}
