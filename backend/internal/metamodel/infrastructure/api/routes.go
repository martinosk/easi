package api

import (
	"net/http"

	authPL "easi/backend/internal/auth/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/metamodel/application/handlers"
	"easi/backend/internal/metamodel/application/projectors"
	"easi/backend/internal/metamodel/application/readmodels"
	"easi/backend/internal/metamodel/infrastructure/repositories"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	platformPL "easi/backend/internal/platform/publishedlanguage"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequirePermission(permission authPL.Permission) func(http.Handler) http.Handler
}

type MetaModelRoutesDeps struct {
	Router         chi.Router
	CommandBus     *cqrs.InMemoryCommandBus
	EventStore     eventstore.EventStore
	EventBus       events.EventBus
	DB             *database.TenantAwareDB
	Hateoas        *sharedAPI.HATEOASLinks
	AuthMiddleware AuthMiddleware
	SessionProvider authPL.SessionProvider
}

func SetupMetaModelRoutes(deps MetaModelRoutesDeps) error {
	configRepo := repositories.NewMetaModelConfigurationRepository(deps.EventStore)

	configReadModel := readmodels.NewMetaModelConfigurationReadModel(deps.DB)

	configProjector := projectors.NewMetaModelConfigurationProjector(configReadModel)

	deps.EventBus.Subscribe(mmPL.MetaModelConfigurationCreated, configProjector)
	deps.EventBus.Subscribe(mmPL.MaturityScaleConfigUpdated, configProjector)
	deps.EventBus.Subscribe(mmPL.MaturityScaleConfigReset, configProjector)
	deps.EventBus.Subscribe(mmPL.StrategyPillarAdded, configProjector)
	deps.EventBus.Subscribe(mmPL.StrategyPillarUpdated, configProjector)
	deps.EventBus.Subscribe(mmPL.StrategyPillarRemoved, configProjector)
	deps.EventBus.Subscribe(mmPL.PillarFitConfigurationUpdated, configProjector)

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
	deps.EventBus.Subscribe(platformPL.TenantCreated, tenantCreatedHandler)

	links := NewMetaModelLinks(deps.Hateoas)
	metaModelHandlers := NewMetaModelHandlers(deps.CommandBus, configReadModel, links, deps.SessionProvider)
	strategyPillarsHandlers := NewStrategyPillarsHandlers(deps.CommandBus, configReadModel, links, deps.SessionProvider)

	deps.Router.Route("/meta-model", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(deps.AuthMiddleware.RequirePermission(authPL.PermMetaModelRead))
			r.Get("/maturity-scale", metaModelHandlers.GetMaturityScale)
			r.Get("/configurations/{id}", metaModelHandlers.GetMaturityScaleByID)
			r.Get("/strategy-pillars", strategyPillarsHandlers.GetStrategyPillars)
			r.Get("/strategy-pillars/{id}", strategyPillarsHandlers.GetStrategyPillarByID)
		})

		r.Group(func(r chi.Router) {
			r.Use(deps.AuthMiddleware.RequirePermission(authPL.PermMetaModelWrite))
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
