package projectors

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/auth/application/readmodels"
	platformEvents "easi/backend/internal/platform/domain/events"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTenantCache struct {
	saved []readmodels.TenantCacheEntry
	err   error
}

func (c *fakeTenantCache) Save(_ context.Context, entry readmodels.TenantCacheEntry) error {
	if c.err != nil {
		return c.err
	}
	c.saved = append(c.saved, entry)
	return nil
}

func tenantCreatedEvent() platformEvents.TenantCreated {
	return platformEvents.NewTenantCreated(platformEvents.TenantDetails{
		ID:              "acme",
		Name:            "Acme Corporation",
		Status:          "active",
		Domains:         []string{"acme.com", "acme.co.uk"},
		FirstAdminEmail: "admin@acme.com",
		OIDC: platformEvents.TenantOIDC{
			DiscoveryURL: "https://login.example.com/v2.0/.well-known/openid-configuration",
			IssuerURL:    "https://login.example.com/v2.0",
			ClientID:     "client-id",
			AuthMethod:   "client_secret",
			Scopes:       "openid email profile",
		},
	})
}

func TestTenantCacheProjector_StoresTenantDomainsAndOIDCConfiguration(t *testing.T) {
	cache := &fakeTenantCache{}
	projector := NewTenantCacheProjector(cache)

	err := projector.Handle(context.Background(), tenantCreatedEvent())

	require.NoError(t, err)
	require.Len(t, cache.saved, 1)
	entry := cache.saved[0]
	assert.Equal(t, "acme", entry.TenantID)
	assert.Equal(t, "Acme Corporation", entry.Name)
	assert.Equal(t, "active", entry.Status)
	assert.Equal(t, []string{"acme.com", "acme.co.uk"}, entry.Domains)
	assert.Equal(t, "https://login.example.com/v2.0/.well-known/openid-configuration", entry.DiscoveryURL)
	assert.Equal(t, "https://login.example.com/v2.0", entry.IssuerURL)
	assert.Equal(t, "client-id", entry.ClientID)
	assert.Equal(t, "client_secret", entry.AuthMethod)
	assert.Equal(t, "openid email profile", entry.Scopes)
}

func TestTenantCacheProjector_PropagatesStoreFailure(t *testing.T) {
	cache := &fakeTenantCache{err: errors.New("write failed")}
	projector := NewTenantCacheProjector(cache)

	err := projector.Handle(context.Background(), tenantCreatedEvent())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "write failed")
}
