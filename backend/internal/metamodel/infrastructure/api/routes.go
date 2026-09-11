package api

import (
	"net/http"

	authPL "easi/backend/internal/auth/publishedlanguage"
	"easi/backend/internal/infrastructure/database"
	"easi/backend/internal/infrastructure/eventstore"
	"easi/backend/internal/metamodel/application/handlers"
	"easi/backend/internal/metamodel/application/projectors"
	"easi/backend/internal/metamodel/application/readmodels"
	mmevents "easi/backend/internal/metamodel/domain/events"
	"easi/backend/internal/metamodel/infrastructure/repositories"
	mmPL "easi/backend/internal/metamodel/publishedlanguage"
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/shared/cqrs"
	"easi/backend/internal/shared/events"

	"github.com/go-chi/chi/v5"
)

type AuthMiddleware interface {
	RequirePermission(permission authPL.Permission) func(http.Handler) http.Handler
}

type MetaModelRoutesDeps struct {
	Router          chi.Router
	CommandBus      *cqrs.InMemoryCommandBus
	EventStore      eventstore.EventStore
	EventBus        events.EventBus
	DB              *database.TenantAwareDB
	Hateoas         *sharedAPI.HATEOASLinks
	AuthMiddleware  AuthMiddleware
	SessionProvider authPL.SessionProvider
}

func SetupMetaModelRoutes(deps MetaModelRoutesDeps) error {
	configRepo := repositories.NewMetaModelConfigurationRepository(deps.EventStore)

	configReadModel := readmodels.NewMetaModelConfigurationReadModel(deps.DB)

	configProjector := projectors.NewMetaModelConfigurationProjector(configReadModel)

	deps.EventBus.Subscribe(mmPL.MetaModelConfigurationCreated, configProjector)
	deps.EventBus.Subscribe(mmPL.MaturityScaleConfigUpdated, configProjector)
	deps.EventBus.Subscribe(mmPL.MaturityScaleConfigReset, configProjector)
	deps.EventBus.Subscribe(mmPL.StrategyPillarAdded, configProjector)
	deps.EventBus.Subscribe(mmPL.StrategyPillarUpdated, configProjector)
	deps.EventBus.Subscribe(mmPL.StrategyPillarRemoved, configProjector)
	deps.EventBus.Subscribe(mmPL.PillarFitConfigurationUpdated, configProjector)

	createConfigHandler := handlers.NewCreateMetaModelConfigurationHandler(configRepo)
	updateScaleHandler := handlers.NewUpdateMaturityScaleHandler(configRepo)
	resetScaleHandler := handlers.NewResetMaturityScaleHandler(configRepo)

	deps.CommandBus.Register("CreateMetaModelConfiguration", createConfigHandler)
	deps.CommandBus.Register("UpdateMaturityScale", updateScaleHandler)
	deps.CommandBus.Register("ResetMaturityScale", resetScaleHandler)

	addPillarHandler := handlers.NewAddStrategyPillarHandler(configRepo)
	updatePillarHandler := handlers.NewUpdateStrategyPillarHandler(configRepo)
	removePillarHandler := handlers.NewRemoveStrategyPillarHandler(configRepo)
	batchUpdatePillarsHandler := handlers.NewBatchUpdateStrategyPillarsHandler(configRepo)
	updatePillarFitConfigHandler := handlers.NewUpdatePillarFitConfigurationHandler(configRepo)

	deps.CommandBus.Register("AddStrategyPillar", addPillarHandler)
	deps.CommandBus.Register("UpdateStrategyPillar", updatePillarHandler)
	deps.CommandBus.Register("RemoveStrategyPillar", removePillarHandler)
	deps.CommandBus.Register("BatchUpdateStrategyPillars", batchUpdatePillarsHandler)
	deps.CommandBus.Register("UpdatePillarFitConfiguration", updatePillarFitConfigHandler)

	tenantCreatedHandler := handlers.NewTenantCreatedHandler(deps.CommandBus)
	deps.EventBus.Subscribe(authPL.TenantCreated, tenantCreatedHandler)

	links := NewMetaModelLinks(deps.Hateoas)
	metaModelHandlers := NewMetaModelHandlers(deps.CommandBus, configReadModel, links, deps.SessionProvider)
	strategyPillarsHandlers := NewStrategyPillarsHandlers(deps.CommandBus, configReadModel, links, deps.SessionProvider)
	attributeHandlers := setupSubjectAttributeSchema(deps, links)

	deps.Router.Route("/meta-model", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(deps.AuthMiddleware.RequirePermission(authPL.PermMetaModelRead))
			r.Get("/maturity-scale", metaModelHandlers.GetMaturityScale)
			r.Get("/configurations/{id}", metaModelHandlers.GetMaturityScaleByID)
			r.Get("/strategy-pillars", strategyPillarsHandlers.GetStrategyPillars)
			r.Get("/strategy-pillars/{id}", strategyPillarsHandlers.GetStrategyPillarByID)
		})

		r.Group(func(r chi.Router) {
			r.Use(deps.AuthMiddleware.RequirePermission(authPL.PermMetaModelWrite))
			r.Put("/maturity-scale", metaModelHandlers.UpdateMaturityScale)
			r.Post("/maturity-scale/reset", metaModelHandlers.ResetMaturityScale)
			r.Patch("/strategy-pillars", strategyPillarsHandlers.BatchUpdateStrategyPillars)
			r.Post("/strategy-pillars", strategyPillarsHandlers.CreateStrategyPillar)
			r.Put("/strategy-pillars/{id}", strategyPillarsHandlers.UpdateStrategyPillar)
			r.Put("/strategy-pillars/{id}/fit-configuration", strategyPillarsHandlers.UpdatePillarFitConfiguration)
			r.Delete("/strategy-pillars/{id}", strategyPillarsHandlers.DeleteStrategyPillar)
		})

		registerSubjectAttributeRoutes(r, attributeHandlers, deps.AuthMiddleware)
	})

	return nil
}

