package api

import (
	"net/http"

	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	authPL "easi/backend/internal/auth/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/enterprisearchitecture/application/handlers"
	"easi/backend/internal/enterprisearchitecture/application/projectors"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/infrastructure/metamodel"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"
	eaPL "easi/backend/internal/enterprisearchitecture/publishedlanguage"
	"easi/backend/internal/infrastructure/api/middleware"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequirePermission(permission authPL.Permission) func(http.Handler) http.Handler
}

type routeRepositories struct {
	capability *repositories.EnterpriseCapabilityRepository
	importance *repositories.EnterpriseStrategicImportanceRepository
}

type routeReadModels struct {
	capability       *readmodels.EnterpriseCapabilityReadModel
	importance       *readmodels.EnterpriseStrategicImportanceReadModel
	metadata         *readmodels.DomainCapabilityMetadataReadModel
	timeSuggestion   *readmodels.TimeSuggestionReadModel
	pillarCache      *readmodels.StrategyPillarCacheReadModel
	realizationCache *readmodels.EARealizationCacheReadModel
	importanceCache  *readmodels.EAImportanceCacheReadModel
	fitScoreCache    *readmodels.EAFitScoreCacheReadModel
}

type routeHTTPHandlers struct {
	enterpriseCapability *EnterpriseCapabilityHandlers
	timeSuggestions      *TimeSuggestionsHandlers
}

type EnterpriseArchRoutesDeps struct {
	Router              chi.Router
	CommandBus          *cqrs.InMemoryCommandBus
	EventStore          eventstore.EventStore
	EventBus            events.EventBus
	DB                  *database.TenantAwareDB
	AuthMiddleware      AuthMiddleware
	SessionProvider     authPL.SessionProvider
	BusinessDomainNames projectors.BusinessDomainNameLookup
}

func SetupEnterpriseArchitectureRoutes(deps EnterpriseArchRoutesDeps) error {
	repos := initializeRepositories(deps.EventStore)
	rm := initializeReadModels(deps.DB)

	setupEventSubscriptions(deps.EventBus, rm, deps.BusinessDomainNames)
	setupCommandHandlers(deps.CommandBus, repos, rm)

	httpHandlers := initializeHTTPHandlers(deps.CommandBus, rm, deps.SessionProvider)
	rateLimiter := middleware.NewRateLimiter(100, 60)
	registerRoutes(deps.Router, httpHandlers, deps.AuthMiddleware, rateLimiter)

	return nil
}

func initializeRepositories(eventStore eventstore.EventStore) *routeRepositories {
	return &routeRepositories{
		capability: repositories.NewEnterpriseCapabilityRepository(eventStore),
		importance: repositories.NewEnterpriseStrategicImportanceRepository(eventStore),
	}
}

func initializeReadModels(db *database.TenantAwareDB) *routeReadModels {
	pillarCache := readmodels.NewStrategyPillarCacheReadModel(db)
	pillarsGateway := metamodel.NewLocalStrategyPillarsGateway(pillarCache)
	capability := readmodels.NewEnterpriseCapabilityReadModel(db)
	metadata := readmodels.NewDomainCapabilityMetadataReadModel(db)
	return &routeReadModels{
		capability:       capability,
		importance:       readmodels.NewEnterpriseStrategicImportanceReadModel(db),
		metadata:         metadata,
		timeSuggestion:   readmodels.NewTimeSuggestionReadModel(db, pillarsGateway),
		pillarCache:      pillarCache,
		realizationCache: readmodels.NewEARealizationCacheReadModel(db),
		importanceCache:  readmodels.NewEAImportanceCacheReadModel(db),
		fitScoreCache:    readmodels.NewEAFitScoreCacheReadModel(db),
	}
}

