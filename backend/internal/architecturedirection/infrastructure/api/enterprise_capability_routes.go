package api

import (
	"easi/backend/internal/architecturedirection/application/handlers"
	"easi/backend/internal/architecturedirection/application/projectors"
	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/infrastructure/repositories"
	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	authPL "easi/backend/internal/auth/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/infrastructure/api/middleware"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

type routeRepositories struct {
	capability *repositories.EnterpriseCapabilityRepository
	importance *repositories.EnterpriseStrategicImportanceRepository
}

type routeReadModels struct {
	capability      *readmodels.EnterpriseCapabilityReadModel
	importance      *readmodels.EnterpriseStrategicImportanceReadModel
	pillarCache     *readmodels.StrategyPillarCacheReadModel
	importanceCache *readmodels.EAImportanceCacheReadModel
	fitScoreCache   *readmodels.EAFitScoreCacheReadModel
}

type routeHTTPHandlers struct {
	enterpriseCapability *EnterpriseCapabilityHandlers
}

func setupEnterpriseCapabilityRoutes(deps RoutesDeps, capability *readmodels.EnterpriseCapabilityReadModel) {
	repos := initializeRepositories(deps.EventStore)
	rm := initializeReadModels(deps.DB, capability)

	setupEventSubscriptions(deps.EventBus, rm)
	setupCommandHandlers(deps.CommandBus, repos, rm)

	httpHandlers := initializeHTTPHandlers(deps.CommandBus, rm, deps.SessionProvider)
	rateLimiter := middleware.NewRateLimiter(100, 60)
	registerEnterpriseArchRoutes(deps.Router, httpHandlers, deps.AuthMiddleware, rateLimiter)
}

func initializeRepositories(eventStore eventstore.EventStore) *routeRepositories {
	return &routeRepositories{
		capability: repositories.NewEnterpriseCapabilityRepository(eventStore),
		importance: repositories.NewEnterpriseStrategicImportanceRepository(eventStore),
	}
}

func initializeReadModels(db *database.TenantAwareDB, capability *readmodels.EnterpriseCapabilityReadModel) *routeReadModels {
	return &routeReadModels{
		capability:      capability,
		importance:      readmodels.NewEnterpriseStrategicImportanceReadModel(db),
		pillarCache:     readmodels.NewStrategyPillarCacheReadModel(db),
		importanceCache: readmodels.NewEAImportanceCacheReadModel(db),
		fitScoreCache:   readmodels.NewEAFitScoreCacheReadModel(db),
	}
}

func setupEventSubscriptions(eventBus events.EventBus, rm *routeReadModels) {
	capabilityProjector := projectors.NewEnterpriseCapabilityProjector(rm.capability)
	importanceProjector := projectors.NewEnterpriseStrategicImportanceProjector(rm.importance)
	pillarCacheProjector := projectors.NewStrategyPillarCacheProjector(rm.pillarCache)
	importanceCacheProjector := projectors.NewEAImportanceCacheProjector(rm.importanceCache)
	fitScoreCacheProjector := projectors.NewEAFitScoreCacheProjector(rm.fitScoreCache)

	subscribeCapabilityEvents(eventBus, capabilityProjector)
	subscribeImportanceEvents(eventBus, importanceProjector)
	subscribePillarCacheEvents(eventBus, pillarCacheProjector)
	subscribeImportanceCacheEvents(eventBus, importanceCacheProjector)
	subscribeFitScoreCacheEvents(eventBus, fitScoreCacheProjector)
}

func subscribeCapabilityEvents(eventBus events.EventBus, projector *projectors.EnterpriseCapabilityProjector) {
	eventTypes := []string{
		pl.EnterpriseCapabilityCreated,
		pl.EnterpriseCapabilityUpdated,
		pl.EnterpriseCapabilityDeleted,
		pl.EnterpriseCapabilityTargetMaturitySet,
	}
	for _, eventType := range eventTypes {
		eventBus.Subscribe(eventType, projector)
	}
}

func subscribeImportanceEvents(eventBus events.EventBus, projector *projectors.EnterpriseStrategicImportanceProjector) {
	eventTypes := []string{
		pl.EnterpriseStrategicImportanceSet,
		pl.EnterpriseStrategicImportanceUpdated,
		pl.EnterpriseStrategicImportanceRemoved,
	}
	for _, eventType := range eventTypes {
		eventBus.Subscribe(eventType, projector)
	}
}

func subscribePillarCacheEvents(eventBus events.EventBus, projector *projectors.StrategyPillarCacheProjector) {
	eventTypes := []string{
		mmPL.MetaModelConfigurationCreated,
		mmPL.StrategyPillarAdded,
		mmPL.StrategyPillarUpdated,
		mmPL.StrategyPillarRemoved,
		mmPL.PillarFitConfigurationUpdated,
	}
	for _, eventType := range eventTypes {
		eventBus.Subscribe(eventType, projector)
	}
}

func subscribeImportanceCacheEvents(eventBus events.EventBus, projector *projectors.EAImportanceCacheProjector) {
	eventBus.Subscribe(cmPL.EffectiveImportanceRecalculated, projector)
}

func subscribeFitScoreCacheEvents(eventBus events.EventBus, projector *projectors.EAFitScoreCacheProjector) {
	eventTypes := []string{
		cmPL.ApplicationFitScoreSet,
		cmPL.ApplicationFitScoreRemoved,
	}
	for _, eventType := range eventTypes {
		eventBus.Subscribe(eventType, projector)
	}
}

func setupCommandHandlers(commandBus *cqrs.InMemoryCommandBus, repos *routeRepositories, rm *routeReadModels) {
	commandBus.Register("CreateEnterpriseCapability", handlers.NewCreateEnterpriseCapabilityHandler(repos.capability, rm.capability))
	commandBus.Register("UpdateEnterpriseCapability", handlers.NewUpdateEnterpriseCapabilityHandler(repos.capability, rm.capability))
	commandBus.Register("DeleteEnterpriseCapability", handlers.NewDeleteEnterpriseCapabilityHandler(repos.capability))
	commandBus.Register("SetTargetMaturity", handlers.NewSetTargetMaturityHandler(repos.capability))

	commandBus.Register("SetEnterpriseStrategicImportance", handlers.NewSetEnterpriseStrategicImportanceHandler(repos.importance, rm.capability, rm.importance))
	commandBus.Register("UpdateEnterpriseStrategicImportance", handlers.NewUpdateEnterpriseStrategicImportanceHandler(repos.importance))
	commandBus.Register("RemoveEnterpriseStrategicImportance", handlers.NewRemoveEnterpriseStrategicImportanceHandler(repos.importance))
}

func initializeHTTPHandlers(commandBus *cqrs.InMemoryCommandBus, rm *routeReadModels, sessionProvider authPL.SessionProvider) *routeHTTPHandlers {
	readModels := &EnterpriseCapabilityReadModels{
		Capability: rm.capability,
		Importance: rm.importance,
	}
	return &routeHTTPHandlers{
		enterpriseCapability: NewEnterpriseCapabilityHandlers(commandBus, readModels, sessionProvider),
	}
}

func registerEnterpriseArchRoutes(r chi.Router, h *routeHTTPHandlers, authMiddleware AuthMiddleware, rateLimiter *middleware.RateLimiter) {
	registerEnterpriseCapabilityRoutes(r, h, authMiddleware, rateLimiter)
}

func registerEnterpriseCapabilityRoutes(r chi.Router, handlers *routeHTTPHandlers, authMiddleware AuthMiddleware, rateLimiter *middleware.RateLimiter) {
	h := handlers.enterpriseCapability
	r.Route("/enterprise-capabilities", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermEnterpriseArchRead))
			r.Get("/", h.GetAllEnterpriseCapabilities)
			r.Get("/{id}", h.GetEnterpriseCapabilityByID)
			r.Get("/{id}/strategic-importance", h.GetStrategicImportance)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermEnterpriseArchWrite))
			r.Use(middleware.RateLimitMiddleware(rateLimiter))
			r.Post("/", h.CreateEnterpriseCapability)
			r.Put("/{id}", h.UpdateEnterpriseCapability)
			r.Put("/{id}/target-maturity", h.SetTargetMaturity)
			r.Post("/{id}/strategic-importance", h.SetStrategicImportance)
			r.Put("/{id}/strategic-importance/{importanceId}", h.UpdateStrategicImportance)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermEnterpriseArchDelete))
			r.Use(middleware.RateLimitMiddleware(rateLimiter))
			r.Delete("/{id}", h.DeleteEnterpriseCapability)
			r.Delete("/{id}/strategic-importance/{importanceId}", h.RemoveStrategicImportance)
		})
	})
}
