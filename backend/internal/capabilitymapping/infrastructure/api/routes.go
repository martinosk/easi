package api

import (
	"net/http"

	authPL "easi/backend/internal/auth/publishedlanguage"
	"easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/application/projectors"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/infrastructure/adapters"
	"easi/backend/internal/capabilitymapping/infrastructure/metamodel"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	archPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	platformAPI "easi/backend/internal/platform/infrastructure/api"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequirePermission(permission authPL.Permission) func(http.Handler) http.Handler
}

type RouteConfig struct {
	Router                 chi.Router
	CommandBus             *cqrs.InMemoryCommandBus
	EventStore             eventstore.EventStore
	EventBus               events.EventBus
	DB                     *database.TenantAwareDB
	HATEOAS                *sharedAPI.HATEOASLinks
	MaturityScaleGateway   metamodel.MaturityScaleGateway
	StrategyPillarsGateway metamodel.StrategyPillarsGateway
	SessionProvider        authPL.SessionProvider
	AuthMiddleware         AuthMiddleware
}

func SetupCapabilityMappingRoutes(config *RouteConfig) error {
	repos := initializeRepositories(config.EventStore)
	rm := initializeReadModels(config.DB)

	if config.StrategyPillarsGateway == nil {
		config.StrategyPillarsGateway = metamodel.NewLocalStrategyPillarsGateway(rm.strategyPillarCache)
	}

	setupEventSubscriptions(config.EventBus, rm, config.StrategyPillarsGateway)
	setupCascadingDeleteHandlers(config.EventBus, config.CommandBus, rm)
	setupCommandHandlers(config.CommandBus, repos, rm, config.StrategyPillarsGateway)
	setupMetaModelEventHandlers(config.EventBus, config.MaturityScaleGateway)

	businessDomainReadModels := &BusinessDomainReadModels{
		Domain:      rm.businessDomain,
		Assignment:  rm.domainAssignment,
		Capability:  rm.capability,
		Realization: rm.realization,
	}

	links := NewCapabilityMappingLinks(config.HATEOAS)
	httpHandlers := &routeHTTPHandlers{
		capability:           NewCapabilityHandlers(config.CommandBus, rm.capability, links),
		dependency:           NewDependencyHandlers(config.CommandBus, rm.dependency, links),
		realization:          NewRealizationHandlers(config.CommandBus, rm.realization, links),
		maturityLevel:        NewMaturityLevelHandlers(config.MaturityScaleGateway),
		businessDomain:       NewBusinessDomainHandlers(config.CommandBus, businessDomainReadModels, links),
		strategyImportance:   NewStrategyImportanceHandlers(config.CommandBus, rm.strategyImportance, links),
		applicationFitScore:  NewApplicationFitScoreHandlers(config.CommandBus, rm.applicationFitScore, links, config.SessionProvider),
		fitComparison:        NewFitComparisonHandlers(rm.componentFitComparison),
		strategicFitAnalysis: NewStrategicFitAnalysisHandlers(rm.strategicFitAnalysis, config.StrategyPillarsGateway, config.SessionProvider),
	}

	rateLimiter := platformAPI.NewRateLimiter(100, 60)

	registerCapabilityRoutes(config.Router, httpHandlers, config.AuthMiddleware)
	registerDependencyRoutes(config.Router, httpHandlers, config.AuthMiddleware)
	registerRealizationRoutes(config.Router, httpHandlers, config.AuthMiddleware)
	registerBusinessDomainRoutes(config.Router, httpHandlers, config.AuthMiddleware)
	registerStrategyImportanceRoutes(config.Router, httpHandlers)
	registerApplicationFitScoreRoutes(config.Router, httpHandlers, config.AuthMiddleware, rateLimiter)
	registerStrategicFitAnalysisRoutes(config.Router, httpHandlers, config.AuthMiddleware)

	return nil
}

