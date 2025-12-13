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

	secretsBasePath := os.Getenv("OIDC_SECRETS_PATH")
	if secretsBasePath == "" {
		secretsBasePath = "/var/run/secrets/oidc"
	}
	secretProvider := secrets.NewFileSecretProvider(secretsBasePath)

	tenantHandlers := NewTenantHandlers(commandBus, tenantRepo, secretProvider)

	platformAdminKey := os.Getenv("PLATFORM_ADMIN_API_KEY")

	rateLimiter := NewRateLimiter(100, 60)

	r.Route("/api/platform/v1", func(r chi.Router) {
		r.Use(RateLimitMiddleware(rateLimiter))
		r.Use(PlatformAdminMiddleware(platformAdminKey))

		r.Post("/tenants", tenantHandlers.CreateTenant)
		r.Get("/tenants", tenantHandlers.ListTenants)
		r.Get("/tenants/{id}", tenantHandlers.GetTenantByID)
	})

	return nil
}
