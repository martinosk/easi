package api

import (
	"context"

	capReadModels "easi/backend/internal/capabilitymapping/application/readmodels"
	eaProjectors "easi/backend/internal/enterprisearchitecture/application/projectors"
	enterpriseArchAPI "easi/backend/internal/enterprisearchitecture/infrastructure/api"

	"github.com/go-chi/chi/v5"
)

func businessDomainNameLookup(readModel *capReadModels.BusinessDomainReadModel) eaProjectors.BusinessDomainNameLookup {
	return func(ctx context.Context, businessDomainID string) (string, error) {
		domain, err := readModel.GetByID(ctx, businessDomainID)
		if err != nil {
			return "", err
		}
		if domain == nil {
			return "", nil
		}
		return domain.Name, nil
	}
}

func setupEnterpriseArchitectureRoutes(r chi.Router, deps routerDependencies) {
	mustSetup(enterpriseArchAPI.SetupEnterpriseArchitectureRoutes(enterpriseArchAPI.EnterpriseArchRoutesDeps{
		Router:              r,
		CommandBus:          deps.commandBus,
		EventStore:          deps.eventStore,
		EventBus:            deps.eventBus,
		DB:                  deps.db,
		AuthMiddleware:      deps.authDeps.AuthMiddleware,
		SessionProvider:     deps.authDeps.SessionManager,
		BusinessDomainNames: businessDomainNameLookup(capReadModels.NewBusinessDomainReadModel(deps.db)),
	}), "enterprise architecture routes")
}
