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

func SetupCapabilityMappingRoutes(
	r chi.Router,
	commandBus *cqrs.InMemoryCommandBus,
	eventStore eventstore.EventStore,
	eventBus events.EventBus,
	db *database.TenantAwareDB,
	hateoas *sharedAPI.HATEOASLinks,
) error {
	capabilityRepo := repositories.NewCapabilityRepository(eventStore)
	dependencyRepo := repositories.NewDependencyRepository(eventStore)
	realizationRepo := repositories.NewRealizationRepository(eventStore)

	capabilityReadModel := readmodels.NewCapabilityReadModel(db)
	dependencyReadModel := readmodels.NewDependencyReadModel(db)
	realizationReadModel := readmodels.NewRealizationReadModel(db)
	componentReadModel := archReadModels.NewApplicationComponentReadModel(db)

	capabilityProjector := projectors.NewCapabilityProjector(capabilityReadModel)
	dependencyProjector := projectors.NewDependencyProjector(dependencyReadModel)
	realizationProjector := projectors.NewRealizationProjector(realizationReadModel)

	eventBus.Subscribe("CapabilityCreated", capabilityProjector)
	eventBus.Subscribe("CapabilityUpdated", capabilityProjector)
	eventBus.Subscribe("CapabilityMetadataUpdated", capabilityProjector)
	eventBus.Subscribe("CapabilityExpertAdded", capabilityProjector)
	eventBus.Subscribe("CapabilityTagAdded", capabilityProjector)

	eventBus.Subscribe("CapabilityDependencyCreated", dependencyProjector)
	eventBus.Subscribe("CapabilityDependencyDeleted", dependencyProjector)

	eventBus.Subscribe("SystemLinkedToCapability", realizationProjector)
	eventBus.Subscribe("SystemRealizationUpdated", realizationProjector)
	eventBus.Subscribe("SystemRealizationDeleted", realizationProjector)

	createCapabilityHandler := handlers.NewCreateCapabilityHandler(capabilityRepo)
	updateCapabilityHandler := handlers.NewUpdateCapabilityHandler(capabilityRepo)
	updateMetadataHandler := handlers.NewUpdateCapabilityMetadataHandler(capabilityRepo)
	addExpertHandler := handlers.NewAddCapabilityExpertHandler(capabilityRepo)
	addTagHandler := handlers.NewAddCapabilityTagHandler(capabilityRepo)

	createDependencyHandler := handlers.NewCreateCapabilityDependencyHandler(dependencyRepo, capabilityRepo)
	deleteDependencyHandler := handlers.NewDeleteCapabilityDependencyHandler(dependencyRepo)

	linkSystemHandler := handlers.NewLinkSystemToCapabilityHandler(realizationRepo, capabilityRepo, componentReadModel)
	updateRealizationHandler := handlers.NewUpdateSystemRealizationHandler(realizationRepo)
	deleteRealizationHandler := handlers.NewDeleteSystemRealizationHandler(realizationRepo)

	commandBus.Register("CreateCapability", createCapabilityHandler)
	commandBus.Register("UpdateCapability", updateCapabilityHandler)
	commandBus.Register("UpdateCapabilityMetadata", updateMetadataHandler)
	commandBus.Register("AddCapabilityExpert", addExpertHandler)
	commandBus.Register("AddCapabilityTag", addTagHandler)

	commandBus.Register("CreateCapabilityDependency", createDependencyHandler)
	commandBus.Register("DeleteCapabilityDependency", deleteDependencyHandler)

	commandBus.Register("LinkSystemToCapability", linkSystemHandler)
	commandBus.Register("UpdateSystemRealization", updateRealizationHandler)
	commandBus.Register("DeleteSystemRealization", deleteRealizationHandler)

	capabilityHandlers := NewCapabilityHandlers(commandBus, capabilityReadModel, hateoas)
	dependencyHandlers := NewDependencyHandlers(commandBus, dependencyReadModel, hateoas)
	realizationHandlers := NewRealizationHandlers(commandBus, realizationReadModel, hateoas)

	r.Route("/capabilities", func(r chi.Router) {
		r.Post("/", capabilityHandlers.CreateCapability)
		r.Get("/", capabilityHandlers.GetAllCapabilities)
		r.Get("/{id}", capabilityHandlers.GetCapabilityByID)
		r.Get("/{id}/children", capabilityHandlers.GetCapabilityChildren)
		r.Get("/{id}/systems", realizationHandlers.GetSystemsByCapability)
		r.Post("/{id}/systems", realizationHandlers.LinkSystemToCapability)
		r.Get("/{id}/dependencies/outgoing", dependencyHandlers.GetOutgoingDependencies)
		r.Get("/{id}/dependencies/incoming", dependencyHandlers.GetIncomingDependencies)
		r.Put("/{id}", capabilityHandlers.UpdateCapability)
		r.Put("/{id}/metadata", capabilityHandlers.UpdateCapabilityMetadata)
		r.Post("/{id}/experts", capabilityHandlers.AddCapabilityExpert)
		r.Post("/{id}/tags", capabilityHandlers.AddCapabilityTag)
	})

	r.Route("/capability-dependencies", func(r chi.Router) {
		r.Post("/", dependencyHandlers.CreateDependency)
		r.Get("/", dependencyHandlers.GetAllDependencies)
		r.Delete("/{id}", dependencyHandlers.DeleteDependency)
	})

	r.Route("/capability-realizations", func(r chi.Router) {
		r.Put("/{id}", realizationHandlers.UpdateRealization)
		r.Delete("/{id}", realizationHandlers.DeleteRealization)
		r.Get("/by-component/{componentId}", realizationHandlers.GetCapabilitiesByComponent)
	})

	return nil
}
