package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveAuthMode_DefaultsToProduction(t *testing.T) {
	for _, raw := range []string{"", "unknown", "production"} {
		mode, err := resolveAuthMode(raw)
		require.NoError(t, err)
		assert.Equal(t, AuthModeProduction, mode)
	}
}

func TestResolveAuthMode_LocalOIDC(t *testing.T) {
	mode, err := resolveAuthMode(" Local_OIDC ")
	require.NoError(t, err)
	assert.Equal(t, AuthModeLocalOIDC, mode)
}

func TestResolveAuthMode_BypassRequiresDevauthBuild(t *testing.T) {
	mode, err := resolveAuthMode("bypass")

	if bypassBuildEnabled {
		require.NoError(t, err)
		assert.Equal(t, AuthModeBypass, mode)
	} else {
		require.Error(t, err, "bypass must be refused in binaries built without -tags devauth")
	}
}

func TestBypassIdentity_IsStableAdminIdentity(t *testing.T) {
	identity := BypassIdentity()

	assert.NotEmpty(t, identity.UserID)
	assert.NotEmpty(t, identity.Email)
	assert.Equal(t, "admin", identity.Role)
	assert.NotEmpty(t, identity.TenantID)
}
