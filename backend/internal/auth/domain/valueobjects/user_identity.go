package valueobjects

import (
	"errors"

	"github.com/google/uuid"

	domain "easi/backend/internal/shared/eventsourcing"
	"easi/backend/internal/shared/eventsourcing/valueobjects"
)

var (
	ErrEmptyEmail = errors.New("email cannot be empty")
	ErrEmptyName  = errors.New("name cannot be empty")
)

type UserIdentity struct {
	userID   uuid.UUID
	email    string
	name     string
	tenantID valueobjects.TenantID
	role     Role
	status   UserStatus
}

func NewUserIdentity(
	userID uuid.UUID,
	email string,
	name string,
	tenantID valueobjects.TenantID,
	role Role,
	status UserStatus,
) (UserIdentity, error) {
	if email == "" {
		return UserIdentity{}, ErrEmptyEmail
	}
	if name == "" {
		return UserIdentity{}, ErrEmptyName
	}

	return UserIdentity{
		userID:   userID,
		email:    email,
		name:     name,
		tenantID: tenantID,
		role:     role,
		status:   status,
	}, nil
}

func (u UserIdentity) UserID() uuid.UUID {
	return u.userID
}

func (u UserIdentity) Email() string {
	return u.email
}

func (u UserIdentity) Name() string {
	return u.name
}

func (u UserIdentity) TenantID() valueobjects.TenantID {
	return u.tenantID
}

func (u UserIdentity) Role() Role {
	return u.role
}

func (u UserIdentity) Status() UserStatus {
	return u.status
}

func (u UserIdentity) Permissions() []Permission {
	return u.role.Permissions()
}

func (u UserIdentity) HasPermission(p Permission) bool {
	return u.role.HasPermission(p)
}

func (u UserIdentity) IsActive() bool {
	return u.status.IsActive()
}

func (u UserIdentity) Equals(other domain.ValueObject) bool {
	otherIdentity, ok := other.(UserIdentity)
	if !ok {
		return false
	}
	return u.userID == otherIdentity.userID &&
		u.email == otherIdentity.email &&
		u.name == otherIdentity.name &&
		u.tenantID.Equals(otherIdentity.tenantID) &&
		u.role.Equals(otherIdentity.role) &&
		u.status.Equals(otherIdentity.status)
}
