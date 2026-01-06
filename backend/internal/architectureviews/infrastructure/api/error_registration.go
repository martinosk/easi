package api

import (
	"easi/backend/internal/architectureviews/application/handlers"
	"easi/backend/internal/architectureviews/domain/aggregates"
	"easi/backend/internal/architectureviews/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
)

func init() {
	registry := sharedAPI.GetErrorRegistry()

	registry.RegisterNotFound(repositories.ErrViewNotFound, "View not found")
	registry.RegisterNotFound(aggregates.ErrComponentNotFound, "Component not found in view")
	registry.RegisterConflict(aggregates.ErrComponentAlreadyInView, "Component already exists in view")
	registry.RegisterConflict(aggregates.ErrCannotDeleteDefaultView, "Cannot delete the default view")
	registry.RegisterConflict(aggregates.ErrViewAlreadyDeleted, "View has been deleted")
	registry.RegisterConflict(aggregates.ErrViewAlreadyPrivate, "View is already private")
	registry.RegisterConflict(aggregates.ErrViewAlreadyPublic, "View is already public")
	registry.RegisterForbidden(aggregates.ErrOnlyOwnerCanMakePrivate, "Only the owner can make a view private")
	registry.RegisterForbidden(handlers.ErrNotAuthorizedToChangeVisibility, "Not authorized to change view visibility")
}
