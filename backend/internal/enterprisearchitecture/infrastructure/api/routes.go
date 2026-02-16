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
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	"easi/backend/internal/infrastructure/api/middleware"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

func init() {
	registry := sharedAPI.GetErrorRegistry()
	registry.RegisterConflict(handlers.ErrCapabilityHasLinks, "Cannot delete enterprise capability: unlink all domain capabilities first")
}

type AuthMiddleware interface {
	RequirePermission(permission authPL.Permission) func(http.Handler) http.Handler
}

type routeRepositories struct {
	capability *repositories.EnterpriseCapabilityRepository
	link       *repositories.EnterpriseCapabilityLinkRepository
	importance *repositories.EnterpriseStrategicImportanceRepository
}

type routeReadModels struct {
	capability       *readmodels.EnterpriseCapabilityReadModel
	link             *readmodels.EnterpriseCapabilityLinkReadModel
	importance       *readmodels.EnterpriseStrategicImportanceReadModel
	metadata         *readmodels.DomainCapabilityMetadataReadModel
	maturityAnalysis *readmodels.MaturityAnalysisReadModel
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
	Router          chi.Router
	CommandBus      *cqrs.InMemoryCommandBus
	EventStore      eventstore.EventStore
	EventBus        events.EventBus
	DB              *database.TenantAwareDB
	AuthMiddleware  AuthMiddleware
	SessionProvider authPL.SessionProvider
}

func SetupEnterpriseArchitectureRoutes(deps EnterpriseArchRoutesDeps) error {
	repos := initializeRepositories(deps.EventStore)
	rm := initializeReadModels(deps.DB)

	setupEventSubscriptions(deps.EventBus, rm)
	setupCommandHandlers(deps.CommandBus, repos, rm)

	httpHandlers := initializeHTTPHandlers(deps.CommandBus, rm, deps.SessionProvider)
	rateLimiter := middleware.NewRateLimiter(100, 60)
	registerRoutes(deps.Router, httpHandlers, deps.AuthMiddleware, rateLimiter)

	return nil
}

func initializeRepositories(eventStore eventstore.EventStore) *routeRepositories {
	return &routeRepositories{
		capability: repositories.NewEnterpriseCapabilityRepository(eventStore),
		link:       repositories.NewEnterpriseCapabilityLinkRepository(eventStore),
		importance: repositories.NewEnterpriseStrategicImportanceRepository(eventStore),
	}
}

func initializeReadModels(db *database.TenantAwareDB) *routeReadModels {
	pillarCache := readmodels.NewStrategyPillarCacheReadModel(db)
	pillarsGateway := metamodel.NewLocalStrategyPillarsGateway(pillarCache)
	return &routeReadModels{
		capability:       readmodels.NewEnterpriseCapabilityReadModel(db),
		link:             readmodels.NewEnterpriseCapabilityLinkReadModel(db),
		importance:       readmodels.NewEnterpriseStrategicImportanceReadModel(db),
		metadata:         readmodels.NewDomainCapabilityMetadataReadModel(db),
		maturityAnalysis: readmodels.NewMaturityAnalysisReadModel(db),
		timeSuggestion:   readmodels.NewTimeSuggestionReadModel(db, pillarsGateway),
		pillarCache:      pillarCache,
		realizationCache: readmodels.NewEARealizationCacheReadModel(db),
		importanceCache:  readmodels.NewEAImportanceCacheReadModel(db),
		fitScoreCache:    readmodels.NewEAFitScoreCacheReadModel(db),
	}
}

func setupEventSubscriptions(eventBus events.EventBus, rm *routeReadModels) {
	capabilityProjector := projectors.NewEnterpriseCapabilityProjector(rm.capability)
	linkProjector := projectors.NewEnterpriseCapabilityLinkProjector(rm.link)
	importanceProjector := projectors.NewEnterpriseStrategicImportanceProjector(rm.importance)
	metadataProjector := projectors.NewDomainCapabilityMetadataProjector(rm.metadata, rm.capability, rm.link)
	pillarCacheProjector := projectors.NewStrategyPillarCacheProjector(rm.pillarCache)
	realizationCacheProjector := projectors.NewEARealizationCacheProjector(rm.realizationCache)
	importanceCacheProjector := projectors.NewEAImportanceCacheProjector(rm.importanceCache)
	fitScoreCacheProjector := projectors.NewEAFitScoreCacheProjector(rm.fitScoreCache)

	subscribeCapabilityEvents(eventBus, capabilityProjector)
	subscribeLinkEvents(eventBus, linkProjector)
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
		eaPL.EnterpriseCapabilityLinked,
		eaPL.EnterpriseCapabilityUnlinked,
		eaPL.EnterpriseCapabilityTargetMaturitySet,
	}
	for _, eventType := range eventTypes {
		eventBus.Subscribe(eventType, projector)
	}
}

