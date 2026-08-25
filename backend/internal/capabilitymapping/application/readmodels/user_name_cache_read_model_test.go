package readmodels

import (
	"context"
	"testing"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserNameCacheReadModel_ResolveEAOwner_PassesUserIDThrough(t *testing.T) {
	rm := NewUserNameCacheReadModel(nil)

	resolved, err := rm.ResolveEAOwner(context.Background(), "2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b")
	require.NoError(t, err)
	assert.Equal(t, "2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b", resolved)
}

func TestUserNameCacheReadModel_ResolveEAOwner_TrimsUserID(t *testing.T) {
	rm := NewUserNameCacheReadModel(nil)

	resolved, err := rm.ResolveEAOwner(context.Background(), " 2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b ")
	require.NoError(t, err)
	assert.Equal(t, "2ec46b70-63b3-4d6d-92f0-1d385f9d4c4b", resolved)
}

func TestUserNameCacheReadModel_ResolveEAOwner_RejectsEmpty(t *testing.T) {
	rm := NewUserNameCacheReadModel(nil)

	_, err := rm.ResolveEAOwner(context.Background(), "   ")
	assert.ErrorIs(t, err, valueobjects.ErrEAOwnerNotUser)
}
