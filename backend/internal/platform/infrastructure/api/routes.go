package api

import (
	"database/sql"
	"os"

	"easi/backend/internal/infrastructure/api/middleware"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/platform/application/handlers"
	"easi/backend/internal/platform/infrastructure/repositories"
	"easi/backend/internal/platform/infrastructure/secrets"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

type PlatformRoutesDeps struct {
	Router     chi.Router
	RawDB      *sql.DB
	TenantDB   *database.TenantAwareDB
	CommandBus *cqrs.InMemoryCommandBus
	EventBus   events.EventBus
}

func SetupPlatformRoutes(deps PlatformRoutesDeps) error {
	tenantRepo := repositories.NewTenantRepository(deps.RawDB)

	tenantEventStore := eventstore.NewPostgresEventStore(deps.TenantDB)
	tenantEventStore.SetEventBus(deps.EventBus)

	createTenantHandler := handlers.NewCreateTenantHandler(tenantRepo, tenantEventStore)
	deps.CommandBus.Register("CreateTenant", createTenantHandler)

	secretProvider := secrets.NewEnvSecretProvider("OIDC_CLIENT_SECRET")

	tenantHandlers := NewTenantHandlers(deps.CommandBus, tenantRepo, secretProvider)

	platformAdminKey := os.Getenv("PLATFORM_ADMIN_API_KEY")

	rateLimiter := middleware.NewRateLimiter(100, 60)

	deps.Router.Route("/platform", func(r chi.Router) {
		r.Use(middleware.RateLimitMiddleware(rateLimiter))
		r.Use(sharedAPI.PlatformAdminMiddleware(platformAdminKey))

		r.Post("/tenants", tenantHandlers.CreateTenant)
		r.Get("/tenants", tenantHandlers.ListTenants)
		r.Get("/tenants/{id}", tenantHandlers.GetTenantByID)
	})

	return nil
}
