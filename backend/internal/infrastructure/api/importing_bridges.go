package api

import (
	archAdapters "easi/backend/internal/architecturemodeling/infrastructure/adapters"
	capAdapters "easi/backend/internal/capabilitymapping/infrastructure/adapters"
	importingAPI "easi/backend/internal/importing/infrastructure/api"
	vsAdapters "easi/backend/internal/valuestreams/infrastructure/adapters"
)

func importingRoutesDeps(deps routerDependencies) importingAPI.ImportingRoutesDeps {
	return importingAPI.ImportingRoutesDeps{
		CommandBus:         deps.commandBus,
		EventStore:         deps.eventStore,
		EventBus:           deps.eventBus,
		DB:                 deps.db,
		ComponentGateway:   archAdapters.NewImportComponentGateway(deps.commandBus),
		CapabilityGateway:  capAdapters.NewImportCapabilityGateway(deps.commandBus),
		ValueStreamGateway: vsAdapters.NewImportValueStreamGateway(deps.commandBus),
		ExecutionContext:   deps.appContext,
	}
}
