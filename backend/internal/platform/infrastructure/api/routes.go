package api

import (
	"database/sql"
	"os"

	"easi/backend/internal/platform/application/handlers"
	"easi/backend/internal/platform/infrastructure/repositories"
	"easi/backend/internal/platform/infrastructure/secrets"
	"easi/backend/internal/shared/cqrs"

	"github.com/go-chi/chi/v5"
)

func SetupPlatformRoutes(r chi.Router, db *sql.DB) error {
	commandBus := cqrs.NewInMemoryCommandBus()

	tenantRepo := repositories.NewTenantRepository(db)

	createTenantHandler := handlers.NewCreateTenantHandler(tenantRepo)
	commandBus.Register("CreateTenant", createTenantHandler)

	secretProvider := secrets.NewEnvSecretProvider("OIDC_CLIENT_SECRET")

	tenantHandlers := NewTenantHandlers(commandBus, tenantRepo, secretProvider)

	platformAdminKey := os.Getenv("PLATFORM_ADMIN_API_KEY")

	rateLimiter := NewRateLimiter(100, 60)

	r.Route("/platform", func(r chi.Router) {
		r.Use(RateLimitMiddleware(rateLimiter))
		r.Use(PlatformAdminMiddleware(platformAdminKey))

		r.Post("/tenants", tenantHandlers.CreateTenant)
		r.Get("/tenants", tenantHandlers.ListTenants)
		r.Get("/tenants/{id}", tenantHandlers.GetTenantByID)
		r.Post("/tenants/{id}/invitations", tenantHandlers.CreateTenantInvitation)
	})

	return nil
}
