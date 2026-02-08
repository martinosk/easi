package api

import (
	"easi/backend/internal/accessdelegation/application/handlers"
	"easi/backend/internal/accessdelegation/application/projectors"
	"easi/backend/internal/accessdelegation/application/readmodels"
	"easi/backend/internal/accessdelegation/infrastructure/repositories"
	archPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	viewsPL "easi/backend/internal/architectureviews/publishedlanguage"
	authAPI "easi/backend/internal/auth/infrastructure/api"
	capPL "easi/backend/internal/capabilitymapping/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

type AccessDelegationRoutesDeps struct {
	CommandBus     *cqrs.InMemoryCommandBus
	EventStore     eventstore.EventStore
	EventBus       *events.InMemoryEventBus
	DB             *database.TenantAwareDB
	HATEOAS        *sharedAPI.HATEOASLinks
	AuthMiddleware *authAPI.AuthMiddleware
}

type AccessDelegationDependencies struct {
	GrantResolver  *readmodels.EditGrantReadModel
	handlers       *EditGrantHandlers
	authMiddleware *authAPI.AuthMiddleware
}

func (d *AccessDelegationDependencies) RegisterRoutes(r chi.Router) {
	registerRoutes(r, d.handlers, d.authMiddleware)
}

func SetupAccessDelegationRoutes(deps AccessDelegationRoutesDeps) (*AccessDelegationDependencies, error) {
	repo := repositories.NewEditGrantRepository(deps.EventStore)
	readModel := readmodels.NewEditGrantReadModel(deps.DB)

	registerCommandHandlers(deps.CommandBus, repo)
	registerEventSubscriptions(deps.EventBus, readModel)
	registerArtifactDeletionSubscriptions(deps.EventBus, readModel, deps.CommandBus)

	httpHandlers := NewEditGrantHandlers(deps.CommandBus, readModel, deps.HATEOAS)

	return &AccessDelegationDependencies{
		GrantResolver:  readModel,
		handlers:       httpHandlers,
		authMiddleware: deps.AuthMiddleware,
	}, nil
}

func registerCommandHandlers(commandBus *cqrs.InMemoryCommandBus, repo *repositories.EditGrantRepository) {
	commandBus.Register("CreateEditGrant", handlers.NewCreateEditGrantHandler(repo))
	commandBus.Register("RevokeEditGrant", handlers.NewRevokeEditGrantHandler(repo))
}

func registerEventSubscriptions(eventBus *events.InMemoryEventBus, readModel *readmodels.EditGrantReadModel) {
	projector := projectors.NewEditGrantProjector(readModel)
	eventBus.Subscribe("EditGrantActivated", projector)
	eventBus.Subscribe("EditGrantRevoked", projector)
	eventBus.Subscribe("EditGrantExpired", projector)
}

func registerArtifactDeletionSubscriptions(eventBus *events.InMemoryEventBus, readModel *readmodels.EditGrantReadModel, commandBus cqrs.CommandBus) {
	capabilityDeletionProjector := projectors.NewArtifactDeletionProjector(readModel, commandBus, "capability")
	componentDeletionProjector := projectors.NewArtifactDeletionProjector(readModel, commandBus, "component")
	viewDeletionProjector := projectors.NewArtifactDeletionProjector(readModel, commandBus, "view")

	eventBus.Subscribe(capPL.CapabilityDeleted, capabilityDeletionProjector)
	eventBus.Subscribe(archPL.ApplicationComponentDeleted, componentDeletionProjector)
	eventBus.Subscribe(viewsPL.ViewDeleted, viewDeletionProjector)
}

func registerRoutes(r chi.Router, h *EditGrantHandlers, authMiddleware *authAPI.AuthMiddleware) {
	r.Route("/edit-grants", func(r chi.Router) {
		r.Use(authMiddleware.RequireAuth())
		r.Post("/", h.CreateEditGrant)
		r.Get("/", h.GetMyEditGrants)
		r.Get("/{id}", h.GetEditGrantByID)
		r.Delete("/{id}", h.RevokeEditGrant)
		r.Get("/artifact/{artifactType}/{artifactId}", h.GetEditGrantsForArtifact)
	})
}
