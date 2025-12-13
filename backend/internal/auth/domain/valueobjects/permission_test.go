package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissionFromString_ValidPermissions(t *testing.T) {
	testCases := []struct {
		input    string
		expected Permission
	}{
		{"components:read", PermComponentsRead},
		{"components:write", PermComponentsWrite},
		{"components:delete", PermComponentsDelete},
		{"views:read", PermViewsRead},
		{"views:write", PermViewsWrite},
		{"views:delete", PermViewsDelete},
		{"capabilities:read", PermCapabilitiesRead},
		{"capabilities:write", PermCapabilitiesWrite},
		{"capabilities:delete", PermCapabilitiesDelete},
		{"domains:read", PermDomainsRead},
		{"domains:write", PermDomainsWrite},
		{"domains:delete", PermDomainsDelete},
		{"users:read", PermUsersRead},
		{"users:manage", PermUsersManage},
		{"invitations:manage", PermInvitationsManage},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			perm, err := PermissionFromString(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, perm)
		})
	}
}

func TestPermissionFromString_InvalidPermission(t *testing.T) {
	_, err := PermissionFromString("invalid:permission")
	assert.ErrorIs(t, err, ErrInvalidPermission)
}

func TestPermissionFromString_EmptyPermission(t *testing.T) {
	_, err := PermissionFromString("")
	assert.ErrorIs(t, err, ErrInvalidPermission)
}

func TestPermission_String(t *testing.T) {
	assert.Equal(t, "components:read", PermComponentsRead.String())
	assert.Equal(t, "users:manage", PermUsersManage.String())
}

func TestPermission_Equals(t *testing.T) {
	perm1 := PermComponentsRead
	perm2, _ := PermissionFromString("components:read")
	perm3 := PermComponentsWrite

	assert.True(t, perm1.Equals(perm2), "same permissions should be equal")
	assert.False(t, perm1.Equals(perm3), "different permissions should not be equal")
}

func TestPermissionsToStrings(t *testing.T) {
	perms := []Permission{PermComponentsRead, PermViewsWrite}
	strings := PermissionsToStrings(perms)

	assert.Equal(t, []string{"components:read", "views:write"}, strings)
}
