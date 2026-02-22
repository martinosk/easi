package api

import (
	"net/http"

	"easi/backend/internal/architecturemodeling/application/handlers"
	"easi/backend/internal/architecturemodeling/application/projectors"
	"easi/backend/internal/architecturemodeling/application/readmodels"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	archPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	authPL "easi/backend/internal/auth/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequirePermission(permission authPL.Permission) func(http.Handler) http.Handler
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
	component           *repositories.ApplicationComponentRepository
	relation            *repositories.ComponentRelationRepository
	acquiredEntity      *repositories.AcquiredEntityRepository
	vendor              *repositories.VendorRepository
	internalTeam        *repositories.InternalTeamRepository
	componentOriginLink *repositories.ComponentOriginLinkRepository
}

type readModelSet struct {
	component      *readmodels.ApplicationComponentReadModel
	relation       *readmodels.ComponentRelationReadModel
	acquiredEntity *readmodels.AcquiredEntityReadModel
	vendor         *readmodels.VendorReadModel
	internalTeam   *readmodels.InternalTeamReadModel
	acquiredVia    *readmodels.AcquiredViaRelationshipReadModel
	purchasedFrom  *readmodels.PurchasedFromRelationshipReadModel
	builtBy        *readmodels.BuiltByRelationshipReadModel
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
		component:           repositories.NewApplicationComponentRepository(eventStore),
		relation:            repositories.NewComponentRelationRepository(eventStore),
		acquiredEntity:      repositories.NewAcquiredEntityRepository(eventStore),
		vendor:              repositories.NewVendorRepository(eventStore),
		internalTeam:        repositories.NewInternalTeamRepository(eventStore),
		componentOriginLink: repositories.NewComponentOriginLinkRepository(eventStore),
	}
}

func newReadModelSet(db *database.TenantAwareDB) *readModelSet {
	return &readModelSet{
		component:      readmodels.NewApplicationComponentReadModel(db),
		relation:       readmodels.NewComponentRelationReadModel(db),
		acquiredEntity: readmodels.NewAcquiredEntityReadModel(db),
		vendor:         readmodels.NewVendorReadModel(db),
		internalTeam:   readmodels.NewInternalTeamReadModel(db),
		acquiredVia:    readmodels.NewAcquiredViaRelationshipReadModel(db),
		purchasedFrom:  readmodels.NewPurchasedFromRelationshipReadModel(db),
		builtBy:        readmodels.NewBuiltByRelationshipReadModel(db),
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
	eventBus.Subscribe(archPL.ApplicationComponentCreated, component)
	eventBus.Subscribe(archPL.ApplicationComponentUpdated, component)
	eventBus.Subscribe(archPL.ApplicationComponentDeleted, component)
	eventBus.Subscribe(archPL.ApplicationComponentExpertAdded, component)
	eventBus.Subscribe(archPL.ApplicationComponentExpertRemoved, component)
	eventBus.Subscribe(archPL.ComponentRelationCreated, relation)
	eventBus.Subscribe(archPL.ComponentRelationUpdated, relation)
	eventBus.Subscribe(archPL.ComponentRelationDeleted, relation)
}

func subscribeOriginEntityProjectors(eventBus events.EventBus, acquired, vendor, team events.EventHandler) {
	eventBus.Subscribe(archPL.AcquiredEntityCreated, acquired)
	eventBus.Subscribe(archPL.AcquiredEntityUpdated, acquired)
	eventBus.Subscribe(archPL.AcquiredEntityDeleted, acquired)
	eventBus.Subscribe(archPL.VendorCreated, vendor)
	eventBus.Subscribe(archPL.VendorUpdated, vendor)
	eventBus.Subscribe(archPL.VendorDeleted, vendor)
	eventBus.Subscribe(archPL.InternalTeamCreated, team)
	eventBus.Subscribe(archPL.InternalTeamUpdated, team)
	eventBus.Subscribe(archPL.InternalTeamDeleted, team)
}

func subscribeOriginRelationshipProjectors(eventBus events.EventBus, projector events.EventHandler) {
	eventBus.Subscribe(archPL.OriginLinkSet, projector)
	eventBus.Subscribe(archPL.OriginLinkReplaced, projector)
	eventBus.Subscribe(archPL.OriginLinkNotesUpdated, projector)
	eventBus.Subscribe(archPL.OriginLinkCleared, projector)
	eventBus.Subscribe(archPL.OriginLinkDeleted, projector)
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

func registerOriginRelationshipCommandHandlers(bus *cqrs.InMemoryCommandBus, repos *repositorySet, _ *readModelSet) {
	bus.Register("SetOriginLink", handlers.NewSetOriginLinkHandler(repos.componentOriginLink))
	bus.Register("ClearOriginLink", handlers.NewClearOriginLinkHandler(repos.componentOriginLink))
}

func newHTTPHandlerSet(bus *cqrs.InMemoryCommandBus, rm *readModelSet, hateoas *sharedAPI.HATEOASLinks) *httpHandlerSet {
	links := NewArchitectureModelingLinks(hateoas)
	return &httpHandlerSet{
		component:      NewComponentHandlers(bus, rm.component, links),
		expert:         NewComponentExpertHandlers(bus, rm.component),
		relation:       NewRelationHandlers(bus, rm.relation, links),
		acquiredEntity: NewAcquiredEntityHandlers(bus, rm.acquiredEntity, links),
		vendor:         NewVendorHandlers(bus, rm.vendor, links),
		internalTeam:   NewInternalTeamHandlers(bus, rm.internalTeam, links),
		originRelationship: NewOriginRelationshipHandlersFromConfig(OriginRelationshipHandlersConfig{
			CommandBus: bus,
			ReadModels: OriginReadModels{
				AcquiredVia:   rm.acquiredVia,
				PurchasedFrom: rm.purchasedFrom,
				BuiltBy:       rm.builtBy,
			},
			HATEOAS: links,
		}),
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
			r.Use(auth.RequirePermission(authPL.PermComponentsRead))
			r.Get("/", h.component.GetAllComponents)
			r.Get("/expert-roles", h.expert.GetExpertRoles)
			r.Get("/{id}", h.component.GetComponentByID)
			r.Get("/{componentId}/origins", h.originRelationship.GetAllOriginsByComponent)
			r.Get("/{componentId}/origin/acquired-via", h.originRelationship.GetAcquiredViaByComponent)
			r.Get("/{componentId}/origin/purchased-from", h.originRelationship.GetPurchasedFromByComponent)
			r.Get("/{componentId}/origin/built-by", h.originRelationship.GetBuiltByByComponent)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authPL.PermComponentsWrite))
			r.Post("/", h.component.CreateApplicationComponent)
			r.Post("/{id}/experts", h.expert.AddComponentExpert)
			r.Put("/{componentId}/origin/acquired-via", h.originRelationship.CreateAcquiredViaRelationship)
			r.Put("/{componentId}/origin/purchased-from", h.originRelationship.CreatePurchasedFromRelationship)
			r.Put("/{componentId}/origin/built-by", h.originRelationship.CreateBuiltByRelationship)
		})
		r.Group(func(r chi.Router) {
			r.Use(sharedAPI.RequireWriteOrEditGrant("components", "id"))
			r.Put("/{id}", h.component.UpdateApplicationComponent)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authPL.PermComponentsDelete))
			r.Delete("/{id}", h.component.DeleteApplicationComponent)
			r.Delete("/{id}/experts", h.expert.RemoveComponentExpert)
			r.Delete("/{componentId}/origin/acquired-via", h.originRelationship.DeleteAcquiredViaRelationship)
			r.Delete("/{componentId}/origin/purchased-from", h.originRelationship.DeletePurchasedFromRelationship)
			r.Delete("/{componentId}/origin/built-by", h.originRelationship.DeleteBuiltByRelationship)
		})
	})
}

