package api

import (
	"net/http"

	authPL "easi/backend/internal/auth/publishedlanguage"
	"easi/backend/internal/archassistant/infrastructure/repositories"
	"easi/backend/internal/infrastructure/database"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequirePermission(permission authPL.Permission) func(http.Handler) http.Handler
}

type ArchAssistantRoutesDeps struct {
	Router         chi.Router
	DB             *database.TenantAwareDB
	AuthMiddleware AuthMiddleware
}

func SetupArchAssistantRoutes(deps ArchAssistantRoutesDeps) error {
	repo := repositories.NewAIConfigurationRepository(deps.DB)
	handlers := NewAIConfigHandlers(repo)

	deps.Router.Route("/assistant-config", func(r chi.Router) {
		r.Use(deps.AuthMiddleware.RequirePermission(authPL.PermMetaModelWrite))
		r.Get("/", handlers.GetConfig)
		r.Put("/", handlers.UpdateConfig)
		r.Post("/test", handlers.TestConnection)
	})

	return nil
}
