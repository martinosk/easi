package api

import (
	"context"

	"easi/backend/internal/importing/application/handlers"
	"easi/backend/internal/importing/application/ports"
	"easi/backend/internal/importing/application/projectors"
	"easi/backend/internal/importing/application/readmodels"
	"easi/backend/internal/importing/application/saga"
	"easi/backend/internal/importing/infrastructure/repositories"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

type ImportingRoutesDeps struct {
	CommandBus         *cqrs.InMemoryCommandBus
	EventStore         eventstore.EventStore
	EventBus           events.EventBus
	DB                 *database.TenantAwareDB
	ComponentGateway   ports.ComponentGateway
	CapabilityGateway  ports.CapabilityGateway
	ValueStreamGateway ports.ValueStreamGateway
	ExecutionContext   context.Context
}

func SetupImportingRoutes(r chi.Router, deps ImportingRoutesDeps) error {
	repository := repositories.NewImportSessionRepository(deps.EventStore)

	readModel := readmodels.NewImportSessionReadModel(deps.DB)

	projector := projectors.NewImportSessionProjector(readModel)
	deps.EventBus.Subscribe("ImportSessionCreated", projector)
	deps.EventBus.Subscribe("ImportStarted", projector)
	deps.EventBus.Subscribe("ImportProgressUpdated", projector)
	deps.EventBus.Subscribe("ImportCompleted", projector)
	deps.EventBus.Subscribe("ImportFailed", projector)
	deps.EventBus.Subscribe("ImportSessionCancelled", projector)

	importSaga := saga.New(deps.ComponentGateway, deps.CapabilityGateway, deps.ValueStreamGateway)

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

	r.Route("/imports", func(r chi.Router) {
		r.Post("/", importHandlers.CreateImportSession)
		r.Get("/{id}", importHandlers.GetImportSession)
		r.Post("/{id}/confirm", importHandlers.ConfirmImport)
		r.Delete("/{id}", importHandlers.DeleteImportSession)
	})

	return nil
}
