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

type RouteConfig struct {
	Router         chi.Router
	CommandBus     *cqrs.InMemoryCommandBus
	EventStore     eventstore.EventStore
	EventBus       events.EventBus
	DB             *database.TenantAwareDB
	HATEOAS        *sharedAPI.HATEOASLinks
	AuthMiddleware AuthMiddleware
}

type repositorySet struct {
	component     *repositories.ApplicationComponentRepository
	relation      *repositories.ComponentRelationRepository
	acquiredEntity *repositories.AcquiredEntityRepository
	vendor        *repositories.VendorRepository
	internalTeam  *repositories.InternalTeamRepository
	acquiredVia   *repositories.AcquiredViaRelationshipRepository
	purchasedFrom *repositories.PurchasedFromRelationshipRepository
	builtBy       *repositories.BuiltByRelationshipRepository
}

type readModelSet struct {
	component     *readmodels.ApplicationComponentReadModel
	relation      *readmodels.ComponentRelationReadModel
	acquiredEntity *readmodels.AcquiredEntityReadModel
	vendor        *readmodels.VendorReadModel
	internalTeam  *readmodels.InternalTeamReadModel
	acquiredVia   *readmodels.AcquiredViaRelationshipReadModel
	purchasedFrom *readmodels.PurchasedFromRelationshipReadModel
	builtBy       *readmodels.BuiltByRelationshipReadModel
}

type httpHandlerSet struct {
	component          *ComponentHandlers
	expert             *ComponentExpertHandlers
	relation           *RelationHandlers
	acquiredEntity     *AcquiredEntityHandlers
	vendor             *VendorHandlers
	internalTeam       *InternalTeamHandlers
	originRelationship *OriginRelationshipHandlers
}

func newRepositorySet(eventStore eventstore.EventStore) *repositorySet {
	return &repositorySet{
		component:     repositories.NewApplicationComponentRepository(eventStore),
		relation:      repositories.NewComponentRelationRepository(eventStore),
		acquiredEntity: repositories.NewAcquiredEntityRepository(eventStore),
		vendor:        repositories.NewVendorRepository(eventStore),
		internalTeam:  repositories.NewInternalTeamRepository(eventStore),
		acquiredVia:   repositories.NewAcquiredViaRelationshipRepository(eventStore),
		purchasedFrom: repositories.NewPurchasedFromRelationshipRepository(eventStore),
		builtBy:       repositories.NewBuiltByRelationshipRepository(eventStore),
	}
}

func newReadModelSet(db *database.TenantAwareDB) *readModelSet {
	return &readModelSet{
		component:     readmodels.NewApplicationComponentReadModel(db),
		relation:      readmodels.NewComponentRelationReadModel(db),
		acquiredEntity: readmodels.NewAcquiredEntityReadModel(db),
		vendor:        readmodels.NewVendorReadModel(db),
		internalTeam:  readmodels.NewInternalTeamReadModel(db),
		acquiredVia:   readmodels.NewAcquiredViaRelationshipReadModel(db),
		purchasedFrom: readmodels.NewPurchasedFromRelationshipReadModel(db),
		builtBy:       readmodels.NewBuiltByRelationshipReadModel(db),
	}
}

func subscribeProjectors(eventBus events.EventBus, rm *readModelSet) {
	componentProjector := projectors.NewApplicationComponentProjector(rm.component)
	relationProjector := projectors.NewComponentRelationProjector(rm.relation)
	acquiredEntityProjector := projectors.NewAcquiredEntityProjector(rm.acquiredEntity)
	vendorProjector := projectors.NewVendorProjector(rm.vendor)
	internalTeamProjector := projectors.NewInternalTeamProjector(rm.internalTeam)
	originRelationshipProjector := projectors.NewOriginRelationshipProjector(rm.acquiredVia, rm.purchasedFrom, rm.builtBy)

	subscribeComponentProjectors(eventBus, componentProjector, relationProjector)
	subscribeOriginEntityProjectors(eventBus, acquiredEntityProjector, vendorProjector, internalTeamProjector)
	subscribeOriginRelationshipProjectors(eventBus, originRelationshipProjector)
}

