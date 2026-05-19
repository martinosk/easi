package api

import (
	"net/http"

	"easi/backend/internal/architecturedirection/application/handlers"
	"easi/backend/internal/architecturedirection/application/projectors"
	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/infrastructure/repositories"
	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	authPL "easi/backend/internal/auth/publishedlanguage"
	cmPL "easi/backend/internal/capabilitymapping/publishedlanguage"
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

type RoutesDeps struct {
	Router          chi.Router
	CommandBus      *cqrs.InMemoryCommandBus
	EventStore      eventstore.EventStore
	EventBus        events.EventBus
	DB              *database.TenantAwareDB
	HATEOAS         *sharedAPI.HATEOASLinks
	AuthMiddleware  AuthMiddleware
	SessionProvider authPL.SessionProvider
}

func SetupRoutes(deps RoutesDeps) error {
	readModel := readmodels.NewDirectionReadModel(deps.DB)
	repo := repositories.NewDirectionRepository(deps.EventStore)

	subscribeEvents(deps.EventBus, readModel)
	registerCommandHandlers(deps.CommandBus, repo, readModel)

	links := NewDirectionLinks(deps.HATEOAS)
	httpHandlers := NewDirectionHandlers(deps.CommandBus, readModel, deps.SessionProvider, links)

	registerRoutes(deps.Router, httpHandlers, deps.AuthMiddleware)
	return nil
}

func subscribeEvents(eventBus events.EventBus, rm *readmodels.DirectionReadModel) {
	directionProjector := projectors.NewDirectionProjector(rm)
	staleProjector := projectors.NewStaleReferenceProjector(rm)

	directionEvents := []string{
		pl.DirectionDrafted,
		pl.DirectionProposed,
		pl.DirectionAgreed,
		pl.DirectionRejected,
		pl.DirectionNarrativeUpdated,
		pl.DirectionHorizonChanged,
		pl.DirectionPlacementsChanged,
		pl.DirectionSourceCapabilitiesChanged,
	}
	for _, eventType := range directionEvents {
		eventBus.Subscribe(eventType, directionProjector)
	}

	eventBus.Subscribe(cmPL.CapabilityDeleted, staleProjector)
}

func registerCommandHandlers(
	commandBus *cqrs.InMemoryCommandBus,
	repo *repositories.DirectionRepository,
	rm *readmodels.DirectionReadModel,
) {
	commandBus.Register("CaptureDirection", handlers.NewCaptureDirectionHandler(repo, rm))
	commandBus.Register("AdvanceDirection", handlers.NewAdvanceDirectionHandler(repo))
	commandBus.Register("RejectDirection", handlers.NewRejectDirectionHandler(repo))
	commandBus.Register("UpdateDirectionNarrative", handlers.NewUpdateDirectionNarrativeHandler(repo))
	commandBus.Register("UpdateDirectionHorizon", handlers.NewUpdateDirectionHorizonHandler(repo))
	commandBus.Register("UpdateDirectionSourceCapabilities", handlers.NewUpdateDirectionSourceCapabilitiesHandler(repo))
	commandBus.Register("UpdateDirectionPlacements", handlers.NewUpdateDirectionPlacementsHandler(repo))
}

func registerRoutes(r chi.Router, h *DirectionHandlers, authMiddleware AuthMiddleware) {
	r.Route("/enterprise-capabilities/{id}/direction", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermArchitectureDirectionRead))
			r.Get("/", h.GetDirectionForEnterpriseCapability)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermArchitectureDirectionWrite))
			r.Post("/", h.CaptureDirection)
		})
	})

	r.Route("/directions", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermArchitectureDirectionRead))
			r.Get("/{id}", h.GetDirection)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermArchitectureDirectionWrite))
			r.Put("/{id}", h.UpdateDirection)
			r.Post("/{id}/advance/{target}", h.AdvanceDirection)
			r.Post("/{id}/reject", h.RejectDirection)
		})
	})
}
