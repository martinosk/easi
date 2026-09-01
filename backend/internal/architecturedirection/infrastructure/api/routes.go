package api

import (
	"net/http"

	"easi/backend/internal/architecturedirection/application/handlers"
	"easi/backend/internal/architecturedirection/application/projectors"
	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/infrastructure/metamodel"
	"easi/backend/internal/architecturedirection/infrastructure/repositories"
	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	authPL "easi/backend/internal/auth/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
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

type RoutesDeps struct {
	Router          chi.Router
	CommandBus      *cqrs.InMemoryCommandBus
	EventStore      eventstore.EventStore
	EventBus        events.EventBus
	DB              *database.TenantAwareDB
	HATEOAS         *sharedAPI.HATEOASLinks
	AuthMiddleware  AuthMiddleware
	SessionProvider authPL.SessionProvider
}

func SetupRoutes(deps RoutesDeps) error {
	nodeCache := readmodels.NewCapabilityNodeCacheReadModel(deps.DB)
	referenceCache := readmodels.NewReferenceCacheReadModel(deps.DB)
	realizationCache := readmodels.NewRealizationCacheReadModel(deps.DB)
	lookups := newReferenceLookups(nodeCache, referenceCache, realizationCache)

	subscribeCacheEvents(deps.EventBus, nodeCache)
	subscribeReferenceCacheEvents(deps.EventBus, referenceCache, realizationCache)
	subscribeSuggestionCacheEvents(deps.EventBus, deps.DB)

	setupTimeAssessmentRoutes(deps, lookups.directRealization)
	setupRealizationRoleRoutes(deps, lookups.directRealization)
	setupCapabilityJourneyRoutes(deps, lookups)
	return nil
}

func subscribeMany(eventBus events.EventBus, handler events.EventHandler, eventTypes ...string) {
	for _, eventType := range eventTypes {
		eventBus.Subscribe(eventType, handler)
	}
}

func setupTimeAssessmentRoutes(deps RoutesDeps, directRealization services.DirectRealizationLookup) {
	readModel := readmodels.NewTimeAssessmentReadModel(deps.DB)
	repo := repositories.NewTimeAssessmentRepository(deps.EventStore)

	subscribeTimeAssessmentEvents(deps.EventBus, readModel)
	deps.CommandBus.Register("AssessRealization", handlers.NewAssessRealizationHandler(repo, readModel, directRealization))
	deps.CommandBus.Register("RemoveTimeAssessment", handlers.NewRemoveTimeAssessmentHandler(repo, readModel))
	deps.EventBus.Subscribe(cmPL.SystemRealizationDeleted, projectors.NewTimeAssessmentDeletionReactor(readModel, deps.CommandBus))

	links := NewTimeAssessmentLinks(deps.HATEOAS)
	view := readmodels.NewTimeAssessmentView(readModel, newTimeSuggestionReadModel(deps.DB))
	httpHandlers := NewTimeAssessmentHandlers(deps.CommandBus, view, links)

	registerTimeAssessmentRoutes(deps.Router, httpHandlers, deps.AuthMiddleware)
}

func newTimeSuggestionReadModel(db *database.TenantAwareDB) *readmodels.TimeSuggestionReadModel {
	pillarsGateway := metamodel.NewLocalStrategyPillarsGateway(readmodels.NewStrategyPillarCacheReadModel(db))
	return readmodels.NewTimeSuggestionReadModel(db, pillarsGateway)
}

func subscribeTimeAssessmentEvents(eventBus events.EventBus, rm *readmodels.TimeAssessmentReadModel) {
	subscribeMany(eventBus, projectors.NewTimeAssessmentProjector(rm),
		pl.TimeAssessmentRecorded, pl.TimeAssessmentRemoved)
	subscribeMany(eventBus, projectors.NewTimeAssessmentReferenceProjector(rm),
		cmPL.CapabilityDeleted, cmPL.CapabilityCreated, cmPL.CapabilityUpdated,
		amPL.ApplicationComponentDeleted, amPL.ApplicationComponentCreated, amPL.ApplicationComponentUpdated,
		authPL.UserCreated)
}

func registerTimeAssessmentRoutes(r chi.Router, h *TimeAssessmentHandlers, authMiddleware AuthMiddleware) {
	registerDomainReadCollection(r, "/time-assessments", authMiddleware, func(r chi.Router) {
		r.Get("/", h.GetTimeAssessments)
		r.Get("/rollups", h.GetTimeAssessmentRollups)
	})
	registerPairResourceRoutes(r, "/capabilities/{id}/components/{componentId}/time-assessment", authMiddleware,
		pairResourceHandlers{get: h.GetTimeAssessment, put: h.PutTimeAssessment, delete: h.DeleteTimeAssessment})
}

func registerDomainReadCollection(r chi.Router, pattern string, authMiddleware AuthMiddleware, register func(chi.Router)) {
	r.Route(pattern, func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermDomainsRead))
			register(r)
		})
	})
}

