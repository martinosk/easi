package api

import (
	"database/sql"

	"easi/backend/internal/architectureviews/application/handlers"
	"easi/backend/internal/architectureviews/application/projectors"
	"easi/backend/internal/architectureviews/application/readmodels"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
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
	db *sql.DB,
	hateoas *sharedAPI.HATEOASLinks,
) error {
	// Initialize repository
	viewRepo := repositories.NewArchitectureViewRepository(eventStore)

	// Initialize read model
	viewReadModel := readmodels.NewArchitectureViewReadModel(db)
	if err := viewReadModel.InitializeSchema(); err != nil {
		return err
	}

	// Initialize projector
	viewProjector := projectors.NewArchitectureViewProjector(viewReadModel)

	// Wire up projector to event bus
	eventBus.Subscribe("ViewCreated", viewProjector)
	eventBus.Subscribe("ComponentAddedToView", viewProjector)
	eventBus.Subscribe("ComponentPositionUpdated", viewProjector)

	// Initialize command handlers
	createViewHandler := handlers.NewCreateViewHandler(viewRepo)
	addComponentHandler := handlers.NewAddComponentToViewHandler(viewRepo)
	updatePositionHandler := handlers.NewUpdateComponentPositionHandler(viewRepo)

	// Register command handlers
	commandBus.Register("CreateView", createViewHandler)
	commandBus.Register("AddComponentToView", addComponentHandler)
	commandBus.Register("UpdateComponentPosition", updatePositionHandler)

	// Initialize HTTP handlers
	viewHandlers := NewViewHandlers(commandBus, viewReadModel, hateoas)

	// Register routes
	r.Route("/views", func(r chi.Router) {
		r.Post("/", viewHandlers.CreateView)
		r.Get("/", viewHandlers.GetAllViews)
		r.Get("/{id}", viewHandlers.GetViewByID)
		r.Post("/{id}/components", viewHandlers.AddComponentToView)
		r.Patch("/{id}/components/{componentId}/position", viewHandlers.UpdateComponentPosition)
	})

	return nil
}
