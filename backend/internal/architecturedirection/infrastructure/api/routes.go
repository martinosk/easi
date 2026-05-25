package api

import (
	"net/http"

	"easi/backend/internal/architecturedirection/application/handlers"
	"easi/backend/internal/architecturedirection/application/projectors"
	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/infrastructure/repositories"
	pl "easi/backend/internal/architecturedirection/publishedlanguage"
	amPL "easi/backend/internal/architecturemodeling/publishedlanguage"
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
	Router           chi.Router
	CommandBus       *cqrs.InMemoryCommandBus
	EventStore       eventstore.EventStore
	EventBus         events.EventBus
	DB               *database.TenantAwareDB
	HATEOAS          *sharedAPI.HATEOASLinks
	AuthMiddleware   AuthMiddleware
	ReferenceChecker *services.ReferenceChecker
}

func SetupRoutes(deps RoutesDeps) error {
	readModel := readmodels.NewDirectionReadModel(deps.DB)
	repo := repositories.NewDirectionRepository(deps.EventStore)

	subscribeEvents(deps.EventBus, readModel)
	registerCommandHandlers(deps.CommandBus, repo, readModel, deps.ReferenceChecker)

	links := NewDirectionLinks(deps.HATEOAS)
	httpHandlers := NewDirectionHandlers(deps.CommandBus, readModel, links)

	registerRoutes(deps.Router, httpHandlers, deps.AuthMiddleware)

	setupStandardApplicationRoutes(deps)
	return nil
}

func setupStandardApplicationRoutes(deps RoutesDeps) {
	readModel := readmodels.NewStandardApplicationReadModel(deps.DB)
	repo := repositories.NewStandardApplicationRepository(deps.EventStore)

	subscribeStandardApplicationEvents(deps.EventBus, readModel)
	deps.CommandBus.Register("SetStandardApplication", handlers.NewSetStandardApplicationHandler(repo, readModel, deps.ReferenceChecker))

	links := NewStandardApplicationLinks(deps.HATEOAS)
	httpHandlers := NewStandardApplicationHandlers(deps.CommandBus, readModel, links)

	registerStandardApplicationRoutes(deps.Router, httpHandlers, deps.AuthMiddleware)
}

func subscribeStandardApplicationEvents(eventBus events.EventBus, rm *readmodels.StandardApplicationReadModel) {
	projector := projectors.NewStandardApplicationProjector(rm)
	staleProjector := projectors.NewStaleApplicationProjector(rm)
	eventBus.Subscribe(pl.StandardApplicationSet, projector)
	eventBus.Subscribe(amPL.ApplicationComponentDeleted, staleProjector)
}

func registerStandardApplicationRoutes(r chi.Router, h *StandardApplicationHandlers, authMiddleware AuthMiddleware) {
	r.Route("/enterprise-capabilities/{id}/standard-application", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermArchitectureDirectionRead))
			r.Get("/", h.GetStandardApplicationForEnterpriseCapability)
			r.Get("/history", h.GetStandardApplicationHistory)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermArchitectureDirectionWrite))
			r.Put("/", h.SetStandardApplication)
		})
	})
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
	refs *services.ReferenceChecker,
) {
	policy := services.NewDirectionReferenceService(refs, rm)
	commandBus.Register("CaptureDirection", handlers.NewCaptureDirectionHandler(repo, policy))
	commandBus.Register("AdvanceDirection", handlers.NewAdvanceDirectionHandler(repo))
	commandBus.Register("RejectDirection", handlers.NewRejectDirectionHandler(repo))
	commandBus.Register("UpdateDirection", handlers.NewUpdateDirectionHandler(repo))
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
			r.Put("/", h.UpdateDirection)
			r.Post("/propose", h.ProposeDirection)
			r.Post("/agree", h.AgreeDirection)
			r.Post("/reject", h.RejectDirection)
		})
	})
}
