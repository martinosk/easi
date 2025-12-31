package api

import (
	"net/http"

	authValueObjects "easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/auth/infrastructure/session"
	"easi/backend/internal/enterprisearchitecture/application/handlers"
	"easi/backend/internal/enterprisearchitecture/application/projectors"
	"easi/backend/internal/enterprisearchitecture/application/readmodels"
	"easi/backend/internal/enterprisearchitecture/infrastructure/repositories"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	platformAPI "easi/backend/internal/platform/infrastructure/api"
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
	RequirePermission(permission authValueObjects.Permission) func(http.Handler) http.Handler
}

type routeRepositories struct {
	capability *repositories.EnterpriseCapabilityRepository
	link       *repositories.EnterpriseCapabilityLinkRepository
	importance *repositories.EnterpriseStrategicImportanceRepository
}

type routeReadModels struct {
	capability *readmodels.EnterpriseCapabilityReadModel
	link       *readmodels.EnterpriseCapabilityLinkReadModel
	importance *readmodels.EnterpriseStrategicImportanceReadModel
	metadata   *readmodels.DomainCapabilityMetadataReadModel
}

type EnterpriseArchRoutesDeps struct {
	Router         chi.Router
	CommandBus     *cqrs.InMemoryCommandBus
	EventStore     eventstore.EventStore
	EventBus       events.EventBus
	DB             *database.TenantAwareDB
	AuthMiddleware AuthMiddleware
	SessionManager *session.SessionManager
}

func SetupEnterpriseArchitectureRoutes(deps EnterpriseArchRoutesDeps) error {
	repos := initializeRepositories(deps.EventStore)
	rm := initializeReadModels(deps.DB)

	setupEventSubscriptions(deps.EventBus, rm)
	setupCommandHandlers(deps.CommandBus, repos, rm)

	httpHandlers := initializeHTTPHandlers(deps.CommandBus, rm, deps.SessionManager)
	rateLimiter := platformAPI.NewRateLimiter(100, 60)
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
	return &routeReadModels{
		capability: readmodels.NewEnterpriseCapabilityReadModel(db),
		link:       readmodels.NewEnterpriseCapabilityLinkReadModel(db),
		importance: readmodels.NewEnterpriseStrategicImportanceReadModel(db),
		metadata:   readmodels.NewDomainCapabilityMetadataReadModel(db),
	}
}

func setupEventSubscriptions(eventBus events.EventBus, rm *routeReadModels) {
	capabilityProjector := projectors.NewEnterpriseCapabilityProjector(rm.capability)
	linkProjector := projectors.NewEnterpriseCapabilityLinkProjector(rm.link)
	importanceProjector := projectors.NewEnterpriseStrategicImportanceProjector(rm.importance)
	metadataProjector := projectors.NewDomainCapabilityMetadataProjector(rm.metadata, rm.capability, rm.link)

	subscribeCapabilityEvents(eventBus, capabilityProjector)
	subscribeLinkEvents(eventBus, linkProjector)
	subscribeImportanceEvents(eventBus, importanceProjector)
	subscribeCapabilityMappingEvents(eventBus, metadataProjector)
}

func subscribeCapabilityEvents(eventBus events.EventBus, projector *projectors.EnterpriseCapabilityProjector) {
	eventTypes := []string{
		"EnterpriseCapabilityCreated",
		"EnterpriseCapabilityUpdated",
		"EnterpriseCapabilityDeleted",
		"EnterpriseCapabilityLinked",
		"EnterpriseCapabilityUnlinked",
	}
	for _, eventType := range eventTypes {
		eventBus.Subscribe(eventType, projector)
	}
}

func subscribeLinkEvents(eventBus events.EventBus, projector *projectors.EnterpriseCapabilityLinkProjector) {
	eventTypes := []string{
		"EnterpriseCapabilityLinked",
		"EnterpriseCapabilityUnlinked",
		"CapabilityParentChanged",
	}
	for _, eventType := range eventTypes {
		eventBus.Subscribe(eventType, projector)
	}
}

func subscribeImportanceEvents(eventBus events.EventBus, projector *projectors.EnterpriseStrategicImportanceProjector) {
	eventTypes := []string{
		"EnterpriseStrategicImportanceSet",
		"EnterpriseStrategicImportanceUpdated",
		"EnterpriseStrategicImportanceRemoved",
	}
	for _, eventType := range eventTypes {
		eventBus.Subscribe(eventType, projector)
	}
}

