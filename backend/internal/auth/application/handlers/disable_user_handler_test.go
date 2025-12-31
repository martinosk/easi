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

type mockDisableUserRepository struct {
	savedUsers []*aggregates.User
	userToLoad *aggregates.User
	getErr     error
	saveErr    error
}

func (m *mockDisableUserRepository) Save(ctx context.Context, user *aggregates.User) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedUsers = append(m.savedUsers, user)
	return nil
}

func (m *mockDisableUserRepository) GetByID(ctx context.Context, id string) (*aggregates.User, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.userToLoad, nil
}

type mockDisableUserReadModel struct {
	isLastAdmin bool
	checkErr    error
}

func (m *mockDisableUserReadModel) IsLastActiveAdmin(ctx context.Context, userID string) (bool, error) {
	if m.checkErr != nil {
		return false, m.checkErr
	}
	return m.isLastAdmin, nil
}

func TestDisableUserHandler_DisablesUserSuccessfully(t *testing.T) {
	user := createUserForTestDisable(t, "architect")
	mockRepo := &mockDisableUserRepository{userToLoad: user}
	mockReadModel := &mockDisableUserReadModel{isLastAdmin: false}

	handler := NewDisableUserHandler(mockRepo, mockReadModel)

	cmd := &commands.DisableUser{
		UserID:       user.ID(),
		DisabledByID: "disabler-123",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedUsers, 1)
	savedUser := mockRepo.savedUsers[0]
	assert.False(t, savedUser.Status().IsActive())
}

func TestDisableUserHandler_CannotDisableSelf(t *testing.T) {
	user := createUserForTestDisable(t, "admin")
	mockRepo := &mockDisableUserRepository{userToLoad: user}
	mockReadModel := &mockDisableUserReadModel{isLastAdmin: false}

	handler := NewDisableUserHandler(mockRepo, mockReadModel)

	cmd := &commands.DisableUser{
		UserID:       user.ID(),
		DisabledByID: user.ID(),
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, aggregates.ErrCannotDisableSelf)
	assert.Empty(t, mockRepo.savedUsers)
}

func TestDisableUserHandler_CannotDisableLastAdmin(t *testing.T) {
	user := createUserForTestDisable(t, "admin")
	mockRepo := &mockDisableUserRepository{userToLoad: user}
	mockReadModel := &mockDisableUserReadModel{isLastAdmin: true}

	handler := NewDisableUserHandler(mockRepo, mockReadModel)

	cmd := &commands.DisableUser{
		UserID:       user.ID(),
		DisabledByID: "other-admin-456",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, aggregates.ErrCannotDisableLastAdmin)
	assert.Empty(t, mockRepo.savedUsers)
}

func TestDisableUserHandler_UserAlreadyDisabled_ReturnsError(t *testing.T) {
	user := createUserForTestDisable(t, "architect")
	user.Disable("previous-disabler", false, false)
	user.MarkChangesAsCommitted()

	mockRepo := &mockDisableUserRepository{userToLoad: user}
	mockReadModel := &mockDisableUserReadModel{isLastAdmin: false}

	handler := NewDisableUserHandler(mockRepo, mockReadModel)

	cmd := &commands.DisableUser{
		UserID:       user.ID(),
		DisabledByID: "disabler-789",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, aggregates.ErrUserAlreadyDisabled)
	assert.Empty(t, mockRepo.savedUsers)
}

func TestDisableUserHandler_UserNotFound_ReturnsError(t *testing.T) {
	mockRepo := &mockDisableUserRepository{getErr: errors.New("not found")}
	mockReadModel := &mockDisableUserReadModel{isLastAdmin: false}

	handler := NewDisableUserHandler(mockRepo, mockReadModel)

	cmd := &commands.DisableUser{
		UserID:       "nonexistent-id",
		DisabledByID: "disabler-999",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}

func TestDisableUserHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockDisableUserRepository{}
	mockReadModel := &mockDisableUserReadModel{}

	handler := NewDisableUserHandler(mockRepo, mockReadModel)

	invalidCmd := &commands.EnableUser{}

	_, err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestDisableUserHandler_ReadModelError_ReturnsError(t *testing.T) {
	user := createUserForTestDisable(t, "admin")
	mockRepo := &mockDisableUserRepository{userToLoad: user}
	mockReadModel := &mockDisableUserReadModel{checkErr: errors.New("database error")}

	handler := NewDisableUserHandler(mockRepo, mockReadModel)

	cmd := &commands.DisableUser{
		UserID:       user.ID(),
		DisabledByID: "disabler-err",
	}

	_, err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockRepo.savedUsers)
}

func TestDisableUserHandler_NonAdminUser_Success(t *testing.T) {
	user := createUserForTestDisable(t, "stakeholder")
	mockRepo := &mockDisableUserRepository{userToLoad: user}
	mockReadModel := &mockDisableUserReadModel{isLastAdmin: false}

	handler := NewDisableUserHandler(mockRepo, mockReadModel)

	cmd := &commands.DisableUser{
		UserID:       user.ID(),
		DisabledByID: "admin-disabler",
	}

	_, err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedUsers, 1)
	savedUser := mockRepo.savedUsers[0]
	assert.False(t, savedUser.Status().IsActive())
}

func createUserForTestDisable(t *testing.T, roleName string) *aggregates.User {
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
