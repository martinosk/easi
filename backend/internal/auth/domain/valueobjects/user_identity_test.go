package valueobjects

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"easi/backend/internal/shared/eventsourcing/valueobjects"
)

func TestNewUserIdentity_ValidInput(t *testing.T) {
	userID := uuid.New()
	tenantID, _ := valueobjects.NewTenantID("acme")

	identity, err := NewUserIdentity(
		userID,
		"john@acme.com",
		"John Doe",
		tenantID,
		RoleArchitect,
		UserStatusActive,
	)

	require.NoError(t, err)
	assert.Equal(t, userID, identity.UserID())
	assert.Equal(t, "john@acme.com", identity.Email())
	assert.Equal(t, "John Doe", identity.Name())
	assert.Equal(t, tenantID, identity.TenantID())
	assert.Equal(t, RoleArchitect, identity.Role())
	assert.Equal(t, UserStatusActive, identity.Status())
}

func TestNewUserIdentity_EmptyEmail(t *testing.T) {
	userID := uuid.New()
	tenantID, _ := valueobjects.NewTenantID("acme")

	_, err := NewUserIdentity(userID, "", "John Doe", tenantID, RoleArchitect, UserStatusActive)
	assert.ErrorIs(t, err, ErrEmptyEmail)
}

func TestNewUserIdentity_EmptyName(t *testing.T) {
	userID := uuid.New()
	tenantID, _ := valueobjects.NewTenantID("acme")

	_, err := NewUserIdentity(userID, "john@acme.com", "", tenantID, RoleArchitect, UserStatusActive)
	assert.ErrorIs(t, err, ErrEmptyName)
}

func TestUserIdentity_Permissions(t *testing.T) {
	userID := uuid.New()
	tenantID, _ := valueobjects.NewTenantID("acme")

	identity, err := NewUserIdentity(userID, "john@acme.com", "John Doe", tenantID, RoleArchitect, UserStatusActive)
	require.NoError(t, err)

	perms := identity.Permissions()
	assert.Contains(t, perms, PermComponentsWrite)
	assert.Contains(t, perms, PermViewsWrite)
	assert.NotContains(t, perms, PermUsersManage)
}

func TestUserIdentity_HasPermission(t *testing.T) {
	userID := uuid.New()
	tenantID, _ := valueobjects.NewTenantID("acme")

	identity, err := NewUserIdentity(userID, "john@acme.com", "John Doe", tenantID, RoleAdmin, UserStatusActive)
	require.NoError(t, err)

	assert.True(t, identity.HasPermission(PermUsersManage))
	assert.True(t, identity.HasPermission(PermComponentsDelete))
}

func TestUserIdentity_IsActive(t *testing.T) {
	userID := uuid.New()
	tenantID, _ := valueobjects.NewTenantID("acme")

	activeUser, _ := NewUserIdentity(userID, "john@acme.com", "John Doe", tenantID, RoleArchitect, UserStatusActive)
	disabledUser, _ := NewUserIdentity(userID, "john@acme.com", "John Doe", tenantID, RoleArchitect, UserStatusDisabled)

	assert.True(t, activeUser.IsActive())
	assert.False(t, disabledUser.IsActive())
}

func TestUserIdentity_Equals(t *testing.T) {
	userID := uuid.New()
	tenantID, _ := valueobjects.NewTenantID("acme")

	identity1, _ := NewUserIdentity(userID, "john@acme.com", "John Doe", tenantID, RoleArchitect, UserStatusActive)
	identity2, _ := NewUserIdentity(userID, "john@acme.com", "John Doe", tenantID, RoleArchitect, UserStatusActive)
	identity3, _ := NewUserIdentity(uuid.New(), "jane@acme.com", "Jane Doe", tenantID, RoleArchitect, UserStatusActive)

	assert.True(t, identity1.Equals(identity2), "same user identities should be equal")
	assert.False(t, identity1.Equals(identity3), "different user identities should not be equal")
}
