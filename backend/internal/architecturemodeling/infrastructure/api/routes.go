package api

import (
	"net/http"

	"easi/backend/internal/architecturemodeling/application/handlers"
	"easi/backend/internal/architecturemodeling/application/projectors"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	authValueObjects "easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequirePermission(permission authValueObjects.Permission) func(http.Handler) http.Handler
}

// SetupArchitectureModelingRoutes initializes and registers architecture modeling routes
func SetupArchitectureModelingRoutes(
	r chi.Router,
	commandBus *cqrs.InMemoryCommandBus,
	eventStore eventstore.EventStore,
	eventBus events.EventBus,
	db *database.TenantAwareDB,
	hateoas *sharedAPI.HATEOASLinks,
	authMiddleware AuthMiddleware,
) error {
	// Initialize repositories
	componentRepo := repositories.NewApplicationComponentRepository(eventStore)
	relationRepo := repositories.NewComponentRelationRepository(eventStore)
	acquiredEntityRepo := repositories.NewAcquiredEntityRepository(eventStore)
	vendorRepo := repositories.NewVendorRepository(eventStore)
	internalTeamRepo := repositories.NewInternalTeamRepository(eventStore)
	acquiredViaRepo := repositories.NewAcquiredViaRelationshipRepository(eventStore)
	purchasedFromRepo := repositories.NewPurchasedFromRelationshipRepository(eventStore)
	builtByRepo := repositories.NewBuiltByRelationshipRepository(eventStore)

	// Initialize read models
	componentReadModel := readmodels.NewApplicationComponentReadModel(db)
	relationReadModel := readmodels.NewComponentRelationReadModel(db)
	acquiredEntityReadModel := readmodels.NewAcquiredEntityReadModel(db)
	vendorReadModel := readmodels.NewVendorReadModel(db)
	internalTeamReadModel := readmodels.NewInternalTeamReadModel(db)
	acquiredViaReadModel := readmodels.NewAcquiredViaRelationshipReadModel(db)
	purchasedFromReadModel := readmodels.NewPurchasedFromRelationshipReadModel(db)
	builtByReadModel := readmodels.NewBuiltByRelationshipReadModel(db)

	// Initialize projectors
	componentProjector := projectors.NewApplicationComponentProjector(componentReadModel)
	relationProjector := projectors.NewComponentRelationProjector(relationReadModel)
	acquiredEntityProjector := projectors.NewAcquiredEntityProjector(acquiredEntityReadModel)
	vendorProjector := projectors.NewVendorProjector(vendorReadModel)
	internalTeamProjector := projectors.NewInternalTeamProjector(internalTeamReadModel)
	originRelationshipProjector := projectors.NewOriginRelationshipProjector(acquiredViaReadModel, purchasedFromReadModel, builtByReadModel)

	// Wire up projectors to event bus
	eventBus.Subscribe("ApplicationComponentCreated", componentProjector)
	eventBus.Subscribe("ApplicationComponentUpdated", componentProjector)
	eventBus.Subscribe("ApplicationComponentDeleted", componentProjector)
	eventBus.Subscribe("ApplicationComponentExpertAdded", componentProjector)
	eventBus.Subscribe("ApplicationComponentExpertRemoved", componentProjector)
	eventBus.Subscribe("ComponentRelationCreated", relationProjector)
	eventBus.Subscribe("ComponentRelationUpdated", relationProjector)
	eventBus.Subscribe("ComponentRelationDeleted", relationProjector)

	// Subscribe origin entity projectors
	eventBus.Subscribe("AcquiredEntityCreated", acquiredEntityProjector)
	eventBus.Subscribe("AcquiredEntityUpdated", acquiredEntityProjector)
	eventBus.Subscribe("AcquiredEntityDeleted", acquiredEntityProjector)
	eventBus.Subscribe("VendorCreated", vendorProjector)
	eventBus.Subscribe("VendorUpdated", vendorProjector)
	eventBus.Subscribe("VendorDeleted", vendorProjector)
	eventBus.Subscribe("InternalTeamCreated", internalTeamProjector)
	eventBus.Subscribe("InternalTeamUpdated", internalTeamProjector)
	eventBus.Subscribe("InternalTeamDeleted", internalTeamProjector)

	// Subscribe origin relationship projectors
	eventBus.Subscribe("AcquiredViaRelationshipCreated", originRelationshipProjector)
	eventBus.Subscribe("AcquiredViaRelationshipDeleted", originRelationshipProjector)
	eventBus.Subscribe("PurchasedFromRelationshipCreated", originRelationshipProjector)
	eventBus.Subscribe("PurchasedFromRelationshipDeleted", originRelationshipProjector)
	eventBus.Subscribe("BuiltByRelationshipCreated", originRelationshipProjector)
	eventBus.Subscribe("BuiltByRelationshipDeleted", originRelationshipProjector)

	// Initialize command handlers
	createComponentHandler := handlers.NewCreateApplicationComponentHandler(componentRepo)
	updateComponentHandler := handlers.NewUpdateApplicationComponentHandler(componentRepo)
	deleteComponentHandler := handlers.NewDeleteApplicationComponentHandler(componentRepo, relationReadModel, commandBus)
	addExpertHandler := handlers.NewAddApplicationComponentExpertHandler(componentRepo)
	removeExpertHandler := handlers.NewRemoveApplicationComponentExpertHandler(componentRepo)
	createRelationHandler := handlers.NewCreateComponentRelationHandler(relationRepo)
	updateRelationHandler := handlers.NewUpdateComponentRelationHandler(relationRepo)
	deleteRelationHandler := handlers.NewDeleteComponentRelationHandler(relationRepo)

	// Origin entity command handlers
	createAcquiredEntityHandler := handlers.NewCreateAcquiredEntityHandler(acquiredEntityRepo)
	updateAcquiredEntityHandler := handlers.NewUpdateAcquiredEntityHandler(acquiredEntityRepo)
	deleteAcquiredEntityHandler := handlers.NewDeleteAcquiredEntityHandler(acquiredEntityRepo, acquiredViaReadModel, commandBus)
	createVendorHandler := handlers.NewCreateVendorHandler(vendorRepo)
	updateVendorHandler := handlers.NewUpdateVendorHandler(vendorRepo)
	deleteVendorHandler := handlers.NewDeleteVendorHandler(vendorRepo, purchasedFromReadModel, commandBus)
	createInternalTeamHandler := handlers.NewCreateInternalTeamHandler(internalTeamRepo)
	updateInternalTeamHandler := handlers.NewUpdateInternalTeamHandler(internalTeamRepo)
	deleteInternalTeamHandler := handlers.NewDeleteInternalTeamHandler(internalTeamRepo, builtByReadModel, commandBus)

	// Origin relationship command handlers
	createAcquiredViaHandler := handlers.NewCreateAcquiredViaRelationshipHandler(acquiredViaRepo)
	deleteAcquiredViaHandler := handlers.NewDeleteAcquiredViaRelationshipHandler(acquiredViaRepo)
	createPurchasedFromHandler := handlers.NewCreatePurchasedFromRelationshipHandler(purchasedFromRepo)
	deletePurchasedFromHandler := handlers.NewDeletePurchasedFromRelationshipHandler(purchasedFromRepo)
	createBuiltByHandler := handlers.NewCreateBuiltByRelationshipHandler(builtByRepo)
	deleteBuiltByHandler := handlers.NewDeleteBuiltByRelationshipHandler(builtByRepo)

	// Register command handlers
	commandBus.Register("CreateApplicationComponent", createComponentHandler)
	commandBus.Register("UpdateApplicationComponent", updateComponentHandler)
	commandBus.Register("DeleteApplicationComponent", deleteComponentHandler)
	commandBus.Register("AddApplicationComponentExpert", addExpertHandler)
	commandBus.Register("RemoveApplicationComponentExpert", removeExpertHandler)
	commandBus.Register("CreateComponentRelation", createRelationHandler)
	commandBus.Register("UpdateComponentRelation", updateRelationHandler)
	commandBus.Register("DeleteComponentRelation", deleteRelationHandler)

	// Register origin entity command handlers
	commandBus.Register("CreateAcquiredEntity", createAcquiredEntityHandler)
	commandBus.Register("UpdateAcquiredEntity", updateAcquiredEntityHandler)
	commandBus.Register("DeleteAcquiredEntity", deleteAcquiredEntityHandler)
	commandBus.Register("CreateVendor", createVendorHandler)
	commandBus.Register("UpdateVendor", updateVendorHandler)
	commandBus.Register("DeleteVendor", deleteVendorHandler)
	commandBus.Register("CreateInternalTeam", createInternalTeamHandler)
	commandBus.Register("UpdateInternalTeam", updateInternalTeamHandler)
	commandBus.Register("DeleteInternalTeam", deleteInternalTeamHandler)

	// Register origin relationship command handlers
	commandBus.Register("CreateAcquiredViaRelationship", createAcquiredViaHandler)
	commandBus.Register("DeleteAcquiredViaRelationship", deleteAcquiredViaHandler)
	commandBus.Register("CreatePurchasedFromRelationship", createPurchasedFromHandler)
	commandBus.Register("DeletePurchasedFromRelationship", deletePurchasedFromHandler)
	commandBus.Register("CreateBuiltByRelationship", createBuiltByHandler)
	commandBus.Register("DeleteBuiltByRelationship", deleteBuiltByHandler)

	// Initialize HTTP handlers
	componentHandlers := NewComponentHandlers(commandBus, componentReadModel, hateoas)
	expertHandlers := NewComponentExpertHandlers(commandBus, componentReadModel)
	relationHandlers := NewRelationHandlers(commandBus, relationReadModel, hateoas)
	acquiredEntityHandlers := NewAcquiredEntityHandlers(commandBus, acquiredEntityReadModel)
	vendorHandlers := NewVendorHandlers(commandBus, vendorReadModel)
	internalTeamHandlers := NewInternalTeamHandlers(commandBus, internalTeamReadModel)
	originRelationshipHandlers := NewOriginRelationshipHandlers(commandBus, acquiredViaReadModel, purchasedFromReadModel, builtByReadModel)

	// Register component routes
	r.Route("/components", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsRead))
			r.Get("/", componentHandlers.GetAllComponents)
			r.Get("/expert-roles", expertHandlers.GetExpertRoles)
			r.Get("/{id}", componentHandlers.GetComponentByID)
			r.Get("/{componentId}/origin/acquired-via", originRelationshipHandlers.GetAcquiredViaByComponent)
			r.Get("/{componentId}/origin/purchased-from", originRelationshipHandlers.GetPurchasedFromByComponent)
			r.Get("/{componentId}/origin/built-by", originRelationshipHandlers.GetBuiltByByComponent)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsWrite))
			r.Post("/", componentHandlers.CreateApplicationComponent)
			r.Put("/{id}", componentHandlers.UpdateApplicationComponent)
			r.Post("/{id}/experts", expertHandlers.AddComponentExpert)
			r.Post("/{componentId}/origin/acquired-via", originRelationshipHandlers.CreateAcquiredViaRelationship)
			r.Post("/{componentId}/origin/purchased-from", originRelationshipHandlers.CreatePurchasedFromRelationship)
			r.Post("/{componentId}/origin/built-by", originRelationshipHandlers.CreateBuiltByRelationship)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsDelete))
			r.Delete("/{id}", componentHandlers.DeleteApplicationComponent)
			r.Delete("/{id}/experts", expertHandlers.RemoveComponentExpert)
		})
	})

	// Register relation routes
	r.Route("/relations", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsRead))
			r.Get("/", relationHandlers.GetAllRelations)
			r.Get("/{id}", relationHandlers.GetRelationByID)
			r.Get("/from/{componentId}", relationHandlers.GetRelationsFromComponent)
			r.Get("/to/{componentId}", relationHandlers.GetRelationsToComponent)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsWrite))
			r.Post("/", relationHandlers.CreateComponentRelation)
			r.Put("/{id}", relationHandlers.UpdateComponentRelation)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsDelete))
			r.Delete("/{id}", relationHandlers.DeleteComponentRelation)
		})
	})

	// Register acquired entity routes
	r.Route("/acquired-entities", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsRead))
			r.Get("/", acquiredEntityHandlers.GetAllAcquiredEntities)
			r.Get("/{id}", acquiredEntityHandlers.GetAcquiredEntityByID)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsWrite))
			r.Post("/", acquiredEntityHandlers.CreateAcquiredEntity)
			r.Put("/{id}", acquiredEntityHandlers.UpdateAcquiredEntity)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsDelete))
			r.Delete("/{id}", acquiredEntityHandlers.DeleteAcquiredEntity)
		})
	})

	// Register vendor routes
	r.Route("/vendors", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsRead))
			r.Get("/", vendorHandlers.GetAllVendors)
			r.Get("/{id}", vendorHandlers.GetVendorByID)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsWrite))
			r.Post("/", vendorHandlers.CreateVendor)
			r.Put("/{id}", vendorHandlers.UpdateVendor)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsDelete))
			r.Delete("/{id}", vendorHandlers.DeleteVendor)
		})
	})

	// Register internal team routes
	r.Route("/internal-teams", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsRead))
			r.Get("/", internalTeamHandlers.GetAllInternalTeams)
			r.Get("/{id}", internalTeamHandlers.GetInternalTeamByID)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsWrite))
			r.Post("/", internalTeamHandlers.CreateInternalTeam)
			r.Put("/{id}", internalTeamHandlers.UpdateInternalTeam)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsDelete))
			r.Delete("/{id}", internalTeamHandlers.DeleteInternalTeam)
		})
	})

	// Register origin relationship routes for delete operations
	r.Route("/origin-relationships", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermComponentsDelete))
			r.Delete("/acquired-via/{id}", originRelationshipHandlers.DeleteAcquiredViaRelationship)
			r.Delete("/purchased-from/{id}", originRelationshipHandlers.DeletePurchasedFromRelationship)
			r.Delete("/built-by/{id}", originRelationshipHandlers.DeleteBuiltByRelationship)
		})
	})

	return nil
}