func setupMetaModelEventHandlers(eventBus events.EventBus, gateway metamodel.MaturityScaleGateway) {
	if gateway == nil {
		return
	}

	maturityScaleUpdatedHandler := handlers.NewMaturityScaleConfigUpdatedHandler(gateway)
	eventBus.Subscribe(mmPL.MaturityScaleConfigUpdated, maturityScaleUpdatedHandler)
	eventBus.Subscribe(mmPL.MaturityScaleConfigReset, maturityScaleUpdatedHandler)
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
	capability                    *readmodels.CapabilityReadModel
	dependency                    *readmodels.DependencyReadModel
	realization                   *readmodels.RealizationReadModel
	businessDomain                *readmodels.BusinessDomainReadModel
	domainAssignment              *readmodels.DomainCapabilityAssignmentReadModel
	strategyImportance            *readmodels.StrategyImportanceReadModel
	applicationFitScore           *readmodels.ApplicationFitScoreReadModel
	strategicFitAnalysis          *readmodels.StrategicFitAnalysisReadModel
	componentFitComparison        *readmodels.ComponentFitComparisonReadModel
	componentCache                *readmodels.ComponentCacheReadModel
	effectiveCapabilityImportance *readmodels.EffectiveCapabilityImportanceReadModel
	strategyPillarCache           *readmodels.StrategyPillarCacheReadModel
	effectiveBusinessDomain       *readmodels.CMEffectiveBusinessDomainReadModel
}

type routeHTTPHandlers struct {
	capability           *CapabilityHandlers
	dependency           *DependencyHandlers
	realization          *RealizationHandlers
	maturityLevel        *MaturityLevelHandlers
	businessDomain       *BusinessDomainHandlers
	strategyImportance   *StrategyImportanceHandlers
	applicationFitScore  *ApplicationFitScoreHandlers
	fitComparison        *FitComparisonHandlers
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
		capability:                    readmodels.NewCapabilityReadModel(db),
		dependency:                    readmodels.NewDependencyReadModel(db),
		realization:                   readmodels.NewRealizationReadModel(db),
		businessDomain:                readmodels.NewBusinessDomainReadModel(db),
		domainAssignment:              readmodels.NewDomainCapabilityAssignmentReadModel(db),
		strategyImportance:            readmodels.NewStrategyImportanceReadModel(db),
		applicationFitScore:           readmodels.NewApplicationFitScoreReadModel(db),
		strategicFitAnalysis:          readmodels.NewStrategicFitAnalysisReadModel(db),
		componentFitComparison:        readmodels.NewComponentFitComparisonReadModel(db),
		componentCache:                readmodels.NewComponentCacheReadModel(db),
		effectiveCapabilityImportance: readmodels.NewEffectiveCapabilityImportanceReadModel(db),
		strategyPillarCache:           readmodels.NewStrategyPillarCacheReadModel(db),
		effectiveBusinessDomain:       readmodels.NewCMEffectiveBusinessDomainReadModel(db),
	}
}

