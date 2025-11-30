package api

import (
	"easi/backend/internal/infrastructure/database"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/events"
	"easi/backend/internal/viewlayouts/application/handlers"
	"easi/backend/internal/viewlayouts/infrastructure/repositories"

	"github.com/go-chi/chi/v5"
)

func SetupViewLayoutsRoutes(
	r chi.Router,
	eventBus events.EventBus,
	db *database.TenantAwareDB,
	hateoas *sharedAPI.HATEOASLinks,
) error {
	repo := repositories.NewLayoutContainerRepository(db)
	layoutHandlers := NewLayoutHandlers(repo, hateoas)

	componentDeletedHandler := handlers.NewComponentDeletedHandler(repo)
	capabilityDeletedHandler := handlers.NewCapabilityDeletedHandler(repo)
	businessDomainDeletedHandler := handlers.NewBusinessDomainDeletedHandler(repo)
	viewDeletedHandler := handlers.NewViewDeletedHandler(repo)

	eventBus.Subscribe("ComponentDeleted", componentDeletedHandler)
	eventBus.Subscribe("CapabilityDeleted", capabilityDeletedHandler)
	eventBus.Subscribe("BusinessDomainDeleted", businessDomainDeletedHandler)
	eventBus.Subscribe("ViewDeleted", viewDeletedHandler)

	r.Route("/layouts/{contextType}/{contextRef}", func(r chi.Router) {
		r.Get("/", layoutHandlers.GetLayout)
		r.Put("/", layoutHandlers.UpsertLayout)
		r.Delete("/", layoutHandlers.DeleteLayout)

		r.Patch("/preferences", layoutHandlers.UpdatePreferences)

		r.Patch("/elements", layoutHandlers.BatchUpdateElements)

		r.Route("/elements/{elementId}", func(r chi.Router) {
			r.Put("/", layoutHandlers.UpsertElementPosition)
			r.Delete("/", layoutHandlers.DeleteElementPosition)
		})
	})

	return nil
}
