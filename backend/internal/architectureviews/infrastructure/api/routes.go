package api

import (
	"easi/backend/internal/architectureviews/application/handlers"
	"easi/backend/internal/architectureviews/application/projectors"
	"easi/backend/internal/architectureviews/application/readmodels"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

// SetupArchitectureViewsRoutes initializes and registers architecture views routes
func SetupArchitectureViewsRoutes(
	r chi.Router,
	commandBus *cqrs.InMemoryCommandBus,
	eventStore eventstore.EventStore,
	eventBus events.EventBus,
	db *database.TenantAwareDB,
	hateoas *sharedAPI.HATEOASLinks,
) error {
	// Initialize repositories
	viewRepo := repositories.NewArchitectureViewRepository(eventStore)
	layoutRepo := repositories.NewViewLayoutRepository(db)

	// Initialize read model
	viewReadModel := readmodels.NewArchitectureViewReadModel(db)

	// Initialize projector
	viewProjector := projectors.NewArchitectureViewProjector(viewReadModel)

	// Wire up projector to event bus
	eventBus.Subscribe("ViewCreated", viewProjector)
	eventBus.Subscribe("ComponentAddedToView", viewProjector)
	eventBus.Subscribe("ComponentPositionUpdated", viewProjector)
	eventBus.Subscribe("ComponentRemovedFromView", viewProjector)
	eventBus.Subscribe("ViewRenamed", viewProjector)
	eventBus.Subscribe("ViewDeleted", viewProjector)
	eventBus.Subscribe("DefaultViewChanged", viewProjector)
	eventBus.Subscribe("ViewEdgeTypeUpdated", viewProjector)
	eventBus.Subscribe("ViewLayoutDirectionUpdated", viewProjector)

	// Initialize cross-context event handlers
	componentDeletedHandler := handlers.NewApplicationComponentDeletedHandler(commandBus, viewReadModel)
	relationDeletedHandler := handlers.NewComponentRelationDeletedHandler()

	// Subscribe to events from ArchitectureModeling context
	eventBus.Subscribe("ApplicationComponentDeleted", componentDeletedHandler)
	eventBus.Subscribe("ComponentRelationDeleted", relationDeletedHandler)

	// Initialize command handlers
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

	// Register command handlers
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

	// Initialize HTTP handlers
	viewHandlers := NewViewHandlers(commandBus, viewReadModel, layoutRepo, hateoas)

	// Register routes
	r.Route("/views", func(r chi.Router) {
		r.Post("/", viewHandlers.CreateView)
		r.Get("/", viewHandlers.GetAllViews)
		r.Get("/{id}", viewHandlers.GetViewByID)
		r.Delete("/{id}", viewHandlers.DeleteView)
		r.Patch("/{id}/name", viewHandlers.RenameView)
		r.Put("/{id}/default", viewHandlers.SetDefaultView)
		r.Patch("/{id}/edge-type", viewHandlers.UpdateEdgeType)
		r.Patch("/{id}/layout-direction", viewHandlers.UpdateLayoutDirection)
		r.Patch("/{id}/color-scheme", viewHandlers.UpdateColorScheme)
		r.Post("/{id}/components", viewHandlers.AddComponentToView)
		r.Delete("/{id}/components/{componentId}", viewHandlers.RemoveComponentFromView)
		r.Patch("/{id}/components/{componentId}/position", viewHandlers.UpdateComponentPosition)
		r.Patch("/{id}/components/{componentId}/color", viewHandlers.UpdateComponentColor)
		r.Delete("/{id}/components/{componentId}/color", viewHandlers.ClearComponentColor)
		r.Patch("/{id}/layout", viewHandlers.UpdateMultiplePositions)
		r.Post("/{id}/capabilities", viewHandlers.AddCapabilityToView)
		r.Delete("/{id}/capabilities/{capabilityId}", viewHandlers.RemoveCapabilityFromView)
		r.Patch("/{id}/capabilities/{capabilityId}/position", viewHandlers.UpdateCapabilityPosition)
		r.Patch("/{id}/capabilities/{capabilityId}/color", viewHandlers.UpdateCapabilityColor)
		r.Delete("/{id}/capabilities/{capabilityId}/color", viewHandlers.ClearCapabilityColor)
	})

	return nil
}