func setupEventSubscriptions(eventBus events.EventBus, rm *routeReadModels, pillarsGateway metamodel.StrategyPillarsGateway) {
	capabilityProjector := projectors.NewCapabilityProjector(rm.capability, rm.domainAssignment)
	dependencyProjector := projectors.NewDependencyProjector(rm.dependency)
	realizationProjector := projectors.NewRealizationProjector(rm.realization, rm.componentCache)
	businessDomainProjector := projectors.NewBusinessDomainProjector(rm.businessDomain)
	domainAssignmentProjector := projectors.NewBusinessDomainAssignmentProjector(rm.domainAssignment, rm.businessDomain, rm.capability)
	strategyImportanceProjector := projectors.NewStrategyImportanceProjector(rm.strategyImportance, rm.businessDomain, rm.capability, pillarsGateway)
	applicationFitScoreProjector := projectors.NewApplicationFitScoreProjector(rm.applicationFitScore, rm.componentCache, pillarsGateway)
	componentCacheProjector := projectors.NewComponentCacheProjector(rm.componentCache)
	pillarCacheProjector := projectors.NewStrategyPillarCacheProjector(rm.strategyPillarCache)

	capabilityLookupAdapter := adapters.NewCapabilityLookupAdapter(rm.capability)
	ratingLookupAdapter := adapters.NewRatingLookupAdapter(rm.strategyImportance)
	hierarchyService := services.NewCapabilityHierarchyService(capabilityLookupAdapter)
	ratingResolver := services.NewHierarchicalRatingResolver(hierarchyService, ratingLookupAdapter, capabilityLookupAdapter)
	recomputer := projectors.NewEffectiveImportanceRecomputer(rm.effectiveCapabilityImportance, ratingResolver, hierarchyService, eventBus)
	importanceChangeProjector := projectors.NewImportanceChangeEffectiveProjector(recomputer, rm.strategyImportance)
	hierarchyChangeProjector := projectors.NewHierarchyChangeEffectiveProjector(recomputer, rm.effectiveCapabilityImportance)
	ancestryChecker := projectors.NewDomainAncestryChecker(hierarchyService, rm.domainAssignment)
	domainAssignmentEffectiveProjector := projectors.NewDomainAssignmentEffectiveProjector(recomputer, ancestryChecker, pillarsGateway)

	effectiveBDProjector := projectors.NewEffectiveBusinessDomainProjector(rm.effectiveBusinessDomain, rm.businessDomain, rm.capability)

	subscribeCapabilityEvents(eventBus, capabilityProjector)
	subscribeEffectiveBusinessDomainEvents(eventBus, effectiveBDProjector)
	subscribeDependencyEvents(eventBus, dependencyProjector)
	subscribeRealizationEvents(eventBus, realizationProjector)
	subscribeBusinessDomainEvents(eventBus, businessDomainProjector)
	subscribeDomainAssignmentEvents(eventBus, domainAssignmentProjector)
	subscribeStrategyImportanceEvents(eventBus, strategyImportanceProjector)
	subscribeApplicationFitScoreEvents(eventBus, applicationFitScoreProjector)
	subscribeComponentCacheEvents(eventBus, componentCacheProjector)
	subscribeImportanceChangeEffectiveEvents(eventBus, importanceChangeProjector)
	subscribeHierarchyChangeEffectiveEvents(eventBus, hierarchyChangeProjector)
	subscribeDomainAssignmentEffectiveEvents(eventBus, domainAssignmentEffectiveProjector)
	subscribeMetaModelEvents(eventBus, pillarCacheProjector)
}

func subscribeCapabilityEvents(eventBus events.EventBus, projector *projectors.CapabilityProjector) {
	events := []string{"CapabilityCreated", "CapabilityUpdated", "CapabilityMetadataUpdated",
		"CapabilityExpertAdded", "CapabilityExpertRemoved", "CapabilityTagAdded", "CapabilityParentChanged", "CapabilityLevelChanged", "CapabilityDeleted"}
	for _, event := range events {
		eventBus.Subscribe(event, projector)
	}
}

func subscribeEffectiveBusinessDomainEvents(eventBus events.EventBus, projector *projectors.EffectiveBusinessDomainProjector) {
	for _, event := range []string{
		"CapabilityCreated",
		"CapabilityDeleted",
		"CapabilityParentChanged",
		"CapabilityLevelChanged",
		"CapabilityAssignedToDomain",
		"CapabilityUnassignedFromDomain",
	} {
		eventBus.Subscribe(event, projector)
	}
}

func subscribeDependencyEvents(eventBus events.EventBus, projector *projectors.DependencyProjector) {
	eventBus.Subscribe("CapabilityDependencyCreated", projector)
	eventBus.Subscribe("CapabilityDependencyDeleted", projector)
}

