package api

import (
	"context"
	"net/http"

	authPL "easi/backend/internal/auth/publishedlanguage"
	"easi/backend/internal/importing/application/handlers"
	"easi/backend/internal/importing/application/projectors"
	"easi/backend/internal/importing/application/readmodels"
	"easi/backend/internal/importing/application/saga"
	"easi/backend/internal/importing/infrastructure/adapters"
	"easi/backend/internal/importing/infrastructure/repositories"
	importPL "easi/backend/internal/importing/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequirePermission(permission authPL.Permission) func(http.Handler) http.Handler
}

var importPermissions = []authPL.Permission{
	authPL.PermComponentsWrite,
	authPL.PermCapabilitiesWrite,
	authPL.PermValueStreamsWrite,
	authPL.PermDomainsWrite,
}

type ImportingRoutesDeps struct {
	AuthMiddleware   AuthMiddleware
	CommandBus       *cqrs.InMemoryCommandBus
	EventStore       eventstore.EventStore
	EventBus         events.EventBus
	DB               *database.TenantAwareDB
	ExecutionContext context.Context
}

func SetupImportingRoutes(r chi.Router, deps ImportingRoutesDeps) error {
	repository := repositories.NewImportSessionRepository(deps.EventStore)

	readModel := readmodels.NewImportSessionReadModel(deps.DB)

	projector := projectors.NewImportSessionProjector(readModel)
	deps.EventBus.Subscribe(importPL.ImportSessionCreated, projector)
	deps.EventBus.Subscribe(importPL.ImportStarted, projector)
	deps.EventBus.Subscribe(importPL.ImportProgressUpdated, projector)
	deps.EventBus.Subscribe(importPL.ImportCompleted, projector)
	deps.EventBus.Subscribe(importPL.ImportFailed, projector)
	deps.EventBus.Subscribe(importPL.ImportSessionCancelled, projector)

	importSaga := saga.New(
		adapters.NewImportComponentGateway(deps.CommandBus),
		adapters.NewImportCapabilityGateway(deps.CommandBus),
		adapters.NewImportValueStreamGateway(deps.CommandBus),
	)

	createHandler := handlers.NewCreateImportSessionHandler(repository)
	confirmHandler := handlers.NewConfirmImportHandlerWithExecutionContext(
		repository,
		importSaga,
		deps.ExecutionContext,
		handlers.DefaultImportExecutionTimeout,
	)
	cancelHandler := handlers.NewCancelImportHandler(repository)

	deps.CommandBus.Register("CreateImportSession", createHandler)
	deps.CommandBus.Register("ConfirmImport", confirmHandler)
	deps.CommandBus.Register("CancelImport", cancelHandler)

	importHandlers := NewImportHandlers(deps.CommandBus, readModel)

	registerRoutes(r, importHandlers, deps.AuthMiddleware)

	return nil
}

func registerRoutes(r chi.Router, importHandlers *ImportHandlers, authMiddleware AuthMiddleware) {
	r.Route("/imports", func(r chi.Router) {
		for _, permission := range importPermissions {
			r.Use(authMiddleware.RequirePermission(permission))
		}
		r.Post("/", importHandlers.CreateImportSession)
		r.Get("/{id}", importHandlers.GetImportSession)
		r.Post("/{id}/confirm", importHandlers.ConfirmImport)
		r.Delete("/{id}", importHandlers.DeleteImportSession)
	})
}