type pairResourceHandlers struct {
	get    http.HandlerFunc
	put    http.HandlerFunc
	delete http.HandlerFunc
}

func registerPairResourceRoutes(r chi.Router, pattern string, authMiddleware AuthMiddleware, h pairResourceHandlers) {
	r.Route(pattern, func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermDomainsRead))
			r.Get("/", h.get)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermArchitectureDirectionWrite))
			r.Put("/", h.put)
			r.Delete("/", h.delete)
		})
	})
}

func setupRealizationRoleRoutes(deps RoutesDeps, directRealization services.DirectRealizationLookup) {
	readModel := readmodels.NewRealizationRoleReadModel(deps.DB)
	repo := repositories.NewRealizationRolesRepository(deps.EventStore)

	subscribeRealizationRoleEvents(deps.EventBus, readModel)
	deps.CommandBus.Register("AssignRealizationRole", handlers.NewAssignRealizationRoleHandler(repo, readModel, directRealization))
	deps.CommandBus.Register("ClearRealizationRole", handlers.NewClearRealizationRoleHandler(repo, readModel))
	deps.EventBus.Subscribe(cmPL.SystemRealizationDeleted, projectors.NewRealizationRoleDeletionReactor(readModel, deps.CommandBus))

	links := NewRealizationRoleLinks(deps.HATEOAS)
	httpHandlers := NewRealizationRoleHandlers(deps.CommandBus, readModel, links)

	registerRealizationRoleRoutes(deps.Router, httpHandlers, deps.AuthMiddleware)
}

func subscribeRealizationRoleEvents(eventBus events.EventBus, rm *readmodels.RealizationRoleReadModel) {
	subscribeMany(eventBus, projectors.NewRealizationRoleProjector(rm),
		pl.RealizationRoleAssigned, pl.RealizationRoleCleared)
	subscribeMany(eventBus, projectors.NewRealizationRoleReferenceProjector(rm),
		cmPL.CapabilityDeleted, cmPL.CapabilityCreated, cmPL.CapabilityUpdated,
		amPL.ApplicationComponentDeleted, amPL.ApplicationComponentCreated, amPL.ApplicationComponentUpdated)
}

func setupCapabilityJourneyRoutes(deps RoutesDeps, lookups referenceLookups) {
	readModel := readmodels.NewCapabilityJourneyReadModel(deps.DB)
	repo := repositories.NewCapabilityJourneyRepository(deps.EventStore)

	subscribeCapabilityJourneyEvents(deps.EventBus, readModel)

	refs := handlers.JourneyReferenceChecks{
		CapabilityExists:              lookups.capabilityExists,
		ComponentExists:               lookups.componentExists,
		DomainExists:                  lookups.domainExists,
		CapabilityEffectivelyInDomain: lookups.capabilityEffectivelyInDomain,
	}
	deps.CommandBus.Register("PlanJourney", handlers.NewPlanJourneyHandler(repo, readModel, refs, lookups.capabilityMaturity))
	deps.CommandBus.Register("StartJourney", handlers.NewStartJourneyHandler(repo))
	deps.CommandBus.Register("CompleteJourney", handlers.NewCompleteJourneyHandler(repo))
	deps.CommandBus.Register("AbandonJourney", handlers.NewAbandonJourneyHandler(repo))
	deps.CommandBus.Register("UpdateJourneyProgress", handlers.NewUpdateJourneyProgressHandler(repo))
	deps.CommandBus.Register("UpdateJourneyDetails", handlers.NewUpdateJourneyDetailsHandler(repo))
	deps.CommandBus.Register("ChangeJourneySourceApplications", handlers.NewChangeJourneySourceApplicationsHandler(repo, lookups.componentExists))
	deps.CommandBus.Register("AddJourneyMilestone", handlers.NewAddJourneyMilestoneHandler(repo))
	deps.CommandBus.Register("UpdateJourneyMilestone", handlers.NewUpdateJourneyMilestoneHandler(repo))
	deps.CommandBus.Register("RemoveJourneyMilestone", handlers.NewRemoveJourneyMilestoneHandler(repo))
	deps.CommandBus.Register("ReorderJourneyMilestones", handlers.NewReorderJourneyMilestonesHandler(repo))

	links := NewCapabilityJourneyLinks(deps.HATEOAS)
	httpHandlers := NewCapabilityJourneyHandlers(deps.CommandBus, readModel, links)

	registerCapabilityJourneyRoutes(deps.Router, httpHandlers, deps.AuthMiddleware)
}

func subscribeCapabilityJourneyEvents(eventBus events.EventBus, rm *readmodels.CapabilityJourneyReadModel) {
	subscribeMany(eventBus, projectors.NewCapabilityJourneyProjector(rm),
		pl.JourneyPlanned, pl.JourneyStarted, pl.JourneyCompleted, pl.JourneyAbandoned,
		pl.JourneyProgressUpdated, pl.JourneyDetailsUpdated, pl.JourneySourceApplicationsChanged,
		pl.JourneyMilestoneAdded, pl.JourneyMilestoneUpdated, pl.JourneyMilestoneRemoved, pl.JourneyMilestonesReordered)
	subscribeMany(eventBus, projectors.NewCapabilityJourneyReferenceProjector(rm),
		cmPL.CapabilityCreated, cmPL.CapabilityUpdated, cmPL.CapabilityDeleted,
		cmPL.BusinessDomainCreated, cmPL.BusinessDomainUpdated, cmPL.BusinessDomainDeleted,
		amPL.ApplicationComponentCreated, amPL.ApplicationComponentUpdated, amPL.ApplicationComponentDeleted,
		authPL.UserCreated)
}

