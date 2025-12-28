package valueobjects

import (
	domain "easi/backend/internal/shared/eventsourcing"
	"errors"
	"strings"
)

var ErrInvalidRole = errors.New("invalid role: must be admin, architect, or stakeholder")

type Role struct {
	value string
}

var (
	RoleAdmin       = Role{value: "admin"}
	RoleArchitect   = Role{value: "architect"}
	RoleStakeholder = Role{value: "stakeholder"}
)

func RoleFromString(s string) (Role, error) {
	switch strings.ToLower(s) {
	case "admin":
		return RoleAdmin, nil
	case "architect":
		return RoleArchitect, nil
	case "stakeholder":
		return RoleStakeholder, nil
	default:
		return Role{}, ErrInvalidRole
	}
}

func (r Role) String() string {
	return r.value
}

func (r Role) Equals(other domain.ValueObject) bool {
	otherRole, ok := other.(Role)
	if !ok {
		return false
	}
	return r.value == otherRole.value
}

func (r Role) Permissions() []Permission {
	switch r.value {
	case "admin":
		return []Permission{
			PermComponentsRead, PermComponentsWrite, PermComponentsDelete,
			PermViewsRead, PermViewsWrite, PermViewsDelete,
			PermCapabilitiesRead, PermCapabilitiesWrite, PermCapabilitiesDelete,
			PermDomainsRead, PermDomainsWrite, PermDomainsDelete,
			PermUsersRead, PermUsersManage,
			PermInvitationsManage,
			PermMetaModelWrite,
		}
	case "architect":
		return []Permission{
			PermComponentsRead, PermComponentsWrite,
			PermViewsRead, PermViewsWrite,
			PermCapabilitiesRead, PermCapabilitiesWrite,
			PermDomainsRead, PermDomainsWrite,
		}
	case "stakeholder":
		return []Permission{
			PermComponentsRead,
			PermViewsRead,
			PermCapabilitiesRead,
			PermDomainsRead,
		}
	default:
		return []Permission{}
	}
}

func (r Role) HasPermission(p Permission) bool {
	for _, perm := range r.Permissions() {
		if perm.Equals(p) {
			return true
		}
	}
	return false
}
