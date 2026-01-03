package api

import (
	"net/http"

	"easi/backend/internal/architecturemodeling/application/handlers"
	"easi/backend/internal/architecturemodeling/application/projectors"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
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

// SetupArchitectureModelingRoutes initializes and registers architecture modeling routes
func SetupArchitectureModelingRoutes(
	r chi.Router,
	commandBus *cqrs.InMemoryCommandBus,
	eventStore eventstore.EventStore,
	eventBus events.EventBus,
	db *database.TenantAwareDB,
	hateoas *sharedAPI.HATEOASLinks,
	authMiddleware AuthMiddleware,
) error {
	// Initialize repositories
	componentRepo := repositories.NewApplicationComponentRepository(eventStore)
	relationRepo := repositories.NewComponentRelationRepository(eventStore)

	// Initialize read models
	componentReadModel := readmodels.NewApplicationComponentReadModel(db)
	relationReadModel := readmodels.NewComponentRelationReadModel(db)

	// Initialize projectors
	componentProjector := projectors.NewApplicationComponentProjector(componentReadModel)
	relationProjector := projectors.NewComponentRelationProjector(relationReadModel)

	// Wire up projectors to event bus
	eventBus.Subscribe("ApplicationComponentCreated", componentProjector)
	eventBus.Subscribe("ApplicationComponentUpdated", componentProjector)
	eventBus.Subscribe("ApplicationComponentDeleted", componentProjector)
	eventBus.Subscribe("ComponentRelationCreated", relationProjector)
	eventBus.Subscribe("ComponentRelationUpdated", relationProjector)
	eventBus.Subscribe("ComponentRelationDeleted", relationProjector)

	// Initialize command handlers
	createComponentHandler := handlers.NewCreateApplicationComponentHandler(componentRepo)
	updateComponentHandler := handlers.NewUpdateApplicationComponentHandler(componentRepo)
	deleteComponentHandler := handlers.NewDeleteApplicationComponentHandler(componentRepo, relationReadModel, commandBus)
	createRelationHandler := handlers.NewCreateComponentRelationHandler(relationRepo)
	updateRelationHandler := handlers.NewUpdateComponentRelationHandler(relationRepo)
	deleteRelationHandler := handlers.NewDeleteComponentRelationHandler(relationRepo)

	// Register command handlers
	commandBus.Register("CreateApplicationComponent", createComponentHandler)
	commandBus.Register("UpdateApplicationComponent", updateComponentHandler)
	commandBus.Register("DeleteApplicationComponent", deleteComponentHandler)
	commandBus.Register("CreateComponentRelation", createRelationHandler)
	commandBus.Register("UpdateComponentRelation", updateRelationHandler)
	commandBus.Register("DeleteComponentRelation", deleteRelationHandler)

	// Initialize HTTP handlers
	componentHandlers := NewComponentHandlers(commandBus, componentReadModel, hateoas)
	relationHandlers := NewRelationHandlers(commandBus, relationReadModel, hateoas)

	// Register component routes
	r.Route("/components", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsRead))
			r.Get("/", componentHandlers.GetAllComponents)
			r.Get("/{id}", componentHandlers.GetComponentByID)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsWrite))
			r.Post("/", componentHandlers.CreateApplicationComponent)
			r.Put("/{id}", componentHandlers.UpdateApplicationComponent)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsDelete))
			r.Delete("/{id}", componentHandlers.DeleteApplicationComponent)
		})
	})

	// Register relation routes
	r.Route("/relations", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsRead))
			r.Get("/", relationHandlers.GetAllRelations)
			r.Get("/{id}", relationHandlers.GetRelationByID)
			r.Get("/from/{componentId}", relationHandlers.GetRelationsFromComponent)
			r.Get("/to/{componentId}", relationHandlers.GetRelationsToComponent)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsWrite))
			r.Post("/", relationHandlers.CreateComponentRelation)
			r.Put("/{id}", relationHandlers.UpdateComponentRelation)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsDelete))
			r.Delete("/{id}", relationHandlers.DeleteComponentRelation)
		})
	})

	return nil
}
