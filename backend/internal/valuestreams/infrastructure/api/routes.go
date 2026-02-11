package api

import (
	"net/http"

	authValueObjects "easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/valuestreams/application/handlers"
	"easi/backend/internal/valuestreams/application/projectors"
	"easi/backend/internal/valuestreams/application/readmodels"
	"easi/backend/internal/valuestreams/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequirePermission(permission authValueObjects.Permission) func(http.Handler) http.Handler
}

type RouteConfig struct {
	Router         chi.Router
	CommandBus     *cqrs.InMemoryCommandBus
	EventStore     eventstore.EventStore
	EventBus       events.EventBus
	DB             *database.TenantAwareDB
	HATEOAS        *sharedAPI.HATEOASLinks
	AuthMiddleware AuthMiddleware
}

func SetupValueStreamsRoutes(config *RouteConfig) error {
	repo := repositories.NewValueStreamRepository(config.EventStore)
	rm := readmodels.NewValueStreamReadModel(config.DB)

	projector := projectors.NewValueStreamProjector(rm)
	config.EventBus.Subscribe("ValueStreamCreated", projector)
	config.EventBus.Subscribe("ValueStreamUpdated", projector)
	config.EventBus.Subscribe("ValueStreamDeleted", projector)

	config.CommandBus.Register("CreateValueStream", handlers.NewCreateValueStreamHandler(repo, rm))
	config.CommandBus.Register("UpdateValueStream", handlers.NewUpdateValueStreamHandler(repo, rm))
	config.CommandBus.Register("DeleteValueStream", handlers.NewDeleteValueStreamHandler(repo))

	links := NewValueStreamsLinks(config.HATEOAS)
	httpHandlers := NewValueStreamHandlers(config.CommandBus, rm, links)

	config.Router.Route("/value-streams", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(config.AuthMiddleware.RequirePermission(authValueObjects.PermValueStreamsRead))
			r.Get("/", httpHandlers.GetAllValueStreams)
			r.Get("/{id}", httpHandlers.GetValueStreamByID)
		})
		r.Group(func(r chi.Router) {
			r.Use(config.AuthMiddleware.RequirePermission(authValueObjects.PermValueStreamsWrite))
			r.Post("/", httpHandlers.CreateValueStream)
			r.Put("/{id}", httpHandlers.UpdateValueStream)
		})
		r.Group(func(r chi.Router) {
			r.Use(config.AuthMiddleware.RequirePermission(authValueObjects.PermValueStreamsDelete))
			r.Delete("/{id}", httpHandlers.DeleteValueStream)
		})
	})

	return nil
}
