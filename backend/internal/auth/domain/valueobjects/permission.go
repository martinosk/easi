package valueobjects

import (
	"errors"

	"easi/backend/internal/shared/domain"
)

var ErrInvalidPermission = errors.New("invalid permission")

type Permission struct {
	value string
}

var (
	PermComponentsRead   = Permission{value: "components:read"}
	PermComponentsWrite  = Permission{value: "components:write"}
	PermComponentsDelete = Permission{value: "components:delete"}

	PermViewsRead   = Permission{value: "views:read"}
	PermViewsWrite  = Permission{value: "views:write"}
	PermViewsDelete = Permission{value: "views:delete"}

	PermCapabilitiesRead   = Permission{value: "capabilities:read"}
	PermCapabilitiesWrite  = Permission{value: "capabilities:write"}
	PermCapabilitiesDelete = Permission{value: "capabilities:delete"}

	PermDomainsRead   = Permission{value: "domains:read"}
	PermDomainsWrite  = Permission{value: "domains:write"}
	PermDomainsDelete = Permission{value: "domains:delete"}

	PermUsersRead   = Permission{value: "users:read"}
	PermUsersManage = Permission{value: "users:manage"}

	PermInvitationsManage = Permission{value: "invitations:manage"}
)

var validPermissions = map[string]Permission{
	"components:read":     PermComponentsRead,
	"components:write":    PermComponentsWrite,
	"components:delete":   PermComponentsDelete,
	"views:read":          PermViewsRead,
	"views:write":         PermViewsWrite,
	"views:delete":        PermViewsDelete,
	"capabilities:read":   PermCapabilitiesRead,
	"capabilities:write":  PermCapabilitiesWrite,
	"capabilities:delete": PermCapabilitiesDelete,
	"domains:read":        PermDomainsRead,
	"domains:write":       PermDomainsWrite,
	"domains:delete":      PermDomainsDelete,
	"users:read":          PermUsersRead,
	"users:manage":        PermUsersManage,
	"invitations:manage":  PermInvitationsManage,
}

func PermissionFromString(s string) (Permission, error) {
	if perm, ok := validPermissions[s]; ok {
		return perm, nil
	}
	return Permission{}, ErrInvalidPermission
}

func (p Permission) String() string {
	return p.value
}

func (p Permission) Equals(other domain.ValueObject) bool {
	otherPerm, ok := other.(Permission)
	if !ok {
		return false
	}
	return p.value == otherPerm.value
}

func PermissionsToStrings(perms []Permission) []string {
	result := make([]string, len(perms))
	for i, p := range perms {
		result[i] = p.String()
	}
	return result
}
