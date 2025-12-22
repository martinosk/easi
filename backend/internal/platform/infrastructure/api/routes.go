package api

import (
	"database/sql"
	"os"

	authHandlers "easi/backend/internal/auth/application/handlers"
	authRepositories "easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/platform/application/handlers"
	"easi/backend/internal/platform/infrastructure/repositories"
	"easi/backend/internal/platform/infrastructure/secrets"
	"easi/backend/internal/shared/cqrs"

	"github.com/go-chi/chi/v5"
)

type PlatformRoutesDeps struct {
	Router     chi.Router
	RawDB      *sql.DB
	TenantDB   *database.TenantAwareDB
	CommandBus *cqrs.InMemoryCommandBus
	EventStore eventstore.EventStore
}

func SetupPlatformRoutes(deps PlatformRoutesDeps) error {
	tenantRepo := repositories.NewTenantRepository(deps.RawDB)
	invitationRepo := authRepositories.NewInvitationRepository(deps.EventStore)

	createTenantHandler := handlers.NewCreateTenantHandler(tenantRepo, deps.CommandBus)
	deps.CommandBus.Register("CreateTenant", createTenantHandler)

	deps.CommandBus.Register("CreateInvitation", authHandlers.NewCreateInvitationHandler(invitationRepo))

	secretProvider := secrets.NewEnvSecretProvider("OIDC_CLIENT_SECRET")

	tenantHandlers := NewTenantHandlers(deps.CommandBus, tenantRepo, secretProvider)

	platformAdminKey := os.Getenv("PLATFORM_ADMIN_API_KEY")

	rateLimiter := NewRateLimiter(100, 60)

	deps.Router.Route("/platform", func(r chi.Router) {
		r.Use(RateLimitMiddleware(rateLimiter))
		r.Use(PlatformAdminMiddleware(platformAdminKey))

		r.Post("/tenants", tenantHandlers.CreateTenant)
		r.Get("/tenants", tenantHandlers.ListTenants)
		r.Get("/tenants/{id}", tenantHandlers.GetTenantByID)
		r.Post("/tenants/{id}/invitations", tenantHandlers.CreateTenantInvitation)
	})

	return nil
}
