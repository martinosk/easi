package api

import (
	"net/http"

	archReadModels "easi/backend/internal/architecturemodeling/application/readmodels"
	authValueObjects "easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/application/projectors"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/infrastructure/adapters"
	"easi/backend/internal/capabilitymapping/infrastructure/metamodel"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	platformAPI "easi/backend/internal/platform/infrastructure/api"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequirePermission(permission authValueObjects.Permission) func(http.Handler) http.Handler
}

type routeConfig struct {
	commandBus             *cqrs.InMemoryCommandBus
	eventStore             eventstore.EventStore
	eventBus               events.EventBus
	db                     *database.TenantAwareDB
	hateoas                *sharedAPI.HATEOASLinks
	maturityScaleGateway   metamodel.MaturityScaleGateway
	strategyPillarsGateway metamodel.StrategyPillarsGateway
	sessionManager         *session.SessionManager
	authMiddleware         AuthMiddleware
}

func SetupCapabilityMappingRoutes(
	r chi.Router,
	commandBus *cqrs.InMemoryCommandBus,
	eventStore eventstore.EventStore,
	eventBus events.EventBus,
	db *database.TenantAwareDB,
	hateoas *sharedAPI.HATEOASLinks,
	sessionManager *session.SessionManager,
	authMiddleware AuthMiddleware,
) error {
	pillarsGateway := metamodel.NewDirectStrategyPillarsGateway(db)
	return SetupCapabilityMappingRoutesWithGateways(r, commandBus, eventStore, eventBus, db, hateoas, nil, pillarsGateway, sessionManager, authMiddleware)
}

func SetupCapabilityMappingRoutesWithGateway(
	r chi.Router,
	commandBus *cqrs.InMemoryCommandBus,
	eventStore eventstore.EventStore,
	eventBus events.EventBus,
	db *database.TenantAwareDB,
	hateoas *sharedAPI.HATEOASLinks,
	gateway metamodel.MaturityScaleGateway,
	sessionManager *session.SessionManager,
	authMiddleware AuthMiddleware,
) error {
	return SetupCapabilityMappingRoutesWithGateways(r, commandBus, eventStore, eventBus, db, hateoas, gateway, nil, sessionManager, authMiddleware)
}

func SetupCapabilityMappingRoutesWithGateways(
	r chi.Router,
	commandBus *cqrs.InMemoryCommandBus,
	eventStore eventstore.EventStore,
	eventBus events.EventBus,
	db *database.TenantAwareDB,
	hateoas *sharedAPI.HATEOASLinks,
	maturityGateway metamodel.MaturityScaleGateway,
	pillarsGateway metamodel.StrategyPillarsGateway,
	sessionManager *session.SessionManager,
	authMiddleware AuthMiddleware,
) error {
	config := &routeConfig{commandBus, eventStore, eventBus, db, hateoas, maturityGateway, pillarsGateway, sessionManager, authMiddleware}

	repos := initializeRepositories(config.eventStore)
	readModels := initializeReadModels(config.db)

	setupEventSubscriptions(config.eventBus, readModels, config.strategyPillarsGateway)
	setupCascadingDeleteHandlers(config.eventBus, config.commandBus, readModels)
	setupCommandHandlers(config.commandBus, repos, readModels, config.strategyPillarsGateway)
	setupMetaModelEventHandlers(config.eventBus, config.maturityScaleGateway)

	rateLimiter := platformAPI.NewRateLimiter(100, 60)
	httpHandlers := initializeHTTPHandlers(config.commandBus, readModels, config.hateoas, config.maturityScaleGateway, config.strategyPillarsGateway, config.sessionManager)
	registerRoutes(r, httpHandlers, config.authMiddleware, rateLimiter)

	return nil
}

func setupMetaModelEventHandlers(eventBus events.EventBus, gateway metamodel.MaturityScaleGateway) {
	if gateway == nil {
		return
	}

	maturityScaleUpdatedHandler := handlers.NewMaturityScaleConfigUpdatedHandler(gateway)
	eventBus.Subscribe("MaturityScaleConfigUpdated", maturityScaleUpdatedHandler)
	eventBus.Subscribe("MaturityScaleConfigReset", maturityScaleUpdatedHandler)
}

type routeRepositories struct {
	capability          *repositories.CapabilityRepository
	dependency          *repositories.DependencyRepository
	realization         *repositories.RealizationRepository
	businessDomain      *repositories.BusinessDomainRepository
	domainAssignment    *repositories.BusinessDomainAssignmentRepository
	strategyImportance  *repositories.StrategyImportanceRepository
	applicationFitScore *repositories.ApplicationFitScoreRepository
}

