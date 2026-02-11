package api

import (
	"easi/backend/internal/valuestreams/application/handlers"
	"easi/backend/internal/valuestreams/domain/valueobjects"
	"easi/backend/internal/valuestreams/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
)

func init() {
	registry := sharedAPI.GetErrorRegistry()

	registry.RegisterNotFound(repositories.ErrValueStreamNotFound, "Value stream not found")
	registry.RegisterNotFound(handlers.ErrValueStreamNotFound, "Value stream not found")

	registry.RegisterValidation(valueobjects.ErrValueStreamNameEmpty, "Value stream name cannot be empty")
	registry.RegisterValidation(valueobjects.ErrValueStreamNameTooLong, "Value stream name cannot exceed 100 characters")
	registry.RegisterValidation(valueobjects.ErrDescriptionTooLong, "Description cannot exceed 500 characters")

	registry.RegisterConflict(handlers.ErrValueStreamNameExists, "Value stream with this name already exists")
}
