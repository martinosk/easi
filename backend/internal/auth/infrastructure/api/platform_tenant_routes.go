package api

import (
	"database/sql"
	"os"

	"easi/backend/internal/auth/application/handlers"
	"easi/backend/internal/auth/infrastructure/repositories"
	"easi/backend/internal/auth/infrastructure/secrets"
	"easi/backend/internal/infrastructure/api/middleware"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

func registerPlatformTenantRoutes(r chi.Router, rawDB *sql.DB, tenantDB *database.TenantAwareDB, commandBus cqrs.CommandBus, eventBus events.EventBus) {
	tenantRepo := repositories.NewTenantRepository(rawDB)

	tenantEventStore := eventstore.NewPostgresEventStore(tenantDB)
	tenantEventStore.SetEventBus(eventBus)

	createTenantHandler := handlers.NewCreateTenantHandler(tenantRepo, tenantEventStore)
	commandBus.Register("CreateTenant", createTenantHandler)

	secretProvider := secrets.NewEnvSecretProvider("OIDC_CLIENT_SECRET")

	tenantHandlers := NewPlatformTenantHandlers(commandBus, tenantRepo, secretProvider)

	platformAdminKey := os.Getenv("PLATFORM_ADMIN_API_KEY")

	rateLimiter := middleware.NewRateLimiter(100, 60)

	r.Route("/platform", func(r chi.Router) {
		r.Use(middleware.RateLimitMiddleware(rateLimiter))
		r.Use(PlatformAdminMiddleware(platformAdminKey))

		r.Post("/tenants", tenantHandlers.CreateTenant)
		r.Get("/tenants", tenantHandlers.ListTenants)
		r.Get("/tenants/{id}", tenantHandlers.GetTenantByID)
	})
}
