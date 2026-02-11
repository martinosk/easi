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

var rolePermissionsList = map[string][]Permission{
	"admin": {
		PermComponentsRead, PermComponentsWrite, PermComponentsDelete,
		PermViewsRead, PermViewsWrite, PermViewsDelete,
		PermCapabilitiesRead, PermCapabilitiesWrite, PermCapabilitiesDelete,
		PermDomainsRead, PermDomainsWrite, PermDomainsDelete,
		PermUsersRead, PermUsersManage,
		PermInvitationsManage,
		PermMetaModelRead, PermMetaModelWrite,
		PermAuditRead,
		PermEnterpriseArchRead, PermEnterpriseArchWrite, PermEnterpriseArchDelete,
		PermEditGrantsManage,
		PermValueStreamsRead, PermValueStreamsWrite, PermValueStreamsDelete,
	},
	"architect": {
		PermComponentsRead, PermComponentsWrite, PermComponentsDelete,
		PermViewsRead, PermViewsWrite, PermViewsDelete,
		PermCapabilitiesRead, PermCapabilitiesWrite, PermCapabilitiesDelete,
		PermDomainsRead, PermDomainsWrite, PermDomainsDelete,
		PermUsersRead,
		PermMetaModelRead,
		PermAuditRead,
		PermEnterpriseArchRead, PermEnterpriseArchWrite, PermEnterpriseArchDelete,
		PermEditGrantsManage,
		PermValueStreamsRead, PermValueStreamsWrite, PermValueStreamsDelete,
	},
	"stakeholder": {
		PermComponentsRead,
		PermViewsRead,
		PermCapabilitiesRead,
		PermDomainsRead,
		PermMetaModelRead,
		PermAuditRead,
		PermEnterpriseArchRead,
		PermValueStreamsRead,
	},
}

func (r Role) Permissions() []Permission {
	if perms, ok := rolePermissionsList[r.value]; ok {
		return perms
	}
	return []Permission{}
}

func (r Role) HasPermission(p Permission) bool {
	for _, perm := range r.Permissions() {
		if perm.Equals(p) {
			return true
		}
	}
	return false
}
