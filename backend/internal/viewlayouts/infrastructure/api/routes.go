package api

import (
	"net/http"

	archPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	avPL "easi/backend/internal/architectureviews/publishedlanguage"
	authValueObjects "easi/backend/internal/auth/domain/valueobjects"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/events"
	"easi/backend/internal/viewlayouts/application/handlers"
	"easi/backend/internal/viewlayouts/infrastructure/repositories"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequirePermission(permission authValueObjects.Permission) func(http.Handler) http.Handler
}

func SubscribeEvents(eventBus events.EventBus, db *database.TenantAwareDB) {
	repo := repositories.NewLayoutContainerRepository(db)

	eventBus.Subscribe(archPL.ApplicationComponentDeleted, handlers.NewComponentDeletedHandler(repo))
	eventBus.Subscribe(cmPL.CapabilityDeleted, handlers.NewCapabilityDeletedHandler(repo))
	eventBus.Subscribe(cmPL.BusinessDomainDeleted, handlers.NewBusinessDomainDeletedHandler(repo))
	eventBus.Subscribe(avPL.ViewDeleted, handlers.NewViewDeletedHandler(repo))
}

func RegisterRoutes(r chi.Router, db *database.TenantAwareDB, hateoas *sharedAPI.HATEOASLinks, authMiddleware AuthMiddleware) {
	repo := repositories.NewLayoutContainerRepository(db)
	layoutHandlers := NewLayoutHandlers(repo, hateoas)

	r.Route("/layouts/{contextType}/{contextRef}", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermViewsRead))
			r.Get("/", layoutHandlers.GetLayout)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermViewsWrite))
			r.Put("/", layoutHandlers.UpsertLayout)
			r.Patch("/preferences", layoutHandlers.UpdatePreferences)
			r.Patch("/elements", layoutHandlers.BatchUpdateElements)
			r.Route("/elements/{elementId}", func(r chi.Router) {
				r.Put("/", layoutHandlers.UpsertElementPosition)
			})
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermViewsDelete))
			r.Delete("/", layoutHandlers.DeleteLayout)
			r.Delete("/elements/{elementId}", layoutHandlers.DeleteElementPosition)
		})
	})
}
