package api

import (
	"easi/backend/internal/architecturedirection/application/handlers"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	"easi/backend/internal/architecturedirection/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
)

func init() {
	registry := sharedAPI.GetErrorRegistry()

	registry.RegisterNotFound(repositories.ErrEnterpriseCapabilityNotFound, "Enterprise capability not found")
	registry.RegisterNotFound(repositories.ErrEnterpriseStrategicImportanceNotFound, "Strategic importance rating not found")

	registry.RegisterConflict(handlers.ErrEnterpriseCapabilityNameExists, "Enterprise capability with this name already exists")
	registry.RegisterConflict(handlers.ErrImportanceAlreadySet, "Strategic importance for this pillar is already set")

	registry.RegisterValidation(valueobjects.ErrEnterpriseCapabilityNameEmpty, "Enterprise capability name cannot be empty")
	registry.RegisterValidation(valueobjects.ErrEnterpriseCapabilityNameTooLong, "Enterprise capability name cannot exceed 200 characters")
	registry.RegisterValidation(valueobjects.ErrDescriptionTooLong, "Description exceeds maximum length of 1000 characters")
	registry.RegisterValidation(valueobjects.ErrCategoryTooLong, "Category cannot exceed 100 characters")
	registry.RegisterValidation(valueobjects.ErrImportanceOutOfRange, "Importance must be between 1 and 5")
	registry.RegisterValidation(valueobjects.ErrRationaleTooLong, "Rationale cannot exceed 2000 characters")
	registry.RegisterValidation(valueobjects.ErrTargetMaturityOutOfRange, "Target maturity must be between 0 and 99")
}
