package api

import (
	"net/http"

	"easi/backend/internal/audit/application/readmodels"
	authPL "easi/backend/internal/auth/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	sharedAPI "easi/backend/internal/shared/api"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequirePermission(permission authPL.Permission) func(http.Handler) http.Handler
}

type AuditRoutesDeps struct {
	Router         chi.Router
	DB             *database.TenantAwareDB
	Hateoas        *sharedAPI.HATEOASLinks
	AuthMiddleware AuthMiddleware
}

func SetupAuditRoutes(deps AuditRoutesDeps) error {
	auditLinks := NewAuditLinks(deps.Hateoas)

	readModel := readmodels.NewAuditHistoryReadModel(deps.DB)
	handlers := NewAuditHandlers(readModel, auditLinks)

	creatorReadModel := readmodels.NewArtifactCreatorReadModel(deps.DB)
	creatorHandlers := NewArtifactCreatorHandlers(creatorReadModel, auditLinks)

	deps.Router.Route("/audit", func(r chi.Router) {
		r.Use(deps.AuthMiddleware.RequirePermission(authPL.PermAuditRead))
		r.Get("/{aggregateId}", handlers.GetAuditHistory)
	})

	deps.Router.Route("/artifact-creators", func(r chi.Router) {
		r.Use(deps.AuthMiddleware.RequirePermission(authPL.PermAuditRead))
		r.Get("/", creatorHandlers.GetArtifactCreators)
	})

	return nil
}
