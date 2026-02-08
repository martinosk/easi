package audit

import (
	"net/http"

	authValueObjects "easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/infrastructure/database"
	sharedAPI "easi/backend/internal/shared/api"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequirePermission(permission authValueObjects.Permission) func(http.Handler) http.Handler
}

type AuditRoutesDeps struct {
	Router         chi.Router
	DB             *database.TenantAwareDB
	Hateoas        *sharedAPI.HATEOASLinks
	AuthMiddleware AuthMiddleware
}

func SetupAuditRoutes(deps AuditRoutesDeps) error {
	readModel := NewAuditHistoryReadModel(deps.DB)
	handlers := NewAuditHandlers(readModel, NewAuditLinks(deps.Hateoas))

	deps.Router.Route("/audit", func(r chi.Router) {
		r.Use(deps.AuthMiddleware.RequirePermission(authValueObjects.PermAuditRead))
		r.Get("/{aggregateId}", handlers.GetAuditHistory)
	})

	return nil
}
