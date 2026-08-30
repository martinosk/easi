//go:build integration
// +build integration

package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"easi/backend/internal/auth/application/handlers"
	"easi/backend/internal/auth/application/projectors"
	"easi/backend/internal/auth/application/readmodels"
	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/auth/infrastructure/secrets"
	authPL "easi/backend/internal/auth/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	domain "easi/backend/internal/shared/eventsourcing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type invitationFault struct {
	fail bool
}

func setupPlatformHandlersWithInvitationFault(db *sql.DB, fault *invitationFault) chi.Router {
	commandBus := cqrs.NewInMemoryCommandBus()
	tenantRepo := repositories.NewTenantRepository(db)
	tenantDB := database.NewTenantAwareDB(db)
	evStore := eventstore.NewPostgresEventStore(tenantDB)

	eventBus := events.NewInMemoryEventBus()
	evStore.SetEventBus(eventBus)

	commandBus.Register("CreateTenant", handlers.NewCreateTenantHandler(tenantRepo, evStore, tenantDB))

	eventBus.Subscribe(authPL.TenantCreated, handlers.NewTenantCreatedReactor(commandBus))

	invitationReadModel := readmodels.NewInvitationReadModel(tenantDB)
	invitationProjector := projectors.NewInvitationProjector(invitationReadModel)
	eventBus.Subscribe("InvitationCreated", invitationProjector)
	eventBus.Subscribe("InvitationCreated", events.EventHandlerFunc(func(context.Context, domain.DomainEvent) error {
		if fault.fail {
			return errors.New("simulated invitation projector failure")
		}
		return nil
	}))
	eventBus.Subscribe("InvitationAccepted", invitationProjector)
	eventBus.Subscribe("InvitationRevoked", invitationProjector)
	eventBus.Subscribe("InvitationExpired", invitationProjector)

	invitationRepo := repositories.NewInvitationRepository(evStore)
	commandBus.Register("CreateInvitation", handlers.NewCreateInvitationHandler(invitationRepo))

	secretProvider := secrets.NewEnvSecretProvider("OIDC_CLIENT_SECRET")
	tenantHandlers := NewPlatformTenantHandlers(commandBus, tenantRepo, secretProvider)

	r := chi.NewRouter()
	r.Use(PlatformAdminMiddleware("test-api-key"))
	r.Post("/tenants", tenantHandlers.CreateTenant)
	r.Get("/tenants", tenantHandlers.ListTenants)
	r.Get("/tenants/{id}", tenantHandlers.GetTenantByID)

	return r
}

func (ctx *platformTestContext) countTenantEvents(t *testing.T, tenantID, eventType string) int {
	t.Helper()

	tx, err := ctx.db.Begin()
	require.NoError(t, err)
	defer func() { _ = tx.Rollback() }()

	_, err = tx.Exec(fmt.Sprintf("SET LOCAL app.current_tenant = '%s'", tenantID))
	require.NoError(t, err)

	var count int
	err = tx.QueryRow(
		"SELECT COUNT(*) FROM infrastructure.events WHERE tenant_id = $1 AND aggregate_id = $2 AND event_type = $3",
		tenantID, tenantID, eventType,
	).Scan(&count)
	require.NoError(t, err)
	return count
}

func (ctx *platformTestContext) tenantExists(t *testing.T, tenantID string) bool {
	t.Helper()
	var exists bool
	require.NoError(t, ctx.db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM auth.tenants WHERE id = $1)", tenantID,
	).Scan(&exists))
	return exists
}

func TestCreateTenant_SubscriberFailure_RollsBackTenant_RetrySucceedsAfterFaultCleared(t *testing.T) {
	ctx, cleanup := setupPlatformTestDB(t)
	defer cleanup()

	fault := &invitationFault{fail: true}
	router := setupPlatformHandlersWithInvitationFault(ctx.db, fault)

	tenantID := fmt.Sprintf("atomic-%d", time.Now().UnixNano())
	ctx.trackTenant(tenantID)

	reqBody := CreateTenantRequest{
		ID:      tenantID,
		Name:    "Atomicity Test Co",
		Domains: []string{tenantID + ".com"},
		OIDCConfig: OIDCConfigRequest{
			DiscoveryURL: "https://login.microsoftonline.com/xxx/v2.0",
			ClientID:     "client-id",
			AuthMethod:   "client_secret",
			Scopes:       "openid email profile",
		},
		FirstAdminEmail: "admin@" + tenantID + ".com",
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	failedResponse := ctx.makeRequest("POST", "/tenants", body, router)
	assert.Equal(t, http.StatusInternalServerError, failedResponse.Code)

	assert.False(t, ctx.tenantExists(t, tenantID), "a failed subscriber must leave no tenant row behind")
	assert.Equal(t, 0, ctx.countTenantEvents(t, tenantID, "TenantCreated"), "a failed subscriber must leave no TenantCreated event behind")

	fault.fail = false

	retryResponse := ctx.makeRequest("POST", "/tenants", body, router)
	assert.Equal(t, http.StatusCreated, retryResponse.Code, "retry after clearing the fault must succeed, not 409")

	assert.True(t, ctx.tenantExists(t, tenantID))
	assert.Equal(t, 1, ctx.countTenantEvents(t, tenantID, "TenantCreated"))
	assert.Equal(t, 1, ctx.countTenantInvitations(t, tenantID, "admin@"+tenantID+".com", "admin"))
}