func setupEventSubscriptions(eventBus events.EventBus, rm *routeReadModels, businessDomainNames projectors.BusinessDomainNameLookup) {
	capabilityProjector := projectors.NewEnterpriseCapabilityProjector(rm.capability)
	importanceProjector := projectors.NewEnterpriseStrategicImportanceProjector(rm.importance)
	metadataProjector := projectors.NewDomainCapabilityMetadataProjector(rm.metadata, businessDomainNames)
	pillarCacheProjector := projectors.NewStrategyPillarCacheProjector(rm.pillarCache)
	realizationCacheProjector := projectors.NewEARealizationCacheProjector(rm.realizationCache)
	importanceCacheProjector := projectors.NewEAImportanceCacheProjector(rm.importanceCache)
	fitScoreCacheProjector := projectors.NewEAFitScoreCacheProjector(rm.fitScoreCache)

	subscribeCapabilityEvents(eventBus, capabilityProjector)
	subscribeImportanceEvents(eventBus, importanceProjector)
	subscribeCapabilityMappingEvents(eventBus, metadataProjector)
	subscribePillarCacheEvents(eventBus, pillarCacheProjector)
	subscribeRealizationCacheEvents(eventBus, realizationCacheProjector)
	subscribeImportanceCacheEvents(eventBus, importanceCacheProjector)
	subscribeFitScoreCacheEvents(eventBus, fitScoreCacheProjector)
}

func subscribeCapabilityEvents(eventBus events.EventBus, projector *projectors.EnterpriseCapabilityProjector) {
	eventTypes := []string{
		eaPL.EnterpriseCapabilityCreated,
		eaPL.EnterpriseCapabilityUpdated,
		eaPL.EnterpriseCapabilityDeleted,
		eaPL.EnterpriseCapabilityTargetMaturitySet,
	}
	for _, eventType := range eventTypes {
		eventBus.Subscribe(eventType, projector)
	}
}

func subscribeImportanceEvents(eventBus events.EventBus, projector *projectors.EnterpriseStrategicImportanceProjector) {
	eventTypes := []string{
		eaPL.EnterpriseStrategicImportanceSet,
		eaPL.EnterpriseStrategicImportanceUpdated,
		eaPL.EnterpriseStrategicImportanceRemoved,
	}
	for _, eventType := range eventTypes {
		eventBus.Subscribe(eventType, projector)
	}
}

func subscribeCapabilityMappingEvents(eventBus events.EventBus, projector *projectors.DomainCapabilityMetadataProjector) {
	eventTypes := []string{
		cmPL.CapabilityCreated,
		cmPL.CapabilityUpdated,
		cmPL.CapabilityDeleted,
		cmPL.CapabilityParentChanged,
		cmPL.CapabilityLevelChanged,
		cmPL.CapabilityAssignedToDomain,
		cmPL.CapabilityUnassignedFromDomain,
		cmPL.CapabilityMetadataUpdated,
		cmPL.BusinessDomainUpdated,
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

func subscribeRealizationCacheEvents(eventBus events.EventBus, projector *projectors.EARealizationCacheProjector) {
	eventTypes := []string{
		cmPL.SystemLinkedToCapability,
		cmPL.SystemRealizationDeleted,
		cmPL.CapabilityDeleted,
		amPL.ApplicationComponentUpdated,
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
	links := NewEnterpriseArchLinks(sharedAPI.NewHATEOASLinks(""))
	return &routeHTTPHandlers{
		enterpriseCapability: NewEnterpriseCapabilityHandlers(commandBus, readModels, sessionProvider),
		timeSuggestions:      NewTimeSuggestionsHandlers(rm.timeSuggestion, links),
	}
}

func registerRoutes(r chi.Router, h *routeHTTPHandlers, authMiddleware AuthMiddleware, rateLimiter *middleware.RateLimiter) {
	registerEnterpriseCapabilityRoutes(r, h, authMiddleware, rateLimiter)
	registerTimeSuggestionsRoutes(r, h.timeSuggestions, authMiddleware)
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

func registerTimeSuggestionsRoutes(r chi.Router, h *TimeSuggestionsHandlers, authMiddleware AuthMiddleware) {
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequirePermission(authPL.PermEnterpriseArchRead))
		r.Get("/time-suggestions", h.GetTimeSuggestions)
	})
}
