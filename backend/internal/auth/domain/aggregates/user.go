package aggregates

import (
	"errors"
	"time"

	"easi/backend/internal/auth/domain/events"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/shared/domain"
)

var (
	ErrCannotDisableSelf       = errors.New("cannot disable your own account")
	ErrCannotDemoteLastAdmin   = errors.New("cannot demote last admin in tenant")
	ErrCannotDisableLastAdmin  = errors.New("cannot disable last admin in tenant")
	ErrUserAlreadyActive       = errors.New("user is already active")
	ErrUserAlreadyDisabled     = errors.New("user is already disabled")
	ErrSameRole                = errors.New("user already has this role")
)

type User struct {
	domain.AggregateRoot
	email      valueobjects.Email
	name       *string
	role       valueobjects.Role
	status     valueobjects.UserStatus
	externalID *string
	createdAt  time.Time
}

func NewUser(
	email valueobjects.Email,
	name string,
	role valueobjects.Role,
	externalID string,
	invitationID string,
) (*User, error) {
	user := &User{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	var namePtr *string
	if name != "" {
		namePtr = &name
	}

	var extIDPtr *string
	if externalID != "" {
		extIDPtr = &externalID
	}

	event := events.NewUserCreated(
		user.ID(),
		email.Value(),
		name,
		role.String(),
		externalID,
		invitationID,
	)

	user.apply(event)
	user.RaiseEvent(event)

	user.name = namePtr
	user.externalID = extIDPtr

	return user, nil
}

func LoadUserFromHistory(evts []domain.DomainEvent) (*User, error) {
	user := &User{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	user.LoadFromHistory(evts, func(event domain.DomainEvent) {
		user.apply(event)
	})

	return user, nil
}

func (u *User) ChangeRole(newRole valueobjects.Role, changedByID string, isLastAdmin bool) error {
	if u.role.Equals(newRole) {
		return ErrSameRole
	}

	if u.role.String() == "admin" && newRole.String() != "admin" && isLastAdmin {
		return ErrCannotDemoteLastAdmin
	}

	oldRole := u.role.String()
	event := events.NewUserRoleChanged(u.ID(), oldRole, newRole.String(), changedByID)
	u.apply(event)
	u.RaiseEvent(event)

	return nil
}

func (u *User) Disable(disabledByID string, isCurrentUser bool, isLastAdmin bool) error {
	if isCurrentUser {
		return ErrCannotDisableSelf
	}

	if !u.status.IsActive() {
		return ErrUserAlreadyDisabled
	}

	if u.role.String() == "admin" && isLastAdmin {
		return ErrCannotDisableLastAdmin
	}

	event := events.NewUserDisabled(u.ID(), disabledByID)
	u.apply(event)
	u.RaiseEvent(event)

	return nil
}

func (u *User) Enable(enabledByID string) error {
	if u.status.IsActive() {
		return ErrUserAlreadyActive
	}

	event := events.NewUserEnabled(u.ID(), enabledByID)
	u.apply(event)
	u.RaiseEvent(event)

	return nil
}

func (u *User) apply(event domain.DomainEvent) {
	switch e := event.(type) {
	case events.UserCreated:
		u.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
		u.email, _ = valueobjects.NewEmail(e.Email)
		u.role, _ = valueobjects.RoleFromString(e.Role)
		u.status = valueobjects.UserStatusActive
		if e.Name != "" {
			u.name = &e.Name
		}
		if e.ExternalID != "" {
			u.externalID = &e.ExternalID
		}
		u.createdAt = e.CreatedAt
	case events.UserRoleChanged:
		u.role, _ = valueobjects.RoleFromString(e.NewRole)
	case events.UserDisabled:
		u.status = valueobjects.UserStatusDisabled
	case events.UserEnabled:
		u.status = valueobjects.UserStatusActive
	}
}

func (u *User) Email() valueobjects.Email {
	return u.email
}

func (u *User) Name() *string {
	return u.name
}

func (u *User) Role() valueobjects.Role {
	return u.role
}

func (u *User) Status() valueobjects.UserStatus {
	return u.status
}

func (u *User) ExternalID() *string {
	return u.externalID
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}
