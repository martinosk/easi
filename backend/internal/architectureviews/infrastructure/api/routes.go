package api

import (
	"net/http"

	archPL "easi/backend/internal/architecturemodeling/publishedlanguage"
	"easi/backend/internal/architectureviews/application/handlers"
	"easi/backend/internal/architectureviews/application/ports"
	"easi/backend/internal/architectureviews/application/projectors"
	"easi/backend/internal/architectureviews/application/readmodels"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	viewsPL "easi/backend/internal/architectureviews/publishedlanguage"
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

func SubscribeEvents(eventBus events.EventBus, commandBus *cqrs.InMemoryCommandBus, db *database.TenantAwareDB) {
	viewReadModel := readmodels.NewArchitectureViewReadModel(db)
	viewProjector := projectors.NewArchitectureViewProjector(viewReadModel)

	eventBus.Subscribe(viewsPL.ViewCreated, viewProjector)
	eventBus.Subscribe(viewsPL.ComponentAddedToView, viewProjector)
	eventBus.Subscribe(viewsPL.ComponentRemovedFromView, viewProjector)
	eventBus.Subscribe(viewsPL.ViewRenamed, viewProjector)
	eventBus.Subscribe(viewsPL.ViewDeleted, viewProjector)
	eventBus.Subscribe(viewsPL.DefaultViewChanged, viewProjector)
	eventBus.Subscribe(viewsPL.ViewVisibilityChanged, viewProjector)

	componentDeletedHandler := handlers.NewApplicationComponentDeletedHandler(commandBus, viewReadModel)
	relationDeletedHandler := handlers.NewComponentRelationDeletedHandler()

	eventBus.Subscribe(archPL.ApplicationComponentDeleted, componentDeletedHandler)
	eventBus.Subscribe(archPL.ComponentRelationDeleted, relationDeletedHandler)
}

func RegisterCommands(commandBus *cqrs.InMemoryCommandBus, eventStore eventstore.EventStore, db *database.TenantAwareDB, userRoleChecker ports.UserRoleChecker) {
	viewRepo := repositories.NewArchitectureViewRepository(eventStore)
	layoutRepo := repositories.NewViewLayoutRepository(db)
	viewReadModel := readmodels.NewArchitectureViewReadModel(db)

	commandBus.Register("CreateView", handlers.NewCreateViewHandler(viewRepo, viewReadModel))
	commandBus.Register("AddComponentToView", handlers.NewAddComponentToViewHandler(viewRepo, layoutRepo))
	commandBus.Register("UpdateComponentPosition", handlers.NewUpdateComponentPositionHandler(layoutRepo))
	commandBus.Register("UpdateMultiplePositions", handlers.NewUpdateMultiplePositionsHandler(layoutRepo))
	commandBus.Register("RenameView", handlers.NewRenameViewHandler(viewRepo))
	commandBus.Register("DeleteView", handlers.NewDeleteViewHandler(viewRepo))
	commandBus.Register("RemoveComponentFromView", handlers.NewRemoveComponentFromViewHandler(viewRepo))
	commandBus.Register("SetDefaultView", handlers.NewSetDefaultViewHandler(viewRepo, viewReadModel))
	commandBus.Register("UpdateViewEdgeType", handlers.NewUpdateViewEdgeTypeHandler(layoutRepo))
	commandBus.Register("UpdateViewLayoutDirection", handlers.NewUpdateViewLayoutDirectionHandler(layoutRepo))
	commandBus.Register("UpdateViewColorScheme", handlers.NewUpdateViewColorSchemeHandler(layoutRepo))
	commandBus.Register("UpdateElementColor", handlers.NewUpdateElementColorHandler(layoutRepo))
	commandBus.Register("ClearElementColor", handlers.NewClearElementColorHandler(layoutRepo))
	commandBus.Register("ChangeViewVisibility", handlers.NewChangeViewVisibilityHandler(viewRepo, userRoleChecker))
}

type HTTPHandlers struct {
	view      *ViewHandlers
	component *ViewComponentHandlers
	element   *ViewElementHandlers
	color     *ViewColorHandlers
}

func NewHTTPHandlers(commandBus *cqrs.InMemoryCommandBus, db *database.TenantAwareDB, hateoas *sharedAPI.HATEOASLinks) *HTTPHandlers {
	viewReadModel := readmodels.NewArchitectureViewReadModel(db)
	layoutRepo := repositories.NewViewLayoutRepository(db)
	links := NewViewLinks(hateoas)
	return &HTTPHandlers{
		view:      NewViewHandlers(commandBus, viewReadModel, links),
		component: NewViewComponentHandlers(commandBus, viewReadModel),
		element:   NewViewElementHandlers(layoutRepo, viewReadModel),
		color:     NewViewColorHandlers(commandBus, viewReadModel, links),
	}
}

func RegisterRoutes(r chi.Router, h *HTTPHandlers, authMiddleware AuthMiddleware) {
	r.Route("/views", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermViewsRead))
			r.Get("/", h.view.GetAllViews)
			r.Get("/{id}", h.view.GetViewByID)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermViewsWrite))
			r.Post("/", h.view.CreateView)
			r.Post("/{id}/components", h.component.AddComponentToView)
			r.Post("/{id}/capabilities", h.element.AddCapabilityToView)
			r.Post("/{id}/origin-entities", h.element.AddOriginEntityToView)
		})
		r.Group(func(r chi.Router) {
			r.Use(sharedAPI.RequireWriteOrEditGrant("views", "id"))
			r.Patch("/{id}/name", h.view.RenameView)
			r.Put("/{id}/default", h.view.SetDefaultView)
			r.Patch("/{id}/edge-type", h.view.UpdateEdgeType)
			r.Patch("/{id}/layout-direction", h.view.UpdateLayoutDirection)
			r.Patch("/{id}/color-scheme", h.color.UpdateColorScheme)
			r.Patch("/{id}/components/{componentId}/position", h.component.UpdateComponentPosition)
			r.Patch("/{id}/components/{componentId}/color", h.color.UpdateComponentColor)
			r.Patch("/{id}/layout", h.component.UpdateMultiplePositions)
			r.Patch("/{id}/capabilities/{capabilityId}/position", h.element.UpdateCapabilityPosition)
			r.Patch("/{id}/capabilities/{capabilityId}/color", h.color.UpdateCapabilityColor)
			r.Patch("/{id}/origin-entities/{originEntityId}/position", h.element.UpdateOriginEntityPosition)
			r.Patch("/{id}/visibility", h.view.ChangeVisibility)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermViewsDelete))
			r.Delete("/{id}", h.view.DeleteView)
			r.Delete("/{id}/components/{componentId}", h.component.RemoveComponentFromView)
			r.Delete("/{id}/components/{componentId}/color", h.color.ClearComponentColor)
			r.Delete("/{id}/capabilities/{capabilityId}", h.element.RemoveCapabilityFromView)
			r.Delete("/{id}/capabilities/{capabilityId}/color", h.color.ClearCapabilityColor)
			r.Delete("/{id}/origin-entities/{originEntityId}", h.element.RemoveOriginEntityFromView)
		})
	})
}