func setupSubjectAttributeSchema(deps MetaModelRoutesDeps, links *MetaModelLinks) *SubjectAttributeHandlers {
	schemaRepo := repositories.NewSubjectAttributeSchemaRepository(deps.EventStore)
	schemaReadModel := readmodels.NewSubjectAttributeSchemaReadModel(deps.DB)

	schemaProjector := projectors.NewSubjectAttributeSchemaProjector(schemaReadModel)
	for _, eventType := range mmevents.SubjectAttributeSchemaEventTypes() {
		deps.EventBus.Subscribe(eventType, schemaProjector)
	}

	deps.CommandBus.Register("CreateSubjectAttributeSchema", handlers.NewCreateSubjectAttributeSchemaHandler(schemaRepo, schemaReadModel))
	deps.CommandBus.Register("DefineSubjectAttribute", handlers.NewDefineSubjectAttributeHandler(schemaRepo))
	deps.CommandBus.Register("RenameSubjectAttribute", handlers.NewRenameSubjectAttributeHandler(schemaRepo))
	deps.CommandBus.Register("RetireSubjectAttribute", handlers.NewRetireSubjectAttributeHandler(schemaRepo))
	deps.CommandBus.Register("ReactivateSubjectAttribute", handlers.NewReactivateSubjectAttributeHandler(schemaRepo))
	deps.CommandBus.Register("AddSubjectAttributeOption", handlers.NewAddSubjectAttributeOptionHandler(schemaRepo))
	deps.CommandBus.Register("RetireSubjectAttributeOption", handlers.NewRetireSubjectAttributeOptionHandler(schemaRepo))
	deps.CommandBus.Register("SetSubjectAttributeBounds", handlers.NewSetSubjectAttributeBoundsHandler(schemaRepo))
	deps.CommandBus.Register("ImportSubjectAttribute", handlers.NewImportSubjectAttributeHandler(schemaRepo, schemaReadModel))

	return NewSubjectAttributeHandlers(deps.CommandBus, schemaReadModel, links, deps.SessionProvider)
}

func registerSubjectAttributeRoutes(r chi.Router, h *SubjectAttributeHandlers, authMiddleware AuthMiddleware) {
	r.Route("/subject-types/{subjectType}/attributes", func(r chi.Router) {
		r.With(authMiddleware.RequirePermission(authPL.PermMetaModelRead)).Get("/", h.GetSchema)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.RequirePermission(authPL.PermMetaModelWrite))
			r.Post("/", h.DefineAttribute)
			r.Put("/{attributeID}", h.RenameAttribute)
			r.Post("/{attributeID}/retire", h.RetireAttribute)
			r.Post("/{attributeID}/reactivate", h.ReactivateAttribute)
			r.Post("/{attributeID}/options", h.AddOption)
			r.Post("/{attributeID}/options/{optionID}/retire", h.RetireOption)
			r.Put("/{attributeID}/bounds", h.SetBounds)
		})
	})
}
