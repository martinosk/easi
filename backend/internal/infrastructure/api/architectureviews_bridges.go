package api

import (
	viewsAPI "easi/backend/internal/architectureviews/infrastructure/api"
	authReadModels "easi/backend/internal/auth/application/readmodels"
	authAdapters "easi/backend/internal/auth/infrastructure/adapters"
)

func registerViewCommands(deps routerDependencies) {
	userRoleChecker := authAdapters.NewUserRoleCheckerAdapter(authReadModels.NewUserReadModel(deps.db))
	viewsAPI.RegisterCommands(deps.commandBus, deps.eventStore, deps.db, userRoleChecker)
}
