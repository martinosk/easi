package api

import (
	"easi/backend/internal/accessdelegation/domain/aggregates"
	"easi/backend/internal/accessdelegation/domain/valueobjects"
	"easi/backend/internal/accessdelegation/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
)

func init() {
	registry := sharedAPI.GetErrorRegistry()

	registry.RegisterNotFound(repositories.ErrEditGrantNotFound, "Edit grant not found")

	registry.RegisterValidation(aggregates.ErrCannotGrantToSelf, "Cannot grant edit access to yourself")
	registry.RegisterValidation(valueobjects.ErrInvalidArtifactType, "Invalid artifact type")
	registry.RegisterValidation(valueobjects.ErrEmptyArtifactID, "Artifact ID is required")
	registry.RegisterValidation(valueobjects.ErrInvalidGrantScope, "Invalid grant scope")

	registry.RegisterConflict(aggregates.ErrGrantAlreadyRevoked, "Edit grant has already been revoked")
	registry.RegisterConflict(aggregates.ErrGrantAlreadyExpired, "Edit grant has already expired")
	registry.RegisterConflict(aggregates.ErrGrantNotActive, "Edit grant is not active")
}
