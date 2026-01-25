package api

import (
	"net/http"

	"easi/backend/internal/architectureviews/application/handlers"
	"easi/backend/internal/architectureviews/application/projectors"
	"easi/backend/internal/architectureviews/application/readmodels"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	authReadModels "easi/backend/internal/auth/application/readmodels"
	authValueObjects "easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequirePermission(permission authValueObjects.Permission) func(http.Handler) http.Handler
}

func SetupArchitectureViewsRoutes(
	r chi.Router,
	commandBus *cqrs.InMemoryCommandBus,
	eventStore eventstore.EventStore,
	eventBus events.EventBus,
	db *database.TenantAwareDB,
	hateoas *sharedAPI.HATEOASLinks,
	authMiddleware AuthMiddleware,
) error {
	userReadModel := authReadModels.NewUserReadModel(db)
	viewRepo := repositories.NewArchitectureViewRepository(eventStore)
	layoutRepo := repositories.NewViewLayoutRepository(db)
	viewReadModel := readmodels.NewArchitectureViewReadModel(db)
	viewProjector := projectors.NewArchitectureViewProjector(viewReadModel)

	eventBus.Subscribe("ViewCreated", viewProjector)
	eventBus.Subscribe("ComponentAddedToView", viewProjector)
	eventBus.Subscribe("ComponentRemovedFromView", viewProjector)
	eventBus.Subscribe("ViewRenamed", viewProjector)
	eventBus.Subscribe("ViewDeleted", viewProjector)
	eventBus.Subscribe("DefaultViewChanged", viewProjector)
	eventBus.Subscribe("ViewVisibilityChanged", viewProjector)

	componentDeletedHandler := handlers.NewApplicationComponentDeletedHandler(commandBus, viewReadModel)
	relationDeletedHandler := handlers.NewComponentRelationDeletedHandler()

	eventBus.Subscribe("ApplicationComponentDeleted", componentDeletedHandler)
	eventBus.Subscribe("ComponentRelationDeleted", relationDeletedHandler)

	createViewHandler := handlers.NewCreateViewHandler(viewRepo, viewReadModel)
	addComponentHandler := handlers.NewAddComponentToViewHandler(viewRepo, layoutRepo)
	updatePositionHandler := handlers.NewUpdateComponentPositionHandler(layoutRepo)
	updateMultiplePositionsHandler := handlers.NewUpdateMultiplePositionsHandler(layoutRepo)
	renameViewHandler := handlers.NewRenameViewHandler(viewRepo)
	deleteViewHandler := handlers.NewDeleteViewHandler(viewRepo)
	removeComponentHandler := handlers.NewRemoveComponentFromViewHandler(viewRepo)
	setDefaultViewHandler := handlers.NewSetDefaultViewHandler(viewRepo, viewReadModel)
	updateEdgeTypeHandler := handlers.NewUpdateViewEdgeTypeHandler(layoutRepo)
	updateLayoutDirectionHandler := handlers.NewUpdateViewLayoutDirectionHandler(layoutRepo)
	updateColorSchemeHandler := handlers.NewUpdateViewColorSchemeHandler(layoutRepo)
	updateElementColorHandler := handlers.NewUpdateElementColorHandler(layoutRepo)
	clearElementColorHandler := handlers.NewClearElementColorHandler(layoutRepo)
	changeVisibilityHandler := handlers.NewChangeViewVisibilityHandler(viewRepo, userReadModel)

	commandBus.Register("CreateView", createViewHandler)
	commandBus.Register("AddComponentToView", addComponentHandler)
	commandBus.Register("UpdateComponentPosition", updatePositionHandler)
	commandBus.Register("UpdateMultiplePositions", updateMultiplePositionsHandler)
	commandBus.Register("RenameView", renameViewHandler)
	commandBus.Register("DeleteView", deleteViewHandler)
	commandBus.Register("RemoveComponentFromView", removeComponentHandler)
	commandBus.Register("SetDefaultView", setDefaultViewHandler)
	commandBus.Register("UpdateViewEdgeType", updateEdgeTypeHandler)
	commandBus.Register("UpdateViewLayoutDirection", updateLayoutDirectionHandler)
	commandBus.Register("UpdateViewColorScheme", updateColorSchemeHandler)
	commandBus.Register("UpdateElementColor", updateElementColorHandler)
	commandBus.Register("ClearElementColor", clearElementColorHandler)
	commandBus.Register("ChangeViewVisibility", changeVisibilityHandler)

	viewHandlers := NewViewHandlers(commandBus, viewReadModel, hateoas)
	componentHandlers := NewViewComponentHandlers(commandBus, viewReadModel)
	elementHandlers := NewViewElementHandlers(layoutRepo, viewReadModel)
	colorHandlers := NewViewColorHandlers(commandBus, viewReadModel, hateoas)

	r.Route("/views", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermViewsRead))
			r.Get("/", viewHandlers.GetAllViews)
			r.Get("/{id}", viewHandlers.GetViewByID)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermViewsWrite))
			r.Post("/", viewHandlers.CreateView)
			r.Patch("/{id}/name", viewHandlers.RenameView)
			r.Put("/{id}/default", viewHandlers.SetDefaultView)
			r.Patch("/{id}/edge-type", viewHandlers.UpdateEdgeType)
			r.Patch("/{id}/layout-direction", viewHandlers.UpdateLayoutDirection)
			r.Patch("/{id}/color-scheme", colorHandlers.UpdateColorScheme)
			r.Post("/{id}/components", componentHandlers.AddComponentToView)
			r.Patch("/{id}/components/{componentId}/position", componentHandlers.UpdateComponentPosition)
			r.Patch("/{id}/components/{componentId}/color", colorHandlers.UpdateComponentColor)
			r.Patch("/{id}/layout", componentHandlers.UpdateMultiplePositions)
			r.Post("/{id}/capabilities", elementHandlers.AddCapabilityToView)
			r.Patch("/{id}/capabilities/{capabilityId}/position", elementHandlers.UpdateCapabilityPosition)
			r.Patch("/{id}/capabilities/{capabilityId}/color", colorHandlers.UpdateCapabilityColor)
			r.Post("/{id}/origin-entities", elementHandlers.AddOriginEntityToView)
			r.Patch("/{id}/origin-entities/{originEntityId}/position", elementHandlers.UpdateOriginEntityPosition)
			r.Patch("/{id}/visibility", viewHandlers.ChangeVisibility)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermViewsDelete))
			r.Delete("/{id}", viewHandlers.DeleteView)
			r.Delete("/{id}/components/{componentId}", componentHandlers.RemoveComponentFromView)
			r.Delete("/{id}/components/{componentId}/color", colorHandlers.ClearComponentColor)
			r.Delete("/{id}/capabilities/{capabilityId}", elementHandlers.RemoveCapabilityFromView)
			r.Delete("/{id}/capabilities/{capabilityId}/color", colorHandlers.ClearCapabilityColor)
			r.Delete("/{id}/origin-entities/{originEntityId}", elementHandlers.RemoveOriginEntityFromView)
		})
	})

	return nil
}
