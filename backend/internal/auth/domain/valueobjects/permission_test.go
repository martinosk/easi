package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissionFromString_ValidPermission(t *testing.T) {
	perm, err := PermissionFromString("components:read")
	require.NoError(t, err)
	assert.Equal(t, PermComponentsRead, perm)
}

func TestPermissionFromString_InvalidPermission(t *testing.T) {
	_, err := PermissionFromString("invalid:permission")
	assert.ErrorIs(t, err, ErrInvalidPermission)
}

func TestPermissionFromString_EmptyPermission(t *testing.T) {
	_, err := PermissionFromString("")
	assert.ErrorIs(t, err, ErrInvalidPermission)
}

func TestPermission_Equals(t *testing.T) {
	perm1 := PermComponentsRead
	perm2, _ := PermissionFromString("components:read")
	perm3 := PermComponentsWrite

	assert.True(t, perm1.Equals(perm2))
	assert.False(t, perm1.Equals(perm3))
}

func TestPermissionsToStrings(t *testing.T) {
	perms := []Permission{PermComponentsRead, PermViewsWrite}
	strings := PermissionsToStrings(perms)

	assert.Equal(t, []string{"components:read", "views:write"}, strings)
}