func registerRelationRoutes(r chi.Router, h *httpHandlerSet, auth AuthMiddleware) {
	r.Route("/relations", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authPL.PermComponentsRead))
			r.Get("/", h.relation.GetAllRelations)
			r.Get("/{id}", h.relation.GetRelationByID)
			r.Get("/from/{componentId}", h.relation.GetRelationsFromComponent)
			r.Get("/to/{componentId}", h.relation.GetRelationsToComponent)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authPL.PermComponentsWrite))
			r.Post("/", h.relation.CreateComponentRelation)
			r.Put("/{id}", h.relation.UpdateComponentRelation)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authPL.PermComponentsDelete))
			r.Delete("/{id}", h.relation.DeleteComponentRelation)
		})
	})
}

func registerOriginEntityRoutes(r chi.Router, h *httpHandlerSet, auth AuthMiddleware) {
	r.Route("/acquired-entities", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authPL.PermComponentsRead))
			r.Get("/", h.acquiredEntity.GetAllAcquiredEntities)
			r.Get("/{id}", h.acquiredEntity.GetAcquiredEntityByID)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authPL.PermComponentsWrite))
			r.Post("/", h.acquiredEntity.CreateAcquiredEntity)
		})
		r.Group(func(r chi.Router) {
			r.Use(sharedAPI.RequireWriteOrEditGrantFor("components", "acquired_entities", "id"))
			r.Put("/{id}", h.acquiredEntity.UpdateAcquiredEntity)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authPL.PermComponentsDelete))
			r.Delete("/{id}", h.acquiredEntity.DeleteAcquiredEntity)
		})
	})

	r.Route("/vendors", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authPL.PermComponentsRead))
			r.Get("/", h.vendor.GetAllVendors)
			r.Get("/{id}", h.vendor.GetVendorByID)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authPL.PermComponentsWrite))
			r.Post("/", h.vendor.CreateVendor)
		})
		r.Group(func(r chi.Router) {
			r.Use(sharedAPI.RequireWriteOrEditGrantFor("components", "vendors", "id"))
			r.Put("/{id}", h.vendor.UpdateVendor)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authPL.PermComponentsDelete))
			r.Delete("/{id}", h.vendor.DeleteVendor)
		})
	})

	r.Route("/internal-teams", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authPL.PermComponentsRead))
			r.Get("/", h.internalTeam.GetAllInternalTeams)
			r.Get("/{id}", h.internalTeam.GetInternalTeamByID)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authPL.PermComponentsWrite))
			r.Post("/", h.internalTeam.CreateInternalTeam)
		})
		r.Group(func(r chi.Router) {
			r.Use(sharedAPI.RequireWriteOrEditGrantFor("components", "internal_teams", "id"))
			r.Put("/{id}", h.internalTeam.UpdateInternalTeam)
		})
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authPL.PermComponentsDelete))
			r.Delete("/{id}", h.internalTeam.DeleteInternalTeam)
		})
	})
}

func registerOriginRelationshipRoutes(r chi.Router, h *httpHandlerSet, auth AuthMiddleware) {
	r.Route("/origin-relationships", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth.RequirePermission(authPL.PermComponentsRead))
			r.Get("/", h.originRelationship.GetAllOriginRelationships)
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