func registerCapabilityJourneyRoutes(r chi.Router, h *CapabilityJourneyHandlers, authMiddleware AuthMiddleware) {
	r.Route("/capabilities/{id}/journey", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermDomainsRead))
			r.Get("/", h.GetJourneyForCapability)
			r.Get("/history", h.GetJourneyHistory)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermArchitectureDirectionWrite))
			r.Post("/", h.CaptureJourney)
		})
	})

	registerDomainReadCollection(r, "/capability-journeys", authMiddleware, func(r chi.Router) {
		r.Get("/", h.GetCapabilityJourneys)
	})

	r.Route("/capability-journeys/{journeyId}", func(r chi.Router) {
		r.Use(authMiddleware.RequirePermission(authPL.PermArchitectureDirectionWrite))
		r.Post("/start", h.StartJourney)
		r.Post("/complete", h.CompleteJourney)
		r.Post("/abandon", h.AbandonJourney)
		r.Put("/details", h.PutJourneyDetails)
		r.Put("/progress", h.PutJourneyProgress)
		r.Put("/source-applications", h.PutJourneySourceApplications)
		r.Post("/milestones", h.PostJourneyMilestone)
		r.Put("/milestones/{milestoneId}", h.PutJourneyMilestone)
		r.Delete("/milestones/{milestoneId}", h.DeleteJourneyMilestone)
		r.Put("/milestone-order", h.PutJourneyMilestoneOrder)
	})
}

func registerRealizationRoleRoutes(r chi.Router, h *RealizationRoleHandlers, authMiddleware AuthMiddleware) {
	registerDomainReadCollection(r, "/realization-roles", authMiddleware, func(r chi.Router) {
		r.Get("/", h.GetRealizationRoles)
	})
	registerPairResourceRoutes(r, "/capabilities/{id}/components/{componentId}/realization-role", authMiddleware,
		pairResourceHandlers{get: h.GetRealizationRole, put: h.PutRealizationRole, delete: h.DeleteRealizationRole})
}

func subscribeCacheEvents(eventBus events.EventBus, nodeCache *readmodels.CapabilityNodeCacheReadModel) {
	subscribeMany(eventBus, projectors.NewCapabilityNodeCacheProjector(nodeCache, nodeCache.BusinessDomainName),
		cmPL.CapabilityCreated, cmPL.CapabilityUpdated, cmPL.CapabilityDeleted,
		cmPL.CapabilityParentChanged, cmPL.CapabilityLevelChanged,
		cmPL.CapabilityAssignedToDomain, cmPL.CapabilityUnassignedFromDomain,
		cmPL.CapabilityMetadataUpdated, cmPL.BusinessDomainUpdated)
}

func subscribeReferenceCacheEvents(eventBus events.EventBus, referenceCache *readmodels.ReferenceCacheReadModel, realizationCache *readmodels.RealizationCacheReadModel) {
	subscribeMany(eventBus, projectors.NewReferenceCacheProjector(referenceCache),
		amPL.ApplicationComponentCreated, amPL.ApplicationComponentUpdated, amPL.ApplicationComponentDeleted,
		cmPL.BusinessDomainCreated, cmPL.BusinessDomainUpdated, cmPL.BusinessDomainDeleted)
	subscribeMany(eventBus, projectors.NewRealizationCacheProjector(realizationCache),
		cmPL.SystemLinkedToCapability, cmPL.SystemRealizationDeleted,
		cmPL.CapabilityDeleted, amPL.ApplicationComponentDeleted)
}

func subscribeSuggestionCacheEvents(eventBus events.EventBus, db *database.TenantAwareDB) {
	subscribeMany(eventBus, projectors.NewStrategyPillarCacheProjector(readmodels.NewStrategyPillarCacheReadModel(db)),
		mmPL.MetaModelConfigurationCreated,
		mmPL.StrategyPillarAdded, mmPL.StrategyPillarUpdated, mmPL.StrategyPillarRemoved,
		mmPL.PillarFitConfigurationUpdated)
	subscribeMany(eventBus, projectors.NewEAImportanceCacheProjector(readmodels.NewEAImportanceCacheReadModel(db)),
		cmPL.EffectiveImportanceRecalculated)
	subscribeMany(eventBus, projectors.NewEAFitScoreCacheProjector(readmodels.NewEAFitScoreCacheReadModel(db)),
		cmPL.ApplicationFitScoreSet, cmPL.ApplicationFitScoreRemoved)
}
