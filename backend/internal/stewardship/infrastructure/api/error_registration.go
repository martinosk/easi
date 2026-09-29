package api

import (
	sharedAPI "easi/backend/internal/shared/api"
	"easi/backend/internal/stewardship/application/handlers"
	"easi/backend/internal/stewardship/domain/aggregates"
	"easi/backend/internal/stewardship/domain/valueobjects"
	"easi/backend/internal/stewardship/infrastructure/repositories"
)

func init() {
	registry := sharedAPI.GetErrorRegistry()

	registry.RegisterNotFound(handlers.ErrDomainNotFound, "Business domain not found")
	registry.RegisterNotFound(repositories.ErrStewardshipNotFound, "Stewardship not found")

	registry.RegisterValidation(valueobjects.ErrUnknownConcern, "Concern must be one of ownership, assessment, documentation, planning, structure")
	registry.RegisterValidation(valueobjects.ErrStewardIDRequired, "Steward user id is required")
	registry.RegisterValidation(valueobjects.ErrStewardNotFound, "Steward is not a user of this tenant")
	registry.RegisterValidation(valueobjects.ErrStewardDisabled, "Steward must be an active user")
	registry.RegisterValidation(aggregates.ErrActorRequired, "Stewardship changes must be attributed to an actor")

	registry.RegisterConflict(aggregates.ErrStewardshipReleased, "Stewardship has been released")
}