type routeReadModels struct {
	capability             *readmodels.CapabilityReadModel
	dependency             *readmodels.DependencyReadModel
	realization            *readmodels.RealizationReadModel
	component              *archReadModels.ApplicationComponentReadModel
	businessDomain         *readmodels.BusinessDomainReadModel
	domainAssignment       *readmodels.DomainCapabilityAssignmentReadModel
	strategyImportance     *readmodels.StrategyImportanceReadModel
	applicationFitScore    *readmodels.ApplicationFitScoreReadModel
	strategicFitAnalysis   *readmodels.StrategicFitAnalysisReadModel
	componentFitComparison *readmodels.ComponentFitComparisonReadModel
}

type routeHTTPHandlers struct {
	capability           *CapabilityHandlers
	dependency           *DependencyHandlers
	realization          *RealizationHandlers
	maturityLevel        *MaturityLevelHandlers
	businessDomain       *BusinessDomainHandlers
	strategyImportance   *StrategyImportanceHandlers
	applicationFitScore  *ApplicationFitScoreHandlers
	strategicFitAnalysis *StrategicFitAnalysisHandlers
}

func initializeRepositories(eventStore eventstore.EventStore) *routeRepositories {
	return &routeRepositories{
		capability:          repositories.NewCapabilityRepository(eventStore),
		dependency:          repositories.NewDependencyRepository(eventStore),
		realization:         repositories.NewRealizationRepository(eventStore),
		businessDomain:      repositories.NewBusinessDomainRepository(eventStore),
		domainAssignment:    repositories.NewBusinessDomainAssignmentRepository(eventStore),
		strategyImportance:  repositories.NewStrategyImportanceRepository(eventStore),
		applicationFitScore: repositories.NewApplicationFitScoreRepository(eventStore),
	}
}

func initializeReadModels(db *database.TenantAwareDB) *routeReadModels {
	return &routeReadModels{
		capability:             readmodels.NewCapabilityReadModel(db),
		dependency:             readmodels.NewDependencyReadModel(db),
		realization:            readmodels.NewRealizationReadModel(db),
		component:              archReadModels.NewApplicationComponentReadModel(db),
		businessDomain:         readmodels.NewBusinessDomainReadModel(db),
		domainAssignment:       readmodels.NewDomainCapabilityAssignmentReadModel(db),
		strategyImportance:     readmodels.NewStrategyImportanceReadModel(db),
		applicationFitScore:    readmodels.NewApplicationFitScoreReadModel(db),
		strategicFitAnalysis:   readmodels.NewStrategicFitAnalysisReadModel(db),
		componentFitComparison: readmodels.NewComponentFitComparisonReadModel(db),
	}
}

func setupEventSubscriptions(eventBus events.EventBus, rm *routeReadModels, pillarsGateway metamodel.StrategyPillarsGateway) {
	capabilityProjector := projectors.NewCapabilityProjector(rm.capability, rm.domainAssignment)
	dependencyProjector := projectors.NewDependencyProjector(rm.dependency)
	realizationProjector := projectors.NewRealizationProjector(rm.realization, rm.capability, rm.component)
	businessDomainProjector := projectors.NewBusinessDomainProjector(rm.businessDomain)
	domainAssignmentProjector := projectors.NewBusinessDomainAssignmentProjector(rm.domainAssignment, rm.businessDomain, rm.capability)
	strategyImportanceProjector := projectors.NewStrategyImportanceProjector(rm.strategyImportance, rm.businessDomain, rm.capability, pillarsGateway)
	applicationFitScoreProjector := projectors.NewApplicationFitScoreProjector(rm.applicationFitScore, rm.component, pillarsGateway)

	subscribeCapabilityEvents(eventBus, capabilityProjector)
	subscribeDependencyEvents(eventBus, dependencyProjector)
	subscribeRealizationEvents(eventBus, realizationProjector)
	subscribeBusinessDomainEvents(eventBus, businessDomainProjector)
	subscribeDomainAssignmentEvents(eventBus, domainAssignmentProjector)
	subscribeStrategyImportanceEvents(eventBus, strategyImportanceProjector)
	subscribeApplicationFitScoreEvents(eventBus, applicationFitScoreProjector)
}

