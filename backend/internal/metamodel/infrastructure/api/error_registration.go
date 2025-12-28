package api

import (
	"easi/backend/internal/metamodel/domain/valueobjects"
	"easi/backend/internal/metamodel/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
)

func init() {
	registry := sharedAPI.GetErrorRegistry()

	registry.RegisterNotFound(repositories.ErrMetaModelConfigurationNotFound, "Meta model configuration not found")

	registry.RegisterConflict(valueobjects.ErrCannotRemoveLastActivePillar, "Cannot remove the last active pillar")
	registry.RegisterConflict(valueobjects.ErrPillarAlreadyInactive, "Pillar is already inactive")

	registry.RegisterValidation(valueobjects.ErrTooManyPillars, "Cannot have more than 20 pillars")
	registry.RegisterValidation(valueobjects.ErrPillarNameDuplicate, "Pillar name already exists")
	registry.RegisterValidation(valueobjects.ErrPillarNotFound, "Pillar not found")
}
