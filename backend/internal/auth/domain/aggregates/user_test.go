package aggregates

import (
	"testing"

	"easi/backend/internal/auth/domain/valueobjects"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUser_CreatesUserWithValidData(t *testing.T) {
	email, err := valueobjects.NewEmail("jane@acme.com")
	require.NoError(t, err)

	role, err := valueobjects.RoleFromString("architect")
	require.NoError(t, err)

	user, err := NewUser(email, "Jane Doe", role, "ext-123", "inv-456")

	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, user.ID())
	assert.Equal(t, email.Value(), user.Email().Value())
	assert.Equal(t, "Jane Doe", *user.Name())
	assert.Equal(t, role.String(), user.Role().String())
	assert.Equal(t, valueobjects.UserStatusActive, user.Status())
	assert.Equal(t, "ext-123", *user.ExternalID())
	assert.NotZero(t, user.CreatedAt())
}

func TestNewUser_RaisesUserCreatedEvent(t *testing.T) {
	email, _ := valueobjects.NewEmail("test@example.com")
	role, _ := valueobjects.RoleFromString("admin")

	user, err := NewUser(email, "Test User", role, "ext-789", "inv-101")
	require.NoError(t, err)

	uncommittedEvents := user.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "UserCreated", uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	assert.Equal(t, user.ID(), eventData["id"])
	assert.Equal(t, "test@example.com", eventData["email"])
	assert.Equal(t, "Test User", eventData["name"])
	assert.Equal(t, "admin", eventData["role"])
	assert.Equal(t, "active", eventData["status"])
	assert.Equal(t, "ext-789", eventData["externalId"])
	assert.Equal(t, "inv-101", eventData["invitationId"])
}

func TestNewUser_WithEmptyName(t *testing.T) {
	email, _ := valueobjects.NewEmail("user@example.com")
	role, _ := valueobjects.RoleFromString("stakeholder")

	user, err := NewUser(email, "", role, "ext-111", "inv-222")
	require.NoError(t, err)
	assert.Nil(t, user.Name())
}

func TestNewUser_WithEmptyExternalID(t *testing.T) {
	email, _ := valueobjects.NewEmail("user@example.com")
	role, _ := valueobjects.RoleFromString("stakeholder")

	user, err := NewUser(email, "User Name", role, "", "inv-333")
	require.NoError(t, err)
	assert.Nil(t, user.ExternalID())
}

func TestUser_ChangeRole_Success(t *testing.T) {
	user := createTestUser(t, "admin")
	user.MarkChangesAsCommitted()

	newRole, err := valueobjects.RoleFromString("architect")
	require.NoError(t, err)

	err = user.ChangeRole(newRole, "changer-123", false)
	require.NoError(t, err)

	assert.Equal(t, "architect", user.Role().String())

	uncommittedEvents := user.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "UserRoleChanged", uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	assert.Equal(t, "admin", eventData["oldRole"])
	assert.Equal(t, "architect", eventData["newRole"])
	assert.Equal(t, "changer-123", eventData["changedById"])
}

func TestUser_ChangeRole_SameRole_ReturnsError(t *testing.T) {
	user := createTestUser(t, "architect")
	user.MarkChangesAsCommitted()

	role, _ := valueobjects.RoleFromString("architect")

	err := user.ChangeRole(role, "changer-456", false)
	assert.ErrorIs(t, err, ErrSameRole)
	assert.Empty(t, user.GetUncommittedChanges())
}

func TestUser_ChangeRole_DemoteLastAdmin_ReturnsError(t *testing.T) {
	user := createTestUser(t, "admin")
	user.MarkChangesAsCommitted()

	newRole, _ := valueobjects.RoleFromString("architect")

	err := user.ChangeRole(newRole, "changer-789", true)
	assert.ErrorIs(t, err, ErrCannotDemoteLastAdmin)
	assert.Equal(t, "admin", user.Role().String())
	assert.Empty(t, user.GetUncommittedChanges())
}

func TestUser_ChangeRole_DemoteNonLastAdmin_Success(t *testing.T) {
	user := createTestUser(t, "admin")
	user.MarkChangesAsCommitted()

	newRole, _ := valueobjects.RoleFromString("stakeholder")

	err := user.ChangeRole(newRole, "changer-999", false)
	require.NoError(t, err)
	assert.Equal(t, "stakeholder", user.Role().String())
}

func TestUser_Disable_Success(t *testing.T) {
	user := createTestUser(t, "architect")
	user.MarkChangesAsCommitted()

	err := user.Disable("disabler-123", false, false)
	require.NoError(t, err)

	assert.False(t, user.Status().IsActive())

	uncommittedEvents := user.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "UserDisabled", uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	assert.Equal(t, user.ID(), eventData["id"])
	assert.Equal(t, "disabler-123", eventData["disabledBy"])
}

func TestUser_Disable_CurrentUser_ReturnsError(t *testing.T) {
	user := createTestUser(t, "admin")
	user.MarkChangesAsCommitted()

	err := user.Disable("disabler-456", true, false)
	assert.ErrorIs(t, err, ErrCannotDisableSelf)
	assert.True(t, user.Status().IsActive())
	assert.Empty(t, user.GetUncommittedChanges())
}

func TestUser_Disable_AlreadyDisabled_ReturnsError(t *testing.T) {
	user := createTestUser(t, "architect")
	user.Disable("disabler-111", false, false)
	user.MarkChangesAsCommitted()

	err := user.Disable("disabler-222", false, false)
	assert.ErrorIs(t, err, ErrUserAlreadyDisabled)
	assert.Empty(t, user.GetUncommittedChanges())
}