func subscribeCapabilityEvents(eventBus events.EventBus, projector *projectors.CapabilityProjector) {
	events := []string{"CapabilityCreated", "CapabilityUpdated", "CapabilityMetadataUpdated",
		"CapabilityExpertAdded", "CapabilityTagAdded", "CapabilityParentChanged", "CapabilityDeleted"}
	for _, event := range events {
		eventBus.Subscribe(event, projector)
	}
}

func subscribeDependencyEvents(eventBus events.EventBus, projector *projectors.DependencyProjector) {
	eventBus.Subscribe("CapabilityDependencyCreated", projector)
	eventBus.Subscribe("CapabilityDependencyDeleted", projector)
}

func subscribeRealizationEvents(eventBus events.EventBus, projector *projectors.RealizationProjector) {
	events := []string{"SystemLinkedToCapability", "SystemRealizationUpdated",
		"SystemRealizationDeleted", "CapabilityParentChanged", "CapabilityUpdated",
		"ApplicationComponentUpdated"}
	for _, event := range events {
		eventBus.Subscribe(event, projector)
	}
}

func subscribeBusinessDomainEvents(eventBus events.EventBus, projector *projectors.BusinessDomainProjector) {
	events := []string{"BusinessDomainCreated", "BusinessDomainUpdated", "BusinessDomainDeleted",
		"CapabilityAssignedToDomain", "CapabilityUnassignedFromDomain"}
	for _, event := range events {
		eventBus.Subscribe(event, projector)
	}
}

func subscribeDomainAssignmentEvents(eventBus events.EventBus, projector *projectors.BusinessDomainAssignmentProjector) {
	events := []string{"CapabilityAssignedToDomain", "CapabilityUnassignedFromDomain"}
	for _, event := range events {
		eventBus.Subscribe(event, projector)
	}
}

func subscribeStrategyImportanceEvents(eventBus events.EventBus, projector *projectors.StrategyImportanceProjector) {
	events := []string{"StrategyImportanceSet", "StrategyImportanceUpdated", "StrategyImportanceRemoved"}
	for _, event := range events {
		eventBus.Subscribe(event, projector)
	}
}

func subscribeApplicationFitScoreEvents(eventBus events.EventBus, projector *projectors.ApplicationFitScoreProjector) {
	events := []string{"ApplicationFitScoreSet", "ApplicationFitScoreUpdated", "ApplicationFitScoreRemoved"}
	for _, event := range events {
		eventBus.Subscribe(event, projector)
	}
}

func setupCascadingDeleteHandlers(eventBus events.EventBus, commandBus *cqrs.InMemoryCommandBus, rm *routeReadModels) {
	onCapabilityDeletedHandler := handlers.NewOnCapabilityDeletedHandler(commandBus, rm.domainAssignment)
	onBusinessDomainDeletedHandler := handlers.NewOnBusinessDomainDeletedHandler(commandBus, rm.domainAssignment)
	onCapabilityParentChangedHandler := handlers.NewOnCapabilityParentChangedHandler(commandBus, rm.domainAssignment, rm.capability)
	onCapabilityDeletedImportanceHandler := handlers.NewOnCapabilityDeletedImportanceHandler(rm.strategyImportance)
	onBusinessDomainDeletedImportanceHandler := handlers.NewOnBusinessDomainDeletedImportanceHandler(rm.strategyImportance)

	eventBus.Subscribe("CapabilityDeleted", onCapabilityDeletedHandler)
	eventBus.Subscribe("CapabilityDeleted", onCapabilityDeletedImportanceHandler)
	eventBus.Subscribe("BusinessDomainDeleted", onBusinessDomainDeletedHandler)
	eventBus.Subscribe("BusinessDomainDeleted", onBusinessDomainDeletedImportanceHandler)
	eventBus.Subscribe("CapabilityParentChanged", onCapabilityParentChangedHandler)
}

func setupCommandHandlers(commandBus *cqrs.InMemoryCommandBus, repos *routeRepositories, rm *routeReadModels, pillarsGateway metamodel.StrategyPillarsGateway) {
	registerCapabilityCommands(commandBus, repos.capability, rm.capability)
	registerDependencyCommands(commandBus, repos.dependency, repos.capability)
	registerRealizationCommands(commandBus, repos.realization, repos.capability, rm.component)
	registerBusinessDomainCommands(commandBus, repos.businessDomain, rm.businessDomain, rm.domainAssignment)
	registerDomainAssignmentCommands(commandBus, repos.domainAssignment, repos.capability, rm.businessDomain, rm.domainAssignment)
	registerStrategyImportanceCommands(commandBus, handlers.StrategyImportanceDeps{
		ImportanceRepo:   repos.strategyImportance,
		DomainReader:     rm.businessDomain,
		CapabilityReader: rm.capability,
		ImportanceReader: rm.strategyImportance,
		PillarsGateway:   pillarsGateway,
	})
	registerApplicationFitScoreCommands(commandBus, handlers.ApplicationFitScoreDeps{
		FitScoreRepo:   repos.applicationFitScore,
		FitScoreReader: rm.applicationFitScore,
		PillarsGateway: pillarsGateway,
	})
}

