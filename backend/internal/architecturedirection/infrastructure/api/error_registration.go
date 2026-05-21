package api

import (
	"easi/backend/internal/architecturedirection/application/handlers"
	"easi/backend/internal/architecturedirection/domain/aggregates"
	"easi/backend/internal/architecturedirection/domain/services"
	"easi/backend/internal/architecturedirection/domain/valueobjects"
	"easi/backend/internal/architecturedirection/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
	sharedvo "easi/backend/internal/shared/eventsourcing/valueobjects"
)

func init() {
	registry := sharedAPI.GetErrorRegistry()

	registry.RegisterNotFound(repositories.ErrDirectionNotFound, "Direction not found")
	registry.RegisterNotFound(services.ErrReferencedEntityNotFound, "A referenced entity does not exist or is not accessible")

	registry.RegisterConflict(services.ErrActiveDirectionAlreadyExists, "An active direction already exists on this enterprise capability")
	registry.RegisterConflict(aggregates.ErrDirectionAgreedImmutable, "Agreed directions are immutable; reject and replace to change")
	registry.RegisterConflict(aggregates.ErrInvalidStatusTransition, "Status transition not allowed from current status")

	registry.RegisterValidation(aggregates.ErrInvalidSourceCardinality, "Source capability count does not match the direction type")
	registry.RegisterValidation(aggregates.ErrInvalidPlacementCardinality, "Placement count does not match the direction type")
	registry.RegisterValidation(aggregates.ErrDuplicateSourceCapabilities, "Source capabilities must be unique")
	registry.RegisterValidation(aggregates.ErrNarrativeRequiredToPropose, "A narrative is required before advancing a direction to proposed")
	registry.RegisterValidation(valueobjects.ErrInvalidDirectionType, "Direction type must be one of consolidate, decompose, stay")
	registry.RegisterValidation(valueobjects.ErrInvalidDirectionStatus, "Direction status must be one of draft, proposed, agreed, rejected")
	registry.RegisterValidation(valueobjects.ErrInvalidHorizon, "Horizon must be one of now, next, later")
	registry.RegisterValidation(sharedvo.ErrDescriptionTooLong, "Narrative cannot exceed 1000 characters")
	registry.RegisterValidation(valueobjects.ErrResultingNameTooLong, "Resulting name cannot exceed 200 characters")
	registry.RegisterValidation(handlers.ErrUnknownAdvanceTarget, "Advance target must be 'proposed' or 'agreed'")
}
