package api

import (
	"easi/backend/internal/accessdelegation/application/handlers"
	"easi/backend/internal/accessdelegation/application/ports"
	"easi/backend/internal/accessdelegation/application/projectors"
	"easi/backend/internal/accessdelegation/application/readmodels"
	"easi/backend/internal/accessdelegation/infrastructure/repositories"
	adServices "easi/backend/internal/accessdelegation/infrastructure/services"
	adPL "easi/backend/internal/accessdelegation/publishedlanguage"
	archPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	viewsPL "easi/backend/internal/architectureviews/publishedlanguage"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/infrastructure/api/middleware"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequireAuth() func(http.Handler) http.Handler
}

type AccessDelegationRoutesDeps struct {
	CommandBus     *cqrs.InMemoryCommandBus
	EventStore     eventstore.EventStore
	EventBus       *events.InMemoryEventBus
	DB             *database.TenantAwareDB
	HATEOAS        *sharedAPI.HATEOASLinks
	AuthMiddleware AuthMiddleware
	NameLookups    adServices.ArtifactNameResolverDeps
	UserLookup     ports.UserEmailLookup
	InvChecker     ports.InvitationChecker
	DomainChecker  ports.DomainAllowlistChecker
}

type AccessDelegationDependencies struct {
	GrantResolver  *readmodels.EditGrantReadModel
	handlers       *EditGrantHandlers
	authMiddleware AuthMiddleware
	rateLimiter    *middleware.RateLimiter
}

func (d *AccessDelegationDependencies) RegisterRoutes(r chi.Router) {
	registerRoutes(r, d.handlers, d.authMiddleware, d.rateLimiter)
}

func SetupAccessDelegationRoutes(deps AccessDelegationRoutesDeps) (*AccessDelegationDependencies, error) {
	repo := repositories.NewEditGrantRepository(deps.EventStore)
	readModel := readmodels.NewEditGrantReadModel(deps.DB)

	registerCommandHandlers(deps.CommandBus, repo)
	registerEventSubscriptions(deps.EventBus, readModel)
	registerArtifactDeletionSubscriptions(deps.EventBus, readModel, deps.CommandBus)

	nameResolver := adServices.NewArtifactNameResolver(deps.NameLookups)

	httpHandlers := NewEditGrantHandlers(EditGrantHandlerDeps{
		CommandBus:    deps.CommandBus,
		ReadModel:     readModel,
		Hateoas:       NewEditGrantLinks(deps.HATEOAS),
		NameResolver:  nameResolver,
		UserLookup:    deps.UserLookup,
		InvChecker:    deps.InvChecker,
		DomainChecker: deps.DomainChecker,
		EventBus:      deps.EventBus,
	})
	rateLimiter := middleware.NewRateLimiter(100, 60)

	return &AccessDelegationDependencies{
		GrantResolver:  readModel,
		handlers:       httpHandlers,
		authMiddleware: deps.AuthMiddleware,
		rateLimiter:    rateLimiter,
	}, nil
}

func registerCommandHandlers(commandBus *cqrs.InMemoryCommandBus, repo *repositories.EditGrantRepository) {
	commandBus.Register("CreateEditGrant", handlers.NewCreateEditGrantHandler(repo))
	commandBus.Register("RevokeEditGrant", handlers.NewRevokeEditGrantHandler(repo))
}

func registerEventSubscriptions(eventBus *events.InMemoryEventBus, readModel *readmodels.EditGrantReadModel) {
	projector := projectors.NewEditGrantProjector(readModel)
	eventBus.Subscribe(adPL.EditGrantActivated, projector)
	eventBus.Subscribe(adPL.EditGrantRevoked, projector)
	eventBus.Subscribe(adPL.EditGrantExpired, projector)
}

func registerArtifactDeletionSubscriptions(eventBus *events.InMemoryEventBus, readModel *readmodels.EditGrantReadModel, commandBus cqrs.CommandBus) {
	capabilityDeletionProjector := projectors.NewArtifactDeletionProjector(readModel, commandBus, "capability")
	componentDeletionProjector := projectors.NewArtifactDeletionProjector(readModel, commandBus, "component")
	viewDeletionProjector := projectors.NewArtifactDeletionProjector(readModel, commandBus, "view")
	domainDeletionProjector := projectors.NewArtifactDeletionProjector(readModel, commandBus, "domain")
	acquiredEntityDeletionProjector := projectors.NewArtifactDeletionProjector(readModel, commandBus, "acquired_entity")
	vendorDeletionProjector := projectors.NewArtifactDeletionProjector(readModel, commandBus, "vendor")
	internalTeamDeletionProjector := projectors.NewArtifactDeletionProjector(readModel, commandBus, "internal_team")

	eventBus.Subscribe(capPL.CapabilityDeleted, capabilityDeletionProjector)
	eventBus.Subscribe(archPL.ApplicationComponentDeleted, componentDeletionProjector)
	eventBus.Subscribe(viewsPL.ViewDeleted, viewDeletionProjector)
	eventBus.Subscribe(capPL.BusinessDomainDeleted, domainDeletionProjector)
	eventBus.Subscribe(archPL.AcquiredEntityDeleted, acquiredEntityDeletionProjector)
	eventBus.Subscribe(archPL.VendorDeleted, vendorDeletionProjector)
	eventBus.Subscribe(archPL.InternalTeamDeleted, internalTeamDeletionProjector)
}

func registerRoutes(r chi.Router, h *EditGrantHandlers, authMiddleware AuthMiddleware, rateLimiter *middleware.RateLimiter) {
	r.Route("/edit-grants", func(r chi.Router) {
		r.Use(authMiddleware.RequireAuth())
		r.Group(func(r chi.Router) {
			r.Use(middleware.RateLimitMiddleware(rateLimiter))
			r.Post("/", h.CreateEditGrant)
		})
		r.Get("/", h.GetMyEditGrants)
		r.Get("/{id}", h.GetEditGrantByID)
		r.Delete("/{id}", h.RevokeEditGrant)
		r.Get("/artifact/{artifactType}/{artifactId}", h.GetEditGrantsForArtifact)
	})
}
