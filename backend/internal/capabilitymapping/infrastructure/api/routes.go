package api

import (
	archReadModels "easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/application/projectors"
	"easi/backend/internal/capabilitymapping/application/readmodels"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	"github.com/go-chi/chi/v5"
)

type routeConfig struct {
	commandBus *cqrs.InMemoryCommandBus
	eventStore eventstore.EventStore
	eventBus   events.EventBus
	db         *database.TenantAwareDB
	hateoas    *sharedAPI.HATEOASLinks
}

func SetupCapabilityMappingRoutes(
	r chi.Router,
	commandBus *cqrs.InMemoryCommandBus,
	eventStore eventstore.EventStore,
	eventBus events.EventBus,
	db *database.TenantAwareDB,
	hateoas *sharedAPI.HATEOASLinks,
) error {
	config := &routeConfig{commandBus, eventStore, eventBus, db, hateoas}

	repos := initializeRepositories(config.eventStore)
	readModels := initializeReadModels(config.db)

	setupEventSubscriptions(config.eventBus, readModels)
	setupCascadingDeleteHandlers(config.eventBus, config.commandBus, readModels)
	setupCommandHandlers(config.commandBus, repos, readModels)

	httpHandlers := initializeHTTPHandlers(config.commandBus, readModels, config.hateoas)
	registerRoutes(r, httpHandlers)

	return nil
}

type routeRepositories struct {
	capability       *repositories.CapabilityRepository
	dependency       *repositories.DependencyRepository
	realization      *repositories.RealizationRepository
	businessDomain   *repositories.BusinessDomainRepository
	domainAssignment *repositories.BusinessDomainAssignmentRepository
}

type routeReadModels struct {
	capability       *readmodels.CapabilityReadModel
	dependency       *readmodels.DependencyReadModel
	realization      *readmodels.RealizationReadModel
	component        *archReadModels.ApplicationComponentReadModel
	businessDomain   *readmodels.BusinessDomainReadModel
	domainAssignment *readmodels.DomainCapabilityAssignmentReadModel
}

type routeHTTPHandlers struct {
	capability     *CapabilityHandlers
	dependency     *DependencyHandlers
	realization    *RealizationHandlers
	maturityLevel  *MaturityLevelHandlers
	businessDomain *BusinessDomainHandlers
}

func initializeRepositories(eventStore eventstore.EventStore) *routeRepositories {
	return &routeRepositories{
		capability:       repositories.NewCapabilityRepository(eventStore),
		dependency:       repositories.NewDependencyRepository(eventStore),
		realization:      repositories.NewRealizationRepository(eventStore),
		businessDomain:   repositories.NewBusinessDomainRepository(eventStore),
		domainAssignment: repositories.NewBusinessDomainAssignmentRepository(eventStore),
	}
}

func initializeReadModels(db *database.TenantAwareDB) *routeReadModels {
	return &routeReadModels{
		capability:       readmodels.NewCapabilityReadModel(db),
		dependency:       readmodels.NewDependencyReadModel(db),
		realization:      readmodels.NewRealizationReadModel(db),
		component:        archReadModels.NewApplicationComponentReadModel(db),
		businessDomain:   readmodels.NewBusinessDomainReadModel(db),
		domainAssignment: readmodels.NewDomainCapabilityAssignmentReadModel(db),
	}
}

