package api

import (
	"easi/backend/internal/capabilitymapping/application/handlers"
	"easi/backend/internal/capabilitymapping/domain/aggregates"
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

	registry.RegisterConflict(handlers.ErrOnlyL1CapabilitiesCanBeAssigned, "Only L1 capabilities can be assigned to business domains")
	registry.RegisterConflict(handlers.ErrAssignmentAlreadyExists, "Capability is already assigned to this domain")
	registry.RegisterConflict(handlers.ErrBusinessDomainNameExists, "Business domain with this name already exists")
	registry.RegisterConflict(handlers.ErrBusinessDomainHasAssignments, "Cannot delete domain with assigned capabilities")
	registry.RegisterConflict(handlers.ErrCapabilityHasChildren, "Cannot delete capability with children")

	registry.RegisterConflict(aggregates.ErrL1CannotHaveParent, "L1 capabilities cannot have a parent")
	registry.RegisterConflict(aggregates.ErrNonL1MustHaveParent, "L2-L4 capabilities must have a parent")
	registry.RegisterConflict(aggregates.ErrParentMustBeOneLevelAbove, "Parent must be exactly one level above")
	registry.RegisterConflict(aggregates.ErrCapabilityCannotBeOwnParent, "Capability cannot be its own parent")
	registry.RegisterConflict(aggregates.ErrWouldCreateCircularReference, "Operation would create circular reference")
	registry.RegisterConflict(aggregates.ErrWouldExceedMaximumDepth, "Operation would create L5+ hierarchy")
	registry.RegisterConflict(aggregates.ErrCannotCreateSelfDependency, "Cannot create self-dependency")
}