func registerCapabilityCommands(commandBus *cqrs.InMemoryCommandBus, repo *repositories.CapabilityRepository, rm *readmodels.CapabilityReadModel) {
	childrenChecker := adapters.NewCapabilityChildrenCheckerAdapter(rm)
	deletionService := services.NewCapabilityDeletionService(childrenChecker)

	commandBus.Register("CreateCapability", handlers.NewCreateCapabilityHandler(repo))
	commandBus.Register("UpdateCapability", handlers.NewUpdateCapabilityHandler(repo))
	commandBus.Register("UpdateCapabilityMetadata", handlers.NewUpdateCapabilityMetadataHandler(repo))
	commandBus.Register("AddCapabilityExpert", handlers.NewAddCapabilityExpertHandler(repo))
	commandBus.Register("AddCapabilityTag", handlers.NewAddCapabilityTagHandler(repo))
	commandBus.Register("ChangeCapabilityParent", handlers.NewChangeCapabilityParentHandler(repo, rm))
	commandBus.Register("DeleteCapability", handlers.NewDeleteCapabilityHandler(repo, deletionService))
}

func registerDependencyCommands(commandBus *cqrs.InMemoryCommandBus, depRepo *repositories.DependencyRepository, capRepo *repositories.CapabilityRepository) {
	commandBus.Register("CreateCapabilityDependency", handlers.NewCreateCapabilityDependencyHandler(depRepo, capRepo))
	commandBus.Register("DeleteCapabilityDependency", handlers.NewDeleteCapabilityDependencyHandler(depRepo))
}

func registerRealizationCommands(commandBus *cqrs.InMemoryCommandBus, realRepo *repositories.RealizationRepository, capRepo *repositories.CapabilityRepository, compRM *archReadModels.ApplicationComponentReadModel) {
	commandBus.Register("LinkSystemToCapability", handlers.NewLinkSystemToCapabilityHandler(realRepo, capRepo, compRM))
	commandBus.Register("UpdateSystemRealization", handlers.NewUpdateSystemRealizationHandler(realRepo))
	commandBus.Register("DeleteSystemRealization", handlers.NewDeleteSystemRealizationHandler(realRepo))
}

func registerBusinessDomainCommands(commandBus *cqrs.InMemoryCommandBus, domainRepo *repositories.BusinessDomainRepository, domainRM *readmodels.BusinessDomainReadModel, assignmentRM *readmodels.DomainCapabilityAssignmentReadModel) {
	assignmentChecker := adapters.NewBusinessDomainAssignmentCheckerAdapter(assignmentRM)
	deletionService := services.NewBusinessDomainDeletionService(assignmentChecker)

	commandBus.Register("CreateBusinessDomain", handlers.NewCreateBusinessDomainHandler(domainRepo, domainRM))
	commandBus.Register("UpdateBusinessDomain", handlers.NewUpdateBusinessDomainHandler(domainRepo, domainRM))
	commandBus.Register("DeleteBusinessDomain", handlers.NewDeleteBusinessDomainHandler(domainRepo, deletionService))
}

func registerDomainAssignmentCommands(commandBus *cqrs.InMemoryCommandBus, assignRepo *repositories.BusinessDomainAssignmentRepository, capRepo *repositories.CapabilityRepository, domainRM *readmodels.BusinessDomainReadModel, assignmentRM *readmodels.DomainCapabilityAssignmentReadModel) {
	commandBus.Register("AssignCapabilityToDomain", handlers.NewAssignCapabilityToDomainHandler(assignRepo, capRepo, domainRM, assignmentRM))
	commandBus.Register("UnassignCapabilityFromDomain", handlers.NewUnassignCapabilityFromDomainHandler(assignRepo))
}

