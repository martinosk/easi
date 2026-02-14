package api

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authvo "easi/backend/internal/auth/domain/valueobjects"
	sharedctx "easi/backend/internal/shared/context"
)

func TestRolePermissionMaps_AreInSync(t *testing.T) {
	roles := []struct {
		name       string
		authRole   string
		sharedRole sharedctx.Role
	}{
		{"admin", "admin", sharedctx.RoleAdmin},
		{"architect", "architect", sharedctx.RoleArchitect},
		{"stakeholder", "stakeholder", sharedctx.RoleStakeholder},
	}

	for _, rc := range roles {
		t.Run(rc.name, func(t *testing.T) {
			authRole, err := authvo.RoleFromString(rc.authRole)
			require.NoError(t, err)
			authPerms := authvo.PermissionsToStrings(authRole.Permissions())
			sort.Strings(authPerms)

			sharedPerms := rc.sharedRole.Permissions()
			sharedPermsList := make([]string, 0, len(sharedPerms))
			for p := range sharedPerms {
				sharedPermsList = append(sharedPermsList, p)
			}
			sort.Strings(sharedPermsList)

			assert.Equal(t, authPerms, sharedPermsList,
				"Permission maps for role %q must be identical between auth/domain/valueobjects.rolePermissionsList and shared/context.rolePermissions",
				rc.name)
		})
	}
}
