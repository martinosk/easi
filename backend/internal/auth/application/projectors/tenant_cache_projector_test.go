package projectors

import (
	"context"
	"errors"
	"testing"
	"time"

	"easi/backend/internal/auth/application/readmodels"
	platformPL "easi/backend/internal/platform/publishedlanguage"
	domain "easi/backend/internal/shared/eventsourcing"

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

type supplierEvent struct {
	aggregateID string
	eventType   string
	data        map[string]interface{}
}

func (e supplierEvent) AggregateID() string               { return e.aggregateID }
func (e supplierEvent) EventType() string                 { return e.eventType }
func (e supplierEvent) OccurredAt() time.Time             { return time.Now() }
func (e supplierEvent) EventData() map[string]interface{} { return e.data }

func tenantCreatedEvent() domain.DomainEvent {
	return supplierEvent{
		aggregateID: "acme",
		eventType:   platformPL.TenantCreated,
		data: map[string]interface{}{
			"id":              "acme",
			"name":            "Acme Corporation",
			"status":          "active",
			"domains":         []string{"acme.com", "acme.co.uk"},
			"firstAdminEmail": "admin@acme.com",
			"discoveryUrl":    "https://login.example.com/v2.0/.well-known/openid-configuration",
			"clientId":        "client-id",
			"authMethod":      "client_secret",
			"scopes":          "openid email profile",
		},
	}
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
