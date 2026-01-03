package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleFromString_ValidRole(t *testing.T) {
	role, err := RoleFromString("admin")
	require.NoError(t, err)
	assert.Equal(t, RoleAdmin, role)
}

func TestRoleFromString_InvalidRole(t *testing.T) {
	_, err := RoleFromString("superuser")
	assert.ErrorIs(t, err, ErrInvalidRole)
}

func TestRoleFromString_EmptyRole(t *testing.T) {
	_, err := RoleFromString("")
	assert.ErrorIs(t, err, ErrInvalidRole)
}

func TestRoleFromString_CaseInsensitive(t *testing.T) {
	role, err := RoleFromString("ADMIN")
	require.NoError(t, err)
	assert.Equal(t, RoleAdmin, role)
}

func TestRole_Permissions(t *testing.T) {
	adminPerms := RoleAdmin.Permissions()
	assert.Contains(t, adminPerms, PermUsersManage)
	assert.Contains(t, adminPerms, PermComponentsDelete)
	assert.Contains(t, adminPerms, PermMetaModelWrite)

	architectPerms := RoleArchitect.Permissions()
	assert.Contains(t, architectPerms, PermComponentsWrite)
	assert.Contains(t, architectPerms, PermComponentsDelete)
	assert.Contains(t, architectPerms, PermViewsWrite)
	assert.Contains(t, architectPerms, PermViewsDelete)
	assert.Contains(t, architectPerms, PermMetaModelRead)
	assert.Contains(t, architectPerms, PermAuditRead)
	assert.NotContains(t, architectPerms, PermUsersManage)
	assert.NotContains(t, architectPerms, PermMetaModelWrite)

	stakeholderPerms := RoleStakeholder.Permissions()
	assert.Contains(t, stakeholderPerms, PermComponentsRead)
	assert.Contains(t, stakeholderPerms, PermViewsRead)
	assert.Contains(t, stakeholderPerms, PermMetaModelRead)
	assert.Contains(t, stakeholderPerms, PermAuditRead)
	assert.NotContains(t, stakeholderPerms, PermComponentsWrite)
	assert.NotContains(t, stakeholderPerms, PermMetaModelWrite)
}

func TestRole_HasPermission(t *testing.T) {
	assert.True(t, RoleAdmin.HasPermission(PermUsersManage))
	assert.True(t, RoleAdmin.HasPermission(PermMetaModelWrite))

	assert.True(t, RoleArchitect.HasPermission(PermComponentsWrite))
	assert.False(t, RoleArchitect.HasPermission(PermUsersManage))
	assert.False(t, RoleArchitect.HasPermission(PermMetaModelWrite))

	assert.True(t, RoleStakeholder.HasPermission(PermComponentsRead))
	assert.False(t, RoleStakeholder.HasPermission(PermComponentsWrite))
	assert.False(t, RoleStakeholder.HasPermission(PermMetaModelWrite))
}
