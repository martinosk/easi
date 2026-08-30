package api

import (
	"context"

	directionServices "easi/backend/internal/architecturedirection/domain/services"
	directionAPI "easi/backend/internal/architecturedirection/infrastructure/api"
	archReadModels "easi/backend/internal/architecturemodeling/application/readmodels"
	capReadModels "easi/backend/internal/capabilitymapping/application/readmodels"

	"github.com/go-chi/chi/v5"
)

func setupArchitectureDirectionRoutes(r chi.Router, deps routerDependencies) {
	capabilities := capReadModels.NewCapabilityReadModel(deps.db)
	mustSetup(directionAPI.SetupRoutes(directionAPI.RoutesDeps{
		Router:                        r,
		CommandBus:                    deps.commandBus,
		EventStore:                    deps.eventStore,
		EventBus:                      deps.eventBus,
		DB:                            deps.db,
		HATEOAS:                       deps.hateoas,
		AuthMiddleware:                deps.authDeps.AuthMiddleware,
		PhysicalCapabilityExists:      existsByID(capabilities.GetByID),
		DirectRealization:             directionServices.DirectRealizationLookup(capReadModels.NewRealizationReadModel(deps.db).GetDirectByCapabilityAndComponent),
		CapabilityExists:              directionServices.CapabilityExists(existsByID(capabilities.GetByID)),
		ComponentExists:               directionServices.ComponentExists(existsByID(archReadModels.NewApplicationComponentReadModel(deps.db).GetByID)),
		DomainExists:                  directionServices.DomainExists(existsByID(capReadModels.NewBusinessDomainReadModel(deps.db).GetByID)),
		CapabilityEffectivelyInDomain: capabilityEffectivelyInDomain(capReadModels.NewCMEffectiveBusinessDomainReadModel(deps.db)),
	}), "architecture direction routes")
}

func existsByID[T any](getByID func(context.Context, string) (*T, error)) directionServices.ExistenceCheck {
	return func(ctx context.Context, id string) (bool, error) {
		dto, err := getByID(ctx, id)
		if err != nil {
			return false, err
		}
		return dto != nil, nil
	}
}

func capabilityEffectivelyInDomain(readModel *capReadModels.CMEffectiveBusinessDomainReadModel) directionServices.CapabilityEffectivelyInDomain {
	return func(ctx context.Context, capabilityID, domainID string) (bool, error) {
		effective, err := readModel.GetByCapabilityID(ctx, capabilityID)
		if err != nil {
			return false, err
		}
		if effective == nil {
			return false, nil
		}
		return effective.BusinessDomainID == domainID, nil
	}
}