func subscribeLinkEvents(eventBus events.EventBus, projector *projectors.EnterpriseCapabilityLinkProjector) {
	eventTypes := []string{
		eaPL.EnterpriseCapabilityLinked,
		eaPL.EnterpriseCapabilityUnlinked,
		cmPL.CapabilityParentChanged,
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
	commandBus.Register("DeleteEnterpriseCapability", handlers.NewDeleteEnterpriseCapabilityHandler(repos.capability, rm.link))
	commandBus.Register("SetTargetMaturity", handlers.NewSetTargetMaturityHandler(repos.capability))

	commandBus.Register("LinkCapability", handlers.NewLinkCapabilityHandler(repos.link, repos.capability, rm.link))
	commandBus.Register("UnlinkCapability", handlers.NewUnlinkCapabilityHandler(repos.link))

	commandBus.Register("SetEnterpriseStrategicImportance", handlers.NewSetEnterpriseStrategicImportanceHandler(repos.importance, rm.capability, rm.importance))
	commandBus.Register("UpdateEnterpriseStrategicImportance", handlers.NewUpdateEnterpriseStrategicImportanceHandler(repos.importance))
	commandBus.Register("RemoveEnterpriseStrategicImportance", handlers.NewRemoveEnterpriseStrategicImportanceHandler(repos.importance))
}

func initializeHTTPHandlers(commandBus *cqrs.InMemoryCommandBus, rm *routeReadModels, sessionProvider authPL.SessionProvider) *routeHTTPHandlers {
	readModels := &EnterpriseCapabilityReadModels{
		Capability:       rm.capability,
		Link:             rm.link,
		Importance:       rm.importance,
		MaturityAnalysis: rm.maturityAnalysis,
	}
	links := NewEnterpriseArchLinks(sharedAPI.NewHATEOASLinks(""))
	return &routeHTTPHandlers{
		enterpriseCapability: NewEnterpriseCapabilityHandlers(commandBus, readModels, sessionProvider),
		timeSuggestions:      NewTimeSuggestionsHandlers(rm.timeSuggestion, links),
	}
}

func registerRoutes(r chi.Router, h *routeHTTPHandlers, authMiddleware AuthMiddleware, rateLimiter *middleware.RateLimiter) {
	registerEnterpriseCapabilityRoutes(r, h.enterpriseCapability, authMiddleware, rateLimiter)
	registerTimeSuggestionsRoutes(r, h.timeSuggestions, authMiddleware)
}

func registerEnterpriseCapabilityRoutes(r chi.Router, h *EnterpriseCapabilityHandlers, authMiddleware AuthMiddleware, rateLimiter *middleware.RateLimiter) {
	r.Route("/enterprise-capabilities", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermEnterpriseArchRead))
			r.Get("/", h.GetAllEnterpriseCapabilities)
			r.Get("/maturity-analysis", h.GetMaturityAnalysisCandidates)
			r.Get("/{id}", h.GetEnterpriseCapabilityByID)
			r.Get("/{id}/links", h.GetLinkedCapabilities)
			r.Get("/{id}/strategic-importance", h.GetStrategicImportance)
			r.Get("/{id}/maturity-gap", h.GetMaturityGapDetail)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermEnterpriseArchWrite))
			r.Use(middleware.RateLimitMiddleware(rateLimiter))
			r.Post("/", h.CreateEnterpriseCapability)
			r.Put("/{id}", h.UpdateEnterpriseCapability)
			r.Put("/{id}/target-maturity", h.SetTargetMaturity)
			r.Post("/{id}/links", h.LinkCapability)
			r.Post("/{id}/strategic-importance", h.SetStrategicImportance)
			r.Put("/{id}/strategic-importance/{importanceId}", h.UpdateStrategicImportance)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermEnterpriseArchDelete))
			r.Use(middleware.RateLimitMiddleware(rateLimiter))
			r.Delete("/{id}", h.DeleteEnterpriseCapability)
			r.Delete("/{id}/links/{linkId}", h.UnlinkCapability)
			r.Delete("/{id}/strategic-importance/{importanceId}", h.RemoveStrategicImportance)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequirePermission(authPL.PermEnterpriseArchRead))
		r.Get("/domain-capabilities/{domainCapabilityId}/enterprise-capability", h.GetEnterpriseCapabilityForDomainCapability)
		r.Get("/domain-capabilities/{domainCapabilityId}/enterprise-link-status", h.GetCapabilityLinkStatus)
		r.Get("/domain-capabilities/enterprise-link-status", h.GetBatchCapabilityLinkStatus)
	})
}

func registerTimeSuggestionsRoutes(r chi.Router, h *TimeSuggestionsHandlers, authMiddleware AuthMiddleware) {
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequirePermission(authPL.PermEnterpriseArchRead))
		r.Get("/time-suggestions", h.GetTimeSuggestions)
	})
}
