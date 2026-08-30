package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/platform/application/commands"
	"easi/backend/internal/platform/infrastructure/repositories"
	sharedctx "easi/backend/internal/shared/context"
	"easi/backend/internal/shared/events"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingTenantStore struct {
	records []repositories.TenantRecord
	err     error
}

func (s *recordingTenantStore) Create(_ context.Context, record repositories.TenantRecord) error {
	if s.err != nil {
		return s.err
	}
	s.records = append(s.records, record)
	return nil
}

type recordingSubscriber struct {
	events   []domain.DomainEvent
	tenantID string
}

func (s *recordingSubscriber) Handle(ctx context.Context, event domain.DomainEvent) error {
	s.events = append(s.events, event)
	if tenant, err := sharedctx.GetTenant(ctx); err == nil {
		s.tenantID = tenant.Value()
	}
	return nil
}

func validCreateTenantCommand() *commands.CreateTenant {
	return &commands.CreateTenant{
		ID:              "acme",
		Name:            "Acme Corporation",
		Domains:         []string{"acme.com"},
		DiscoveryURL:    "https://login.example.com/v2.0/.well-known/openid-configuration",
		ClientID:        "client-id",
		AuthMethod:      "client_secret",
		Scopes:          "openid email profile",
		FirstAdminEmail: "admin@acme.com",
	}
}

func setupCreateTenantHandler() (*CreateTenantHandler, *recordingTenantStore, *recordingSubscriber) {
	store := &recordingTenantStore{}
	subscriber := &recordingSubscriber{}
	eventBus := events.NewInMemoryEventBus()
	eventBus.Subscribe("TenantCreated", subscriber)
	return NewCreateTenantHandler(store, eventBus), store, subscriber
}

func TestCreateTenantHandler_PublishesTenantCreatedWithOIDCConfiguration(t *testing.T) {
	handler, store, subscriber := setupCreateTenantHandler()

	result, err := handler.Handle(context.Background(), validCreateTenantCommand())

	require.NoError(t, err)
	assert.Equal(t, "acme", result.CreatedID)
	require.Len(t, store.records, 1)
	require.Len(t, subscriber.events, 1)

	data := subscriber.events[0].EventData()
	assert.Equal(t, "acme", data["id"])
	assert.Equal(t, "admin@acme.com", data["firstAdminEmail"])
	assert.Equal(t, "https://login.example.com/v2.0/.well-known/openid-configuration", data["discoveryUrl"])
	assert.Equal(t, "client-id", data["clientId"])
	assert.Equal(t, "client_secret", data["authMethod"])
	assert.Equal(t, "openid email profile", data["scopes"])
}

func TestCreateTenantHandler_PublishesInTheTenantsOwnContext(t *testing.T) {
	handler, _, subscriber := setupCreateTenantHandler()

	_, err := handler.Handle(context.Background(), validCreateTenantCommand())

	require.NoError(t, err)
	assert.Equal(t, "acme", subscriber.tenantID)
}

func TestCreateTenantHandler_DoesNotPublishWhenPersistenceFails(t *testing.T) {
	subscriber := &recordingSubscriber{}
	eventBus := events.NewInMemoryEventBus()
	eventBus.Subscribe("TenantCreated", subscriber)
	store := &recordingTenantStore{err: errors.New("duplicate tenant")}
	handler := NewCreateTenantHandler(store, eventBus)

	_, err := handler.Handle(context.Background(), validCreateTenantCommand())

	require.Error(t, err)
	assert.Empty(t, subscriber.events)
}
