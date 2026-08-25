package readmodels

import (
	"context"
	"testing"

	"easi/backend/internal/capabilitymapping/domain/valueobjects"

	"github.com/stretchr/testify/assert"
)

func TestUserNameCacheReadModel_ResolveEAOwner_RejectsBlankWithoutLookup(t *testing.T) {
	rm := NewUserNameCacheReadModel(nil)

	for _, value := range []string{"", "   "} {
		_, err := rm.ResolveEAOwner(context.Background(), value)
		assert.ErrorIs(t, err, valueobjects.ErrEAOwnerNotUser)
	}
}
