package api

import (
	"easi/backend/internal/valuestreams/application/handlers"
	"easi/backend/internal/valuestreams/domain/aggregates"
	"easi/backend/internal/valuestreams/domain/valueobjects"
	"easi/backend/internal/valuestreams/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
)

func init() {
	registry := sharedAPI.GetErrorRegistry()

	registry.RegisterNotFound(repositories.ErrValueStreamNotFound, "Value stream not found")
	registry.RegisterNotFound(handlers.ErrValueStreamNotFound, "Value stream not found")
	registry.RegisterNotFound(handlers.ErrStageNotFound, "Stage not found")
	registry.RegisterNotFound(handlers.ErrCapabilityNotFound, "Capability not found")
	registry.RegisterNotFound(aggregates.ErrStageNotFound, "Stage not found")

	registry.RegisterValidation(valueobjects.ErrValueStreamNameEmpty, "Value stream name cannot be empty")
	registry.RegisterValidation(valueobjects.ErrValueStreamNameTooLong, "Value stream name cannot exceed 100 characters")
	registry.RegisterValidation(valueobjects.ErrDescriptionTooLong, "Description cannot exceed 500 characters")
	registry.RegisterValidation(valueobjects.ErrStageNameEmpty, "Stage name cannot be empty")
	registry.RegisterValidation(valueobjects.ErrStageNameTooLong, "Stage name cannot exceed 100 characters")
	registry.RegisterValidation(valueobjects.ErrStagePositionInvalid, "Stage position must be a positive integer")
	registry.RegisterValidation(aggregates.ErrInvalidStagePositions, "Invalid stage positions")

	registry.RegisterConflict(handlers.ErrValueStreamNameExists, "Value stream with this name already exists")
	registry.RegisterConflict(handlers.ErrStageNameExists, "Stage with this name already exists in this value stream")
	registry.RegisterConflict(aggregates.ErrStageNameExists, "Stage with this name already exists in this value stream")
	registry.RegisterConflict(aggregates.ErrCapabilityAlreadyMapped, "Capability is already mapped to this stage")
}
