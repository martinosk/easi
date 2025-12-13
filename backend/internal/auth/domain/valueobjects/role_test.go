package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoleFromString_ValidRoles(t *testing.T) {
	testCases := []struct {
		input    string
		expected Role
	}{
		{"admin", RoleAdmin},
		{"architect", RoleArchitect},
		{"stakeholder", RoleStakeholder},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			role, err := RoleFromString(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, role)
		})
	}
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

	role, err = RoleFromString("Architect")
	require.NoError(t, err)
	assert.Equal(t, RoleArchitect, role)
}

func TestRole_String(t *testing.T) {
	assert.Equal(t, "admin", RoleAdmin.String())
	assert.Equal(t, "architect", RoleArchitect.String())
	assert.Equal(t, "stakeholder", RoleStakeholder.String())
}

func TestRole_Permissions(t *testing.T) {
	adminPerms := RoleAdmin.Permissions()
	assert.Contains(t, adminPerms, PermUsersManage)
	assert.Contains(t, adminPerms, PermComponentsDelete)

	architectPerms := RoleArchitect.Permissions()
	assert.Contains(t, architectPerms, PermComponentsWrite)
	assert.Contains(t, architectPerms, PermViewsWrite)
	assert.NotContains(t, architectPerms, PermUsersManage)
	assert.NotContains(t, architectPerms, PermComponentsDelete)

	stakeholderPerms := RoleStakeholder.Permissions()
	assert.Contains(t, stakeholderPerms, PermComponentsRead)
	assert.Contains(t, stakeholderPerms, PermViewsRead)
	assert.NotContains(t, stakeholderPerms, PermComponentsWrite)
}

func TestRole_Equals(t *testing.T) {
	role1 := RoleAdmin
	role2, _ := RoleFromString("admin")
	role3 := RoleArchitect

	assert.True(t, role1.Equals(role2), "same roles should be equal")
	assert.False(t, role1.Equals(role3), "different roles should not be equal")
}

func TestRole_HasPermission(t *testing.T) {
	assert.True(t, RoleAdmin.HasPermission(PermUsersManage))
	assert.True(t, RoleAdmin.HasPermission(PermComponentsDelete))
	assert.True(t, RoleAdmin.HasPermission(PermComponentsRead))

	assert.True(t, RoleArchitect.HasPermission(PermComponentsWrite))
	assert.False(t, RoleArchitect.HasPermission(PermUsersManage))
	assert.False(t, RoleArchitect.HasPermission(PermComponentsDelete))

	assert.True(t, RoleStakeholder.HasPermission(PermComponentsRead))
	assert.False(t, RoleStakeholder.HasPermission(PermComponentsWrite))
}
