package api

import (
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/domain/valueobjects"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
)

func init() {
	registry := sharedAPI.GetErrorRegistry()

	registry.RegisterNotFound(repositories.ErrComponentNotFound, "Component not found")
	registry.RegisterNotFound(repositories.ErrRelationNotFound, "Relation not found")
	registry.RegisterNotFound(repositories.ErrAcquiredEntityNotFound, "Acquired entity not found")
	registry.RegisterNotFound(repositories.ErrVendorNotFound, "Vendor not found")
	registry.RegisterNotFound(repositories.ErrInternalTeamNotFound, "Internal team not found")
	registry.RegisterNotFound(repositories.ErrComponentOriginsNotFound, "Component origins not found")

	registry.RegisterConflict(aggregates.ErrSelfReference, "Component cannot have a relation to itself")

	registry.RegisterValidation(valueobjects.ErrEntityNameEmpty, "Name cannot be empty")
	registry.RegisterValidation(valueobjects.ErrEntityNameTooLong, "Name exceeds maximum length of 100 characters")
	registry.RegisterValidation(valueobjects.ErrNotesTooLong, "Notes exceeds maximum length of 500 characters")
	registry.RegisterValidation(valueobjects.ErrInvalidIntegrationStatus, "Invalid integration status")
}
