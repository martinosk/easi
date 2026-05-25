package aggregates

import (
	"errors"
	"fmt"
	"time"

	"easi/backend/internal/auth/domain/events"
	"easi/backend/internal/auth/domain/valueobjects"
	domain "easi/backend/internal/shared/eventsourcing"
)

var (
	ErrCannotDisableSelf      = errors.New("cannot disable your own account")
	ErrCannotDemoteLastAdmin  = errors.New("cannot demote last admin in tenant")
	ErrCannotDisableLastAdmin = errors.New("cannot disable last admin in tenant")
	ErrUserAlreadyActive      = errors.New("user is already active")
	ErrUserAlreadyDisabled    = errors.New("user is already disabled")
	ErrSameRole               = errors.New("user already has this role")
)

type User struct {
	domain.AggregateRoot
	email     valueobjects.Email
	profile   valueobjects.ExternalProfile
	role      valueobjects.Role
	status    valueobjects.UserStatus
	createdAt time.Time
}

func NewUser(
	email valueobjects.Email,
	profile valueobjects.ExternalProfile,
	role valueobjects.Role,
	invitationID string,
) (*User, error) {
	user := &User{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	event := events.NewUserCreated(
		user.ID(),
		email.Value(),
		profile.Name(),
		role.String(),
		profile.ExternalID(),
		invitationID,
	)

	if err := user.apply(event); err != nil {
		return nil, err
	}
	user.RaiseEvent(event)

	return user, nil
}

func LoadUserFromHistory(evts []domain.DomainEvent) (*User, error) {
	user := &User{
		AggregateRoot: domain.NewAggregateRoot(),
	}

	var applyErr error
	user.LoadFromHistory(evts, func(event domain.DomainEvent) {
		if applyErr != nil {
			return
		}
		applyErr = user.apply(event)
	})
	if applyErr != nil {
		return nil, applyErr
	}

	return user, nil
}

func (u *User) ChangeRole(newRole valueobjects.Role, changedBy valueobjects.UserID, isLastAdmin bool) error {
	if u.role.Equals(newRole) {
		return ErrSameRole
	}

	if u.isDemotionOfLastAdmin(newRole, isLastAdmin) {
		return ErrCannotDemoteLastAdmin
	}

	oldRole := u.role.String()
	event := events.NewUserRoleChanged(u.ID(), oldRole, newRole.String(), changedBy.Value())
	if err := u.apply(event); err != nil {
		return err
	}
	u.RaiseEvent(event)

	return nil
}

func (u *User) Disable(disabledBy valueobjects.UserID, isCurrentUser bool, isLastAdmin bool) error {
	if isCurrentUser {
		return ErrCannotDisableSelf
	}

	if !u.status.IsActive() {
		return ErrUserAlreadyDisabled
	}

	if u.role.IsAdmin() && isLastAdmin {
		return ErrCannotDisableLastAdmin
	}

	event := events.NewUserDisabled(u.ID(), disabledBy.Value())
	if err := u.apply(event); err != nil {
		return err
	}
	u.RaiseEvent(event)

	return nil
}

func (u *User) Enable(enabledBy valueobjects.UserID) error {
	if u.status.IsActive() {
		return ErrUserAlreadyActive
	}

	event := events.NewUserEnabled(u.ID(), enabledBy.Value())
	if err := u.apply(event); err != nil {
		return err
	}
	u.RaiseEvent(event)

	return nil
}

func (u *User) isDemotionOfLastAdmin(newRole valueobjects.Role, isLastAdmin bool) bool {
	return u.role.IsAdmin() && !newRole.IsAdmin() && isLastAdmin
}

func (u *User) apply(event domain.DomainEvent) error {
	switch e := event.(type) {
	case events.UserCreated:
		return u.applyCreated(e)
	case events.UserRoleChanged:
		return u.applyRoleChanged(e)
	case events.UserDisabled:
		u.status = valueobjects.UserStatusDisabled
	case events.UserEnabled:
		u.status = valueobjects.UserStatusActive
	}
	return nil
}

func (u *User) applyCreated(e events.UserCreated) error {
	email, err := valueobjects.NewEmail(e.Email)
	if err != nil {
		return fmt.Errorf("%w: email %q: %v", domain.ErrCorruptedEvent, e.Email, err)
	}
	role, err := valueobjects.RoleFromString(e.Role)
	if err != nil {
		return fmt.Errorf("%w: role %q: %v", domain.ErrCorruptedEvent, e.Role, err)
	}
	u.AggregateRoot = domain.NewAggregateRootWithID(e.ID)
	u.email = email
	u.profile = valueobjects.NewExternalProfile(e.Name, e.ExternalID)
	u.role = role
	u.status = valueobjects.UserStatusActive
	u.createdAt = e.CreatedAt
	return nil
}

func (u *User) applyRoleChanged(e events.UserRoleChanged) error {
	role, err := valueobjects.RoleFromString(e.NewRole)
	if err != nil {
		return fmt.Errorf("%w: role %q: %v", domain.ErrCorruptedEvent, e.NewRole, err)
	}
	u.role = role
	return nil
}

func (u *User) Email() valueobjects.Email {
	return u.email
}

func (u *User) Name() *string {
	if u.profile.HasName() {
		name := u.profile.Name()
		return &name
	}
	return nil
}

func (u *User) Profile() valueobjects.ExternalProfile {
	return u.profile
}

func (u *User) Role() valueobjects.Role {
	return u.role
}

func (u *User) Status() valueobjects.UserStatus {
	return u.status
}

func (u *User) ExternalID() *string {
	if u.profile.HasExternalID() {
		id := u.profile.ExternalID()
		return &id
	}
	return nil
}

func (u *User) CreatedAt() time.Time {
	return u.createdAt
}
