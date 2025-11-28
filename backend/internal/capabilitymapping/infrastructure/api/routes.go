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
	setupCommandHandlers(config.commandBus, repos, readModels)

	httpHandlers := initializeHTTPHandlers(config.commandBus, readModels, config.hateoas)
	registerRoutes(r, httpHandlers)

	return nil
}

type routeRepositories struct {
	capability  *repositories.CapabilityRepository
	dependency  *repositories.DependencyRepository
	realization *repositories.RealizationRepository
}

type routeReadModels struct {
	capability  *readmodels.CapabilityReadModel
	dependency  *readmodels.DependencyReadModel
	realization *readmodels.RealizationReadModel
	component   *archReadModels.ApplicationComponentReadModel
}

type routeHTTPHandlers struct {
	capability    *CapabilityHandlers
	dependency    *DependencyHandlers
	realization   *RealizationHandlers
	maturityLevel *MaturityLevelHandlers
}

func initializeRepositories(eventStore eventstore.EventStore) *routeRepositories {
	return &routeRepositories{
		capability:  repositories.NewCapabilityRepository(eventStore),
		dependency:  repositories.NewDependencyRepository(eventStore),
		realization: repositories.NewRealizationRepository(eventStore),
	}
}

func initializeReadModels(db *database.TenantAwareDB) *routeReadModels {
	return &routeReadModels{
		capability:  readmodels.NewCapabilityReadModel(db),
		dependency:  readmodels.NewDependencyReadModel(db),
		realization: readmodels.NewRealizationReadModel(db),
		component:   archReadModels.NewApplicationComponentReadModel(db),
	}
}

func setupEventSubscriptions(eventBus events.EventBus, rm *routeReadModels) {
	capabilityProjector := projectors.NewCapabilityProjector(rm.capability)
	dependencyProjector := projectors.NewDependencyProjector(rm.dependency)
	realizationProjector := projectors.NewRealizationProjector(rm.realization, rm.capability)

	subscribeCapabilityEvents(eventBus, capabilityProjector)
	subscribeDependencyEvents(eventBus, dependencyProjector)
	subscribeRealizationEvents(eventBus, realizationProjector)
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
		"SystemRealizationDeleted", "CapabilityParentChanged"}
	for _, event := range events {
		eventBus.Subscribe(event, projector)
	}
}

func setupCommandHandlers(commandBus *cqrs.InMemoryCommandBus, repos *routeRepositories, rm *routeReadModels) {
	registerCapabilityCommands(commandBus, repos.capability, rm.capability)
	registerDependencyCommands(commandBus, repos.dependency, repos.capability)
	registerRealizationCommands(commandBus, repos.realization, repos.capability, rm.component)
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

func initializeHTTPHandlers(commandBus *cqrs.InMemoryCommandBus, rm *routeReadModels, hateoas *sharedAPI.HATEOASLinks) *routeHTTPHandlers {
	return &routeHTTPHandlers{
		capability:    NewCapabilityHandlers(commandBus, rm.capability, hateoas),
		dependency:    NewDependencyHandlers(commandBus, rm.dependency, hateoas),
		realization:   NewRealizationHandlers(commandBus, rm.realization, hateoas),
		maturityLevel: NewMaturityLevelHandlers(),
	}
}

func registerRoutes(r chi.Router, h *routeHTTPHandlers) {
	registerCapabilityRoutes(r, h)
	registerDependencyRoutes(r, h)
	registerRealizationRoutes(r, h)
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
