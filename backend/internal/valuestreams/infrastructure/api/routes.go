package api

import (
	"net/http"

	authValueObjects "easi/backend/internal/auth/domain/valueobjects"
	capReadModels "easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/valuestreams/application/handlers"
	"easi/backend/internal/valuestreams/application/projectors"
	"easi/backend/internal/valuestreams/application/readmodels"
	"easi/backend/internal/valuestreams/infrastructure/gateways"
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
	config.EventBus.Subscribe("ValueStreamStageAdded", projector)
	config.EventBus.Subscribe("ValueStreamStageUpdated", projector)
	config.EventBus.Subscribe("ValueStreamStageRemoved", projector)
	config.EventBus.Subscribe("ValueStreamStagesReordered", projector)
	config.EventBus.Subscribe("ValueStreamStageCapabilityAdded", projector)
	config.EventBus.Subscribe("ValueStreamStageCapabilityRemoved", projector)

	capRM := capReadModels.NewCapabilityReadModel(config.DB)
	capGateway := gateways.NewCapabilityGateway(capRM)

	config.CommandBus.Register("CreateValueStream", handlers.NewCreateValueStreamHandler(repo, rm))
	config.CommandBus.Register("UpdateValueStream", handlers.NewUpdateValueStreamHandler(repo, rm))
	config.CommandBus.Register("DeleteValueStream", handlers.NewDeleteValueStreamHandler(repo))
	config.CommandBus.Register("AddStage", handlers.NewAddStageHandler(repo))
	config.CommandBus.Register("UpdateStage", handlers.NewUpdateStageHandler(repo))
	config.CommandBus.Register("RemoveStage", handlers.NewRemoveStageHandler(repo))
	config.CommandBus.Register("ReorderStages", handlers.NewReorderStagesHandler(repo))
	config.CommandBus.Register("AddStageCapability", handlers.NewAddStageCapabilityHandler(repo, capGateway))
	config.CommandBus.Register("RemoveStageCapability", handlers.NewRemoveStageCapabilityHandler(repo))

	links := NewValueStreamsLinks(config.HATEOAS)
	httpHandlers := NewValueStreamHandlers(config.CommandBus, rm, links)
	stageHttpHandlers := NewStageHandlers(config.CommandBus, rm, links)

	config.Router.Route("/value-streams", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(config.AuthMiddleware.RequirePermission(authValueObjects.PermValueStreamsRead))
			r.Get("/", httpHandlers.GetAllValueStreams)
			r.Get("/{id}", httpHandlers.GetValueStreamByID)
			r.Get("/{id}/capabilities", stageHttpHandlers.GetValueStreamCapabilities)
		})
		r.Group(func(r chi.Router) {
			r.Use(config.AuthMiddleware.RequirePermission(authValueObjects.PermValueStreamsWrite))
			r.Post("/", httpHandlers.CreateValueStream)
			r.Put("/{id}", httpHandlers.UpdateValueStream)
			r.Post("/{id}/stages", stageHttpHandlers.CreateStage)
			r.Put("/{id}/stages/positions", stageHttpHandlers.ReorderStages)
			r.Put("/{id}/stages/{stageId}", stageHttpHandlers.UpdateStage)
			r.Post("/{id}/stages/{stageId}/capabilities", stageHttpHandlers.AddStageCapability)
		})
		r.Group(func(r chi.Router) {
			r.Use(config.AuthMiddleware.RequirePermission(authValueObjects.PermValueStreamsDelete))
			r.Delete("/{id}", httpHandlers.DeleteValueStream)
			r.Delete("/{id}/stages/{stageId}", stageHttpHandlers.DeleteStage)
			r.Delete("/{id}/stages/{stageId}/capabilities/{capabilityId}", stageHttpHandlers.RemoveStageCapability)
		})
	})

	return nil
}