func registerStrategyImportanceCommands(commandBus *cqrs.InMemoryCommandBus, deps handlers.StrategyImportanceDeps) {
	commandBus.Register("SetStrategyImportance", handlers.NewSetStrategyImportanceHandler(deps))
	commandBus.Register("UpdateStrategyImportance", handlers.NewUpdateStrategyImportanceHandler(deps.ImportanceRepo))
	commandBus.Register("RemoveStrategyImportance", handlers.NewRemoveStrategyImportanceHandler(deps.ImportanceRepo))
}

func registerApplicationFitScoreCommands(commandBus *cqrs.InMemoryCommandBus, deps handlers.ApplicationFitScoreDeps) {
	commandBus.Register("SetApplicationFitScore", handlers.NewSetApplicationFitScoreHandler(deps))
	commandBus.Register("UpdateApplicationFitScore", handlers.NewUpdateApplicationFitScoreHandler(deps.FitScoreRepo))
	commandBus.Register("RemoveApplicationFitScore", handlers.NewRemoveApplicationFitScoreHandler(deps.FitScoreRepo))
}

func initializeHTTPHandlers(commandBus *cqrs.InMemoryCommandBus, rm *routeReadModels, hateoas *sharedAPI.HATEOASLinks, maturityGateway metamodel.MaturityScaleGateway, pillarsGateway metamodel.StrategyPillarsGateway, sessionManager *session.SessionManager) *routeHTTPHandlers {
	businessDomainReadModels := &BusinessDomainReadModels{
		Domain:      rm.businessDomain,
		Assignment:  rm.domainAssignment,
		Capability:  rm.capability,
		Realization: rm.realization,
	}

	return &routeHTTPHandlers{
		capability:           NewCapabilityHandlers(commandBus, rm.capability, hateoas),
		dependency:           NewDependencyHandlers(commandBus, rm.dependency, hateoas),
		realization:          NewRealizationHandlers(commandBus, rm.realization, hateoas),
		maturityLevel:        NewMaturityLevelHandlers(maturityGateway),
		businessDomain:       NewBusinessDomainHandlers(commandBus, businessDomainReadModels, hateoas),
		strategyImportance:   NewStrategyImportanceHandlers(commandBus, rm.strategyImportance, hateoas),
		applicationFitScore:  NewApplicationFitScoreHandlers(commandBus, rm.applicationFitScore, rm.componentFitComparison, hateoas, sessionManager),
		strategicFitAnalysis: NewStrategicFitAnalysisHandlers(rm.strategicFitAnalysis, pillarsGateway, sessionManager),
	}
}

func registerRoutes(r chi.Router, h *routeHTTPHandlers, authMiddleware AuthMiddleware, rateLimiter *platformAPI.RateLimiter) {
	registerCapabilityRoutes(r, h, authMiddleware)
	registerDependencyRoutes(r, h, authMiddleware)
	registerRealizationRoutes(r, h, authMiddleware)
	registerBusinessDomainRoutes(r, h, authMiddleware)
	registerStrategyImportanceRoutes(r, h)
	registerApplicationFitScoreRoutes(r, h, authMiddleware, rateLimiter)
	registerStrategicFitAnalysisRoutes(r, h, authMiddleware)
}

func registerCapabilityRoutes(r chi.Router, h *routeHTTPHandlers, authMiddleware AuthMiddleware) {
	r.Route("/capabilities", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermCapabilitiesRead))
			r.Get("/metadata", h.maturityLevel.GetMetadataIndex)
			r.Get("/metadata/maturity-levels", h.maturityLevel.GetMaturityLevels)
			r.Get("/metadata/statuses", h.maturityLevel.GetStatuses)
			r.Get("/metadata/ownership-models", h.maturityLevel.GetOwnershipModels)
			r.Get("/", h.capability.GetAllCapabilities)
			r.Get("/{id}", h.capability.GetCapabilityByID)
			r.Get("/{id}/children", h.capability.GetCapabilityChildren)
			r.Get("/{id}/systems", h.realization.GetSystemsByCapability)
			r.Get("/{id}/dependencies/outgoing", h.dependency.GetOutgoingDependencies)
			r.Get("/{id}/dependencies/incoming", h.dependency.GetIncomingDependencies)
			r.Get("/{id}/business-domains", h.businessDomain.GetDomainsForCapability)
			r.Get("/{id}/importance", h.strategyImportance.GetImportanceByCapability)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermCapabilitiesWrite))
			r.Post("/", h.capability.CreateCapability)
			r.Post("/{id}/systems", h.realization.LinkSystemToCapability)
			r.Put("/{id}", h.capability.UpdateCapability)
			r.Put("/{id}/metadata", h.capability.UpdateCapabilityMetadata)
			r.Patch("/{id}/parent", h.capability.ChangeCapabilityParent)
			r.Post("/{id}/experts", h.capability.AddCapabilityExpert)
			r.Post("/{id}/tags", h.capability.AddCapabilityTag)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermCapabilitiesDelete))
			r.Delete("/{id}", h.capability.DeleteCapability)
		})
	})
}

