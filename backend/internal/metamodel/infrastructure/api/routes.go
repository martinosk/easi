package api

import (
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/metamodel/application/handlers"
	"easi/backend/internal/metamodel/application/projectors"
	"easi/backend/internal/metamodel/application/readmodels"
	"easi/backend/internal/metamodel/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

func SetupMetaModelRoutes(
	r chi.Router,
	commandBus *cqrs.InMemoryCommandBus,
	eventStore eventstore.EventStore,
	eventBus events.EventBus,
	db *database.TenantAwareDB,
	hateoas *sharedAPI.HATEOASLinks,
) error {
	configRepo := repositories.NewMetaModelConfigurationRepository(eventStore)

	configReadModel := readmodels.NewMetaModelConfigurationReadModel(db)

	configProjector := projectors.NewMetaModelConfigurationProjector(configReadModel)

	eventBus.Subscribe("MetaModelConfigurationCreated", configProjector)
	eventBus.Subscribe("MaturityScaleConfigUpdated", configProjector)
	eventBus.Subscribe("MaturityScaleConfigReset", configProjector)

	createConfigHandler := handlers.NewCreateMetaModelConfigurationHandler(configRepo)
	updateScaleHandler := handlers.NewUpdateMaturityScaleHandler(configRepo)
	resetScaleHandler := handlers.NewResetMaturityScaleHandler(configRepo)

	commandBus.Register("CreateMetaModelConfiguration", createConfigHandler)
	commandBus.Register("UpdateMaturityScale", updateScaleHandler)
	commandBus.Register("ResetMaturityScale", resetScaleHandler)

	tenantCreatedHandler := handlers.NewTenantCreatedHandler(commandBus)
	eventBus.Subscribe("TenantCreated", tenantCreatedHandler)

	metaModelHandlers := NewMetaModelHandlers(commandBus, configReadModel, hateoas)

	r.Route("/metamodel", func(r chi.Router) {
		r.Get("/maturity-scale", metaModelHandlers.GetMaturityScale)
		r.Put("/maturity-scale", metaModelHandlers.UpdateMaturityScale)
		r.Put("/maturity-scale/reset", metaModelHandlers.ResetMaturityScale)
		r.Get("/configurations/{id}", metaModelHandlers.GetMaturityScaleByID)
	})

	return nil
}
