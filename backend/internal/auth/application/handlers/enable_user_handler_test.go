package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/domain/aggregates"
	"easi/backend/internal/auth/domain/valueobjects"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEnableUserRepository struct {
	savedUsers []*aggregates.User
	userToLoad *aggregates.User
	getErr     error
	saveErr    error
}

func (m *mockEnableUserRepository) Save(ctx context.Context, user *aggregates.User) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedUsers = append(m.savedUsers, user)
	return nil
}

func (m *mockEnableUserRepository) GetByID(ctx context.Context, id string) (*aggregates.User, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.userToLoad, nil
}

func TestEnableUserHandler_EnablesUserSuccessfully(t *testing.T) {
	user := createUserForTestEnable(t, "architect")
	user.Disable("disabler-111", false, false)
	user.MarkChangesAsCommitted()

	mockRepo := &mockEnableUserRepository{userToLoad: user}

	handler := NewEnableUserHandler(mockRepo)

	cmd := &commands.EnableUser{
		UserID:      user.ID(),
		EnabledByID: "enabler-123",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedUsers, 1)
	savedUser := mockRepo.savedUsers[0]
	assert.True(t, savedUser.Status().IsActive())
}

func TestEnableUserHandler_UserAlreadyActive_ReturnsError(t *testing.T) {
	user := createUserForTestEnable(t, "architect")
	mockRepo := &mockEnableUserRepository{userToLoad: user}

	handler := NewEnableUserHandler(mockRepo)

	cmd := &commands.EnableUser{
		UserID:      user.ID(),
		EnabledByID: "enabler-456",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, aggregates.ErrUserAlreadyActive)
	assert.Empty(t, mockRepo.savedUsers)
}

func TestEnableUserHandler_UserNotFound_ReturnsError(t *testing.T) {
	mockRepo := &mockEnableUserRepository{getErr: errors.New("not found")}

	handler := NewEnableUserHandler(mockRepo)

	cmd := &commands.EnableUser{
		UserID:      "nonexistent-id",
		EnabledByID: "enabler-789",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}

func TestEnableUserHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockEnableUserRepository{}

	handler := NewEnableUserHandler(mockRepo)

	invalidCmd := &commands.DisableUser{}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestEnableUserHandler_SaveError_ReturnsError(t *testing.T) {
	user := createUserForTestEnable(t, "architect")
	user.Disable("disabler-222", false, false)
	user.MarkChangesAsCommitted()

	mockRepo := &mockEnableUserRepository{
		userToLoad: user,
		saveErr:    errors.New("save failed"),
	}

	handler := NewEnableUserHandler(mockRepo)

	cmd := &commands.EnableUser{
		UserID:      user.ID(),
		EnabledByID: "enabler-err",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}

func TestEnableUserHandler_EnableDisabledAdmin(t *testing.T) {
	user := createUserForTestEnable(t, "admin")
	user.Disable("disabler-333", false, false)
	user.MarkChangesAsCommitted()

	mockRepo := &mockEnableUserRepository{userToLoad: user}

	handler := NewEnableUserHandler(mockRepo)

	cmd := &commands.EnableUser{
		UserID:      user.ID(),
		EnabledByID: "enabler-999",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedUsers, 1)
	savedUser := mockRepo.savedUsers[0]
	assert.True(t, savedUser.Status().IsActive())
	assert.Equal(t, "admin", savedUser.Role().String())
}

func createUserForTestEnable(t *testing.T, roleName string) *aggregates.User {
	t.Helper()

	email, err := valueobjects.NewEmail("test@example.com")
	require.NoError(t, err)

	role, err := valueobjects.RoleFromString(roleName)
	require.NoError(t, err)

	user, err := aggregates.NewUser(email, "Test User", role, "ext-test", "inv-test")
	require.NoError(t, err)

	user.MarkChangesAsCommitted()

	return user
}