func subscribeComponentProjectors(eventBus events.EventBus, component, relation events.EventHandler) {
	eventBus.Subscribe("ApplicationComponentCreated", component)
	eventBus.Subscribe("ApplicationComponentUpdated", component)
	eventBus.Subscribe("ApplicationComponentDeleted", component)
	eventBus.Subscribe("ApplicationComponentExpertAdded", component)
	eventBus.Subscribe("ApplicationComponentExpertRemoved", component)
	eventBus.Subscribe("ComponentRelationCreated", relation)
	eventBus.Subscribe("ComponentRelationUpdated", relation)
	eventBus.Subscribe("ComponentRelationDeleted", relation)
}

func subscribeOriginEntityProjectors(eventBus events.EventBus, acquired, vendor, team events.EventHandler) {
	eventBus.Subscribe("AcquiredEntityCreated", acquired)
	eventBus.Subscribe("AcquiredEntityUpdated", acquired)
	eventBus.Subscribe("AcquiredEntityDeleted", acquired)
	eventBus.Subscribe("VendorCreated", vendor)
	eventBus.Subscribe("VendorUpdated", vendor)
	eventBus.Subscribe("VendorDeleted", vendor)
	eventBus.Subscribe("InternalTeamCreated", team)
	eventBus.Subscribe("InternalTeamUpdated", team)
	eventBus.Subscribe("InternalTeamDeleted", team)
}

func subscribeOriginRelationshipProjectors(eventBus events.EventBus, projector events.EventHandler) {
	eventBus.Subscribe("AcquiredViaRelationshipCreated", projector)
	eventBus.Subscribe("AcquiredViaRelationshipDeleted", projector)
	eventBus.Subscribe("PurchasedFromRelationshipCreated", projector)
	eventBus.Subscribe("PurchasedFromRelationshipDeleted", projector)
	eventBus.Subscribe("BuiltByRelationshipCreated", projector)
	eventBus.Subscribe("BuiltByRelationshipDeleted", projector)
}

func registerCommandHandlers(bus *cqrs.InMemoryCommandBus, repos *repositorySet, rm *readModelSet) {
	registerComponentCommandHandlers(bus, repos, rm)
	registerOriginEntityCommandHandlers(bus, repos, rm)
	registerOriginRelationshipCommandHandlers(bus, repos, rm)
}

func registerComponentCommandHandlers(bus *cqrs.InMemoryCommandBus, repos *repositorySet, rm *readModelSet) {
	bus.Register("CreateApplicationComponent", handlers.NewCreateApplicationComponentHandler(repos.component))
	bus.Register("UpdateApplicationComponent", handlers.NewUpdateApplicationComponentHandler(repos.component))
	bus.Register("DeleteApplicationComponent", handlers.NewDeleteApplicationComponentHandler(repos.component, rm.relation, bus))
	bus.Register("AddApplicationComponentExpert", handlers.NewAddApplicationComponentExpertHandler(repos.component))
	bus.Register("RemoveApplicationComponentExpert", handlers.NewRemoveApplicationComponentExpertHandler(repos.component))
	bus.Register("CreateComponentRelation", handlers.NewCreateComponentRelationHandler(repos.relation))
	bus.Register("UpdateComponentRelation", handlers.NewUpdateComponentRelationHandler(repos.relation))
	bus.Register("DeleteComponentRelation", handlers.NewDeleteComponentRelationHandler(repos.relation))
}

func registerOriginEntityCommandHandlers(bus *cqrs.InMemoryCommandBus, repos *repositorySet, rm *readModelSet) {
	bus.Register("CreateAcquiredEntity", handlers.NewCreateAcquiredEntityHandler(repos.acquiredEntity))
	bus.Register("UpdateAcquiredEntity", handlers.NewUpdateAcquiredEntityHandler(repos.acquiredEntity))
	bus.Register("DeleteAcquiredEntity", handlers.NewDeleteAcquiredEntityHandler(repos.acquiredEntity, rm.acquiredVia, bus))
	bus.Register("CreateVendor", handlers.NewCreateVendorHandler(repos.vendor))
	bus.Register("UpdateVendor", handlers.NewUpdateVendorHandler(repos.vendor))
	bus.Register("DeleteVendor", handlers.NewDeleteVendorHandler(repos.vendor, rm.purchasedFrom, bus))
	bus.Register("CreateInternalTeam", handlers.NewCreateInternalTeamHandler(repos.internalTeam))
	bus.Register("UpdateInternalTeam", handlers.NewUpdateInternalTeamHandler(repos.internalTeam))
	bus.Register("DeleteInternalTeam", handlers.NewDeleteInternalTeamHandler(repos.internalTeam, rm.builtBy, bus))
}

