package api

import (
	"easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
	"easi/backend/internal/capabilitymapping/domain/services"
	"easi/backend/internal/capabilitymapping/domain/valueobjects"
	"easi/backend/internal/capabilitymapping/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
)

func init() {
	registry := sharedAPI.GetErrorRegistry()

	registry.RegisterNotFound(repositories.ErrCapabilityNotFound, "Capability not found")
	registry.RegisterNotFound(repositories.ErrBusinessDomainNotFound, "Business domain not found")
	registry.RegisterNotFound(repositories.ErrAssignmentNotFound, "Assignment not found")
	registry.RegisterNotFound(repositories.ErrRealizationNotFound, "Realization not found")
	registry.RegisterNotFound(repositories.ErrDependencyNotFound, "Dependency not found")

	registry.RegisterNotFound(handlers.ErrCapabilityNotFound, "Capability not found")
	registry.RegisterNotFound(handlers.ErrBusinessDomainNotFound, "Business domain not found")
	registry.RegisterNotFound(handlers.ErrParentCapabilityNotFound, "Parent capability not found")
	registry.RegisterNotFound(handlers.ErrComponentNotFound, "Component not found")
	registry.RegisterNotFound(handlers.ErrCapabilityNotFoundForRealization, "Capability not found")
	registry.RegisterNotFound(handlers.ErrSourceCapabilityNotFound, "Source capability not found")
	registry.RegisterNotFound(handlers.ErrTargetCapabilityNotFound, "Target capability not found")

	registry.RegisterValidation(valueobjects.ErrDomainNameEmpty, "Domain name cannot be empty")
	registry.RegisterValidation(valueobjects.ErrDomainNameTooLong, "Domain name cannot exceed 100 characters")
	registry.RegisterValidation(valueobjects.ErrCapabilityNameEmpty, "Capability name cannot be empty")

	registry.RegisterConflict(aggregates.ErrOnlyL1CanBeAssignedToDomain, "Only L1 capabilities can be assigned to business domains")
	registry.RegisterConflict(handlers.ErrAssignmentAlreadyExists, "Capability is already assigned to this domain")
	registry.RegisterConflict(handlers.ErrBusinessDomainNameExists, "Business domain with this name already exists")
	registry.RegisterConflict(services.ErrBusinessDomainHasAssignments, "Cannot delete domain with assigned capabilities")
	registry.RegisterConflict(services.ErrCapabilityHasChildren, "Cannot delete capability with children")

	registry.RegisterConflict(aggregates.ErrL1CannotHaveParent, "L1 capabilities cannot have a parent")
	registry.RegisterConflict(aggregates.ErrNonL1MustHaveParent, "L2-L4 capabilities must have a parent")
	registry.RegisterConflict(aggregates.ErrParentMustBeOneLevelAbove, "Parent must be exactly one level above")
	registry.RegisterConflict(aggregates.ErrCapabilityCannotBeOwnParent, "Capability cannot be its own parent")
	registry.RegisterConflict(aggregates.ErrWouldCreateCircularReference, "Operation would create circular reference")
	registry.RegisterConflict(aggregates.ErrWouldExceedMaximumDepth, "Operation would create L5+ hierarchy")
	registry.RegisterConflict(aggregates.ErrCannotCreateSelfDependency, "Cannot create self-dependency")

	registry.RegisterNotFound(repositories.ErrApplicationFitScoreNotFound, "Application fit score not found")
	registry.RegisterNotFound(handlers.ErrFitScoreNotFound, "Application fit score not found")
	registry.RegisterNotFound(handlers.ErrPillarNotFound, "Strategy pillar not found or inactive")

	registry.RegisterValidation(handlers.ErrInvalidFitScoreValue, "Fit score must be between 1 and 5")
	registry.RegisterValidation(handlers.ErrPillarFitScoringDisabled, "Fit scoring is not enabled for this pillar")
	registry.RegisterValidation(valueobjects.ErrFitScoreOutOfRange, "Fit score must be between 1 and 5")
	registry.RegisterValidation(valueobjects.ErrFitRationaleTooLong, "Fit rationale cannot exceed 500 characters")

	registry.RegisterConflict(handlers.ErrFitScoreAlreadyExists, "Fit score already exists for this component and pillar")

	registry.RegisterNotFound(repositories.ErrStrategyImportanceNotFound, "Strategy importance not found")
	registry.RegisterNotFound(handlers.ErrImportanceNotFound, "Strategy importance not found")

	registry.RegisterValidation(handlers.ErrInvalidImportanceValue, "Importance must be between 1 and 5")
	registry.RegisterValidation(valueobjects.ErrImportanceOutOfRange, "Importance must be between 1 and 5")
	registry.RegisterValidation(valueobjects.ErrRationaleTooLong, "Rationale cannot exceed 500 characters")

	registry.RegisterConflict(handlers.ErrImportanceAlreadyExists, "Importance rating already exists for this combination")
}