func setupEventSubscriptions(eventBus events.EventBus, rm *routeReadModels) {
	capabilityProjector := projectors.NewCapabilityProjector(rm.capability, rm.domainAssignment)
	dependencyProjector := projectors.NewDependencyProjector(rm.dependency)
	realizationProjector := projectors.NewRealizationProjector(rm.realization, rm.capability, rm.component)
	businessDomainProjector := projectors.NewBusinessDomainProjector(rm.businessDomain)
	domainAssignmentProjector := projectors.NewBusinessDomainAssignmentProjector(rm.domainAssignment, rm.businessDomain, rm.capability)

	subscribeCapabilityEvents(eventBus, capabilityProjector)
	subscribeDependencyEvents(eventBus, dependencyProjector)
	subscribeRealizationEvents(eventBus, realizationProjector)
	subscribeBusinessDomainEvents(eventBus, businessDomainProjector)
	subscribeDomainAssignmentEvents(eventBus, domainAssignmentProjector)
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

func setupCascadingDeleteHandlers(eventBus events.EventBus, commandBus *cqrs.InMemoryCommandBus, rm *routeReadModels) {
	onCapabilityDeletedHandler := handlers.NewOnCapabilityDeletedHandler(commandBus, rm.domainAssignment)
	onBusinessDomainDeletedHandler := handlers.NewOnBusinessDomainDeletedHandler(commandBus, rm.domainAssignment)
	onCapabilityParentChangedHandler := handlers.NewOnCapabilityParentChangedHandler(commandBus, rm.domainAssignment, rm.capability)

	eventBus.Subscribe("CapabilityDeleted", onCapabilityDeletedHandler)
	eventBus.Subscribe("BusinessDomainDeleted", onBusinessDomainDeletedHandler)
	eventBus.Subscribe("CapabilityParentChanged", onCapabilityParentChangedHandler)
}

func setupCommandHandlers(commandBus *cqrs.InMemoryCommandBus, repos *routeRepositories, rm *routeReadModels) {
	registerCapabilityCommands(commandBus, repos.capability, rm.capability)
	registerDependencyCommands(commandBus, repos.dependency, repos.capability)
	registerRealizationCommands(commandBus, repos.realization, repos.capability, rm.component)
	registerBusinessDomainCommands(commandBus, repos.businessDomain, rm.businessDomain, rm.domainAssignment)
	registerDomainAssignmentCommands(commandBus, repos.domainAssignment, rm.businessDomain, rm.capability, rm.domainAssignment)
}

func registerCapabilityCommands(commandBus *cqrs.InMemoryCommandBus, repo *repositories.CapabilityRepository, rm *readmodels.CapabilityReadModel) {
	commandBus.Register("CreateCapability", handlers.NewCreateCapabilityHandler(repo))
	commandBus.Register("UpdateCapability", handlers.NewUpdateCapabilityHandler(repo))
	commandBus.Register("UpdateCapabilityMetadata", handlers.NewUpdateCapabilityMetadataHandler(repo))
	commandBus.Register("AddCapabilityExpert", handlers.NewAddCapabilityExpertHandler(repo))
	commandBus.Register("AddCapabilityTag", handlers.NewAddCapabilityTagHandler(repo))
	commandBus.Register("ChangeCapabilityParent", handlers.NewChangeCapabilityParentHandler(repo, rm))
	commandBus.Register("DeleteCapability", handlers.NewDeleteCapabilityHandler(repo, rm))
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
	commandBus.Register("CreateBusinessDomain", handlers.NewCreateBusinessDomainHandler(domainRepo, domainRM))
	commandBus.Register("UpdateBusinessDomain", handlers.NewUpdateBusinessDomainHandler(domainRepo, domainRM))
	commandBus.Register("DeleteBusinessDomain", handlers.NewDeleteBusinessDomainHandler(domainRepo, assignmentRM))
}

func registerDomainAssignmentCommands(commandBus *cqrs.InMemoryCommandBus, assignRepo *repositories.BusinessDomainAssignmentRepository, domainRM *readmodels.BusinessDomainReadModel, capRM *readmodels.CapabilityReadModel, assignmentRM *readmodels.DomainCapabilityAssignmentReadModel) {
	commandBus.Register("AssignCapabilityToDomain", handlers.NewAssignCapabilityToDomainHandler(assignRepo, domainRM, capRM, assignmentRM))
	commandBus.Register("UnassignCapabilityFromDomain", handlers.NewUnassignCapabilityFromDomainHandler(assignRepo))
}

func initializeHTTPHandlers(commandBus *cqrs.InMemoryCommandBus, rm *routeReadModels, hateoas *sharedAPI.HATEOASLinks) *routeHTTPHandlers {
	businessDomainReadModels := &BusinessDomainReadModels{
		Domain:      rm.businessDomain,
		Assignment:  rm.domainAssignment,
		Capability:  rm.capability,
		Realization: rm.realization,
	}

	return &routeHTTPHandlers{
		capability:     NewCapabilityHandlers(commandBus, rm.capability, hateoas),
		dependency:     NewDependencyHandlers(commandBus, rm.dependency, hateoas),
		realization:    NewRealizationHandlers(commandBus, rm.realization, hateoas),
		maturityLevel:  NewMaturityLevelHandlers(),
		businessDomain: NewBusinessDomainHandlers(commandBus, businessDomainReadModels, hateoas),
	}
}

func registerRoutes(r chi.Router, h *routeHTTPHandlers) {
	registerCapabilityRoutes(r, h)
	registerDependencyRoutes(r, h)
	registerRealizationRoutes(r, h)
	registerBusinessDomainRoutes(r, h)
}

func registerCapabilityRoutes(r chi.Router, h *routeHTTPHandlers) {
	r.Route("/capabilities", func(r chi.Router) {
		r.Get("/metadata", h.maturityLevel.GetMetadataIndex)
		r.Get("/metadata/maturity-levels", h.maturityLevel.GetMaturityLevels)
		r.Get("/metadata/statuses", h.maturityLevel.GetStatuses)
		r.Get("/metadata/ownership-models", h.maturityLevel.GetOwnershipModels)
		r.Get("/metadata/strategy-pillars", h.maturityLevel.GetStrategyPillars)
		r.Post("/", h.capability.CreateCapability)
		r.Get("/", h.capability.GetAllCapabilities)
		r.Get("/{id}", h.capability.GetCapabilityByID)
		r.Get("/{id}/children", h.capability.GetCapabilityChildren)
		r.Get("/{id}/systems", h.realization.GetSystemsByCapability)
		r.Post("/{id}/systems", h.realization.LinkSystemToCapability)
		r.Get("/{id}/dependencies/outgoing", h.dependency.GetOutgoingDependencies)
		r.Get("/{id}/dependencies/incoming", h.dependency.GetIncomingDependencies)
		r.Get("/{id}/business-domains", h.businessDomain.GetDomainsForCapability)
		r.Put("/{id}", h.capability.UpdateCapability)
		r.Put("/{id}/metadata", h.capability.UpdateCapabilityMetadata)
		r.Patch("/{id}/parent", h.capability.ChangeCapabilityParent)
		r.Post("/{id}/experts", h.capability.AddCapabilityExpert)
		r.Post("/{id}/tags", h.capability.AddCapabilityTag)
		r.Delete("/{id}", h.capability.DeleteCapability)
	})
}

func registerDependencyRoutes(r chi.Router, h *routeHTTPHandlers) {
	r.Route("/capability-dependencies", func(r chi.Router) {
		r.Post("/", h.dependency.CreateDependency)
		r.Get("/", h.dependency.GetAllDependencies)
		r.Delete("/{id}", h.dependency.DeleteDependency)
	})
}

func registerRealizationRoutes(r chi.Router, h *routeHTTPHandlers) {
	r.Route("/capability-realizations", func(r chi.Router) {
		r.Put("/{id}", h.realization.UpdateRealization)
		r.Delete("/{id}", h.realization.DeleteRealization)
		r.Get("/by-component/{componentId}", h.realization.GetCapabilitiesByComponent)
	})
}

func registerBusinessDomainRoutes(r chi.Router, h *routeHTTPHandlers) {
	r.Route("/business-domains", func(r chi.Router) {
		r.Post("/", h.businessDomain.CreateBusinessDomain)
		r.Get("/", h.businessDomain.GetAllBusinessDomains)
		r.Get("/{id}", h.businessDomain.GetBusinessDomainByID)
		r.Put("/{id}", h.businessDomain.UpdateBusinessDomain)
		r.Delete("/{id}", h.businessDomain.DeleteBusinessDomain)
		r.Get("/{id}/capabilities", h.businessDomain.GetCapabilitiesInDomain)
		r.Post("/{id}/capabilities", h.businessDomain.AssignCapabilityToDomain)
		r.Get("/{id}/capability-realizations", h.businessDomain.GetCapabilityRealizationsByDomain)
		r.Delete("/{domainId}/capabilities/{capabilityId}", h.businessDomain.RemoveCapabilityFromDomain)
	})
}