func subscribeCapabilityMappingEvents(eventBus events.EventBus, projector *projectors.DomainCapabilityMetadataProjector) {
	eventTypes := []string{
		"CapabilityCreated",
		"CapabilityUpdated",
		"CapabilityDeleted",
		"CapabilityParentChanged",
		"CapabilityAssignedToDomain",
		"CapabilityUnassignedFromDomain",
	}
	for _, eventType := range eventTypes {
		eventBus.Subscribe(eventType, projector)
	}
}

func setupCommandHandlers(commandBus *cqrs.InMemoryCommandBus, repos *routeRepositories, rm *routeReadModels) {
	commandBus.Register("CreateEnterpriseCapability", handlers.NewCreateEnterpriseCapabilityHandler(repos.capability, rm.capability))
	commandBus.Register("UpdateEnterpriseCapability", handlers.NewUpdateEnterpriseCapabilityHandler(repos.capability, rm.capability))
	commandBus.Register("DeleteEnterpriseCapability", handlers.NewDeleteEnterpriseCapabilityHandler(repos.capability, rm.link))

	commandBus.Register("LinkCapability", handlers.NewLinkCapabilityHandler(repos.link, repos.capability, rm.link))
	commandBus.Register("UnlinkCapability", handlers.NewUnlinkCapabilityHandler(repos.link))

	commandBus.Register("SetEnterpriseStrategicImportance", handlers.NewSetEnterpriseStrategicImportanceHandler(repos.importance, rm.capability, rm.importance))
	commandBus.Register("UpdateEnterpriseStrategicImportance", handlers.NewUpdateEnterpriseStrategicImportanceHandler(repos.importance))
	commandBus.Register("RemoveEnterpriseStrategicImportance", handlers.NewRemoveEnterpriseStrategicImportanceHandler(repos.importance))
}

func initializeHTTPHandlers(commandBus *cqrs.InMemoryCommandBus, rm *routeReadModels, sessionManager *session.SessionManager) *EnterpriseCapabilityHandlers {
	readModels := &EnterpriseCapabilityReadModels{
		Capability: rm.capability,
		Link:       rm.link,
		Importance: rm.importance,
	}
	return NewEnterpriseCapabilityHandlers(commandBus, readModels, sessionManager)
}

func registerRoutes(r chi.Router, h *EnterpriseCapabilityHandlers, authMiddleware AuthMiddleware, rateLimiter *platformAPI.RateLimiter) {
	r.Route("/enterprise-capabilities", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermEnterpriseArchRead))
			r.Get("/", h.GetAllEnterpriseCapabilities)
			r.Get("/{id}", h.GetEnterpriseCapabilityByID)
			r.Get("/{id}/links", h.GetLinkedCapabilities)
			r.Get("/{id}/strategic-importance", h.GetStrategicImportance)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermEnterpriseArchWrite))
			r.Use(platformAPI.RateLimitMiddleware(rateLimiter))
			r.Post("/", h.CreateEnterpriseCapability)
			r.Put("/{id}", h.UpdateEnterpriseCapability)
			r.Post("/{id}/links", h.LinkCapability)
			r.Post("/{id}/strategic-importance", h.SetStrategicImportance)
			r.Put("/{id}/strategic-importance/{importanceId}", h.UpdateStrategicImportance)
		})

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authValueObjects.PermEnterpriseArchDelete))
			r.Use(platformAPI.RateLimitMiddleware(rateLimiter))
			r.Delete("/{id}", h.DeleteEnterpriseCapability)
			r.Delete("/{id}/links/{linkId}", h.UnlinkCapability)
			r.Delete("/{id}/strategic-importance/{importanceId}", h.RemoveStrategicImportance)
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.RequirePermission(authValueObjects.PermEnterpriseArchRead))
		r.Get("/domain-capabilities/{domainCapabilityId}/enterprise-capability", h.GetEnterpriseCapabilityForDomainCapability)
		r.Get("/domain-capabilities/{domainCapabilityId}/enterprise-link-status", h.GetCapabilityLinkStatus)
		r.Get("/domain-capabilities/enterprise-link-status", h.GetBatchCapabilityLinkStatus)
	})
}
