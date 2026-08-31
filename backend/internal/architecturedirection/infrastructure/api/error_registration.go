package api

import (
	"easi/backend/internal/architecturedirection/application/handlers"
	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/entities"
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	"easi/backend/internal/architecturedirection/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

func init() {
	registry := sharedAPI.GetErrorRegistry()

	registry.RegisterNotFound(repositories.ErrTimeAssessmentNotFound, "Time assessment not found")
	registry.RegisterNotFound(handlers.ErrTimeAssessmentNotFoundForPair, "No time assessment exists for this capability and component pair")
	registry.RegisterNotFound(repositories.ErrDirectionNotFound, "Direction not found")
	registry.RegisterNotFound(repositories.ErrStandardApplicationNotFound, "Standard application not found")
	registry.RegisterNotFound(services.ErrReferencedEntityNotFound, "A referenced entity does not exist or is not accessible")
	registry.RegisterNotFound(aggregates.ErrSourceCapabilityNotInDirection, "Capability is not a source of this direction")
	registry.RegisterNotFound(aggregates.ErrInvalidStatusTransition, "No active direction in a status that allows this transition")
	registry.RegisterNotFound(repositories.ErrRealizationRolesNotFound, "Realization roles not found")
	registry.RegisterNotFound(handlers.ErrRealizationRoleNotFoundForPair, "No realization role exists for this capability and component pair")
	registry.RegisterNotFound(aggregates.ErrNoRoleToClear, "No realization role exists for this capability and component pair")
	registry.RegisterNotFound(repositories.ErrCapabilityJourneyNotFound, "Capability journey not found")
	registry.RegisterNotFound(aggregates.ErrInvalidJourneyTransition, "No journey in a status that allows this transition")
	registry.RegisterNotFound(aggregates.ErrJourneyMilestoneNotFound, "Milestone not found on this journey")

	registry.RegisterConflict(readmodels.ErrTimeAssessmentAlreadyExists, "A time assessment already exists for this capability and component pair")
	registry.RegisterConflict(aggregates.ErrTimeAssessmentAlreadyRemoved, "This time assessment has already been removed")
	registry.RegisterConflict(services.ErrActiveDirectionAlreadyExists, "An active direction already exists on this enterprise capability")
	registry.RegisterConflict(services.ErrEnterpriseCapabilityInactive, "Directions can only be captured on active enterprise capabilities.")
	registry.RegisterConflict(aggregates.ErrDirectionAgreedImmutable, "Agreed directions are immutable; reject and replace to change")
	registry.RegisterConflict(readmodels.ErrStandardApplicationAlreadyExists, "A standard application already exists for this enterprise capability")
	registry.RegisterConflict(readmodels.ErrRealizationRolesAggregateConflict, "A different realization roles aggregate is already registered for this capability")
	registry.RegisterConflict(aggregates.ErrJourneyFrozen, "This journey is terminal and can no longer be edited")
	registry.RegisterConflict(aggregates.ErrJourneyMilestoneOrderUnchanged, "The milestone order is unchanged")
	registry.RegisterConflict(readmodels.ErrActiveCapabilityJourneyExists, "An active journey already exists for this capability")

	registry.RegisterValidation(valueobjects.ErrInvalidTimeGrade, "Grade must be one of Invest, Tolerate, Migrate, Eliminate")
	registry.RegisterValidation(valueobjects.ErrInvalidRealizationRole, "Role must be one of standard, legacy")
	registry.RegisterValidation(aggregates.ErrNarrativeRequiredForStandardApplication, "A narrative is required when setting or changing the standard application")
	registry.RegisterValidation(aggregates.ErrInvalidSourceCardinality, "Source capability count does not match the direction type")
	registry.RegisterValidation(aggregates.ErrDuplicateSourceCapabilities, "Source capabilities must be unique")
	registry.RegisterValidation(aggregates.ErrNarrativeRequiredToPropose, "A narrative is required before advancing a direction to proposed")
	registry.RegisterValidation(valueobjects.ErrInvalidDirectionType, "Direction type must be one of consolidate, decompose, stay")
	registry.RegisterValidation(valueobjects.ErrInvalidDirectionStatus, "Direction status must be one of draft, proposed, agreed, rejected")
	registry.RegisterValidation(valueobjects.ErrInvalidHorizon, "Horizon must be one of now, next, later")
	registry.RegisterValidation(sharedvo.ErrDescriptionTooLong, "Narrative cannot exceed 1000 characters")
	registry.RegisterValidation(valueobjects.ErrResultingNameTooLong, "Resulting name cannot exceed 200 characters")
	registry.RegisterValidation(handlers.ErrUnknownAdvanceTarget, "Advance target must be 'proposed' or 'agreed'")

	registry.RegisterValidation(aggregates.ErrJourneyTargetAmongSources, "The target application must not be among the from-applications")
	registry.RegisterValidation(aggregates.ErrJourneyMoveRequiresTargetDomain, "A move journey requires a target business domain")
	registry.RegisterValidation(services.ErrTargetParentNotInTargetDomain, "The target parent capability must belong to the target business domain")
	registry.RegisterValidation(aggregates.ErrJourneyMoveFieldsOnNonMove, "Move fields are only valid for move journeys")
	registry.RegisterValidation(valueobjects.ErrInvalidJourneyKind, "Kind must be one of migration, consolidation, carve-out, move")
	registry.RegisterValidation(valueobjects.ErrInvalidSourceApplicationCount, "Source application count does not match the journey kind")
	registry.RegisterValidation(valueobjects.ErrInvalidJourneyStatus, "Journey status must be one of planned, in-flight, done, abandoned")
	registry.RegisterValidation(valueobjects.ErrInvalidMilestoneStatus, "Milestone status must be one of planned, in-flight, done")
	registry.RegisterValidation(valueobjects.ErrInvalidTargetPeriodYear, "Target period year must be between 2000 and 2100")
	registry.RegisterValidation(valueobjects.ErrInvalidTargetPeriodQuarter, "Target period quarter must be between 1 and 4")
	registry.RegisterValidation(valueobjects.ErrInvalidJourneyProgress, "Progress must be between 0 and 100")
	registry.RegisterValidation(valueobjects.ErrResultingCapabilityNameRequired, "Resulting name is required for move journeys")
	registry.RegisterValidation(valueobjects.ErrResultingCapabilityNameTooLong, "Resulting name cannot exceed 200 characters")
	registry.RegisterValidation(entities.ErrMilestoneLabelRequired, "Milestone label is required")
	registry.RegisterValidation(entities.ErrMilestoneLabelTooLong, "Milestone label cannot exceed 200 characters")
	registry.RegisterValidation(aggregates.ErrJourneyMilestoneOrderIncomplete, "The milestone order must list every milestone of the journey exactly once")
	registry.RegisterValidation(aggregates.ErrJourneyMilestoneOrderDuplicate, "The milestone order must not repeat a milestone")
	registry.RegisterValidation(handlers.ErrTargetPeriodRequiresBoth, "Target period requires both year and quarter, or neither")
}
