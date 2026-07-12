package api

import (
	"easi/backend/internal/architecturedirection/application/handlers"
	"easi/backend/internal/architecturedirection/application/readmodels"
	"easi/backend/internal/architecturedirection/domain/aggregates"
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

	registry.RegisterConflict(readmodels.ErrTimeAssessmentAlreadyExists, "A time assessment already exists for this capability and component pair")
	registry.RegisterConflict(aggregates.ErrTimeAssessmentAlreadyRemoved, "This time assessment has already been removed")
	registry.RegisterConflict(services.ErrActiveDirectionAlreadyExists, "An active direction already exists on this enterprise capability")
	registry.RegisterConflict(services.ErrEnterpriseCapabilityInactive, "Directions can only be captured on active enterprise capabilities.")
	registry.RegisterConflict(aggregates.ErrDirectionAgreedImmutable, "Agreed directions are immutable; reject and replace to change")
	registry.RegisterConflict(readmodels.ErrStandardApplicationAlreadyExists, "A standard application already exists for this enterprise capability")

	registry.RegisterValidation(valueobjects.ErrInvalidTimeGrade, "Grade must be one of Invest, Tolerate, Migrate, Eliminate")
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
}