func registerOriginRelationshipCommandHandlers(bus *cqrs.InMemoryCommandBus, repos *repositorySet, rm *readModelSet) {
	bus.Register("CreateAcquiredViaRelationship", handlers.NewCreateAcquiredViaRelationshipHandler(repos.acquiredVia, rm.acquiredVia))
	bus.Register("DeleteAcquiredViaRelationship", handlers.NewDeleteAcquiredViaRelationshipHandler(repos.acquiredVia))
	bus.Register("CreatePurchasedFromRelationship", handlers.NewCreatePurchasedFromRelationshipHandler(repos.purchasedFrom, rm.purchasedFrom))
	bus.Register("DeletePurchasedFromRelationship", handlers.NewDeletePurchasedFromRelationshipHandler(repos.purchasedFrom))
	bus.Register("CreateBuiltByRelationship", handlers.NewCreateBuiltByRelationshipHandler(repos.builtBy, rm.builtBy))
	bus.Register("DeleteBuiltByRelationship", handlers.NewDeleteBuiltByRelationshipHandler(repos.builtBy))
}

func newHTTPHandlerSet(bus *cqrs.InMemoryCommandBus, rm *readModelSet, hateoas *sharedAPI.HATEOASLinks) *httpHandlerSet {
	return &httpHandlerSet{
		component:          NewComponentHandlers(bus, rm.component, hateoas),
		expert:             NewComponentExpertHandlers(bus, rm.component),
		relation:           NewRelationHandlers(bus, rm.relation, hateoas),
		acquiredEntity:     NewAcquiredEntityHandlers(bus, rm.acquiredEntity, hateoas),
		vendor:             NewVendorHandlers(bus, rm.vendor, hateoas),
		internalTeam:       NewInternalTeamHandlers(bus, rm.internalTeam, hateoas),
		originRelationship: NewOriginRelationshipHandlers(bus, rm.acquiredVia, rm.purchasedFrom, rm.builtBy, hateoas),
	}
}

func registerRoutes(r chi.Router, h *httpHandlerSet, auth AuthMiddleware) {
	registerComponentRoutes(r, h, auth)
	registerRelationRoutes(r, h, auth)
	registerOriginEntityRoutes(r, h, auth)
	registerOriginRelationshipRoutes(r, h, auth)
}

func registerComponentRoutes(r chi.Router, h *httpHandlerSet, auth AuthMiddleware) {
	r.Route("/components", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authValueObjects.PermComponentsRead))
			r.Get("/", h.component.GetAllComponents)
			r.Get("/expert-roles", h.expert.GetExpertRoles)
			r.Get("/{id}", h.component.GetComponentByID)
			r.Get("/{componentId}/origins", h.originRelationship.GetAllOriginsByComponent)
			r.Get("/{componentId}/origin/acquired-via", h.originRelationship.GetAcquiredViaByComponent)
			r.Get("/{componentId}/origin/purchased-from", h.originRelationship.GetPurchasedFromByComponent)
			r.Get("/{componentId}/origin/built-by", h.originRelationship.GetBuiltByByComponent)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authValueObjects.PermComponentsWrite))
			r.Post("/", h.component.CreateApplicationComponent)
			r.Put("/{id}", h.component.UpdateApplicationComponent)
			r.Post("/{id}/experts", h.expert.AddComponentExpert)
			r.Post("/{componentId}/origin/acquired-via", h.originRelationship.CreateAcquiredViaRelationship)
			r.Post("/{componentId}/origin/purchased-from", h.originRelationship.CreatePurchasedFromRelationship)
			r.Post("/{componentId}/origin/built-by", h.originRelationship.CreateBuiltByRelationship)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authValueObjects.PermComponentsDelete))
			r.Delete("/{id}", h.component.DeleteApplicationComponent)
			r.Delete("/{id}/experts", h.expert.RemoveComponentExpert)
		})
	})
}

