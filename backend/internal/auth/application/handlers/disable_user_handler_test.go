package handlers

import (
	"context"
	"errors"
	"testing"

	"easi/backend/internal/auth/application/commands"
	"easi/backend/internal/auth/domain/aggregates"
	"easi/backend/internal/shared/cqrs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserRepositoryForDisable struct {
	savedUsers []*aggregates.User
	userToLoad *aggregates.User
	getErr     error
	saveErr    error
}

func (m *mockUserRepositoryForDisable) Save(ctx context.Context, user *aggregates.User) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.savedUsers = append(m.savedUsers, user)
	return nil
}

func (m *mockUserRepositoryForDisable) GetByID(ctx context.Context, id string) (*aggregates.User, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.userToLoad, nil
}

type mockUserReadModelForDisable struct {
	isLastAdmin bool
	checkErr    error
}

func (m *mockUserReadModelForDisable) IsLastActiveAdmin(ctx context.Context, userID string) (bool, error) {
	if m.checkErr != nil {
		return false, m.checkErr
	}
	return m.isLastAdmin, nil
}

type userRepositoryForDisable interface {
	Save(ctx context.Context, user *aggregates.User) error
	GetByID(ctx context.Context, id string) (*aggregates.User, error)
}

type userReadModelForDisable interface {
	IsLastActiveAdmin(ctx context.Context, userID string) (bool, error)
}

type testableDisableUserHandler struct {
	repository userRepositoryForDisable
	readModel  userReadModelForDisable
}

func newTestableDisableUserHandler(
	repository userRepositoryForDisable,
	readModel userReadModelForDisable,
) *testableDisableUserHandler {
	return &testableDisableUserHandler{
		repository: repository,
		readModel:  readModel,
	}
}

func (h *testableDisableUserHandler) Handle(ctx context.Context, cmd cqrs.Command) error {
	command, ok := cmd.(*commands.DisableUser)
	if !ok {
		return cqrs.ErrInvalidCommand
	}

	isCurrentUser := command.UserID == command.DisabledByID

	isLastAdmin, err := h.readModel.IsLastActiveAdmin(ctx, command.UserID)
	if err != nil {
		return err
	}

	user, err := h.repository.GetByID(ctx, command.UserID)
	if err != nil {
		return err
	}

	if err := user.Disable(command.DisabledByID, isCurrentUser, isLastAdmin); err != nil {
		return err
	}

	return h.repository.Save(ctx, user)
}

func TestDisableUserHandler_DisablesUserSuccessfully(t *testing.T) {
	user := createUserForTestDisable(t, "architect")
	mockRepo := &mockUserRepositoryForDisable{userToLoad: user}
	mockReadModel := &mockUserReadModelForDisable{isLastAdmin: false}

	handler := newTestableDisableUserHandler(mockRepo, mockReadModel)

	cmd := &commands.DisableUser{
		UserID:       user.ID(),
		DisabledByID: "disabler-123",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedUsers, 1)
	savedUser := mockRepo.savedUsers[0]
	assert.False(t, savedUser.Status().IsActive())
}

func TestDisableUserHandler_CannotDisableSelf(t *testing.T) {
	user := createUserForTestDisable(t, "admin")
	mockRepo := &mockUserRepositoryForDisable{userToLoad: user}
	mockReadModel := &mockUserReadModelForDisable{isLastAdmin: false}

	handler := newTestableDisableUserHandler(mockRepo, mockReadModel)

	cmd := &commands.DisableUser{
		UserID:       user.ID(),
		DisabledByID: user.ID(),
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, aggregates.ErrCannotDisableSelf)
	assert.Empty(t, mockRepo.savedUsers)
}

func TestDisableUserHandler_CannotDisableLastAdmin(t *testing.T) {
	user := createUserForTestDisable(t, "admin")
	mockRepo := &mockUserRepositoryForDisable{userToLoad: user}
	mockReadModel := &mockUserReadModelForDisable{isLastAdmin: true}

	handler := newTestableDisableUserHandler(mockRepo, mockReadModel)

	cmd := &commands.DisableUser{
		UserID:       user.ID(),
		DisabledByID: "other-admin-456",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, aggregates.ErrCannotDisableLastAdmin)
	assert.Empty(t, mockRepo.savedUsers)
}

func TestDisableUserHandler_UserAlreadyDisabled_ReturnsError(t *testing.T) {
	user := createUserForTestDisable(t, "architect")
	user.Disable("previous-disabler", false, false)
	user.MarkChangesAsCommitted()

	mockRepo := &mockUserRepositoryForDisable{userToLoad: user}
	mockReadModel := &mockUserReadModelForDisable{isLastAdmin: false}

	handler := newTestableDisableUserHandler(mockRepo, mockReadModel)

	cmd := &commands.DisableUser{
		UserID:       user.ID(),
		DisabledByID: "disabler-789",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.ErrorIs(t, err, aggregates.ErrUserAlreadyDisabled)
	assert.Empty(t, mockRepo.savedUsers)
}

func TestDisableUserHandler_UserNotFound_ReturnsError(t *testing.T) {
	mockRepo := &mockUserRepositoryForDisable{getErr: errors.New("not found")}
	mockReadModel := &mockUserReadModelForDisable{isLastAdmin: false}

	handler := newTestableDisableUserHandler(mockRepo, mockReadModel)

	cmd := &commands.DisableUser{
		UserID:       "nonexistent-id",
		DisabledByID: "disabler-999",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
}

func TestDisableUserHandler_InvalidCommand_ReturnsError(t *testing.T) {
	mockRepo := &mockUserRepositoryForDisable{}
	mockReadModel := &mockUserReadModelForDisable{}

	handler := newTestableDisableUserHandler(mockRepo, mockReadModel)

	invalidCmd := &commands.EnableUser{}

	err := handler.Handle(context.Background(), invalidCmd)
	assert.ErrorIs(t, err, cqrs.ErrInvalidCommand)
}

func TestDisableUserHandler_ReadModelError_ReturnsError(t *testing.T) {
	user := createUserForTestDisable(t, "admin")
	mockRepo := &mockUserRepositoryForDisable{userToLoad: user}
	mockReadModel := &mockUserReadModelForDisable{checkErr: errors.New("database error")}

	handler := newTestableDisableUserHandler(mockRepo, mockReadModel)

	cmd := &commands.DisableUser{
		UserID:       user.ID(),
		DisabledByID: "disabler-err",
	}

	err := handler.Handle(context.Background(), cmd)
	assert.Error(t, err)
	assert.Empty(t, mockRepo.savedUsers)
}

func TestDisableUserHandler_NonAdminUser_Success(t *testing.T) {
	user := createUserForTestDisable(t, "stakeholder")
	mockRepo := &mockUserRepositoryForDisable{userToLoad: user}
	mockReadModel := &mockUserReadModelForDisable{isLastAdmin: false}

	handler := newTestableDisableUserHandler(mockRepo, mockReadModel)

	cmd := &commands.DisableUser{
		UserID:       user.ID(),
		DisabledByID: "admin-disabler",
	}

	err := handler.Handle(context.Background(), cmd)
	require.NoError(t, err)

	require.Len(t, mockRepo.savedUsers, 1)
	savedUser := mockRepo.savedUsers[0]
	assert.False(t, savedUser.Status().IsActive())
}

func createUserForTestDisable(t *testing.T, roleName string) *aggregates.User {
	t.Helper()
	return createUserForTest(t, roleName)
}