func subscribeRealizationEvents(eventBus events.EventBus, projector *projectors.RealizationProjector) {
	for _, event := range []string{
		"SystemLinkedToCapability",
		"SystemRealizationUpdated",
		"SystemRealizationDeleted",
		"CapabilityRealizationsInherited",
		"CapabilityRealizationsUninherited",
		"CapabilityUpdated",
		archPL.ApplicationComponentUpdated,
		archPL.ApplicationComponentDeleted,
	} {
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

func subscribeComponentCacheEvents(eventBus events.EventBus, projector *projectors.ComponentCacheProjector) {
	for _, event := range []string{
		archPL.ApplicationComponentCreated,
		archPL.ApplicationComponentUpdated,
		archPL.ApplicationComponentDeleted,
	} {
		eventBus.Subscribe(event, projector)
	}
}

func subscribeImportanceChangeEffectiveEvents(eventBus events.EventBus, projector *projectors.ImportanceChangeEffectiveProjector) {
	for _, event := range []string{"StrategyImportanceSet", "StrategyImportanceUpdated", "StrategyImportanceRemoved"} {
		eventBus.Subscribe(event, projector)
	}
}

func subscribeHierarchyChangeEffectiveEvents(eventBus events.EventBus, projector *projectors.HierarchyChangeEffectiveProjector) {
	for _, event := range []string{"CapabilityParentChanged", "CapabilityDeleted"} {
		eventBus.Subscribe(event, projector)
	}
}

func subscribeDomainAssignmentEffectiveEvents(eventBus events.EventBus, projector *projectors.DomainAssignmentEffectiveProjector) {
	for _, event := range []string{"CapabilityAssignedToDomain", "CapabilityUnassignedFromDomain"} {
		eventBus.Subscribe(event, projector)
	}
}

func subscribeMetaModelEvents(eventBus events.EventBus, projector *projectors.StrategyPillarCacheProjector) {
	for _, event := range []string{
		mmPL.MetaModelConfigurationCreated,
		mmPL.StrategyPillarAdded,
		mmPL.StrategyPillarUpdated,
		mmPL.StrategyPillarRemoved,
		mmPL.PillarFitConfigurationUpdated,
	} {
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
	registerCapabilityCommands(commandBus, repos.capability, rm.capability, rm.realization)
	registerDependencyCommands(commandBus, repos.dependency, repos.capability)
	registerRealizationCommands(commandBus, repos, rm)
	registerBusinessDomainCommands(commandBus, repos.businessDomain, rm.businessDomain, rm.domainAssignment)
	commandBus.Register("AssignCapabilityToDomain", handlers.NewAssignCapabilityToDomainHandler(repos.domainAssignment, repos.capability, rm.businessDomain, rm.domainAssignment))
	commandBus.Register("UnassignCapabilityFromDomain", handlers.NewUnassignCapabilityFromDomainHandler(repos.domainAssignment))
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

func registerCapabilityCommands(commandBus *cqrs.InMemoryCommandBus, repo *repositories.CapabilityRepository, capabilityRM *readmodels.CapabilityReadModel, realizationRM *readmodels.RealizationReadModel) {
	childrenChecker := adapters.NewCapabilityChildrenCheckerAdapter(capabilityRM)
	deletionService := services.NewCapabilityDeletionService(childrenChecker)
	reparentingService := services.NewCapabilityReparentingService(adapters.NewCapabilityLookupAdapter(capabilityRM))

	commandBus.Register("CreateCapability", handlers.NewCreateCapabilityHandler(repo))
	commandBus.Register("UpdateCapability", handlers.NewUpdateCapabilityHandler(repo))
	commandBus.Register("UpdateCapabilityMetadata", handlers.NewUpdateCapabilityMetadataHandler(repo))
	commandBus.Register("AddCapabilityExpert", handlers.NewAddCapabilityExpertHandler(repo))
	commandBus.Register("RemoveCapabilityExpert", handlers.NewRemoveCapabilityExpertHandler(repo))
	commandBus.Register("AddCapabilityTag", handlers.NewAddCapabilityTagHandler(repo))
	commandBus.Register("ChangeCapabilityParent", handlers.NewChangeCapabilityParentHandler(repo, capabilityRM, realizationRM, reparentingService))
	commandBus.Register("DeleteCapability", handlers.NewDeleteCapabilityHandler(repo, deletionService))
}

func registerDependencyCommands(commandBus *cqrs.InMemoryCommandBus, depRepo *repositories.DependencyRepository, capRepo *repositories.CapabilityRepository) {
	commandBus.Register("CreateCapabilityDependency", handlers.NewCreateCapabilityDependencyHandler(depRepo, capRepo))
	commandBus.Register("DeleteCapabilityDependency", handlers.NewDeleteCapabilityDependencyHandler(depRepo))
}

func registerRealizationCommands(commandBus *cqrs.InMemoryCommandBus, repos *routeRepositories, rm *routeReadModels) {
	commandBus.Register("LinkSystemToCapability", handlers.NewLinkSystemToCapabilityHandler(repos.realization, repos.capability, rm.capability, rm.componentCache))
	commandBus.Register("UpdateSystemRealization", handlers.NewUpdateSystemRealizationHandler(repos.realization))
	commandBus.Register("DeleteSystemRealization", handlers.NewDeleteSystemRealizationHandler(repos.realization))
}

func registerBusinessDomainCommands(commandBus *cqrs.InMemoryCommandBus, domainRepo *repositories.BusinessDomainRepository, domainRM *readmodels.BusinessDomainReadModel, assignmentRM *readmodels.DomainCapabilityAssignmentReadModel) {
	assignmentChecker := adapters.NewBusinessDomainAssignmentCheckerAdapter(assignmentRM)
	deletionService := services.NewBusinessDomainDeletionService(assignmentChecker)

	commandBus.Register("CreateBusinessDomain", handlers.NewCreateBusinessDomainHandler(domainRepo, domainRM))
	commandBus.Register("UpdateBusinessDomain", handlers.NewUpdateBusinessDomainHandler(domainRepo, domainRM))
	commandBus.Register("DeleteBusinessDomain", handlers.NewDeleteBusinessDomainHandler(domainRepo, deletionService))
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

func registerCapabilityRoutes(r chi.Router, h *routeHTTPHandlers, authMiddleware AuthMiddleware) {
	r.Route("/capabilities", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermCapabilitiesRead))
			r.Get("/metadata", h.maturityLevel.GetMetadataIndex)
			r.Get("/metadata/maturity-levels", h.maturityLevel.GetMaturityLevels)
			r.Get("/metadata/statuses", h.maturityLevel.GetStatuses)
			r.Get("/metadata/ownership-models", h.maturityLevel.GetOwnershipModels)
			r.Get("/expert-roles", h.capability.GetExpertRoles)
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
			r.Use(authMiddleware.RequirePermission(authPL.PermCapabilitiesWrite))
			r.Post("/", h.capability.CreateCapability)
			r.Post("/{id}/systems", h.realization.LinkSystemToCapability)
			r.Post("/{id}/experts", h.capability.AddCapabilityExpert)
			r.Delete("/{id}/experts", h.capability.RemoveCapabilityExpert)
			r.Post("/{id}/tags", h.capability.AddCapabilityTag)
		})
		r.Group(func(r chi.Router) {
			r.Use(sharedAPI.RequireWriteOrEditGrant("capabilities", "id"))
			r.Put("/{id}", h.capability.UpdateCapability)
			r.Put("/{id}/metadata", h.capability.UpdateCapabilityMetadata)
			r.Patch("/{id}/parent", h.capability.ChangeCapabilityParent)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermCapabilitiesDelete))
			r.Delete("/{id}", h.capability.DeleteCapability)
		})
	})
}

