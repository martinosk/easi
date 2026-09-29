package projectors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authPL "easi/backend/internal/auth/publishedlanguage"
	"easi/backend/internal/stewardship/application/readmodels"
)

type fakeUserCache struct {
	users map[string]readmodels.CachedUser
}

func (f *fakeUserCache) Upsert(_ context.Context, u readmodels.CachedUser) error {
	f.users[u.ID] = u
	return nil
}

func (f *fakeUserCache) SetActive(_ context.Context, userID string, active bool) error {
	if u, ok := f.users[userID]; ok {
		u.Active = active
		f.users[userID] = u
	}
	return nil
}

func TestUserCacheProjector_CachesCreatedUsers(t *testing.T) {
	cache := &fakeUserCache{users: map[string]readmodels.CachedUser{}}
	event := storedEvent(t, "mette", authPL.UserCreated, map[string]any{"id": "mette", "name": "Mette Gram", "email": "mette@example.com", "status": "active"})

	require.NoError(t, NewUserCacheProjector(cache).Handle(ctx, event))

	assert.Equal(t, readmodels.CachedUser{ID: "mette", Name: "Mette Gram", Email: "mette@example.com", Active: true}, cache.users["mette"])
}

func TestUserCacheProjector_StatusFollowsDisabledAndEnabled(t *testing.T) {
	cache := &fakeUserCache{users: map[string]readmodels.CachedUser{"mette": {ID: "mette", Active: true}}}
	projector := NewUserCacheProjector(cache)

	require.NoError(t, projector.Handle(ctx, storedEvent(t, "mette", authPL.UserDisabled, map[string]any{"id": "mette"})))
	assert.False(t, cache.users["mette"].Active)

	require.NoError(t, projector.Handle(ctx, storedEvent(t, "mette", authPL.UserEnabled, map[string]any{"id": "mette"})))
	assert.True(t, cache.users["mette"].Active)
}

func TestUserCacheProjector_IgnoresOtherEvents(t *testing.T) {
	cache := &fakeUserCache{users: map[string]readmodels.CachedUser{}}

	require.NoError(t, NewUserCacheProjector(cache).Handle(ctx, storedEvent(t, "mette", authPL.UserRoleChanged, map[string]any{"id": "mette"})))

	assert.Empty(t, cache.users)
}
