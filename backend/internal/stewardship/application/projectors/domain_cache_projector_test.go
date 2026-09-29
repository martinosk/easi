package projectors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/stewardship/application/readmodels"
)

type fakeDomainCache struct {
	domains map[string]readmodels.CachedDomain
}

func (f *fakeDomainCache) Upsert(_ context.Context, d readmodels.CachedDomain) error {
	f.domains[d.ID] = d
	return nil
}

func TestDomainCacheProjector_CachesCreatedAndUpdatedDomains(t *testing.T) {
	for _, eventType := range []string{capPL.BusinessDomainCreated, capPL.BusinessDomainUpdated} {
		t.Run(eventType, func(t *testing.T) {
			cache := &fakeDomainCache{domains: map[string]readmodels.CachedDomain{}}
			event := storedEvent(t, "ce", eventType, map[string]any{"id": "ce", "name": "Customer Engagement", "domainArchitectId": "alice"})

			require.NoError(t, NewDomainCacheProjector(cache).Handle(ctx, event))

			assert.Equal(t, readmodels.CachedDomain{ID: "ce", Name: "Customer Engagement", DomainArchitectID: "alice"}, cache.domains["ce"])
		})
	}
}

func TestDomainCacheProjector_FallsBackToTheAggregateID(t *testing.T) {
	cache := &fakeDomainCache{domains: map[string]readmodels.CachedDomain{}}
	event := storedEvent(t, "ce", capPL.BusinessDomainCreated, map[string]any{"name": "Customer Engagement"})

	require.NoError(t, NewDomainCacheProjector(cache).Handle(ctx, event))

	assert.Equal(t, "Customer Engagement", cache.domains["ce"].Name)
}

func TestDomainCacheProjector_IgnoresOtherEvents(t *testing.T) {
	cache := &fakeDomainCache{domains: map[string]readmodels.CachedDomain{}}

	require.NoError(t, NewDomainCacheProjector(cache).Handle(ctx, storedEvent(t, "ce", capPL.BusinessDomainDeleted, map[string]any{"id": "ce"})))

	assert.Empty(t, cache.domains)
}