func TestUser_Disable_LastAdmin_ReturnsError(t *testing.T) {
	user := createTestUser(t, "admin")
	user.MarkChangesAsCommitted()

	err := user.Disable("disabler-333", false, true)
	assert.ErrorIs(t, err, ErrCannotDisableLastAdmin)
	assert.True(t, user.Status().IsActive())
	assert.Empty(t, user.GetUncommittedChanges())
}

func TestUser_Disable_NonLastAdmin_Success(t *testing.T) {
	user := createTestUser(t, "admin")
	user.MarkChangesAsCommitted()

	err := user.Disable("disabler-444", false, false)
	require.NoError(t, err)
	assert.False(t, user.Status().IsActive())
}

func TestUser_Enable_Success(t *testing.T) {
	user := createTestUser(t, "stakeholder")
	user.Disable("disabler-555", false, false)
	user.MarkChangesAsCommitted()

	err := user.Enable("enabler-666")
	require.NoError(t, err)

	assert.True(t, user.Status().IsActive())

	uncommittedEvents := user.GetUncommittedChanges()
	require.Len(t, uncommittedEvents, 1)
	assert.Equal(t, "UserEnabled", uncommittedEvents[0].EventType())

	eventData := uncommittedEvents[0].EventData()
	assert.Equal(t, user.ID(), eventData["id"])
	assert.Equal(t, "enabler-666", eventData["enabledBy"])
}

func TestUser_Enable_AlreadyActive_ReturnsError(t *testing.T) {
	user := createTestUser(t, "architect")
	user.MarkChangesAsCommitted()

	err := user.Enable("enabler-777")
	assert.ErrorIs(t, err, ErrUserAlreadyActive)
	assert.Empty(t, user.GetUncommittedChanges())
}

func TestUser_LoadFromHistory_PreservesState(t *testing.T) {
	email, _ := valueobjects.NewEmail("history@example.com")
	role, _ := valueobjects.RoleFromString("admin")

	user, _ := NewUser(email, "History User", role, "ext-history", "inv-history")

	newRole, _ := valueobjects.RoleFromString("architect")
	user.ChangeRole(newRole, "changer-hist", false)
	user.Disable("disabler-hist", false, false)

	allEvents := user.GetUncommittedChanges()

	loadedUser, err := LoadUserFromHistory(allEvents)
	require.NoError(t, err)

	assert.Equal(t, user.ID(), loadedUser.ID())
	assert.Equal(t, email.Value(), loadedUser.Email().Value())
	assert.Equal(t, "History User", *loadedUser.Name())
	assert.Equal(t, "architect", loadedUser.Role().String())
	assert.False(t, loadedUser.Status().IsActive())
	assert.Equal(t, "ext-history", *loadedUser.ExternalID())
}

func TestUser_LoadFromHistory_MultipleRoleChanges(t *testing.T) {
	email, _ := valueobjects.NewEmail("multi@example.com")
	role, _ := valueobjects.RoleFromString("stakeholder")

	user, _ := NewUser(email, "Multi User", role, "", "")

	role2, _ := valueobjects.RoleFromString("architect")
	user.ChangeRole(role2, "changer-1", false)

	role3, _ := valueobjects.RoleFromString("admin")
	user.ChangeRole(role3, "changer-2", false)

	allEvents := user.GetUncommittedChanges()

	loadedUser, err := LoadUserFromHistory(allEvents)
	require.NoError(t, err)

	assert.Equal(t, "admin", loadedUser.Role().String())
}

func TestUser_LoadFromHistory_EnableAfterDisable(t *testing.T) {
	email, _ := valueobjects.NewEmail("toggle@example.com")
	role, _ := valueobjects.RoleFromString("architect")

	user, _ := NewUser(email, "Toggle User", role, "", "")
	user.Disable("disabler", false, false)
	user.Enable("enabler")

	allEvents := user.GetUncommittedChanges()

	loadedUser, err := LoadUserFromHistory(allEvents)
	require.NoError(t, err)

	assert.True(t, loadedUser.Status().IsActive())
}

func TestUser_ApplyEvents_PreservesOtherState(t *testing.T) {
	email, _ := valueobjects.NewEmail("preserve@example.com")
	role, _ := valueobjects.RoleFromString("admin")

	user, _ := NewUser(email, "Preserve User", role, "ext-pres", "inv-pres")
	originalID := user.ID()
	originalCreatedAt := user.CreatedAt()
	user.MarkChangesAsCommitted()

	newRole, _ := valueobjects.RoleFromString("architect")
	user.ChangeRole(newRole, "changer", false)

	assert.Equal(t, originalID, user.ID())
	assert.Equal(t, email.Value(), user.Email().Value())
	assert.Equal(t, "Preserve User", *user.Name())
	assert.Equal(t, "ext-pres", *user.ExternalID())
	assert.Equal(t, originalCreatedAt, user.CreatedAt())
}

func createTestUser(t *testing.T, roleName string) *User {
	t.Helper()

	email, err := valueobjects.NewEmail("test@example.com")
	require.NoError(t, err)

	role, err := valueobjects.RoleFromString(roleName)
	require.NoError(t, err)

	user, err := NewUser(email, "Test User", role, "ext-test", "inv-test")
	require.NoError(t, err)

	return user
}