func registerDependencyRoutes(r chi.Router, h *routeHTTPHandlers, authMiddleware AuthMiddleware) {
	r.Route("/capability-dependencies", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermCapabilitiesRead))
			r.Get("/", h.dependency.GetAllDependencies)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermCapabilitiesWrite))
			r.Post("/", h.dependency.CreateDependency)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermCapabilitiesDelete))
			r.Delete("/{id}", h.dependency.DeleteDependency)
		})
	})
}

func registerRealizationRoutes(r chi.Router, h *routeHTTPHandlers, authMiddleware AuthMiddleware) {
	r.Route("/capability-realizations", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermCapabilitiesRead))
			r.Get("/by-component/{componentId}", h.realization.GetCapabilitiesByComponent)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermCapabilitiesWrite))
			r.Put("/{id}", h.realization.UpdateRealization)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermCapabilitiesDelete))
			r.Delete("/{id}", h.realization.DeleteRealization)
		})
	})
}

func registerBusinessDomainRoutes(r chi.Router, h *routeHTTPHandlers, authMiddleware AuthMiddleware) {
	r.Route("/business-domains", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermDomainsRead))
			r.Get("/", h.businessDomain.GetAllBusinessDomains)
			r.Get("/{id}", h.businessDomain.GetBusinessDomainByID)
			r.Get("/{id}/capabilities", h.businessDomain.GetCapabilitiesInDomain)
			r.Get("/{id}/capability-realizations", h.businessDomain.GetCapabilityRealizationsByDomain)
			r.Get("/{id}/importance", h.strategyImportance.GetImportanceByDomain)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermDomainsWrite))
			r.Post("/", h.businessDomain.CreateBusinessDomain)
			r.Post("/{id}/capabilities", h.businessDomain.AssignCapabilityToDomain)
		})
		r.Group(func(r chi.Router) {
			r.Use(sharedAPI.RequireWriteOrEditGrant("domains", "id"))
			r.Put("/{id}", h.businessDomain.UpdateBusinessDomain)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermDomainsDelete))
			r.Delete("/{id}", h.businessDomain.DeleteBusinessDomain)
			r.Delete("/{id}/capabilities/{capabilityId}", h.businessDomain.RemoveCapabilityFromDomain)
		})
		r.Route("/{id}/capabilities/{capabilityId}/importance", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequirePermission(authPL.PermDomainsRead))
				r.Get("/", h.strategyImportance.GetImportanceByDomainAndCapability)
			})
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequirePermission(authPL.PermDomainsWrite))
				r.Post("/", h.strategyImportance.SetImportance)
				r.Put("/{importanceId}", h.strategyImportance.UpdateImportance)
			})
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.RequirePermission(authPL.PermDomainsDelete))
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
			r.Use(authMiddleware.RequirePermission(authPL.PermComponentsRead))
			r.Get("/", h.applicationFitScore.GetFitScoresByComponent)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermComponentsWrite))
			r.Use(platformAPI.RateLimitMiddleware(rateLimiter))
			r.Put("/{pillarId}", h.applicationFitScore.SetFitScore)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermComponentsDelete))
			r.Use(platformAPI.RateLimitMiddleware(rateLimiter))
			r.Delete("/{pillarId}", h.applicationFitScore.RemoveFitScore)
		})
	})
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequirePermission(authPL.PermComponentsRead))
		r.Get("/components/{id}/fit-comparisons", h.fitComparison.GetFitComparisons)
		r.Get("/strategy-pillars/{pillarId}/fit-scores", h.applicationFitScore.GetFitScoresByPillar)
	})
}

func registerStrategicFitAnalysisRoutes(r chi.Router, h *routeHTTPHandlers, authMiddleware AuthMiddleware) {
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequirePermission(authPL.PermEnterpriseArchRead))
		r.Get("/strategic-fit-analysis/{pillarId}", h.strategicFitAnalysis.GetStrategicFitAnalysis)
	})
}
