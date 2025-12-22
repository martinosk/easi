package api

import (
	"easi/backend/internal/architecturemodeling/domain/aggregates"
	"easi/backend/internal/architecturemodeling/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
)

func init() {
	registry := sharedAPI.GetErrorRegistry()

	registry.RegisterNotFound(repositories.ErrComponentNotFound, "Component not found")
	registry.RegisterNotFound(repositories.ErrRelationNotFound, "Relation not found")
	registry.RegisterConflict(aggregates.ErrSelfReference, "Component cannot have a relation to itself")
}
