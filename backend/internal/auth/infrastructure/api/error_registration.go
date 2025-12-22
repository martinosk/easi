package api

import (
	"easi/backend/internal/auth/application/handlers"
	"easi/backend/internal/auth/application/services"
	"easi/backend/internal/auth/domain/aggregates"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/auth/infrastructure/repositories"
	sharedAPI "easi/backend/internal/shared/api"
)

func init() {
	registry := sharedAPI.GetErrorRegistry()

	registry.RegisterNotFound(repositories.ErrUserAggregateNotFound, "User not found")
	registry.RegisterNotFound(repositories.ErrUserNotFound, "User not found")
	registry.RegisterNotFound(repositories.ErrInvitationNotFound, "Invitation not found")
	registry.RegisterNotFound(repositories.ErrDomainNotFound, "Domain not registered")
	registry.RegisterNotFound(repositories.ErrTenantNotFound, "Tenant not found")

	registry.RegisterNotFound(handlers.ErrNoPendingInvitation, "No pending invitation found for this email")
	registry.RegisterNotFound(services.ErrNoValidInvitation, "No valid invitation found for this email")

	registry.RegisterConflict(aggregates.ErrCannotDisableSelf, "Cannot disable your own account")
	registry.RegisterConflict(aggregates.ErrCannotDemoteLastAdmin, "Cannot demote the last admin in tenant")
	registry.RegisterConflict(aggregates.ErrCannotDisableLastAdmin, "Cannot disable the last admin in tenant")
	registry.RegisterConflict(aggregates.ErrUserAlreadyActive, "User is already active")
	registry.RegisterConflict(aggregates.ErrUserAlreadyDisabled, "User is already disabled")
	registry.RegisterConflict(aggregates.ErrSameRole, "User already has this role")

	registry.RegisterConflict(aggregates.ErrInvitationAlreadyAccepted, "Invitation has already been accepted")
	registry.RegisterConflict(aggregates.ErrInvitationAlreadyRevoked, "Invitation has already been revoked")
	registry.RegisterConflict(aggregates.ErrInvitationAlreadyExpired, "Invitation has already expired")
	registry.RegisterConflict(aggregates.ErrInvitationNotPending, "Invitation is not pending")
	registry.RegisterConflict(aggregates.ErrInvitationExpired, "Invitation has expired")

	registry.RegisterValidation(valueobjects.ErrInvalidRole, "Invalid role")
	registry.RegisterValidation(valueobjects.ErrInvalidEmailFormat, "Invalid email format")
	registry.RegisterValidation(valueobjects.ErrEmailEmpty, "Email cannot be empty")

	registry.Register(sharedAPI.ErrorRegistration{Error: repositories.ErrTenantInactive, Category: sharedAPI.CategoryForbidden, Message: "Tenant is not active"})
	registry.Register(sharedAPI.ErrorRegistration{Error: services.ErrUserDisabled, Category: sharedAPI.CategoryForbidden, Message: "User account is disabled"})
}
