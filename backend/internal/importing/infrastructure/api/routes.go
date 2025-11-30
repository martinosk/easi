package api

import (
	"easi/backend/internal/importing/application/handlers"
	"easi/backend/internal/importing/application/orchestrator"
	"easi/backend/internal/importing/application/projectors"
	"easi/backend/internal/importing/application/readmodels"
	"easi/backend/internal/importing/infrastructure/repositories"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

func SetupImportingRoutes(
	r chi.Router,
	commandBus *cqrs.InMemoryCommandBus,
	eventStore eventstore.EventStore,
	eventBus events.EventBus,
	db *database.TenantAwareDB,
) error {
	repository := repositories.NewImportSessionRepository(eventStore)

	readModel := readmodels.NewImportSessionReadModel(db)

	projector := projectors.NewImportSessionProjector(readModel)
	eventBus.Subscribe("ImportSessionCreated", projector)
	eventBus.Subscribe("ImportStarted", projector)
	eventBus.Subscribe("ImportProgressUpdated", projector)
	eventBus.Subscribe("ImportCompleted", projector)
	eventBus.Subscribe("ImportFailed", projector)
	eventBus.Subscribe("ImportSessionCancelled", projector)

	importOrchestrator := orchestrator.NewImportOrchestrator(commandBus, repository)

	createHandler := handlers.NewCreateImportSessionHandler(repository)
	confirmHandler := handlers.NewConfirmImportHandler(repository, importOrchestrator)
	cancelHandler := handlers.NewCancelImportHandler(repository)

	commandBus.Register("CreateImportSession", createHandler)
	commandBus.Register("ConfirmImport", confirmHandler)
	commandBus.Register("CancelImport", cancelHandler)

	importHandlers := NewImportHandlers(commandBus, readModel)

	r.Route("/imports", func(r chi.Router) {
		r.Post("/", importHandlers.CreateImportSession)
		r.Get("/{id}", importHandlers.GetImportSession)
		r.Post("/{id}/confirm", importHandlers.ConfirmImport)
		r.Delete("/{id}", importHandlers.DeleteImportSession)
	})

	return nil
}