func registerRelationRoutes(r chi.Router, h *httpHandlerSet, auth AuthMiddleware) {
	r.Route("/relations", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authValueObjects.PermComponentsRead))
			r.Get("/", h.relation.GetAllRelations)
			r.Get("/{id}", h.relation.GetRelationByID)
			r.Get("/from/{componentId}", h.relation.GetRelationsFromComponent)
			r.Get("/to/{componentId}", h.relation.GetRelationsToComponent)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authValueObjects.PermComponentsWrite))
			r.Post("/", h.relation.CreateComponentRelation)
			r.Put("/{id}", h.relation.UpdateComponentRelation)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authValueObjects.PermComponentsDelete))
			r.Delete("/{id}", h.relation.DeleteComponentRelation)
		})
	})
}

func registerOriginEntityRoutes(r chi.Router, h *httpHandlerSet, auth AuthMiddleware) {
	r.Route("/acquired-entities", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authValueObjects.PermComponentsRead))
			r.Get("/", h.acquiredEntity.GetAllAcquiredEntities)
			r.Get("/{id}", h.acquiredEntity.GetAcquiredEntityByID)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authValueObjects.PermComponentsWrite))
			r.Post("/", h.acquiredEntity.CreateAcquiredEntity)
			r.Put("/{id}", h.acquiredEntity.UpdateAcquiredEntity)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authValueObjects.PermComponentsDelete))
			r.Delete("/{id}", h.acquiredEntity.DeleteAcquiredEntity)
		})
	})

	r.Route("/vendors", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authValueObjects.PermComponentsRead))
			r.Get("/", h.vendor.GetAllVendors)
			r.Get("/{id}", h.vendor.GetVendorByID)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authValueObjects.PermComponentsWrite))
			r.Post("/", h.vendor.CreateVendor)
			r.Put("/{id}", h.vendor.UpdateVendor)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authValueObjects.PermComponentsDelete))
			r.Delete("/{id}", h.vendor.DeleteVendor)
		})
	})

	r.Route("/internal-teams", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authValueObjects.PermComponentsRead))
			r.Get("/", h.internalTeam.GetAllInternalTeams)
			r.Get("/{id}", h.internalTeam.GetInternalTeamByID)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authValueObjects.PermComponentsWrite))
			r.Post("/", h.internalTeam.CreateInternalTeam)
			r.Put("/{id}", h.internalTeam.UpdateInternalTeam)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authValueObjects.PermComponentsDelete))
			r.Delete("/{id}", h.internalTeam.DeleteInternalTeam)
		})
	})
}

func registerOriginRelationshipRoutes(r chi.Router, h *httpHandlerSet, auth AuthMiddleware) {
	r.Route("/origin-relationships", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authValueObjects.PermComponentsRead))
			r.Get("/", h.originRelationship.GetAllOriginRelationships)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authValueObjects.PermComponentsDelete))
			r.Delete("/acquired-via/{id}", h.originRelationship.DeleteAcquiredViaRelationship)
			r.Delete("/purchased-from/{id}", h.originRelationship.DeletePurchasedFromRelationship)
			r.Delete("/built-by/{id}", h.originRelationship.DeleteBuiltByRelationship)
		})
	})
}

func SetupArchitectureModelingRoutes(cfg RouteConfig) error {
	repos := newRepositorySet(cfg.EventStore)
	rm := newReadModelSet(cfg.DB)

	subscribeProjectors(cfg.EventBus, rm)
	registerCommandHandlers(cfg.CommandBus, repos, rm)

	handlers := newHTTPHandlerSet(cfg.CommandBus, rm, cfg.HATEOAS)
	registerRoutes(cfg.Router, handlers, cfg.AuthMiddleware)

	return nil
}
