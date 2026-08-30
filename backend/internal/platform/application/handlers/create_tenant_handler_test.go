package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/platform/application/commands"
	"easi/backend/internal/platform/infrastructure/repositories"
	sharedctx "easi/backend/internal/shared/context"
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

type savedTenantEvents struct {
	aggregateID     string
	events          []domain.DomainEvent
	expectedVersion int
	tenantID        string
}

type fakeTenantEventStore struct {
	saved   []savedTenantEvents
	saveErr error
}

func (s *fakeTenantEventStore) SaveEvents(ctx context.Context, aggregateID string, evts []domain.DomainEvent, expectedVersion int) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	tenantID := ""
	if tenant, err := sharedctx.GetTenant(ctx); err == nil {
		tenantID = tenant.Value()
	}
	s.saved = append(s.saved, savedTenantEvents{
		aggregateID:     aggregateID,
		events:          evts,
		expectedVersion: expectedVersion,
		tenantID:        tenantID,
	})
	return nil
}

func (s *fakeTenantEventStore) GetEvents(_ context.Context, _ string) ([]domain.DomainEvent, error) {
	return nil, nil
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

func setupCreateTenantHandler() (*CreateTenantHandler, *recordingTenantStore, *fakeTenantEventStore) {
	store := &recordingTenantStore{}
	eventStore := &fakeTenantEventStore{}
	return NewCreateTenantHandler(store, eventStore), store, eventStore
}

func TestCreateTenantHandler_PersistsTenantCreatedThroughTheEventStore(t *testing.T) {
	handler, store, eventStore := setupCreateTenantHandler()

	result, err := handler.Handle(context.Background(), validCreateTenantCommand())

	require.NoError(t, err)
	assert.Equal(t, "acme", result.CreatedID)
	require.Len(t, store.records, 1)
	require.Len(t, eventStore.saved, 1)

	saved := eventStore.saved[0]
	assert.Equal(t, "acme", saved.aggregateID)
	assert.Equal(t, 0, saved.expectedVersion)
	require.Len(t, saved.events, 1)

	data := saved.events[0].EventData()
	assert.Equal(t, "acme", data["id"])
	assert.Equal(t, "admin@acme.com", data["firstAdminEmail"])
	assert.Equal(t, "https://login.example.com/v2.0/.well-known/openid-configuration", data["discoveryUrl"])
	assert.Equal(t, "client-id", data["clientId"])
	assert.Equal(t, "client_secret", data["authMethod"])
	assert.Equal(t, "openid email profile", data["scopes"])
}

func TestCreateTenantHandler_SavesEventsInTheTenantsOwnContext(t *testing.T) {
	handler, _, eventStore := setupCreateTenantHandler()

	_, err := handler.Handle(context.Background(), validCreateTenantCommand())

	require.NoError(t, err)
	require.Len(t, eventStore.saved, 1)
	assert.Equal(t, "acme", eventStore.saved[0].tenantID)
}

func TestCreateTenantHandler_DoesNotSaveEventsWhenRelationalWriteFails(t *testing.T) {
	store := &recordingTenantStore{err: errors.New("duplicate tenant")}
	eventStore := &fakeTenantEventStore{}
	handler := NewCreateTenantHandler(store, eventStore)

	_, err := handler.Handle(context.Background(), validCreateTenantCommand())

	require.Error(t, err)
	assert.Empty(t, eventStore.saved)
}

func TestCreateTenantHandler_FailsWhenEventStoreCannotPersist(t *testing.T) {
	store := &recordingTenantStore{}
	eventStore := &fakeTenantEventStore{saveErr: errors.New("db unavailable")}
	handler := NewCreateTenantHandler(store, eventStore)

	_, err := handler.Handle(context.Background(), validCreateTenantCommand())

	require.Error(t, err)
	require.Len(t, store.records, 1)
}
