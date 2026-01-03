package api

import (
	"net/http"

	authValueObjects "easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/metamodel/application/handlers"
	"easi/backend/internal/metamodel/application/projectors"
	"easi/backend/internal/metamodel/application/readmodels"
	"easi/backend/internal/metamodel/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequirePermission(permission authValueObjects.Permission) func(http.Handler) http.Handler
}

type MetaModelRoutesDeps struct {
	Router         chi.Router
	CommandBus     *cqrs.InMemoryCommandBus
	EventStore     eventstore.EventStore
	EventBus       events.EventBus
	DB             *database.TenantAwareDB
	Hateoas        *sharedAPI.HATEOASLinks
	AuthMiddleware AuthMiddleware
	SessionManager *session.SessionManager
}

func SetupMetaModelRoutes(deps MetaModelRoutesDeps) error {
	configRepo := repositories.NewMetaModelConfigurationRepository(deps.EventStore)

	configReadModel := readmodels.NewMetaModelConfigurationReadModel(deps.DB)

	configProjector := projectors.NewMetaModelConfigurationProjector(configReadModel)

	deps.EventBus.Subscribe("MetaModelConfigurationCreated", configProjector)
	deps.EventBus.Subscribe("MaturityScaleConfigUpdated", configProjector)
	deps.EventBus.Subscribe("MaturityScaleConfigReset", configProjector)
	deps.EventBus.Subscribe("StrategyPillarAdded", configProjector)
	deps.EventBus.Subscribe("StrategyPillarUpdated", configProjector)
	deps.EventBus.Subscribe("StrategyPillarRemoved", configProjector)
	deps.EventBus.Subscribe("PillarFitConfigurationUpdated", configProjector)

	createConfigHandler := handlers.NewCreateMetaModelConfigurationHandler(configRepo)
	updateScaleHandler := handlers.NewUpdateMaturityScaleHandler(configRepo)
	resetScaleHandler := handlers.NewResetMaturityScaleHandler(configRepo)

	deps.CommandBus.Register("CreateMetaModelConfiguration", createConfigHandler)
	deps.CommandBus.Register("UpdateMaturityScale", updateScaleHandler)
	deps.CommandBus.Register("ResetMaturityScale", resetScaleHandler)

	addPillarHandler := handlers.NewAddStrategyPillarHandler(configRepo)
	updatePillarHandler := handlers.NewUpdateStrategyPillarHandler(configRepo)
	removePillarHandler := handlers.NewRemoveStrategyPillarHandler(configRepo)
	batchUpdatePillarsHandler := handlers.NewBatchUpdateStrategyPillarsHandler(configRepo)
	updatePillarFitConfigHandler := handlers.NewUpdatePillarFitConfigurationHandler(configRepo)

	deps.CommandBus.Register("AddStrategyPillar", addPillarHandler)
	deps.CommandBus.Register("UpdateStrategyPillar", updatePillarHandler)
	deps.CommandBus.Register("RemoveStrategyPillar", removePillarHandler)
	deps.CommandBus.Register("BatchUpdateStrategyPillars", batchUpdatePillarsHandler)
	deps.CommandBus.Register("UpdatePillarFitConfiguration", updatePillarFitConfigHandler)

	tenantCreatedHandler := handlers.NewTenantCreatedHandler(deps.CommandBus)
	deps.EventBus.Subscribe("TenantCreated", tenantCreatedHandler)

	metaModelHandlers := NewMetaModelHandlers(deps.CommandBus, configReadModel, deps.Hateoas, deps.SessionManager)
	strategyPillarsHandlers := NewStrategyPillarsHandlers(deps.CommandBus, configReadModel, deps.Hateoas, deps.SessionManager)

	deps.Router.Route("/meta-model", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(deps.AuthMiddleware.RequirePermission(authValueObjects.PermMetaModelRead))
			r.Get("/maturity-scale", metaModelHandlers.GetMaturityScale)
			r.Get("/configurations/{id}", metaModelHandlers.GetMaturityScaleByID)
			r.Get("/strategy-pillars", strategyPillarsHandlers.GetStrategyPillars)
			r.Get("/strategy-pillars/{id}", strategyPillarsHandlers.GetStrategyPillarByID)
		})

		r.Group(func(r chi.Router) {
			r.Use(deps.AuthMiddleware.RequirePermission(authValueObjects.PermMetaModelWrite))
			r.Put("/maturity-scale", metaModelHandlers.UpdateMaturityScale)
			r.Post("/maturity-scale/reset", metaModelHandlers.ResetMaturityScale)
			r.Patch("/strategy-pillars", strategyPillarsHandlers.BatchUpdateStrategyPillars)
			r.Post("/strategy-pillars", strategyPillarsHandlers.CreateStrategyPillar)
			r.Put("/strategy-pillars/{id}", strategyPillarsHandlers.UpdateStrategyPillar)
			r.Put("/strategy-pillars/{id}/fit-configuration", strategyPillarsHandlers.UpdatePillarFitConfiguration)
			r.Delete("/strategy-pillars/{id}", strategyPillarsHandlers.DeleteStrategyPillar)
		})
	})

	return nil
}
