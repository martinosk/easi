package api

import (
	"database/sql"

	"easi/backend/internal/architecturemodeling/application/handlers"
	"easi/backend/internal/architecturemodeling/application/projectors"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	"github.com/go-chi/chi/v5"
)

// SetupArchitectureModelingRoutes initializes and registers architecture modeling routes
func SetupArchitectureModelingRoutes(
	r chi.Router,
	commandBus *cqrs.InMemoryCommandBus,
	eventStore eventstore.EventStore,
	eventBus events.EventBus,
	db *sql.DB,
	hateoas *sharedAPI.HATEOASLinks,
) error {
	// Initialize repositories
	componentRepo := repositories.NewApplicationComponentRepository(eventStore)
	relationRepo := repositories.NewComponentRelationRepository(eventStore)

	// Initialize read models
	componentReadModel := readmodels.NewApplicationComponentReadModel(db)
	if err := componentReadModel.InitializeSchema(); err != nil {
		return err
	}

	relationReadModel := readmodels.NewComponentRelationReadModel(db)
	if err := relationReadModel.InitializeSchema(); err != nil {
		return err
	}

	// Initialize projectors
	componentProjector := projectors.NewApplicationComponentProjector(componentReadModel)
	relationProjector := projectors.NewComponentRelationProjector(relationReadModel)

	// Wire up projectors to event bus
	eventBus.Subscribe("ApplicationComponentCreated", componentProjector)
	eventBus.Subscribe("ComponentRelationCreated", relationProjector)

	// Initialize command handlers
	createComponentHandler := handlers.NewCreateApplicationComponentHandler(componentRepo)
	createRelationHandler := handlers.NewCreateComponentRelationHandler(relationRepo)

	// Register command handlers
	commandBus.Register("CreateApplicationComponent", createComponentHandler)
	commandBus.Register("CreateComponentRelation", createRelationHandler)

	// Initialize HTTP handlers
	componentHandlers := NewComponentHandlers(commandBus, componentReadModel, hateoas)
	relationHandlers := NewRelationHandlers(commandBus, relationReadModel, hateoas)

	// Register component routes
	r.Route("/components", func(r chi.Router) {
		r.Post("/", componentHandlers.CreateApplicationComponent)
		r.Get("/", componentHandlers.GetAllComponents)
		r.Get("/{id}", componentHandlers.GetComponentByID)
	})

	// Register relation routes
	r.Route("/relations", func(r chi.Router) {
		r.Post("/", relationHandlers.CreateComponentRelation)
		r.Get("/", relationHandlers.GetAllRelations)
		r.Get("/{id}", relationHandlers.GetRelationByID)
		r.Get("/from/{componentId}", relationHandlers.GetRelationsFromComponent)
		r.Get("/to/{componentId}", relationHandlers.GetRelationsToComponent)
	})

	return nil
}