func registerDependencyRoutes(r chi.Router, h *routeHTTPHandlers, authMiddleware AuthMiddleware) {
	r.Route("/capability-dependencies", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermCapabilitiesRead))
			r.Get("/", h.dependency.GetAllDependencies)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermCapabilitiesWrite))
			r.Post("/", h.dependency.CreateDependency)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermCapabilitiesDelete))
			r.Delete("/{id}", h.dependency.DeleteDependency)
		})
	})
}

func registerRealizationRoutes(r chi.Router, h *routeHTTPHandlers, authMiddleware AuthMiddleware) {
	r.Route("/capability-realizations", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermCapabilitiesRead))
			r.Get("/by-component/{componentId}", h.realization.GetCapabilitiesByComponent)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermCapabilitiesWrite))
			r.Put("/{id}", h.realization.UpdateRealization)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermCapabilitiesDelete))
			r.Delete("/{id}", h.realization.DeleteRealization)
		})
	})
}

func registerBusinessDomainRoutes(r chi.Router, h *routeHTTPHandlers, authMiddleware AuthMiddleware) {
	r.Route("/business-domains", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermDomainsRead))
			r.Get("/", h.businessDomain.GetAllBusinessDomains)
			r.Get("/{id}", h.businessDomain.GetBusinessDomainByID)
			r.Get("/{id}/capabilities", h.businessDomain.GetCapabilitiesInDomain)
			r.Get("/{id}/capability-realizations", h.businessDomain.GetCapabilityRealizationsByDomain)
			r.Get("/{id}/importance", h.strategyImportance.GetImportanceByDomain)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermDomainsWrite))
			r.Post("/", h.businessDomain.CreateBusinessDomain)
			r.Put("/{id}", h.businessDomain.UpdateBusinessDomain)
			r.Post("/{id}/capabilities", h.businessDomain.AssignCapabilityToDomain)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermDomainsDelete))
			r.Delete("/{id}", h.businessDomain.DeleteBusinessDomain)
			r.Delete("/{id}/capabilities/{capabilityId}", h.businessDomain.RemoveCapabilityFromDomain)
		})
		r.Route("/{id}/capabilities/{capabilityId}/importance", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequirePermission(authValueObjects.PermDomainsRead))
				r.Get("/", h.strategyImportance.GetImportanceByDomainAndCapability)
			})
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequirePermission(authValueObjects.PermDomainsWrite))
				r.Post("/", h.strategyImportance.SetImportance)
				r.Put("/{importanceId}", h.strategyImportance.UpdateImportance)
			})
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequirePermission(authValueObjects.PermDomainsDelete))
				r.Delete("/{importanceId}", h.strategyImportance.RemoveImportance)
			})
		})
	})
}

func registerStrategyImportanceRoutes(r chi.Router, h *routeHTTPHandlers) {
}

func registerApplicationFitScoreRoutes(r chi.Router, h *routeHTTPHandlers, authMiddleware AuthMiddleware, rateLimiter *platformAPI.RateLimiter) {
	r.Route("/components/{id}/fit-scores", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsRead))
			r.Get("/", h.applicationFitScore.GetFitScoresByComponent)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsWrite))
			r.Use(platformAPI.RateLimitMiddleware(rateLimiter))
			r.Put("/{pillarId}", h.applicationFitScore.SetFitScore)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsDelete))
			r.Use(platformAPI.RateLimitMiddleware(rateLimiter))
			r.Delete("/{pillarId}", h.applicationFitScore.RemoveFitScore)
		})
	})
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsRead))
		r.Get("/components/{id}/fit-comparisons", h.applicationFitScore.GetFitComparisons)
		r.Get("/strategy-pillars/{pillarId}/fit-scores", h.applicationFitScore.GetFitScoresByPillar)
	})
}

func registerStrategicFitAnalysisRoutes(r chi.Router, h *routeHTTPHandlers, authMiddleware AuthMiddleware) {
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequirePermission(authValueObjects.PermEnterpriseArchRead))
		r.Get("/strategic-fit-analysis/{pillarId}", h.strategicFitAnalysis.GetStrategicFitAnalysis)
	})
}
