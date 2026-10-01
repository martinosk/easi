package api

import (
	"net/http"

	authPL "easi/backend/internal/auth/publishedlanguage"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	"easi/backend/internal/stewardship/application/handlers"
	"easi/backend/internal/stewardship/application/projectors"
	"easi/backend/internal/stewardship/application/readmodels"
	"easi/backend/internal/stewardship/infrastructure/repositories"
	pl "easi/backend/internal/stewardship/publishedlanguage"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequirePermission(permission authPL.Permission) func(http.Handler) http.Handler
}

type RoutesDeps struct {
	Router         chi.Router
	CommandBus     cqrs.CommandBus
	EventStore     eventstore.EventStore
	EventBus       events.EventBus
	DB             *database.TenantAwareDB
	HATEOAS        *sharedAPI.HATEOASLinks
	AuthMiddleware AuthMiddleware
}

type stewardshipModels struct {
	stewardships *readmodels.StewardshipReadModel
	domains      *readmodels.DomainCacheReadModel
	users        *readmodels.UserCacheReadModel
}

func SetupRoutes(deps RoutesDeps) error {
	models := stewardshipModels{
		stewardships: readmodels.NewStewardshipReadModel(deps.DB),
		domains:      readmodels.NewDomainCacheReadModel(deps.DB),
		users:        readmodels.NewUserCacheReadModel(deps.DB),
	}
	registerCommandHandlers(deps.CommandBus, repositories.NewStewardshipRepository(deps.EventStore), models)
	subscribeEvents(deps.EventBus, deps.CommandBus, models)

	httpHandlers := NewStewardshipHandlers(
		deps.CommandBus,
		readmodels.NewStewardshipView(models.domains, models.stewardships),
		NewStewardshipLinks(deps.HATEOAS),
	)
	RegisterRoutes(deps.Router, httpHandlers, deps.AuthMiddleware)
	setupHome(deps)
	return nil
}

func registerCommandHandlers(commandBus cqrs.CommandBus, repo *repositories.StewardshipRepository, models stewardshipModels) {
	commandBus.Register("AssignSteward", handlers.NewAssignStewardHandler(handlers.AssignStewardDeps{
		Repository: repo,
		Lookup:     models.stewardships,
		Domains:    models.domains,
		Users:      models.users,
	}))
	commandBus.Register("ReleaseSteward", handlers.NewReleaseStewardHandler(handlers.ReleaseStewardDeps{
		Repository: repo,
		Lookup:     models.stewardships,
		Domains:    models.domains,
	}))
}

func subscribeEvents(eventBus events.EventBus, commandBus cqrs.CommandBus, models stewardshipModels) {
	subscribeMany(eventBus, projectors.NewStewardshipProjector(models.stewardships), pl.StewardAssigned, pl.StewardReleased)
	subscribeMany(eventBus, projectors.NewDomainCacheProjector(models.domains), capPL.BusinessDomainCreated, capPL.BusinessDomainUpdated)
	subscribeMany(eventBus, projectors.NewDomainDeletionReactor(models.stewardships, commandBus, models.domains), capPL.BusinessDomainDeleted)
	subscribeMany(eventBus, projectors.NewUserCacheProjector(models.users), authPL.UserCreated, authPL.UserDisabled, authPL.UserEnabled)
}

func subscribeMany(eventBus events.EventBus, handler events.EventHandler, eventTypes ...string) {
	for _, eventType := range eventTypes {
		eventBus.Subscribe(eventType, handler)
	}
}

func RegisterRoutes(r chi.Router, h *StewardshipHandlers, authMiddleware AuthMiddleware) {
	r.Route(string(stewardshipsPath), func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermDomainsRead))
			r.Get("/", h.GetStewardships)
			r.Get("/{domainId}/{concern}", h.GetStewardship)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermDomainsWrite))
			r.Put("/{domainId}/{concern}", h.AssignSteward)
			r.Delete("/{domainId}/{concern}", h.ReleaseSteward)
		})
	})
}
